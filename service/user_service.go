package service

import (
	"errors"

	"github.com/mycodeLife01/qa/models"
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
			return nil, errors.New("invalid credentials")
		}
		return nil, err
	}
	if !utils.CheckPasswordHash(password, loginUser.PasswordHashed) {
		return nil, errors.New("invalid credentials")
	}
	return &loginUser, nil
}
