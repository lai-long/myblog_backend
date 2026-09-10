// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package interact

import (
	"context"
	"errors"
	"time"

	"myblog_backend/blog/internal/model"
	"myblog_backend/blog/internal/svc"
	"myblog_backend/blog/internal/types"
	"myblog_backend/pkg/errx"

	"github.com/zeromicro/go-zero/core/logx"
)

type CommentListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCommentListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CommentListLogic {
	return &CommentListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CommentListLogic) CommentList(req *types.CommentListReq) (resp *types.CommentListResp, err error) {
	// 1. 文章必须存在且已发布，草稿/隐藏不提供评论
	a, err := l.svcCtx.ArticleModel.FindOneBySlug(l.ctx, req.Slug)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, errx.New(errx.NotFound, "文章不存在")
		}
		return nil, errx.Wrap(err, errx.ServerError, "查询文章失败")
	}
	if a.Status != 1 {
		return nil, errx.New(errx.NotFound, "文章不存在")
	}

	// 2. 只取已通过的评论（FindApprovedByArticleId 里写死了 status=1）
	comments, err := l.svcCtx.CommentModel.FindApprovedByArticleId(l.ctx, a.Id)
	if err != nil {
		return nil, errx.Wrap(err, errx.ServerError, "查询评论失败")
	}

	list := make([]types.Comment, 0, len(comments))
	for _, c := range comments {
		list = append(list, types.Comment{
			Id:        c.Id,
			ParentId:  c.ParentId,
			Nickname:  c.Nickname,
			AvatarUrl: c.AvatarUrl,
			Content:   c.Content,
			CreatedAt: c.CreatedAt.Format(time.RFC3339),
		})
	}

	return &types.CommentListResp{
		List:  list,
		Total: int64(len(list)),
	}, nil
}
