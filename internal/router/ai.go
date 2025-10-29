package router

import (
	"github.com/gin-gonic/gin"
	"github.com/mycodeLife01/qa/internal/handler"
)

func SetupAiRouterGroup(g *gin.RouterGroup, aiHandler *handler.AiHandler) {
	g.POST("/ask", aiHandler.Ask)
}
