package api

type AppError struct {
	Code    int    // 业务自定义错误码
	Message string // 错误信息
	Err     error  // 原始的 error，方便记录日志
}

func (e *AppError) Error() string {
	return e.Message
}

func NewAppError(code int, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

var (
	ErrUserExists           = NewAppError(1001, "user already exists", nil)
	ErrUserInvalid          = NewAppError(1002, "user invalid", nil)
	ErrUploadFileExtInvalid = NewAppError(1003, "file extension invalid", nil)
)
