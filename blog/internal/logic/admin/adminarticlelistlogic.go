// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package admin

import (
	"context"

	"myblog_backend/blog/internal/svc"
	"myblog_backend/blog/internal/types"

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

func (l *AdminArticleListLogic) AdminArticleList(req *types.AdminArticleListReq) (resp *types.ArticleListResp, err error) {
	// todo: add your logic here and delete this line

	return
}
