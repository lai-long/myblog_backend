// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package admin

import (
	"context"
	"errors"

	"myblog_backend/blog/internal/model"
	"myblog_backend/blog/internal/svc"
	"myblog_backend/blog/internal/types"
	"myblog_backend/pkg/errx"

	"github.com/zeromicro/go-zero/core/logx"
)

type ArticleUpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewArticleUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ArticleUpdateLogic {
	return &ArticleUpdateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ArticleUpdateLogic) ArticleUpdate(req *types.ArticleSaveReq) (resp *types.EmptyResp, err error) {
	// 1. 先取原记录：views / published_at / created_at 都得以它为准，不能被覆盖
	old, err := l.svcCtx.ArticleModel.FindOne(l.ctx, req.Id)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, errx.New(errx.NotFound, "文章不存在")
		}
		return nil, errx.Wrap(err, errx.ServerError, "查询文章失败")
	}

	// 2. 只有真的改了 slug 才查重，没改就省这次 SQL
	if req.Slug != old.Slug {
		if _, err := l.svcCtx.ArticleModel.FindOneBySlug(l.ctx, req.Slug); err == nil {
			return nil, errx.New(errx.ParamError, "slug 已存在")
		} else if !errors.Is(err, model.ErrNotFound) {
			return nil, errx.Wrap(err, errx.ServerError, "查询文章失败")
		}
	}

	// 3. 以原记录为底，只改前端能改的字段（updated_at 由 model.Update 写入）
	data := *old
	data.Title = req.Title
	data.Slug = req.Slug
	data.Summary = req.Summary
	data.Content = req.Content
	data.CoverUrl = req.CoverUrl
	data.Status = int64(req.Status)

	if err := l.svcCtx.ArticleModel.Update(l.ctx, &data); err != nil {
		return nil, errx.Wrap(err, errx.ServerError, "更新文章失败")
	}

	return &types.EmptyResp{}, nil
}
