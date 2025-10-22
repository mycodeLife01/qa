package auth

import (
	"errors"

	"github.com/mycodeLife01/qa/internal/model"
	"github.com/mycodeLife01/qa/pkg/api"
	"github.com/mycodeLife01/qa/pkg/security"
	"gorm.io/gorm"
)

type authService struct {
	DB *gorm.DB
}

func NewAuthService(db *gorm.DB) AuthService {
	return &authService{DB: db}
}

func (as *authService) IsValidUser(username, password string) (*model.User, error) {
	var loginUser model.User
	err := as.DB.Where("username = ?", username).First(&loginUser).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 为了安全，不明确提示是“用户不存在”还是“密码错误”
			return nil, api.ErrUserInvalid
		}
		return nil, err
	}
	if !security.CheckPasswordHash(password, loginUser.PasswordHashed) {
		return nil, api.ErrUserInvalid
	}
	return &loginUser, nil
}

func (as *authService) Register(username, password, email string) (*model.User, error) {
	// 检查用户是否已存在
	var existingUser model.User
	err := as.DB.Where("username=?", username).First(&existingUser).Error
	if err == nil {
		return nil, api.ErrUserExists
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// 创建用户
	hashedPassword, hashErr := security.HashPassword(password)
	if hashErr != nil {
		return nil, hashErr
	}
	newUser := model.User{
		Username:       username,
		PasswordHashed: hashedPassword,
		Email:          email,
	}
	resultErr := as.DB.Create(&newUser).Error
	if resultErr != nil {
		return nil, resultErr
	}
	return &model.User{
		Username: newUser.Username,
		Email:    newUser.Email,
	}, nil
}
