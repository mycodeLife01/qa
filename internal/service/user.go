package service

import "github.com/mycodeLife01/qa/internal/model"

type UserService interface {
	FindAllUser() ([]model.User, error)
	FindUserByName(name string) ([]model.User, error)
	AddUser(user model.User) (*model.User, error)
	UpdateUser(user model.User) (*model.User, error)
	DeleteUserById(id uint) (bool, error)
}
