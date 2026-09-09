// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package svc

import (
	"myblog_backend/blog/internal/config"
	"myblog_backend/blog/internal/model"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config     config.Config
	AdminModel model.AdminModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewSqlConn("sqlite", c.SqliteDSN)
	return &ServiceContext{
		Config:     c,
		AdminModel: model.NewAdminModel(conn),
	}
}
