// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package config

import (
	"myblog_backend/pkg/logx"

	"github.com/zeromicro/go-zero/rest"
)

type Config struct {
	rest.RestConf
	Log logx.Config
}
