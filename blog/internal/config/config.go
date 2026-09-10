// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package config

import (
	"myblog_backend/pkg/logx"

	"github.com/zeromicro/go-zero/rest"
)

type Config struct {
	rest.RestConf
	ZeroLog   logx.Config
	SqliteDSN string
	Auth      struct {
		AccessSecret string
		AccessExpire int64
	}
	OSS struct {
		// RustFS / S3 兼容对象存储配置
		Endpoint        string // 例如 https://s3.example.com
		Region          string // 例如 us-east-1
		Bucket          string
		AccessKeyID     string
		SecretAccessKey string
		UsePathStyle    bool   // 自建 S3 兼容服务通常为 true
		PublicBaseURL   string // 对象对外访问前缀，例如 https://s3.example.com/myblog
	}
}
