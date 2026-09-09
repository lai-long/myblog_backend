// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package admin

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"myblog_backend/blog/internal/logic/admin"
	"myblog_backend/blog/internal/svc"
	"myblog_backend/blog/internal/types"
)

func ProfileSaveHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ProfileSaveReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := admin.NewProfileSaveLogic(r.Context(), svcCtx)
		resp, err := l.ProfileSave(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
