// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package admin

import (
	"context"
	"errors"
	"time"

	"myblog_backend/blog/internal/model"
	"myblog_backend/blog/internal/svc"
	"myblog_backend/blog/internal/types"
	"myblog_backend/pkg/errx"

	"github.com/zeromicro/go-zero/core/logx"
)

type ArticleCreateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewArticleCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ArticleCreateLogic {
	return &ArticleCreateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ArticleCreateLogic) ArticleCreate(req *types.ArticleSaveReq) (resp *types.EmptyResp, err error) {
	// 1. slug 唯一性检查：先查一次，比直接撞数据库的 UNIQUE 约束再解析错误要清爽
	_, err = l.svcCtx.ArticleModel.FindOneBySlug(l.ctx, req.Slug)
	if err == nil {
		return nil, errx.New(errx.ParamError, "slug 已存在")
	}
	if !errors.Is(err, model.ErrNotFound) {
		return nil, errx.Wrap(err, errx.ServerError, "查询文章失败")
	}

	// 2. 入库。id / created_at / updated_at 由数据库填，published_at 我们自己给
	res, err := l.svcCtx.ArticleModel.Insert(l.ctx, &model.Article{
		Title:       req.Title,
		Slug:        req.Slug,
		Summary:     req.Summary,
		Content:     req.Content,
		CoverUrl:    req.CoverUrl,
		Status:      int64(req.Status), // .api 里是 int，model 是 int64，必须显式转
		PublishedAt: time.Now().UTC(),  // 全库统一存 UTC：SQLite 的 CURRENT_TIMESTAMP 也是 UTC，两边才对得上
	})
	if err != nil {
		return nil, errx.Wrap(err, errx.ServerError, "创建文章失败")
	}

	// 3. 标签：没传就跳过；传了则按名字找或建，再关联
	if req.Tags == nil {
		return &types.EmptyResp{}, nil
	}
	articleId, err := res.LastInsertId()
	if err != nil {
		return nil, errx.Wrap(err, errx.ServerError, "获取新文章 id 失败")
	}
	if err := l.saveTags(req.Tags, articleId); err != nil {
		return nil, err
	}

	return &types.EmptyResp{}, nil
}

// saveTags 解析标签名（不存在自动创建）并整体替换文章关联；create/update 共用
func (l *ArticleCreateLogic) saveTags(names []string, articleId int64) error {
	tags, err := l.svcCtx.TagModel.FindOrCreate(l.ctx, names)
	if err != nil {
		return errx.Wrap(err, errx.ServerError, "保存标签失败")
	}
	ids := make([]int64, 0, len(tags))
	for _, t := range tags {
		ids = append(ids, t.Id)
	}
	if err := l.svcCtx.TagModel.SetArticleTags(l.ctx, articleId, ids); err != nil {
		return errx.Wrap(err, errx.ServerError, "关联标签失败")
	}
	return nil
}
