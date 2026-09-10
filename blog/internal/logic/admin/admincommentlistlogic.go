// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package admin

import (
	"context"
	"time"

	"myblog_backend/blog/internal/model"
	"myblog_backend/blog/internal/svc"
	"myblog_backend/blog/internal/types"
	"myblog_backend/pkg/errx"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminCommentListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminCommentListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminCommentListLogic {
	return &AdminCommentListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminCommentListLogic) AdminCommentList(req *types.AdminCommentListReq) (resp *types.AdminCommentListResp, err error) {
	cond := model.AdminCommentListCond{
		Offset: (req.Page - 1) * req.Size,
		Size:   req.Size,
	}
	// Status = -1 表示不过滤，看全部
	if req.Status >= 0 {
		status := int64(req.Status)
		cond.Status = &status
	}

	rows, err := l.svcCtx.CommentModel.FindAdminPage(l.ctx, cond)
	if err != nil {
		return nil, errx.Wrap(err, errx.ServerError, "查询评论列表失败")
	}

	total, err := l.svcCtx.CommentModel.CountAdmin(l.ctx, cond)
	if err != nil {
		return nil, errx.Wrap(err, errx.ServerError, "查询评论总数失败")
	}

	list := make([]types.AdminComment, 0, len(rows))
	for _, r := range rows {
		list = append(list, types.AdminComment{
			Id:           r.Id,
			ArticleId:    r.ArticleId,
			ArticleTitle: r.ArticleTitle,
			ParentId:     r.ParentId,
			Nickname:     r.Nickname,
			Content:      r.Content,
			Status:       int(r.Status),
			CreatedAt:    r.CreatedAt.Format(time.RFC3339),
		})
	}

	return &types.AdminCommentListResp{
		List:  list,
		Total: total,
	}, nil
}
