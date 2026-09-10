// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package admin

import (
	"io"
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"myblog_backend/blog/internal/logic/admin"
	"myblog_backend/blog/internal/svc"
	"myblog_backend/pkg/errx"
)

func UploadHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// go-zero 的 httpx.Parse 不支持文件上传，这里手动读 multipart 的 "file" 字段
		file, header, err := r.FormFile("file")
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, errx.New(errx.ParamError, "缺少 file 字段"))
			return
		}
		defer file.Close()

		data, err := io.ReadAll(file)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, errx.Wrap(err, errx.ServerError, "读取文件失败"))
			return
		}

		l := admin.NewUploadLogic(r.Context(), svcCtx)
		resp, err := l.Upload(data, header.Header.Get("Content-Type"), header.Filename)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
