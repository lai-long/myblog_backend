package handler

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net"
	"net/http"
	"strings"

	"myblog_backend/blog/internal/model"
	"myblog_backend/blog/internal/svc"

	"github.com/zeromicro/go-zero/rest"
)

// 页面上报浏览记录：POST /v1/track {"path": "/articles/hello-world"}
// 需要 *http.Request 拿客户端 IP，goctl 生成的 logic 拿不到，所以走裸路由。

type trackReq struct {
	Path string `json:"path"`
}

// clientIP 优先取 X-Forwarded-For 第一跳（nginx 反代会带上），否则用直连地址
func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return strings.TrimSpace(strings.Split(xff, ",")[0])
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func trackHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req trackReq
		// body 解析失败或 path 为空都不报错——统计是附属功能，不影响访客体验
		if err := json.NewDecoder(r.Body).Decode(&req); err == nil && req.Path != "" {
			// 用 AccessSecret 当盐做哈希：能去重，又反推不出原始 IP
			sum := sha256.Sum256([]byte(svcCtx.Config.Auth.AccessSecret + "|" + clientIP(r)))
			ipHash := hex.EncodeToString(sum[:8]) // 8 字节对小站足够区分访客
			// 记录失败只记日志，不影响响应
			if err := svcCtx.VisitModel.Record(r.Context(), model.Today(), ipHash, req.Path); err != nil {
				// 统计写库失败不值得让访客看到错误
				_ = err
			}
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"code":0,"message":"ok"}`))
	}
}

// RegisterVisitRoutes 注册访问上报裸路由
func RegisterVisitRoutes(server *rest.Server, svcCtx *svc.ServiceContext) {
	server.AddRoute(rest.Route{Method: http.MethodPost, Path: "/v1/track", Handler: trackHandler(svcCtx)})
}
