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
		userRoutes.GET("", userHandler.GetUsers)
	}
}
