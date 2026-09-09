// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package admin

import (
	"context"

	"myblog_backend/blog/internal/svc"
	"myblog_backend/blog/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ProfileGetLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewProfileGetLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ProfileGetLogic {
	return &ProfileGetLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ProfileGetLogic) ProfileGet() (resp *types.ProfileResp, err error) {
	// todo: add your logic here and delete this line

	return
}
