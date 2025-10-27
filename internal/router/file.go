package router

import (
	"github.com/gin-gonic/gin"
	"github.com/mycodeLife01/qa/internal/handler"
)

func SetupFileRouterGroup(rg *gin.RouterGroup, fileHandler *handler.FileHandler) {
	{
		rg.POST("/upload", fileHandler.UploadFile)
	}
}
