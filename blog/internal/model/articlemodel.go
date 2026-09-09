package model

import (
	"context"
	"fmt"
	"strings"

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
