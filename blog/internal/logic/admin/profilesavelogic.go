// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package admin

import (
	"context"

	"myblog_backend/blog/internal/svc"
	"myblog_backend/blog/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ProfileSaveLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewProfileSaveLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ProfileSaveLogic {
	return &ProfileSaveLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ProfileSaveLogic) ProfileSave(req *types.ProfileSaveReq) (resp *types.EmptyResp, err error) {
	// todo: add your logic here and delete this line

	return
}
