// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package interact

import (
	"context"

	"myblog_backend/blog/internal/svc"
	"myblog_backend/blog/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CommentCreateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCommentCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CommentCreateLogic {
	return &CommentCreateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CommentCreateLogic) CommentCreate(req *types.CommentSaveReq) (resp *types.EmptyResp, err error) {
	// todo: add your logic here and delete this line

	return
}
