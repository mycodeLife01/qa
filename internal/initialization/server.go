package initialization

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mycodeLife01/qa/config"
)

// InitHTTPServer 初始化HTTP服务器
func InitHTTPServer(engine *gin.Engine) *http.Server {
	server := &http.Server{
		Addr:        fmt.Sprintf(":%d", config.C.Server.Port),
		Handler:     engine,
		ReadTimeout: time.Duration(config.C.Server.ReadTimeout) * time.Second,
		IdleTimeout: 60 * time.Second,
	}
	return server
}
