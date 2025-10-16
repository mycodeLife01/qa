package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/mycodeLife01/qa/models"
	"github.com/mycodeLife01/qa/service"
	"github.com/mycodeLife01/qa/types"
	"github.com/mycodeLife01/qa/utils"
	"gorm.io/gorm"
)

// func TestHandler(c *gin.Context) {
// 	c.JSON(http.StatusOK, gin.H{"message": "Hello, World!"})
// }

type UserHandler struct {
	DB          *gorm.DB
	AuthService *service.AuthService
}

func NewUserHandler(db *gorm.DB) *UserHandler {
	return &UserHandler{DB: db, AuthService: service.NewAuthService(db)}
}

func (uh *UserHandler) GetUsers(c *gin.Context) {
	var users []models.User
	uh.DB.Find(&users)
	c.JSON(http.StatusOK, users)
}

func (uh *UserHandler) Register(c *gin.Context) {
	// 解析请求体
	var req types.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 检查用户是否已存在
	var existingUser models.User
	err := uh.DB.Where("username=?", req.Username).First(&existingUser).Error
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "username already exists"})
		return
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check username"})
	}

	// 创建用户
	hashedPassword, hashErr := utils.HashPassword(req.Password)
	if hashErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}
	newUser := models.User{
		Username:       req.Username,
		PasswordHashed: hashedPassword,
		Email:          req.Email,
	}
	result := uh.DB.Create(&newUser)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "user created successfully"})

}

func (uh *UserHandler) SayHello(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Hello, World!"})
}
