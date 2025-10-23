package initialization

import (
	"github.com/mycodeLife01/qa/internal/middleware"
)

// InitMiddleware 初始化中间件
func InitMiddleware(services *Services) (*Middlewares, error) {
	authMiddleware, err := middleware.NewAuthMiddleware(services.AuthService)
	if err != nil {
		return nil, err
	}
	return &Middlewares{
		AuthMiddleware:  authMiddleware,
		ResponseHandler: middleware.ResponseHandler(),
	}, nil
}
