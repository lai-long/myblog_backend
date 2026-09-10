package model

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ CommentModel = (*customCommentModel)(nil)

// articleTable 后台列表要 join 文章表拿标题。这里没法复用 ArticleModel 的 table 字段，只能写死
const articleTable = "`article`"

// AdminCommentListCond 后台评论列表条件。Status 为 nil 表示不过滤
type AdminCommentListCond struct {
	Status *int64
	Offset int
	Size   int
}

// AdminCommentRow 后台列表的一行：评论 + 所属文章标题
type AdminCommentRow struct {
	Id           int64     `db:"id"`
	ArticleId    int64     `db:"article_id"`
	ArticleTitle string    `db:"article_title"`
	ArticleSlug  string    `db:"article_slug"`
	ParentId     int64     `db:"parent_id"`
	Nickname     string    `db:"nickname"`
	Content      string    `db:"content"`
	Status       int64     `db:"status"`
	CreatedAt    time.Time `db:"created_at"`
}

type (
	// CommentModel is an interface to be customized, add more methods here,
	// and implement the added methods in customCommentModel.
	CommentModel interface {
		commentModel
		withSession(session sqlx.Session) CommentModel
		FindApprovedByArticleId(ctx context.Context, articleId int64) ([]*Comment, error)
		FindAdminPage(ctx context.Context, cond AdminCommentListCond) ([]*AdminCommentRow, error)
		CountAdmin(ctx context.Context, cond AdminCommentListCond) (int64, error)
		UpdateStatus(ctx context.Context, id int64, status int64) error
	}

	customCommentModel struct {
		*defaultCommentModel
	}
)

// NewCommentModel returns a model for the database table.
func NewCommentModel(conn sqlx.SqlConn) CommentModel {
	return &customCommentModel{
		defaultCommentModel: newCommentModel(conn),
	}
}

func (m *customCommentModel) withSession(session sqlx.Session) CommentModel {
	return NewCommentModel(sqlx.NewSqlConnFromSession(session))
}

// FindApprovedByArticleId 公开端用：某篇文章下"已通过"且未删除的评论，按时间正序
// 写死 status=1，游客永远看不到待审核和垃圾评论
func (m *customCommentModel) FindApprovedByArticleId(ctx context.Context, articleId int64) ([]*Comment, error) {
	query := fmt.Sprintf("SELECT %s FROM %s "+
		"WHERE `article_id` = ? AND `status` = 1 AND `deleted_at` IS NULL "+
		"ORDER BY `created_at` ASC", commentRows, m.table)

	var resp []*Comment
	if err := m.conn.QueryRowsCtx(ctx, &resp, query, articleId); err != nil {
		return nil, err
	}
	return resp, nil
}

// findAdminWhere 后台列表的条件，分页和计数共用。
// 列名必须带 `c.` 前缀：article 表也有 status / deleted_at，join 后不写前缀会歧义报错
func findAdminWhere(cond AdminCommentListCond) (string, []any) {
	var (
		where = []string{"`c`.`deleted_at` IS NULL"}
		args  []any
	)
	if cond.Status != nil {
		where = append(where, "`c`.`status` = ?")
		args = append(args, *cond.Status)
	}
	return "WHERE " + strings.Join(where, " AND "), args
}

func (m *customCommentModel) FindAdminPage(ctx context.Context, cond AdminCommentListCond) ([]*AdminCommentRow, error) {
	where, args := findAdminWhere(cond)
	query := fmt.Sprintf("SELECT `c`.`id`, `c`.`article_id`, `a`.`title` AS `article_title`, `a`.`slug` AS `article_slug`, "+
		"`c`.`parent_id`, `c`.`nickname`, `c`.`content`, `c`.`status`, `c`.`created_at` "+
		"FROM %s `c` LEFT JOIN %s `a` ON `c`.`article_id` = `a`.`id` %s "+
		"ORDER BY `c`.`created_at` DESC LIMIT ? OFFSET ?", m.table, articleTable, where)
	args = append(args, cond.Size, cond.Offset)

	var resp []*AdminCommentRow
	if err := m.conn.QueryRowsCtx(ctx, &resp, query, args...); err != nil {
		return nil, err
	}
	return resp, nil
}

func (m *customCommentModel) CountAdmin(ctx context.Context, cond AdminCommentListCond) (int64, error) {
	where, args := findAdminWhere(cond)
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s `c` %s", m.table, where)

	var total int64
	if err := m.conn.QueryRowCtx(ctx, &total, query, args...); err != nil {
		return 0, err
	}
	return total, nil
}

// UpdateStatus 审核：只改 status，不碰其他字段
func (m *customCommentModel) UpdateStatus(ctx context.Context, id int64, status int64) error {
	query := fmt.Sprintf("UPDATE %s SET `status` = ? WHERE `id` = ? AND `deleted_at` IS NULL", m.table)
	_, err := m.conn.ExecCtx(ctx, query, status, id)
	return err
}

// Delete 改成软删除：只打标记，数据保留
func (m *customCommentModel) Delete(ctx context.Context, id int64) error {
	query := fmt.Sprintf("UPDATE %s SET `deleted_at` = ? WHERE `id` = ? AND `deleted_at` IS NULL", m.table)
	_, err := m.conn.ExecCtx(ctx, query, time.Now().UTC(), id)
	return err
}

// FindOne 覆盖生成版本，加上未删除条件
func (m *customCommentModel) FindOne(ctx context.Context, id int64) (*Comment, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE `id` = ? AND `deleted_at` IS NULL LIMIT 1",
		commentRows, m.table)
	var resp Comment
	if err := m.conn.QueryRowCtx(ctx, &resp, query, id); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &resp, nil
}
