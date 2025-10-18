package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mycodeLife01/qa/api"
	"github.com/mycodeLife01/qa/config"
	"github.com/mycodeLife01/qa/middleware"
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

	// 创建Gin Router
	router := gin.Default()
	router.Use(middleware.ResponseHandler())
	api.SetupRouter(router, db)

	// 配置http服务
	server := &http.Server{
		Addr:        fmt.Sprintf(":%d", config.C.Server.Port),
		Handler:     router,
		ReadTimeout: time.Duration(config.C.Server.ReadTimeout) * time.Second,
		IdleTimeout: 60 * time.Second,
	}

	// 启动服务
	fmt.Printf("Server is running on port %d", config.C.Server.Port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
