// Package validate 提供各 logic 共用的请求参数校验。
package validate

import "myblog_backend/pkg/errx"

// MaxPageSize 单次查询上限，防止 size 过大一次拉全表
const MaxPageSize = 100

// Pagination 校验分页参数：page 从 1 开始，size 限制在 1~MaxPageSize。
// 非法参数返回 40001 业务错误，避免负 OFFSET 触发 SQLite 报错被包装成 500，
// 以及 LIMIT -1 在 SQLite 中等价于无限制导致全表返回。
func Pagination(page, size int) error {
	if page < 1 {
		return errx.New(errx.ParamError, "page 必须从 1 开始")
	}
	if size < 1 || size > MaxPageSize {
		return errx.New(errx.ParamError, "size 需在 1~100 之间")
	}
	return nil
}
