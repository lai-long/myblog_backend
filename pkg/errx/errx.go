package errx

import (
	"errors"
	"fmt"
)

// Error 业务错误：对外暴露 Code + Msg，对内保留原始错误链
type Error struct {
	Code int    // 业务码：40001 参数错误、40101 未认证…
	Msg  string // 给前端看的友好提示
	err  error  // 原始错误，仅日志用，不暴露给前端
}

func (e *Error) Error() string {
	if e.err != nil {
		return fmt.Sprintf("code=%d msg=%s: %v", e.Code, e.Msg, e.err)
	}
	return fmt.Sprintf("code=%d msg=%s", e.Code, e.Msg)
}

// Unwrap 支持 errors.Is / errors.As 追溯原始错误
func (e *Error) Unwrap() error { return e.err }

// New 新建业务错误（无底层错误时用）
func New(code int, msg string) *Error {
	return &Error{Code: code, Msg: msg}
}

// Wrap 包装底层错误：对外只说友好提示，原始错误留在链上供日志追溯
func Wrap(err error, code int, msg string) *Error {
	return &Error{Code: code, Msg: msg, err: err}
}

// CodeOf 从任意 error 提取业务码，非业务错误返回 500
func CodeOf(err error) int {
	var e *Error
	if errors.As(err, &e) {
		return e.Code
	}
	return 50000
}

// MsgOf 提取对外提示语
func MsgOf(err error) string {
	var e *Error
	if errors.As(err, &e) {
		return e.Msg
	}
	return "服务器开小差了"
}
