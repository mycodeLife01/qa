package router

import (
	jwt "github.com/appleboy/gin-jwt/v3"
	"github.com/gin-gonic/gin"
	"github.com/mycodeLife01/qa/internal/handler"
)

func SetupAppRouter(r *gin.Engine, authHandler *handler.AuthHandler, authMiddleware *jwt.GinJWTMiddleware) {
	authRouterGroup := r.Group("/auth")
	SetupAuthRouterGroup(authRouterGroup, authHandler, authMiddleware)

}
