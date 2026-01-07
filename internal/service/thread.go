package service

import "github.com/mycodeLife01/qa/internal/model"

type ThreadService interface {
	CreateThread(userID uint, fileID uint, title string) (*model.Thread, error)
	GetThreadByUUID(uuid string) (*model.Thread, error)
	GetUserThreads(userID uint) ([]model.Thread, error)
}
