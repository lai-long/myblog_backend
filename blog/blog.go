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

// errResp 统一错误响应体：HTTP 状态码恒为 200，业务成败只看 code
type errResp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

var configFile = flag.String("f", "etc/blog-api.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	// 统一错误处理：logic 返回的 error 一律转成 {code, message}
	httpx.SetErrorHandlerCtx(func(ctx context.Context, err error) (int, interface{}) {
		return http.StatusOK, errResp{
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
