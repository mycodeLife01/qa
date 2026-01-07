package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/mycodeLife01/qa/internal/dto"
	"github.com/mycodeLife01/qa/internal/model"
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

	forceStr := c.DefaultQuery("force", "false")
	force := forceStr == "true"

	// 从 JWT 中获取当前登录用户
	user, exists := c.Get("username")
	if !exists {
		_ = c.Error(api.ErrUnauthorized)
		return
	}
	result, err := fh.FileService.Upload(user.(*model.User).Username, fileHeader, force)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.Set(api.ResponseDataKey, result)
}

func (fh *FileHandler) DeleteFileByHash(c *gin.Context) {
	var req dto.DeleteFileByHashRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		_ = c.Error(err)
		return
	}
	// 从 JWT 中获取当前登录用户
	user, exists := c.Get("username")
	if !exists {
		_ = c.Error(api.ErrUnauthorized)
		return
	}
	result, err := fh.FileService.DeleteFileByHash(user.(*model.User).Username, req.ContentHash)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.Set(api.ResponseDataKey, result)
}

func (fh *FileHandler) GetAllUserFiles(c *gin.Context) {
	// 从 JWT 中获取当前登录用户
	user, exists := c.Get("username")
	if !exists {
		_ = c.Error(api.ErrUnauthorized)
		return
	}
	username := user.(*model.User).Username
	result, err := fh.FileService.GetAllUserFiles(username)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.Set(api.ResponseDataKey, result)
}
