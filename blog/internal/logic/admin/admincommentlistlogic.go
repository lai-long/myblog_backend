// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package admin

import (
	"context"

	"myblog_backend/blog/internal/svc"
	"myblog_backend/blog/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminCommentListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminCommentListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminCommentListLogic {
	return &AdminCommentListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminCommentListLogic) AdminCommentList(req *types.AdminCommentListReq) (resp *types.AdminCommentListResp, err error) {
	// todo: add your logic here and delete this line

	return
}
