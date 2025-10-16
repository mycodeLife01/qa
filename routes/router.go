package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/mycodeLife01/qa/handlers"
	"gorm.io/gorm"
)

func SetupRouter(router *gin.Engine, db *gorm.DB) {
	userHandler := handlers.NewUserHandler(db)
	userRoutes := router.Group("/users")
	{
		userRoutes.POST("/register", userHandler.Register)
		userRoutes.POST("/login", userHandler.AuthService.InitJwtMiddleware().LoginHandler)
	}
	authRoutes := router.Group("/auth", userHandler.AuthService.InitJwtMiddleware().MiddlewareFunc())
	{
		authRoutes.GET("/hello", userHandler.SayHello)
	}
}
