package initialization

import (
	jwt "github.com/appleboy/gin-jwt/v3"
	"github.com/mycodeLife01/qa/internal/handler"
)

// InitHandlers 初始化所有处理器
func InitHandlers(services *Services, authMiddleware *jwt.GinJWTMiddleware) *Handlers {
	return &Handlers{
		AuthHandler: handler.NewAuthHandler(services.AuthService, authMiddleware),
	}
}
