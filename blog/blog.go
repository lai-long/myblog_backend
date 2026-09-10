// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package main

import (
	"context"

	"database/sql"
	"flag"
	"fmt"
	"myblog_backend/migrations"
	"myblog_backend/pkg/logx"
	"net/http"
	"strings"

	"myblog_backend/blog/internal/config"
	"myblog_backend/blog/internal/handler"
	"myblog_backend/blog/internal/svc"
	"myblog_backend/pkg/errx"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/httpx"
	_ "modernc.org/sqlite"
)

// body 统一响应体：HTTP 状态码恒为 200，业务成败只看 code
// 成功：{"code":0,"message":"ok","data":...}
// 失败：{"code":40101,"message":"用户名或密码错误"}（data 省略）
type body struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

var configFile = flag.String("f", "etc/blog-api.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	server := rest.MustNewServer(c.RestConf,
		// JWT 鉴权失败时返回统一 {code,message}，而不是空 body
		rest.WithUnauthorizedCallback(func(w http.ResponseWriter, r *http.Request, err error) {
			code := errx.TokenInvalid
			msg := "未登录或登录已失效"
			if err != nil && strings.Contains(err.Error(), "expired") {
				code = errx.TokenExpired
				msg = "登录已过期，请重新登录"
			}
			httpx.WriteJson(w, http.StatusOK, body{Code: code, Message: msg})
		}),
	)
	defer server.Stop()

	// 统一成功响应：logic 返回的数据包进 data，补上 code=0
	httpx.SetOkHandler(func(ctx context.Context, v any) any {
		return body{
			Code:    errx.OK,
			Message: "ok",
			Data:    v,
		}
	})

	// 统一错误响应：logic 返回的 error 转成 {code, message}
	httpx.SetErrorHandlerCtx(func(ctx context.Context, err error) (int, interface{}) {
		return http.StatusOK, body{
			Code:    errx.CodeOf(err),
			Message: errx.MsgOf(err),
		}
	})

	// Must be called after MustNewServer, which calls logx.SetUp internally
	// and would otherwise overwrite our zerolog writer.
	if err := logx.Init(c.ZeroLog); err != nil {
		panic(err)
	}

	// 启动时自动执行未应用的迁移（幂等），空数据目录也能直接拉起
	db, err := sql.Open("sqlite", c.SqliteDSN)
	if err != nil {
		panic(err)
	}
	if err := migrations.Apply(db); err != nil {
		panic(err)
	}
	db.Close()

	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
