package admin

import (
	"context"
	"encoding/json"
)

// uidFromCtx 从 context 取当前登录用户 id。
// go-zero 的 JWT 中间件把 claims 放进 context，且用 WithJSONNumber 解析，
// 所以数字 claim 是 json.Number 类型（不是 float64）。
func uidFromCtx(ctx context.Context) int64 {
	switch v := ctx.Value("uid").(type) {
	case json.Number:
		uid, _ := v.Int64()
		return uid
	case float64:
		return int64(v)
	}
	return 0
}
