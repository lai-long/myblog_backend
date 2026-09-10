// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package interact

import (
	"context"
	"errors"
	"strings"

	"myblog_backend/blog/internal/model"
	"myblog_backend/blog/internal/svc"
	"myblog_backend/blog/internal/types"
	"myblog_backend/pkg/errx"

	"github.com/zeromicro/go-zero/core/logx"
)

// defaultNickname 游客不填昵称时的兜底
const defaultNickname = "匿名用户"

// 长度上限：数据库字段有限，超长直接拒绝，避免落库报错
const (
	maxCommentContent  = 2000
	maxCommentNickname = 32
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
	// 1. 按 slug 找文章。草稿/隐藏的不能评论，对外统一说"文章不存在"
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

	// 2. 回复必须挂在同篇文章下，否则拿别人的 parentId 就能串到别的文章
	if req.ParentId > 0 {
		parent, err := l.svcCtx.CommentModel.FindOne(l.ctx, req.ParentId)
		if err != nil {
			if errors.Is(err, model.ErrNotFound) {
				return nil, errx.New(errx.ParamError, "父评论不存在")
			}
			return nil, errx.Wrap(err, errx.ServerError, "查询父评论失败")
		}
		if parent.ArticleId != a.Id {
			return nil, errx.New(errx.ParamError, "父评论不属于这篇文章")
		}
	}

	content := strings.TrimSpace(req.Content)
	if content == "" {
		return nil, errx.New(errx.ParamError, "评论内容不能为空")
	}
	if len([]rune(content)) > maxCommentContent {
		return nil, errx.New(errx.ParamError, "评论内容过长（上限 2000 字）")
	}

	nickname := strings.TrimSpace(req.Nickname)
	if nickname == "" {
		nickname = defaultNickname
	}
	if len([]rune(nickname)) > maxCommentNickname {
		return nil, errx.New(errx.ParamError, "昵称过长（上限 32 字）")
	}

	// 3. 入库，status=0 待审核，管理员通过后才会公开
	_, err = l.svcCtx.CommentModel.Insert(l.ctx, &model.Comment{
		ArticleId: a.Id,
		ParentId:  req.ParentId,
		Nickname:  nickname,
		Content:   content,
		Status:    0,
	})
	if err != nil {
		return nil, errx.Wrap(err, errx.ServerError, "提交评论失败")
	}

	return &types.EmptyResp{}, nil
}
