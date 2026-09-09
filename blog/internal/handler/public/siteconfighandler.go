// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package public

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"myblog_backend/blog/internal/logic/public"
	"myblog_backend/blog/internal/svc"
)

func SiteConfigHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := public.NewSiteConfigLogic(r.Context(), svcCtx)
		resp, err := l.SiteConfig()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
