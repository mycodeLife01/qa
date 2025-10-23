package initialization

import (
	"github.com/mycodeLife01/qa/internal/handler"
)

// InitHandlers 初始化所有处理器
func InitHandlers(services *Services) *Handlers {
	return &Handlers{
		AuthHandler: handler.NewAuthHandler(services.AuthService),
	}
}
