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

type CommentDeleteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCommentDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CommentDeleteLogic {
	return &CommentDeleteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CommentDeleteLogic) CommentDelete(req *types.CommentIdReq) (resp *types.EmptyResp, err error) {
	// 先确认存在（FindOne 已排除软删的，删过的再次请求就是 404）
	if _, err := l.svcCtx.CommentModel.FindOne(l.ctx, req.Id); err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, errx.New(errx.NotFound, "评论不存在")
		}
		return nil, errx.Wrap(err, errx.ServerError, "查询评论失败")
	}

	// 软删除：只写 deleted_at
	if err := l.svcCtx.CommentModel.Delete(l.ctx, req.Id); err != nil {
		return nil, errx.Wrap(err, errx.ServerError, "删除评论失败")
	}

	return &types.EmptyResp{}, nil
}
