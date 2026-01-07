package router

import (
	jwt "github.com/appleboy/gin-jwt/v3"
	"github.com/gin-gonic/gin"
	"github.com/mycodeLife01/qa/internal/handler"
)

func SetupAppRouter(r *gin.Engine, authHandler *handler.AuthHandler, userHandler *handler.UserHandler, fileHandler *handler.FileHandler, aiHandler *handler.AiHandler, healthHandler *handler.HealthHandler, internalHandler *handler.InternalHandler, authMiddleware *jwt.GinJWTMiddleware) {

	// 健康检查接口（无需认证）
	r.GET("/health", healthHandler.HealthCheck)

	// 内部接口（无需认证，建议配合防火墙或Internal Token）
	internalRouterGroup := r.Group("/internal")
	internalRouterGroup.POST("/file/status", internalHandler.UpdateFileStatus)

	authRouterGroup := r.Group("/auth")
	userRouterGroup := r.Group("/user", authMiddleware.MiddlewareFunc())
	fileRouterGroup := r.Group("/file", authMiddleware.MiddlewareFunc())
	aiRouterGroup := r.Group("/ai", authMiddleware.MiddlewareFunc())

	SetupAuthRouterGroup(authRouterGroup, authHandler, authMiddleware)
	SetupUserRouterGroup(userRouterGroup, userHandler)
	SetupFileRouterGroup(fileRouterGroup, fileHandler)
	SetupAiRouterGroup(aiRouterGroup, aiHandler)
}
