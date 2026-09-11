package model

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ ArticleModel = (*customArticleModel)(nil)

// ArticleListCond 列表查询条件，公开端和管理端共用。
// Status 为 nil 表示不过滤状态（管理端看全部）；Keyword 为空表示不搜索。
type ArticleListCond struct {
	Status  *int64
	Keyword string
	TagSlug string // 非空时按标签 slug 过滤
	Offset  int
	Size    int
}

type (
	// ArticleModel is an interface to be customized, add more methods here,
	// and implement the added methods in customArticleModel.
	ArticleModel interface {
		articleModel
		withSession(session sqlx.Session) ArticleModel
		FindPage(ctx context.Context, cond ArticleListCond) ([]*Article, error)
		Count(ctx context.Context, cond ArticleListCond) (int64, error)
		IncrViews(ctx context.Context, id int64) error
	}

	customArticleModel struct {
		*defaultArticleModel
	}
)

// NewArticleModel returns a model for the database table.
func NewArticleModel(conn sqlx.SqlConn) ArticleModel {
	return &customArticleModel{
		defaultArticleModel: newArticleModel(conn),
	}
}

func (m *customArticleModel) withSession(session sqlx.Session) ArticleModel {
	return NewArticleModel(sqlx.NewSqlConnFromSession(session))
}

// buildListWhere 拼 WHERE 子句（不含全文检索条件），FindPage 和 Count 共用，避免两处条件写歪。
// 返回的 ftsQuery 非空表示关键词走 FTS5 索引，由调用方决定怎么接（FindPage 用 JOIN 排序，Count 用 IN 子查询）。
func buildListWhere(cond ArticleListCond) (where string, args []any, ftsQuery string) {
	var clauses []string
	clauses = append(clauses, "`deleted_at` IS NULL") // 软删除：所有列表查询都排除已删除的
	if cond.Status != nil {
		clauses = append(clauses, "`status` = ?")
		args = append(args, *cond.Status)
	}
	if cond.Keyword != "" {
		// trigram 索引至少要 3 个字符才能命中，短词回退 LIKE
		if utf8.RuneCountInString(cond.Keyword) >= 3 {
			ftsQuery = fts5MatchQuery(cond.Keyword)
		} else {
			clauses = append(clauses, "(`title` LIKE ? OR `content` LIKE ?)")
			kw := "%" + cond.Keyword + "%"
			args = append(args, kw, kw)
		}
	}
	if cond.TagSlug != "" {
		// EXISTS 子查询过滤：不动主查询的 FROM，避免 JOIN 引入重复行
		clauses = append(clauses, "EXISTS (SELECT 1 FROM article_tag at JOIN tag t ON t.id = at.tag_id "+
			"WHERE at.article_id = "+articleTable+".`id` AND t.slug = ?)")
		args = append(args, cond.TagSlug)
	}
	if len(clauses) == 0 {
		return "", nil, ftsQuery
	}
	return "WHERE " + strings.Join(clauses, " AND "), args, ftsQuery
}

// fts5MatchQuery 把用户输入转成 FTS5 查询串：按空白分词，每个词加引号当短语，
// 词之间是 AND 关系；引号转义防注入语法错误
func fts5MatchQuery(keyword string) string {
	terms := strings.Fields(keyword)
	quoted := make([]string, 0, len(terms))
	for _, t := range terms {
		quoted = append(quoted, `"`+strings.ReplaceAll(t, `"`, `""`)+`"`)
	}
	return strings.Join(quoted, " ")
}

// qualifiedArticleRows article 表全字段，每列带表名前缀（JOIN article_fts 后列名会歧义）
func (m *customArticleModel) qualifiedRows() string {
	cols := strings.Split(articleRows, ",")
	for i, c := range cols {
		cols[i] = m.table + "." + c
	}
	return strings.Join(cols, ",")
}

func (m *customArticleModel) FindPage(ctx context.Context, cond ArticleListCond) ([]*Article, error) {
	where, args, ftsQuery := buildListWhere(cond)

	// 有 FTS 关键词时 JOIN 索引表（MATCH 放 JOIN 条件里，bm25 才有效），按相关度排序；
	// 否则置顶优先、再按发布时间
	var query string
	if ftsQuery != "" {
		query = fmt.Sprintf("SELECT %s FROM %s JOIN article_fts ON article_fts.rowid = %s.`id` AND article_fts MATCH ? %s "+
			"ORDER BY bm25(article_fts) LIMIT ? OFFSET ?",
			m.qualifiedRows(), m.table, m.table, where)
		args = append([]any{ftsQuery}, args...)
	} else {
		query = fmt.Sprintf("SELECT %s FROM %s %s ORDER BY `is_top` DESC, `published_at` DESC LIMIT ? OFFSET ?",
			articleRows, m.table, where)
	}
	args = append(args, cond.Size, cond.Offset)

	var resp []*Article
	if err := m.conn.QueryRowsCtx(ctx, &resp, query, args...); err != nil {
		return nil, err
	}
	return resp, nil
}

