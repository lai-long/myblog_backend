package model

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ ArticleModel = (*customArticleModel)(nil)

// ArticleListCond 列表查询条件，公开端和管理端共用。
// Status 为 nil 表示不过滤状态（管理端看全部）；Keyword 为空表示不搜索。
type ArticleListCond struct {
	Status  *int64
	Keyword string
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

// buildListWhere 拼 WHERE 子句，FindPage 和 Count 共用，避免两处条件写歪
func buildListWhere(cond ArticleListCond) (string, []any) {
	var (
		where []string
		args  []any
	)
	if cond.Status != nil {
		where = append(where, "`status` = ?")
		args = append(args, *cond.Status)
	}
	if cond.Keyword != "" {
		where = append(where, "(`title` LIKE ? OR `content` LIKE ?)")
		kw := "%" + cond.Keyword + "%"
		args = append(args, kw, kw)
	}
	if len(where) == 0 {
		return "", nil
	}
	return "WHERE " + strings.Join(where, " AND "), args
}

func (m *customArticleModel) FindPage(ctx context.Context, cond ArticleListCond) ([]*Article, error) {
	where, args := buildListWhere(cond)
	query := fmt.Sprintf("SELECT %s FROM %s %s ORDER BY `published_at` DESC LIMIT ? OFFSET ?",
		articleRows, m.table, where)
	args = append(args, cond.Size, cond.Offset)

	var resp []*Article
	if err := m.conn.QueryRowsCtx(ctx, &resp, query, args...); err != nil {
		return nil, err
	}
	return resp, nil
}

func (m *customArticleModel) Count(ctx context.Context, cond ArticleListCond) (int64, error) {
	where, args := buildListWhere(cond)
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s %s", m.table, where)

	var total int64
	if err := m.conn.QueryRowCtx(ctx, &total, query, args...); err != nil {
		return 0, err
	}
	return total, nil
}

// Update 覆盖 goctl 生成的版本。
// 原版把 updated_at 当"自动字段"排除在 SET 外（那是 MySQL ON UPDATE 的假设），
// 而 SQLite 没有这个机制，updated_at 会永远停在创建时间。这里补写它。
// 注意：Article 结构体增删字段时，下面 ExecCtx 的参数顺序要跟着改。
func (m *customArticleModel) Update(ctx context.Context, data *Article) error {
	data.UpdatedAt = time.Now().UTC()
	query := fmt.Sprintf("UPDATE %s SET %s, `updated_at` = ? WHERE `id` = ?",
		m.table, articleRowsWithPlaceHolder)
	_, err := m.conn.ExecCtx(ctx, query,
		data.Title, data.Slug, data.Summary, data.Content, data.CoverUrl,
		data.Status, data.Views, data.PublishedAt, data.UpdatedAt, data.Id)
	return err
}

// IncrViews 浏览量 +1。用 SQL 自增而不是"读出来+1再写回"，避免并发下互相覆盖
func (m *customArticleModel) IncrViews(ctx context.Context, id int64) error {
	query := fmt.Sprintf("UPDATE %s SET `views` = `views` + 1 WHERE `id` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, id)
	return err
}
