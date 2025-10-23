package initialization

import (
	jwt "github.com/appleboy/gin-jwt/v3"
	"github.com/mycodeLife01/qa/internal/middleware"
)

// InitMiddleware 初始化中间件
func InitMiddleware(services *Services) (*jwt.GinJWTMiddleware, error) {
	authMiddleware, err := middleware.NewAuthMiddleware(services.AuthService)
	if err != nil {
		return nil, err
	}
	return authMiddleware, nil
}
