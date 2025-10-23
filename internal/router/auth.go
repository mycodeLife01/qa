package router

import (
	jwt "github.com/appleboy/gin-jwt/v3"
	"github.com/gin-gonic/gin"
	"github.com/mycodeLife01/qa/internal/handler"
)

func SetupAuthRouterGroup(rg *gin.RouterGroup, authHandler *handler.AuthHandler, authMiddleware *jwt.GinJWTMiddleware) {
	{
		rg.POST("/login", authMiddleware.LoginHandler)
		rg.POST("/logout", authMiddleware.LogoutHandler)
		rg.GET("/refresh_token", authMiddleware.RefreshHandler)
		rg.POST("/register", authHandler.Register)
	}
}
