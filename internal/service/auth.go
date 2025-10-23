package service

import "github.com/mycodeLife01/qa/internal/model"

type AuthService interface {
	IsValidUser(username, password string) (*model.User, error)
	Register(username, password, email string) (*model.User, error)
}
