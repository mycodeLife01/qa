package api

import (
	jwt "github.com/appleboy/gin-jwt/v3"
	"github.com/gin-gonic/gin"
	"github.com/mycodeLife01/qa/handlers"
)

func SetupPublicUserRoutes(rg *gin.RouterGroup, userHandler *handlers.UserHandler, jwtMiddleware *jwt.GinJWTMiddleware) {
	rg.POST("/register", userHandler.Register)
	rg.POST("/login", jwtMiddleware.LoginHandler) // login 不需要 token，放在 public 路由
}

func SetupProtectedUserRoutes(rg *gin.RouterGroup, userHandler *handlers.UserHandler, jwtMiddleware *jwt.GinJWTMiddleware) {
	rg.POST("/logout", jwtMiddleware.LogoutHandler) // logout 需要 token，放在 protected 路由
}
