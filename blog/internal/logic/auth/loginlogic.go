// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package auth

import (
	"context"
	"time"

	"myblog_backend/blog/internal/svc"
	"myblog_backend/blog/internal/types"
	"myblog_backend/pkg/errx"
	"myblog_backend/pkg/pwd"

	"github.com/golang-jwt/jwt/v4"
	"github.com/zeromicro/go-zero/core/logx"
)

type LoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LoginLogic) Login(req *types.LoginReq) (resp *types.LoginResp, err error) {
	// 1. 按用户名查 admin
	admin, err := l.svcCtx.AdminModel.FindOneByUsername(l.ctx, req.Username)
	if err != nil {
		return nil, errx.Wrap(err, errx.LoginFailed, "用户名或密码错误")
	}

	// 2. 验证密码（用我们写的 pkg/pwd）
	if !pwd.Compare(admin.PasswordHash, req.Password) {
		return nil, errx.ErrLoginFailed
	}

	// 3. 签发 JWT
	now := time.Now()
	expire := l.svcCtx.Config.Auth.AccessExpire
	claims := jwt.MapClaims{
		"uid": admin.Id,
		"iat": now.Unix(),
		"exp": now.Add(time.Duration(expire) * time.Second).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessToken, err := token.SignedString([]byte(l.svcCtx.Config.Auth.AccessSecret))
	if err != nil {
		return nil, err
	}

	return &types.LoginResp{
		AccessToken:  accessToken,
		RefreshToken: "", // refresh token 下一步单独做，先留空
		ExpiresIn:    expire,
	}, nil
}
