// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package public

import (
	"context"

	"myblog_backend/blog/internal/svc"
	"myblog_backend/blog/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SiteConfigLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSiteConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SiteConfigLogic {
	return &SiteConfigLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SiteConfigLogic) SiteConfig() (resp *types.SiteConfigResp, err error) {
	// todo: add your logic here and delete this line

	return
}
