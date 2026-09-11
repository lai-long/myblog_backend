// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package admin

import (
	"context"

	"myblog_backend/blog/internal/svc"
	"myblog_backend/blog/internal/types"
	"myblog_backend/pkg/errx"

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
	uid := uidFromCtx(l.ctx)
	if uid <= 0 {
		return nil, errx.New(errx.TokenInvalid, "无法识别登录身份")
	}

	admin, err := l.svcCtx.AdminModel.FindOne(l.ctx, uid)
	if err != nil {
		return nil, errx.Wrap(err, errx.ServerError, "查询用户失败")
	}

	admin.Nickname = req.Nickname
	admin.AvatarUrl = req.AvatarUrl
	if err := l.svcCtx.AdminModel.Update(l.ctx, admin); err != nil {
		return nil, errx.Wrap(err, errx.ServerError, "保存失败")
	}

	return &types.EmptyResp{}, nil
}
