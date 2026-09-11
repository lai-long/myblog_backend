// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package admin

import (
	"context"

	"myblog_backend/blog/internal/svc"
	"myblog_backend/blog/internal/types"
	"myblog_backend/pkg/errx"
	"myblog_backend/pkg/pwd"

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
	if len(req.NewPassword) < 6 {
		return nil, errx.New(errx.ParamError, "新密码至少 6 位")
	}

	uid := uidFromCtx(l.ctx)
	if uid <= 0 {
		return nil, errx.New(errx.TokenInvalid, "无法识别登录身份")
	}

	admin, err := l.svcCtx.AdminModel.FindOne(l.ctx, uid)
	if err != nil {
		return nil, errx.Wrap(err, errx.ServerError, "查询用户失败")
	}

	// 旧密码校验：不对就按登录失败处理，不暴露"用户存在但密码错"之外的信息
	if !pwd.Compare(admin.PasswordHash, req.OldPassword) {
		return nil, errx.ErrLoginFailed
	}

	hash, err := pwd.Hash(req.NewPassword)
	if err != nil {
		return nil, errx.Wrap(err, errx.ServerError, "密码加密失败")
	}
	admin.PasswordHash = hash
	if err := l.svcCtx.AdminModel.Update(l.ctx, admin); err != nil {
		return nil, errx.Wrap(err, errx.ServerError, "密码更新失败")
	}

	return &types.EmptyResp{}, nil
}
