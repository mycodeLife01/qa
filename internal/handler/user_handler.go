package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/mycodeLife01/qa/internal/dto"
	"github.com/mycodeLife01/qa/internal/model"
	"github.com/mycodeLife01/qa/internal/pkg/api"
	"github.com/mycodeLife01/qa/internal/pkg/security"
	"github.com/mycodeLife01/qa/internal/service"
)

type UserHandler struct {
	UserService service.UserService
}

func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{UserService: userService}
}

func (uh *UserHandler) GetAllUsers(c *gin.Context) {
	users, err := uh.UserService.FindAllUser()
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.Set(api.ResponseDataKey, users)
}

func (uh *UserHandler) GetUserByName(c *gin.Context) {
	name := c.Query("name")
	users, err := uh.UserService.FindUserByName(name)
	if err != nil {
		_ = c.Error(err)
		return
	}
	safeUsers := make([]dto.UserResponse, len(users))
	for i, user := range users {
		safeUsers[i] = dto.UserResponse{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
		}
	}
	c.Set(api.ResponseDataKey, safeUsers)
}
func (uh *UserHandler) AddUser(c *gin.Context) {
	var userInfo dto.AddUserRequest
	if err := c.ShouldBindJSON(&userInfo); err != nil {
		_ = c.Error(err)
		return
	}
	hashed, err := security.HashPassword(userInfo.Password)
	if err != nil {
		_ = c.Error(err)
		return
	}
	user := model.User{
		Username:       userInfo.Username,
		PasswordHashed: hashed,
		Email:          userInfo.Email,
		Role:           userInfo.Role,
	}
	newUser, err := uh.UserService.AddUser(user)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.Set(api.ResponseDataKey, newUser)
}
func (uh *UserHandler) UpdateUser(c *gin.Context) {
	var userInfo dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&userInfo); err != nil {
		_ = c.Error(err)
		return
	}
	var updatedUser model.User
	if userInfo.Password != nil {
		hashed, err := security.HashPassword(*userInfo.Password)
		if err != nil {
			_ = c.Error(err)
			return
		}
		updatedUser.PasswordHashed = hashed
	}
	if userInfo.Username != nil {
		updatedUser.Username = *userInfo.Username
	}
	if userInfo.Email != nil {
		updatedUser.Email = *userInfo.Email
	}
	if userInfo.Role != nil {
		updatedUser.Role = *userInfo.Role
	}

	updatedUser.ID = userInfo.ID
	user, err := uh.UserService.UpdateUser(updatedUser)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.Set(api.ResponseDataKey, user)
}

func (uh *UserHandler) DeleteUserById(c *gin.Context) {
	id := c.Query("id")
	uintId, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		_ = c.Error(err)
		return
	}
	deleted, err := uh.UserService.DeleteUserById(uint(uintId))
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.Set(api.ResponseDataKey, deleted)
}
