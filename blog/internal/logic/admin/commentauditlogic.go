// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package admin

import (
	"context"
	"errors"

	"myblog_backend/blog/internal/model"
	"myblog_backend/blog/internal/svc"
	"myblog_backend/blog/internal/types"
	"myblog_backend/pkg/errx"

	"github.com/zeromicro/go-zero/core/logx"
)

type CommentAuditLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCommentAuditLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CommentAuditLogic {
	return &CommentAuditLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CommentAuditLogic) CommentAudit(req *types.CommentAuditReq) (resp *types.EmptyResp, err error) {
	// 只接受 1=通过 2=垃圾，其他值（含 0 待审核）一律拒绝
	if req.Status != 1 && req.Status != 2 {
		return nil, errx.New(errx.ParamError, "状态只能是 1（通过）或 2（垃圾）")
	}

	if _, err := l.svcCtx.CommentModel.FindOne(l.ctx, req.Id); err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, errx.New(errx.NotFound, "评论不存在")
		}
		return nil, errx.Wrap(err, errx.ServerError, "查询评论失败")
	}

	if err := l.svcCtx.CommentModel.UpdateStatus(l.ctx, req.Id, int64(req.Status)); err != nil {
		return nil, errx.Wrap(err, errx.ServerError, "审核失败")
	}

	return &types.EmptyResp{}, nil
}
