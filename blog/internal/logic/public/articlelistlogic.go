// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package public

import (
	"context"
	"time"

	"myblog_backend/blog/internal/logic/validate"
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
	if err := validate.Pagination(req.Page, req.Size); err != nil {
		return nil, err
	}

	// 游客只看已发布的，status 写死 1，不给外部传的余地
	published := int64(1)
	cond := model.ArticleListCond{
		Status:  &published,
		Keyword: req.Keyword,
		TagSlug: req.Tag,
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

	// 批量取这一页文章的标签（一次 SQL，避免 N+1）
	ids := make([]int64, 0, len(articles))
	for _, a := range articles {
		ids = append(ids, a.Id)
	}
	tagMap, err := l.svcCtx.TagModel.FindByArticleIds(l.ctx, ids)
	if err != nil {
		return nil, errx.Wrap(err, errx.ServerError, "查询文章标签失败")
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
			IsTop:       int(a.IsTop),
			PublishedAt: a.PublishedAt.Format(time.RFC3339), // 带时区，前端 new Date() 才能解析对
			Tags:        toTagVos(tagMap[a.Id]),
		})
	}

	return &types.ArticleListResp{
		List:  list,
		Total: total,
	}, nil
}
