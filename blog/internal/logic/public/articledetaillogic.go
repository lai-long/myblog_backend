// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package public

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

type ArticleDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewArticleDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ArticleDetailLogic {
	return &ArticleDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ArticleDetailLogic) ArticleDetail(req *types.ArticleDetailReq) (resp *types.ArticleDetail, err error) {
	a, err := l.svcCtx.ArticleModel.FindOneBySlug(l.ctx, req.Slug)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, errx.New(errx.NotFound, "文章不存在")
		}
		return nil, errx.Wrap(err, errx.ServerError, "查询文章失败")
	}

	// 草稿/隐藏对游客等同于不存在，不能靠猜 slug 看到未发布内容
	if a.Status != 1 {
		return nil, errx.New(errx.NotFound, "文章不存在")
	}

	// 浏览量 +1；自增失败只记日志，不影响把文章返回给读者
	if err := l.svcCtx.ArticleModel.IncrViews(l.ctx, a.Id); err != nil {
		l.Errorf("浏览量自增失败 id=%d: %v", a.Id, err)
	}

	return &types.ArticleDetail{
		Id:          a.Id,
		Title:       a.Title,
		Slug:        a.Slug,
		Content:     a.Content,
		CoverUrl:    a.CoverUrl,
		Views:       a.Views + 1, // 直接带上刚 +1 的值，前端不用再刷一次
		PublishedAt: a.PublishedAt.Format(time.RFC3339),
	}, nil
}
