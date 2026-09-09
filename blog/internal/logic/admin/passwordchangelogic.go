// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package admin

import (
	"context"

	"myblog_backend/blog/internal/svc"
	"myblog_backend/blog/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type PasswordChangeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPasswordChangeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PasswordChangeLogic {
	return &PasswordChangeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PasswordChangeLogic) PasswordChange(req *types.PasswordReq) (resp *types.EmptyResp, err error) {
	// todo: add your logic here and delete this line

	return
}
