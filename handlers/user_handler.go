package handlers

import (
	"github.com/gin-gonic/gin"

	"github.com/mycodeLife01/qa/middleware"
	"github.com/mycodeLife01/qa/service"
	"github.com/mycodeLife01/qa/types"
)

type UserHandler struct {
	UserService *service.UserService
}

func NewUserHandler(jwtService *service.JwtService, userService *service.UserService) *UserHandler {
	return &UserHandler{UserService: userService}
}

func (uh *UserHandler) Register(c *gin.Context) {
	// 解析请求体
	var req types.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(err)
		return
	}

	//调用UserService
	registerUser, err := uh.UserService.Register(req.Username, req.Password, req.Email)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.Set(middleware.ResponseDataKey, registerUser)

}
