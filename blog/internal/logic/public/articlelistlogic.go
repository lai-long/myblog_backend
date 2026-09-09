// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package public

import (
	"context"
	"time"

	"myblog_backend/blog/internal/model"
	"myblog_backend/blog/internal/svc"
	"myblog_backend/blog/internal/types"
	"myblog_backend/pkg/errx"

	"github.com/zeromicro/go-zero/core/logx"
)

type ArticleListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewArticleListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ArticleListLogic {
	return &ArticleListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ArticleListLogic) ArticleList(req *types.ArticleListReq) (resp *types.ArticleListResp, err error) {
	// 游客只看已发布的，status 写死 1，不给外部传的余地
	published := int64(1)
	cond := model.ArticleListCond{
		Status:  &published,
		Keyword: req.Keyword,
		Offset:  (req.Page - 1) * req.Size,
		Size:    req.Size,
	}

	articles, err := l.svcCtx.ArticleModel.FindPage(l.ctx, cond)
	if err != nil {
		return nil, errx.Wrap(err, errx.ServerError, "查询文章列表失败")
	}

	// 复用同一个 cond，总数和列表的过滤条件必然一致
	total, err := l.svcCtx.ArticleModel.Count(l.ctx, cond)
	if err != nil {
		return nil, errx.Wrap(err, errx.ServerError, "查询文章总数失败")
	}

	list := make([]types.ArticleSummary, 0, len(articles))
	for _, a := range articles {
		list = append(list, types.ArticleSummary{
			Id:          a.Id,
			Title:       a.Title,
			Slug:        a.Slug,
			Summary:     a.Summary,
			CoverUrl:    a.CoverUrl,
			Views:       a.Views,
			PublishedAt: a.PublishedAt.Format(time.RFC3339), // 带时区，前端 new Date() 才能解析对
		})
	}

	return &types.ArticleListResp{
		List:  list,
		Total: total,
	}, nil
}
