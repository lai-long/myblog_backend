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
	uid := uidFromCtx(l.ctx)
	if uid <= 0 {
		return nil, errx.New(errx.TokenInvalid, "无法识别登录身份")
	}

	admin, err := l.svcCtx.AdminModel.FindOne(l.ctx, uid)
	if err != nil {
		return nil, errx.Wrap(err, errx.ServerError, "查询用户失败")
	}

	return &types.ProfileResp{
		Username:  admin.Username,
		Nickname:  admin.Nickname,
		AvatarUrl: admin.AvatarUrl,
	}, nil
}