func (m *customArticleModel) Count(ctx context.Context, cond ArticleListCond) (int64, error) {
	where, args, ftsQuery := buildListWhere(cond)
	if ftsQuery != "" {
		if where == "" {
			where = "WHERE `id` IN (SELECT rowid FROM article_fts WHERE article_fts MATCH ?)"
		} else {
			where += " AND `id` IN (SELECT rowid FROM article_fts WHERE article_fts MATCH ?)"
		}
		args = append(args, ftsQuery)
	}
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s %s", m.table, where)

	var total int64
	if err := m.conn.QueryRowCtx(ctx, &total, query, args...); err != nil {
		return 0, err
	}
	return total, nil
}

// FindOne 覆盖 goctl 生成的版本，加上"未删除"条件，否则软删后还能被查出来
func (m *customArticleModel) FindOne(ctx context.Context, id int64) (*Article, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE `id` = ? AND `deleted_at` IS NULL LIMIT 1",
		articleRows, m.table)
	var resp Article
	if err := m.conn.QueryRowCtx(ctx, &resp, query, id); err != nil {
		return nil, errFromNotFoundErr(err)
	}
	return &resp, nil
}

// FindOneBySlug 同理，加"未删除"条件
func (m *customArticleModel) FindOneBySlug(ctx context.Context, slug string) (*Article, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE `slug` = ? AND `deleted_at` IS NULL LIMIT 1",
		articleRows, m.table)
	var resp Article
	if err := m.conn.QueryRowCtx(ctx, &resp, query, slug); err != nil {
		return nil, errFromNotFoundErr(err)
	}
	return &resp, nil
}

// errFromNotFoundErr 把驱动的"没查到"统一转成 model.ErrNotFound，logic 才能用 errors.Is 判断
func errFromNotFoundErr(err error) error {
	if errors.Is(err, sqlx.ErrNotFound) {
		return ErrNotFound
	}
	return err
}

// Update 覆盖 goctl 生成的版本。
// 原版把 updated_at 当"自动字段"排除在 SET 外（那是 MySQL ON UPDATE 的假设），
// 而 SQLite 没有这个机制，updated_at 会永远停在创建时间。这里补写它。
// deleted_at 刻意不写：更新操作不应该改变删除状态。
// 注意：Article 结构体增删字段时，下面 ExecCtx 的参数顺序要跟着改。
func (m *customArticleModel) Update(ctx context.Context, data *Article) error {
	data.UpdatedAt = time.Now().UTC()
	query := fmt.Sprintf("UPDATE %s SET `title` = ?, `slug` = ?, `summary` = ?, `content` = ?, "+
		"`cover_url` = ?, `status` = ?, `is_top` = ?, `views` = ?, `published_at` = ?, `updated_at` = ? "+
		"WHERE `id` = ? AND `deleted_at` IS NULL", m.table)
	_, err := m.conn.ExecCtx(ctx, query,
		data.Title, data.Slug, data.Summary, data.Content, data.CoverUrl,
		data.Status, data.IsTop, data.Views, data.PublishedAt.UTC().Format(time.RFC3339), data.UpdatedAt, data.Id)
	return err
}

// Delete 改成软删除：只打删除标记，数据还在库里。
// 加 deleted_at IS NULL 条件，重复删同一篇不会覆盖第一次的删除时间。
func (m *customArticleModel) Delete(ctx context.Context, id int64) error {
	query := fmt.Sprintf("UPDATE %s SET `deleted_at` = ? WHERE `id` = ? AND `deleted_at` IS NULL",
		m.table)
	_, err := m.conn.ExecCtx(ctx, query, time.Now().UTC(), id)
	return err
}

// IncrViews 浏览量 +1。用 SQL 自增而不是"读出来+1再写回"，避免并发下互相覆盖
func (m *customArticleModel) IncrViews(ctx context.Context, id int64) error {
	query := fmt.Sprintf("UPDATE %s SET `views` = `views` + 1 WHERE `id` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, id)
	return err
}
