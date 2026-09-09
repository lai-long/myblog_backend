// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package admin

import (
	"context"

	"myblog_backend/blog/internal/svc"
	"myblog_backend/blog/internal/types"

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
	// todo: add your logic here and delete this line

	return
}
