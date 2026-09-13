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

type AdminArticleDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminArticleDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminArticleDetailLogic {
	return &AdminArticleDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 后台详情：编辑页用。与公开详情的区别——不限 status（草稿/隐藏也要能打开）、不增加浏览量
func (l *AdminArticleDetailLogic) AdminArticleDetail(req *types.ArticleIdReq) (resp *types.ArticleDetail, err error) {
	a, err := l.svcCtx.ArticleModel.FindOne(l.ctx, req.Id)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, errx.New(errx.NotFound, "文章不存在")
		}
		return nil, errx.Wrap(err, errx.ServerError, "查询文章失败")
	}

	tags, err := l.svcCtx.TagModel.FindByArticleIds(l.ctx, []int64{a.Id})
	if err != nil {
		return nil, errx.Wrap(err, errx.ServerError, "查询文章标签失败")
	}

	tagVos := make([]types.TagVo, 0, len(tags[a.Id]))
	for _, t := range tags[a.Id] {
		tagVos = append(tagVos, types.TagVo{Name: t.Name, Slug: t.Slug})
	}

	return &types.ArticleDetail{
		Id:          a.Id,
		Title:       a.Title,
		Slug:        a.Slug,
		Summary:     a.Summary,
		Content:     a.Content,
		CoverUrl:    a.CoverUrl,
		Status:      int(a.Status),
		Views:       a.Views,
		IsTop:       int(a.IsTop),
		PublishedAt: a.PublishedAt.Format(time.RFC3339),
		Tags:        tagVos,
	}, nil
}
