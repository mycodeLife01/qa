package initialization

import (
	jwt "github.com/appleboy/gin-jwt/v3"
	"github.com/mycodeLife01/qa/internal/handler"
	"github.com/mycodeLife01/qa/internal/service"
	"gorm.io/gorm"
)

// Services 包含所有业务服务
type Services struct {
	AuthService service.AuthService
}

// Handlers 包含所有HTTP处理器
type Handlers struct {
	AuthHandler *handler.AuthHandler
}

// Dependencies 包含所有初始化后的依赖
type Dependencies struct {
	DB             *gorm.DB
	Services       *Services
	Handlers       *Handlers
	AuthMiddleware *jwt.GinJWTMiddleware
}
