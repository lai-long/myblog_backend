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

type TagListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewTagListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TagListLogic {
	return &TagListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *TagListLogic) TagList() (resp *types.TagListResp, err error) {
	rows, err := l.svcCtx.TagModel.FindAllWithCount(l.ctx)
	if err != nil {
		return nil, errx.Wrap(err, errx.ServerError, "查询标签列表失败")
	}

	list := make([]types.TagWithCount, 0, len(rows))
	for _, r := range rows {
		list = append(list, types.TagWithCount{
			Name:  r.Name,
			Slug:  r.Slug,
			Count: r.Count,
		})
	}
	return &types.TagListResp{List: list}, nil
}
