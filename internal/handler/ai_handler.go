package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/mycodeLife01/qa/internal/pkg/api"
	"github.com/mycodeLife01/qa/internal/service"
)

type AiHandler struct {
	AiService service.AiService
}

func NewAiHandler(aiService service.AiService) *AiHandler {
	return &AiHandler{AiService: aiService}
}

type AskRequest struct {
	Question        string `json:"question" binding:"required"`
	FileContentHash string `json:"file_content_hash" binding:"required"`
}

type AskResponse struct {
	Answer string `json:"answer"`
}

// Ask 处理用户问答请求
func (ah *AiHandler) Ask(c *gin.Context) {
	var req AskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(err)
		return
	}

	answer, err := ah.AiService.Ask(req.Question, req.FileContentHash)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.Set(api.ResponseDataKey, AskResponse{Answer: answer})
}
