package response

// Response 标准响应格式
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Success 成功响应
func Success(data interface{}) Response {
	return Response{
		Code:    200,
		Message: "成功",
		Data:    data,
	}
}

// Created 创建成功响应
func Created(data interface{}) Response {
	return Response{
		Code:    201,
		Message: "创建成功",
		Data:    data,
	}
}

// Error 错误响应
func Error(code int, message string) Response {
	return Response{
		Code:    code,
		Message: message,
	}
}
