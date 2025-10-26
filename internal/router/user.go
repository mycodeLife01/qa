package router

import (
	"github.com/gin-gonic/gin"
	"github.com/mycodeLife01/qa/internal/handler"
)

func SetupUserRouterGroup(rg *gin.RouterGroup, userHandler *handler.UserHandler) {
	{
		rg.GET("/all", userHandler.GetAllUsers)
		rg.GET("/find", userHandler.GetUserByName)
		rg.POST("/add", userHandler.AddUser)
		rg.PUT("/update", userHandler.UpdateUser)
		rg.DELETE("/remove", userHandler.DeleteUserById)
	}
}
