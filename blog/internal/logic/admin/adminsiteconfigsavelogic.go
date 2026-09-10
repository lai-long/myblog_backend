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

type AdminSiteConfigSaveLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminSiteConfigSaveLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminSiteConfigSaveLogic {
	return &AdminSiteConfigSaveLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminSiteConfigSaveLogic) AdminSiteConfigSave(req *types.SiteConfigSaveReq) (resp *types.EmptyResp, err error) {
	// 三个字段都是 optional，语义约定为"传了就更新，空串表示清空"：
	// 前端设置页总是全量提交，缺省字段解析成空串等于清空，行为一致
	kvs := []struct {
		key   string
		value string
	}{
		{"siteTitle", req.SiteTitle},
		{"icp", req.Icp},
		{"githubUrl", req.GithubUrl},
	}
	for _, kv := range kvs {
		if err := l.svcCtx.SettingsModel.Set(l.ctx, kv.key, kv.value); err != nil {
			return nil, errx.Wrap(err, errx.ServerError, "保存站点配置失败")
		}
	}
	return &types.EmptyResp{}, nil
}
