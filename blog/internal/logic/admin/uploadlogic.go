// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package admin

import (
	"context"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"myblog_backend/blog/internal/svc"
	"myblog_backend/blog/internal/types"
	"myblog_backend/pkg/errx"

	"github.com/zeromicro/go-zero/core/logx"
)

type UploadLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUploadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UploadLogic {
	return &UploadLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 允许的图片类型及其对应扩展名
var allowedImageTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/gif":  ".gif",
	"image/webp": ".webp",
}

const maxImageSize = 5 << 20 // 5MB

func (l *UploadLogic) Upload(data []byte, contentType, filename string) (resp *types.UploadResp, err error) {
	// 1. 校验类型
	ext, ok := allowedImageTypes[contentType]
	if !ok {
		return nil, errx.New(errx.ParamError, "仅支持 jpg/png/gif/webp 图片")
	}
	// 2. 校验大小
	if len(data) > maxImageSize {
		return nil, errx.New(errx.ParamError, "图片大小不能超过 5MB")
	}

	// 3. 生成对象 key：日期目录/uuid.后缀
	if e := strings.ToLower(filepath.Ext(filename)); e != "" {
		ext = e // 优先用原始扩展名，缺失时回退到类型推导
	}
	key := time.Now().UTC().Format("2006/01/02") + "/" + uuid.NewString() + ext

	// 4. 上传到 RustFS / S3 兼容存储
	url, err := l.svcCtx.OSS.Upload(l.ctx, key, data, contentType)
	if err != nil {
		return nil, errx.Wrap(err, errx.ServerError, "上传失败")
	}

	return &types.UploadResp{Url: url}, nil
}
