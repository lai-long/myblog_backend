// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package admin

import (
	"context"
	"time"

	"myblog_backend/blog/internal/model"
	"myblog_backend/blog/internal/svc"
	"myblog_backend/blog/internal/types"
	"myblog_backend/pkg/errx"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminArticleListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminArticleListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminArticleListLogic {
	return &AdminArticleListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminArticleListLogic) AdminArticleList(req *types.AdminArticleListReq) (resp *types.AdminArticleListResp, err error) {
	cond := model.ArticleListCond{
		Offset: (req.Page - 1) * req.Size,
		Size:   req.Size,
	}
	// Status = -1 表示不过滤，看全部；否则按状态过滤
	if req.Status >= 0 {
		status := int64(req.Status)
		cond.Status = &status
	}

	articles, err := l.svcCtx.ArticleModel.FindPage(l.ctx, cond)
	if err != nil {
		return nil, errx.Wrap(err, errx.ServerError, "查询文章列表失败")
	}

	total, err := l.svcCtx.ArticleModel.Count(l.ctx, cond)
	if err != nil {
		return nil, errx.Wrap(err, errx.ServerError, "查询文章总数失败")
	}

	list := make([]types.AdminArticleSummary, 0, len(articles))
	for _, a := range articles {
		list = append(list, types.AdminArticleSummary{
			Id:          a.Id,
			Title:       a.Title,
			Slug:        a.Slug,
			Status:      int(a.Status),
			Views:       a.Views,
			PublishedAt: a.PublishedAt.Format(time.RFC3339), // 带时区，前端 new Date() 才能解析对
			UpdatedAt:   a.UpdatedAt.Format(time.RFC3339),
		})
	}

	return &types.AdminArticleListResp{
		List:  list,
		Total: total,
	}, nil
}
