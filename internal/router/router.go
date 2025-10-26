package router

import (
	jwt "github.com/appleboy/gin-jwt/v3"
	"github.com/gin-gonic/gin"
	"github.com/mycodeLife01/qa/internal/handler"
)

func SetupAppRouter(r *gin.Engine, authHandler *handler.AuthHandler, userHandler *handler.UserHandler, authMiddleware *jwt.GinJWTMiddleware) {

	authRouterGroup := r.Group("/auth")
	userRouterGroup := r.Group("/user", authMiddleware.MiddlewareFunc())

	SetupAuthRouterGroup(authRouterGroup, authHandler, authMiddleware)
	SetupUserRouterGroup(userRouterGroup, userHandler)
}
