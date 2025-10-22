package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/mycodeLife01/qa/internal/handler/dto"
	"github.com/mycodeLife01/qa/pkg/api"
)

type AuthHandler struct {
	AuthService AuthService
}

func NewAuthHandler(authService AuthService) *AuthHandler {
	return &AuthHandler{AuthService: authService}
}

func (ah *AuthHandler) Register(c *gin.Context) {
	// 解析请求体
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(err)
		return
	}

	//调用AuthService
	registerUser, err := ah.AuthService.Register(req.Username, req.Password, req.Email)
	if err != nil {
		_ = c.Error(err)
		return
	}
	userResponse := dto.UserResponse{
		Username: registerUser.Username,
		Email:    registerUser.Email,
	}
	c.Set(api.ResponseDataKey, userResponse)
}
