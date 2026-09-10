// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package public

import (
	"context"

	"myblog_backend/blog/internal/svc"
	"myblog_backend/blog/internal/types"
	"myblog_backend/pkg/errx"

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
	// 始终返回非空对象（没配过的字段就是空串），前端不用处理 data: null
	resp = &types.SiteConfigResp{}
	if resp.SiteTitle, err = l.svcCtx.SettingsModel.Get(l.ctx, "siteTitle"); err != nil {
		return nil, errx.Wrap(err, errx.ServerError, "读取站点配置失败")
	}
	if resp.Icp, err = l.svcCtx.SettingsModel.Get(l.ctx, "icp"); err != nil {
		return nil, errx.Wrap(err, errx.ServerError, "读取站点配置失败")
	}
	if resp.GithubUrl, err = l.svcCtx.SettingsModel.Get(l.ctx, "githubUrl"); err != nil {
		return nil, errx.Wrap(err, errx.ServerError, "读取站点配置失败")
	}
	return
}
