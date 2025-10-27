package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/mycodeLife01/qa/internal/pkg/api"
	"github.com/mycodeLife01/qa/internal/service"
)

type FileHandler struct {
	FileService service.FileService
}

func NewFileHandler(fileService service.FileService) *FileHandler {
	return &FileHandler{FileService: fileService}
}

func (fh *FileHandler) UploadFile(c *gin.Context) {
	fileHeader, err := c.FormFile("qa_file")
	if err != nil {
		_ = c.Error(err)
		return
	}
	result, err := fh.FileService.Upload(fileHeader)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.Set(api.ResponseDataKey, result)
}
