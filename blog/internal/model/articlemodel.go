package model

import (
	"context"
	"errors"
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
	where = append(where, "`deleted_at` IS NULL") // 软删除：所有列表查询都排除已删除的
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
		"`cover_url` = ?, `status` = ?, `views` = ?, `published_at` = ?, `updated_at` = ? "+
		"WHERE `id` = ? AND `deleted_at` IS NULL", m.table)
	_, err := m.conn.ExecCtx(ctx, query,
		data.Title, data.Slug, data.Summary, data.Content, data.CoverUrl,
		data.Status, data.Views, data.PublishedAt, data.UpdatedAt, data.Id)
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
