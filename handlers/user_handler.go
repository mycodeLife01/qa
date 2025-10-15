package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mycodeLife01/qa/models"
	"gorm.io/gorm"
)

// func TestHandler(c *gin.Context) {
// 	c.JSON(http.StatusOK, gin.H{"message": "Hello, World!"})
// }

type UserHandler struct {
	DB *gorm.DB
}

func NewUserHandler(db *gorm.DB) *UserHandler {
	return &UserHandler{DB: db}
}

func (uh *UserHandler) GetUsers(c *gin.Context) {
	var users []models.User
	uh.DB.Find(&users)
	c.JSON(http.StatusOK, users)
}
