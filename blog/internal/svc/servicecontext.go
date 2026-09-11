// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package svc

import (
	"myblog_backend/blog/internal/config"
	"myblog_backend/blog/internal/model"
	"myblog_backend/pkg/oss"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
	_ "modernc.org/sqlite"
)

type ServiceContext struct {
	Config        config.Config
	AdminModel    model.AdminModel
	ArticleModel  model.ArticleModel
	CommentModel  model.CommentModel
	TagModel      *model.TagModel
	SettingsModel *model.SettingsModel
	VisitModel    *model.VisitModel
	OSS           *oss.Client
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewSqlConn("sqlite", c.SqliteDSN)

	ossClient, err := oss.New(oss.Config{
		Endpoint:        c.OSS.Endpoint,
		Region:          c.OSS.Region,
		Bucket:          c.OSS.Bucket,
		AccessKeyID:     c.OSS.AccessKeyID,
		SecretAccessKey: c.OSS.SecretAccessKey,
		UsePathStyle:    c.OSS.UsePathStyle,
		PublicBaseURL:   c.OSS.PublicBaseURL,
	})
	if err != nil {
		// 配置错误属于启动期致命问题，直接失败比带病运行更安全
		panic("oss init failed: " + err.Error())
	}

	return &ServiceContext{
		Config:        c,
		AdminModel:    model.NewAdminModel(conn),
		ArticleModel:  model.NewArticleModel(conn),
		CommentModel:  model.NewCommentModel(conn),
		TagModel:      model.NewTagModel(conn),
		SettingsModel: model.NewSettingsModel(conn),
		VisitModel:    model.NewVisitModel(conn),
		OSS:           ossClient,
	}
}
