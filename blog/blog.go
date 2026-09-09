// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package main

import (
	"context"
	"flag"
	"fmt"
	"myblog_backend/pkg/logx"
	"net/http"

	"myblog_backend/blog/internal/config"
	"myblog_backend/blog/internal/handler"
	"myblog_backend/blog/internal/svc"
	"myblog_backend/pkg/errx"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/httpx"
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

	server := rest.MustNewServer(c.RestConf)
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

	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
