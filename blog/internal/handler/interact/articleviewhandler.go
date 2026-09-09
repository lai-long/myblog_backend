// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package interact

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"myblog_backend/blog/internal/logic/interact"
	"myblog_backend/blog/internal/svc"
)

func ArticleViewHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := interact.NewArticleViewLogic(r.Context(), svcCtx)
		resp, err := l.ArticleView()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
