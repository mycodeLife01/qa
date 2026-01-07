package service

import (
	"mime/multipart"

	"github.com/mycodeLife01/qa/internal/model"
)

type FileService interface {
	Upload(username string, file *multipart.FileHeader, force bool) (map[string]string, error)
	UpdateStatus(contentHash string, status string) error
	DeleteFileByHash(username string, hash string) (bool, error)
	GetAllUserFiles(username string) ([]map[string]string, error)
	GetFileByHash(hash string) (*model.File, error)
}
