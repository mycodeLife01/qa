package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/mycodeLife01/qa/internal/pkg/api"
	"github.com/mycodeLife01/qa/internal/service"
)

type InternalHandler struct {
	FileService service.FileService
}

func NewInternalHandler(fileService service.FileService) *InternalHandler {
	return &InternalHandler{FileService: fileService}
}

type UpdateStatusRequest struct {
	ContentHash string `json:"content_hash" binding:"required"`
	Status      string `json:"status" binding:"required"`
	Message     string `json:"message"`
}

func (h *InternalHandler) UpdateFileStatus(c *gin.Context) {
	var req UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(err)
		return
	}

	// 简单的状态验证
	if req.Status != "SUCCESS" && req.Status != "FAILURE" {
		_ = c.Error(api.ErrInvalidParams)
		return
	}

	err := h.FileService.UpdateStatus(req.ContentHash, req.Status)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(200, gin.H{"status": "ok"})
}
