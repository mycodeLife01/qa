package api

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/mycodeLife01/qa/handlers"
	"github.com/mycodeLife01/qa/service"
	"gorm.io/gorm"
)

func SetupRouter(router *gin.Engine, db *gorm.DB) error {
	userService := service.NewUserService(db)
	jwtService := service.NewJwtService(db)

	jwtMiddleware, err := jwtService.InitJwtMiddleware()
	if err != nil {
		log.Fatal("Failed to initialize JWT middleware:", err)
		return err
	}

	userHandler := handlers.NewUserHandler(jwtService, userService)

	publicUserRoutes := router.Group("/users")
	protectedUserRoutes := router.Group("/users", jwtMiddleware.MiddlewareFunc())

	SetupPublicUserRoutes(publicUserRoutes, userHandler, jwtMiddleware)
	SetupProtectedUserRoutes(protectedUserRoutes, userHandler, jwtMiddleware)
	return nil
}
