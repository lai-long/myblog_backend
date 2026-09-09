// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package admin

import (
	"context"

	"myblog_backend/blog/internal/svc"
	"myblog_backend/blog/internal/types"

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
	// todo: add your logic here and delete this line

	return
}
