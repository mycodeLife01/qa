package initialization

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/mycodeLife01/qa/config"
	"github.com/mycodeLife01/qa/internal/router"
)

// InitApp 初始化并启动应用
func InitApp() error {
	// 1. 初始化配置
	if err := InitConfig(); err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	fmt.Printf("JWT SECRET: %s\n", config.C.JWT.JWTSecretKey)

	// 2. 初始化数据库
	db, err := InitDatabase()
	if err != nil {
		return fmt.Errorf("failed to init database: %w", err)
	}

	// 3. 初始化服务层
	services := InitServices(db)

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
	router.SetupAppRouter(r, handlers.AuthHandler, middlewares.AuthMiddleware)

	// 7. 启动HTTP服务器
	server := InitHTTPServer(r)
	fmt.Printf("Server is running on port %d\n", config.C.Server.Port)

	if err := server.ListenAndServe(); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}

// InitDependencies 初始化所有依赖（供测试使用）
func InitDependencies() (*Dependencies, error) {
	// 初始化配置
	if err := InitConfig(); err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// 初始化数据库
	db, err := InitDatabase()
	if err != nil {
		return nil, fmt.Errorf("failed to init database: %w", err)
	}

	// 初始化服务
	services := InitServices(db)

	// 初始化中间件
	middlewares, err := InitMiddleware(services)
	if err != nil {
		return nil, fmt.Errorf("failed to create auth middleware: %w", err)
	}

	// 初始化处理器
	handlers := InitHandlers(services, middlewares.AuthMiddleware)

	return &Dependencies{
		DB:          db,
		Services:    services,
		Handlers:    handlers,
		Middlewares: middlewares,
	}, nil
}
