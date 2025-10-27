package initialization

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/mycodeLife01/qa/config"
	"github.com/mycodeLife01/qa/internal/pkg/client"
	"github.com/mycodeLife01/qa/internal/router"
)

// InitApp 初始化并启动应用
func InitApp() error {
	// 1. 初始化配置
	if err := InitConfig(); err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	fmt.Printf("JWT SECRET: %s\n", config.C.JWT.JWTSecretKey)
	fmt.Printf("COS SECRET ID: %s\n", config.C.COS.SecretID)
	fmt.Printf("COS SECRET KEY: %s\n", config.C.COS.SecretKey)

	// 2. 初始化数据库
	db, err := InitDatabase()
	if err != nil {
		return fmt.Errorf("failed to init database: %w", err)
	}

	// 3. 初始化服务层
	CosClient := client.InitCosClient()
	services := InitServices(db, CosClient)

	// 4. 初始化中间件
	middlewares, err := InitMiddleware(services)
	if err != nil {
		return fmt.Errorf("failed to create auth middleware: %w", err)
	}

	// 5. 初始化处理器
	handlers := InitHandlers(services, middlewares.AuthMiddleware)

	// 6. 设置路由
	r := gin.Default()
	r.Use(middlewares.ResponseHandler)
	router.SetupAppRouter(r, handlers.AuthHandler, handlers.UserHandler, handlers.FileHandler, middlewares.AuthMiddleware)

	// 7. 启动HTTP服务器
	server := InitHTTPServer(r)
	fmt.Printf("Server is running on port %d\n", config.C.Server.Port)

	if err := server.ListenAndServe(); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}
