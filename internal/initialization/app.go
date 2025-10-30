package initialization

import (
	"fmt"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/mycodeLife01/qa/config"
	"github.com/mycodeLife01/qa/internal/model"
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
	db.AutoMigrate(&model.User{}, &model.File{})

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

	corsConfig := cors.Config{
		// AllowOrigins:     []string{"http://localhost:3000", "https://your-frontend.com"}, // 允许的源
		AllowOrigins:     []string{"*"}, // 允许所有源，生产环境请务必指定特定源
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Accept"}, // 允许的请求头
		ExposeHeaders:    []string{"Content-Length"},                                    // 允许前端访问的响应头
		AllowCredentials: true,                                                          // 是否允许发送Cookie
		MaxAge:           12 * time.Hour,                                                // 预检请求（OPTIONS）的缓存时间
	}
	r.Use(cors.New(corsConfig))

	router.SetupAppRouter(r, handlers.AuthHandler, handlers.UserHandler, handlers.FileHandler, handlers.AiHandler, handlers.HealthHandler, middlewares.AuthMiddleware)

	// 7. 启动HTTP服务器
	server := InitHTTPServer(r)
	fmt.Printf("Server is running on port %d\n", config.C.Server.Port)

	if err := server.ListenAndServe(); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}
