package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mycodeLife01/qa/config"
	"github.com/mycodeLife01/qa/internal/auth"
	"github.com/mycodeLife01/qa/internal/middleware"
	"github.com/mycodeLife01/qa/pkg/api"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	// 加载Viper配置
	if err := config.LoadConfig(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 连接数据库
	dsn := config.C.Database.DatabaseURL
	fmt.Println("Database URL: ", dsn)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database")
	}
	// 初始化服务
	authService := auth.NewAuthService(db)
	authHandler := auth.NewAuthHandler(authService)
	authMiddleware, err := middleware.NewAuthMiddleware(authService)
	if err != nil {
		log.Fatalf("Failed to create auth middleware: %v", err)
	}

	// 创建Gin Router
	r := gin.Default()
	// 配置统一返回中间件
	r.Use(middleware.ResponseHandler())

	// 配置auth路由组
	authRoutes := r.Group("/auth")
	{
		authRoutes.POST("/login", authMiddleware.LoginHandler)
		authRoutes.POST("/logout", authMiddleware.LogoutHandler)
		authRoutes.GET("/refresh_token", authMiddleware.RefreshHandler)
		authRoutes.POST("/register", authHandler.Register)
	}

	// 配置受保护路由组
	privateGroup := r.Group("/api/v1", authMiddleware.MiddlewareFunc())
	{
		privateGroup.POST("/test", func(c *gin.Context) {
			c.Set(api.ResponseDataKey, "user authenticated")
		})
	}

	// 配置http服务
	server := &http.Server{
		Addr:        fmt.Sprintf(":%d", config.C.Server.Port),
		Handler:     r,
		ReadTimeout: time.Duration(config.C.Server.ReadTimeout) * time.Second,
		IdleTimeout: 60 * time.Second,
	}

	// 启动服务
	fmt.Printf("Server is running on port %d", config.C.Server.Port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
