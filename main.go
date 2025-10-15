package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/mycodeLife01/qa/routes"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	router := gin.Default()
	dsn := "root:89757@tcp(127.0.0.1:3306)/qa?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database")
	}
	routes.SetupRouter(router, db)
	router.Run()
}
