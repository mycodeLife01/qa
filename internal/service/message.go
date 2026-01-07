package service

import "github.com/mycodeLife01/qa/internal/model"

type MessageService interface {
	CreateMessage(threadID uint, role string, content string) (*model.Message, error)
	GetThreadMessages(threadID uint) ([]model.Message, error)
}
