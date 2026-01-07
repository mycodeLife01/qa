package impl

import (
	"github.com/mycodeLife01/qa/internal/model"
	"github.com/mycodeLife01/qa/internal/service"
	"gorm.io/gorm"
)

type messageService struct {
	DB *gorm.DB
}

func NewMessageService(db *gorm.DB) service.MessageService {
	return &messageService{DB: db}
}

func (s *messageService) CreateMessage(threadID uint, role string, content string) (*model.Message, error) {
	message := &model.Message{
		ThreadID: threadID,
		Role:     role,
		Content:  content,
	}
	if err := s.DB.Create(message).Error; err != nil {
		return nil, err
	}
	return message, nil
}

func (s *messageService) GetThreadMessages(threadID uint) ([]model.Message, error) {
	var messages []model.Message
	if err := s.DB.Where("thread_id = ?", threadID).Order("created_at asc").Find(&messages).Error; err != nil {
		return nil, err
	}
	return messages, nil
}
