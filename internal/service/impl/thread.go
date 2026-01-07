package impl

import (
	"github.com/google/uuid"
	"github.com/mycodeLife01/qa/internal/model"
	"github.com/mycodeLife01/qa/internal/service"
	"gorm.io/gorm"
)

type threadService struct {
	DB *gorm.DB
}

func NewThreadService(db *gorm.DB) service.ThreadService {
	return &threadService{DB: db}
}

func (s *threadService) CreateThread(userID uint, fileID uint, title string) (*model.Thread, error) {
	thread := &model.Thread{
		UUID:   uuid.New().String(),
		UserID: userID,
		FileID: fileID,
		Title:  title,
	}
	if err := s.DB.Create(thread).Error; err != nil {
		return nil, err
	}
	return thread, nil
}

func (s *threadService) GetThreadByUUID(uuid string) (*model.Thread, error) {
	var thread model.Thread
	if err := s.DB.Where("uuid = ?", uuid).First(&thread).Error; err != nil {
		return nil, err
	}
	return &thread, nil
}

func (s *threadService) GetUserThreads(userID uint) ([]model.Thread, error) {
	var threads []model.Thread
	if err := s.DB.Where("user_id = ?", userID).Order("created_at desc").Find(&threads).Error; err != nil {
		return nil, err
	}
	return threads, nil
}
