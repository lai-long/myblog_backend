// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package admin

import (
	"context"

	"myblog_backend/blog/internal/svc"
	"myblog_backend/blog/internal/model"
	"myblog_backend/blog/internal/types"
	"myblog_backend/pkg/errx"

	"github.com/zeromicro/go-zero/core/logx"
)

type VisitStatsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewVisitStatsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *VisitStatsLogic {
	return &VisitStatsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *VisitStatsLogic) VisitStats() (resp *types.VisitStatsResp, err error) {
	today := model.Today()

	todayPv, err := l.svcCtx.VisitModel.PvOn(l.ctx, today)
	if err != nil {
		return nil, errx.Wrap(err, errx.ServerError, "统计查询失败")
	}
	todayUv, err := l.svcCtx.VisitModel.UvOn(l.ctx, today)
	if err != nil {
		return nil, errx.Wrap(err, errx.ServerError, "统计查询失败")
	}
	totalPv, err := l.svcCtx.VisitModel.PvOn(l.ctx, "")
	if err != nil {
		return nil, errx.Wrap(err, errx.ServerError, "统计查询失败")
	}
	totalUv, err := l.svcCtx.VisitModel.UvOn(l.ctx, "")
	if err != nil {
		return nil, errx.Wrap(err, errx.ServerError, "统计查询失败")
	}
	trend, err := l.svcCtx.VisitModel.Trend(l.ctx, 7)
	if err != nil {
		return nil, errx.Wrap(err, errx.ServerError, "统计查询失败")
	}

	resp = &types.VisitStatsResp{
		TodayPv: todayPv,
		TodayUv: todayUv,
		TotalPv: totalPv,
		TotalUv: totalUv,
		Trend:   make([]types.DayStat, 0, len(trend)),
	}
	for _, s := range trend {
		resp.Trend = append(resp.Trend, types.DayStat{Date: s.Date, Pv: s.Pv, Uv: s.Uv})
	}
	return resp, nil
}
