package impl

import (
	"errors"

	"github.com/mycodeLife01/qa/internal/model"
	"github.com/mycodeLife01/qa/internal/pkg/api"
	"github.com/mycodeLife01/qa/internal/service"
	"gorm.io/gorm"
)

type userService struct {
	DB *gorm.DB
}

func NewUserService(db *gorm.DB) service.UserService {
	return &userService{DB: db}
}

func (us *userService) FindAllUser() ([]model.User, error) {
	users := []model.User{}
	err := us.DB.Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (us *userService) FindUserByName(name string) ([]model.User, error) {
	users := []model.User{}
	err := us.DB.Where("username like ?", "%"+name+"%").Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (us *userService) AddUser(user model.User) (*model.User, error) {
	userInDB := model.User{}
	err := us.DB.First(&userInDB, user.ID).Error
	if err == nil {
		return nil, api.ErrUserExists
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		createErr := us.DB.Create(&user).Error
		if createErr != nil {
			return nil, createErr
		}
		return &user, nil
	} else {
		return nil, err
	}
}

func (us *userService) UpdateUser(user model.User) (*model.User, error) {
	userInDB := model.User{}
	err := us.DB.First(&userInDB, user.ID).Error
	if err == nil {
		updateErr := us.DB.Model(&userInDB).Updates(user).Error
		if updateErr != nil {
			return nil, updateErr
		}
		return &userInDB, nil
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, api.ErrUserInvalid
	} else {
		return nil, err
	}
}

func (us *userService) DeleteUserById(id uint) (bool, error) {
	err := us.DB.Delete(&model.User{}, id).Error
	if err != nil {
		return false, err
	}
	return true, nil
}
