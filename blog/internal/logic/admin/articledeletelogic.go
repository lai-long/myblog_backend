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

type ArticleDeleteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewArticleDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ArticleDeleteLogic {
	return &ArticleDeleteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ArticleDeleteLogic) ArticleDelete(req *types.ArticleIdReq) (resp *types.EmptyResp, err error) {
	// 先确认文章存在（FindOne 已排除软删的，删过的再次请求就是 404）
	if _, err := l.svcCtx.ArticleModel.FindOne(l.ctx, req.Id); err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, errx.New(errx.NotFound, "文章不存在")
		}
		return nil, errx.Wrap(err, errx.ServerError, "查询文章失败")
	}

	// 软删除：只写 deleted_at，数据仍留在库里
	if err := l.svcCtx.ArticleModel.Delete(l.ctx, req.Id); err != nil {
		return nil, errx.Wrap(err, errx.ServerError, "删除文章失败")
	}

	return &types.EmptyResp{}, nil
}
