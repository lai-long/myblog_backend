package errx

// 通用
const (
	OK          = 0
	ParamError  = 40001 // 参数错误
	NotFound    = 40401 // 资源不存在
	ServerError = 50000
)

// 认证相关
const (
	LoginFailed  = 40101 // 用户名或密码错误
	TokenExpired = 40102
	TokenInvalid = 40103
)

// 预定义常用错误
var (
	ErrLoginFailed = New(LoginFailed, "用户名或密码错误")
)
