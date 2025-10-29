package service

import "mime/multipart"

type FileService interface {
	Upload(file *multipart.FileHeader) (string, error)
}
