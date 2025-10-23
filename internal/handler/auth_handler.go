package handler

import (
	jwt "github.com/appleboy/gin-jwt/v3"
	"github.com/gin-gonic/gin"
	"github.com/mycodeLife01/qa/internal/dto"
	"github.com/mycodeLife01/qa/internal/pkg/api"
	"github.com/mycodeLife01/qa/internal/service"
)

type AuthHandler struct {
	AuthService    service.AuthService
	authMiddleware *jwt.GinJWTMiddleware
}

func NewAuthHandler(authService service.AuthService, authMiddleware *jwt.GinJWTMiddleware) *AuthHandler {
	return &AuthHandler{AuthService: authService, authMiddleware: authMiddleware}
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

	// 注册成功后自动生成 JWT token
	tokenPair, err := ah.authMiddleware.TokenGenerator(registerUser)
	if err != nil {
		_ = c.Error(err)
		return
	}

	// 设置 cookie（如果启用了 SendCookie 配置）
	ah.authMiddleware.SetCookie(c, tokenPair.AccessToken)

	// 构造响应数据，包含用户信息和 token
	response := map[string]any{
		"user": dto.UserResponse{
			Username: registerUser.Username,
			Email:    registerUser.Email,
		},
		"access_token":  tokenPair.AccessToken,
		"token_type":    tokenPair.TokenType,
		"refresh_token": tokenPair.RefreshToken,
		"expires_in":    tokenPair.ExpiresIn(),
	}
	c.Set(api.ResponseDataKey, response)
}
