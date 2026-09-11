package model

import (
	"context"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// Tag 标签。slug 用于 URL（/tags/:slug），name 用于展示
type Tag struct {
	Id   int64  `db:"id"`
	Name string `db:"name"`
	Slug string `db:"slug"`
}

// TagWithCount 标签墙用：标签 + 已发布文章数
type TagWithCount struct {
	Tag
	Count int64 `db:"count"`
}

type TagModel struct {
	conn sqlx.SqlConn
}

func NewTagModel(conn sqlx.SqlConn) *TagModel {
	return &TagModel{conn: conn}
}

// slugify 从标签名生成 slug：小写、空格转连字符。中文名原样保留（URL 会 percent-encode，不影响使用）
func slugify(name string) string {
	return strings.ReplaceAll(strings.ToLower(strings.TrimSpace(name)), " ", "-")
}

// FindOrCreate 按名字找标签，不存在就创建。文章保存时调用：前端只传标签名，id 的解析收在这里
func (m *TagModel) FindOrCreate(ctx context.Context, names []string) ([]Tag, error) {
	tags := make([]Tag, 0, len(names))
	for _, name := range names {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		slug := slugify(name)
		// INSERT OR IGNORE：已存在（name 或 slug 撞 UNIQUE）就静默跳过，接下来 SELECT 拿到既有记录
		if _, err := m.conn.ExecCtx(ctx,
			"INSERT OR IGNORE INTO tag (name, slug) VALUES (?, ?)", name, slug); err != nil {
			return nil, err
		}
		var t Tag
		if err := m.conn.QueryRowCtx(ctx, &t,
			"SELECT id, name, slug FROM tag WHERE slug = ? LIMIT 1", slug); err != nil {
			return nil, err
		}
		tags = append(tags, t)
	}
	return tags, nil
}

// SetArticleTags 整体替换一篇文章的标签关联（先删后插，一个事务里完成）
func (m *TagModel) SetArticleTags(ctx context.Context, articleId int64, tagIds []int64) error {
	return m.conn.TransactCtx(ctx, func(ctx context.Context, s sqlx.Session) error {
		if _, err := s.ExecCtx(ctx, "DELETE FROM article_tag WHERE article_id = ?", articleId); err != nil {
			return err
		}
		for _, tagId := range tagIds {
			if _, err := s.ExecCtx(ctx,
				"INSERT OR IGNORE INTO article_tag (article_id, tag_id) VALUES (?, ?)",
				articleId, tagId); err != nil {
				return err
			}
		}
		return nil
	})
}

// FindByArticleIds 批量查多篇文章的标签，一次 SQL 搞定（避免列表页 N+1 查询）
func (m *TagModel) FindByArticleIds(ctx context.Context, articleIds []int64) (map[int64][]Tag, error) {
	result := make(map[int64][]Tag, len(articleIds))
	if len(articleIds) == 0 {
		return result, nil
	}
	marks := strings.TrimSuffix(strings.Repeat("?,", len(articleIds)), ",")
	args := make([]any, 0, len(articleIds))
	for _, id := range articleIds {
		args = append(args, id)
	}
	var rows []struct {
		ArticleId int64  `db:"article_id"`
		Id        int64  `db:"id"`
		Name      string `db:"name"`
		Slug      string `db:"slug"`
	}
	err := m.conn.QueryRowsCtx(ctx, &rows, fmt.Sprintf(
		"SELECT at.article_id, t.id, t.name, t.slug FROM article_tag at "+
			"JOIN tag t ON t.id = at.tag_id WHERE at.article_id IN (%s) ORDER BY t.id", marks), args...)
	if err != nil {
		return nil, err
	}
	for _, r := range rows {
		result[r.ArticleId] = append(result[r.ArticleId], Tag{Id: r.Id, Name: r.Name, Slug: r.Slug})
	}
	return result, nil
}

// FindAllWithCount 标签墙：全部标签 + 各自已发布文章数（只统计已发布且未删除的）
func (m *TagModel) FindAllWithCount(ctx context.Context) ([]*TagWithCount, error) {
	var rows []*TagWithCount
	err := m.conn.QueryRowsCtx(ctx, &rows,
		"SELECT t.id, t.name, t.slug, COUNT(a.id) AS `count` FROM tag t "+
			"LEFT JOIN article_tag at ON at.tag_id = t.id "+
			// 统计条件放在 JOIN 条件里而不是 WHERE：这样没有文章的标签也会以 0 出现
			"LEFT JOIN article a ON a.id = at.article_id AND a.status = 1 AND a.deleted_at IS NULL "+
			"GROUP BY t.id ORDER BY `count` DESC, t.id")
	if err != nil {
		return nil, err
	}
	return rows, nil
}
