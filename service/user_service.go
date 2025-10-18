package service

import (
	"errors"

	"github.com/mycodeLife01/qa/models"
	"github.com/mycodeLife01/qa/types"
	"github.com/mycodeLife01/qa/utils"
	"gorm.io/gorm"
)

type UserService struct {
	DB *gorm.DB
}

func NewUserService(db *gorm.DB) *UserService {
	return &UserService{DB: db}
}

func (us *UserService) IsValidUser(username, password string) (*models.User, error) {
	var loginUser models.User
	err := us.DB.Where("username = ?", username).First(&loginUser).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 为了安全，不明确提示是“用户不存在”还是“密码错误”
			return nil, types.ErrUserInvalid
		}
		return nil, err
	}
	if !utils.CheckPasswordHash(password, loginUser.PasswordHashed) {
		return nil, types.ErrUserInvalid
	}
	return &loginUser, nil
}

func (us *UserService) Register(username, password, email string) (*types.UserResponse, error) {
	// 检查用户是否已存在
	var existingUser models.User
	err := us.DB.Where("username=?", username).First(&existingUser).Error
	if err == nil {
		return nil, types.ErrUserExists
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// 创建用户
	hashedPassword, hashErr := utils.HashPassword(password)
	if hashErr != nil {
		return nil, hashErr
	}
	newUser := models.User{
		Username:       username,
		PasswordHashed: hashedPassword,
		Email:          email,
	}
	resultErr := us.DB.Create(&newUser).Error
	if resultErr != nil {
		return nil, resultErr
	}
	return &types.UserResponse{
		Username: newUser.Username,
		Email:    newUser.Email,
	}, nil
}
