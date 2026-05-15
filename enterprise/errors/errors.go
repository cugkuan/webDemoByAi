package errors

import "net/http"

// AppError 应用错误
type AppError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Err     error  `json:"-"`
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

// 预定义错误
var (
	ErrNotFound       = New(404, "资源不存在")
	ErrBadRequest     = New(400, "无效的请求")
	ErrInternalServer = New(500, "服务器内部错误")
	ErrTitleRequired  = New(400, "标题不能为空")
	ErrInvalidID      = New(400, "无效的任务 ID")
	ErrRateLimited    = New(429, "请求过于频繁，请稍后再试")
)

// New 创建应用错误
func New(code int, message string) *AppError {
	return &AppError{Code: code, Message: message}
}

// Wrap 包装已有错误
func Wrap(code int, message string, err error) *AppError {
	return &AppError{Code: code, Message: message, Err: err}
}

// HTTPStatus 获取 HTTP 状态码
func HTTPStatus(code int) int {
	switch {
	case code >= 1000:
		return http.StatusInternalServerError
	default:
		return code
	}
}
