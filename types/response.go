package types

// ResponseData 是统一的 API 响应结构体
type ResponseData struct {
	Code int    `json:"code"` // 业务状态码
	Msg  string `json:"msg"`  // 响应消息
	Data any    `json:"data"` // 响应数据
}

// Success 创建一个成功的响应实例
func Success(data any) *ResponseData {
	return &ResponseData{
		Code: 0, // 约定 0 为成功
		Msg:  "success",
		Data: data,
	}
}

// Fail 创建一个失败的响应实例（用于业务逻辑错误，非系统panic）
func Fail(msg string, code int) *ResponseData {
	return &ResponseData{
		Code: code, // 约定非 0 为各种错误
		Msg:  msg,
		Data: nil,
	}
}
