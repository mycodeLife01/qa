package handler

import (
	"io"

	"github.com/gin-gonic/gin"
	"github.com/mycodeLife01/qa/internal/model"
	"github.com/mycodeLife01/qa/internal/pkg/api"
	"github.com/mycodeLife01/qa/internal/service"
)

type AiHandler struct {
	AiService      service.AiService
	ThreadService  service.ThreadService
	MessageService service.MessageService
	FileService    service.FileService // Need to look up file ID
}

func NewAiHandler(aiService service.AiService, threadService service.ThreadService, messageService service.MessageService, fileService service.FileService) *AiHandler {
	return &AiHandler{
		AiService:      aiService,
		ThreadService:  threadService,
		MessageService: messageService,
		FileService:    fileService,
	}
}

type AskRequest struct {
	Question        string `json:"question" binding:"required"`
	FileContentHash string `json:"file_content_hash"` // Optional if thread_id is provided
	ThreadID        string `json:"thread_id"`         // Optional if file_content_hash is provided
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

	userID := c.GetUint("userID")
	var threadUUID string
	var fileHash string
	var threadID uint
	var threadObj *model.Thread

	// Determine Thread and File
	if req.ThreadID != "" {
		var err error
		threadObj, err = ah.ThreadService.GetThreadByUUID(req.ThreadID)
		if err != nil {
			_ = c.Error(api.ErrInvalidParams)
			return
		}
		if threadObj.UserID != userID {
			_ = c.Error(api.ErrUserInvalid)
			return
		}
		if req.FileContentHash == "" {
			_ = c.Error(api.ErrInvalidParams)
			return
		}
	} else {
		if req.FileContentHash == "" {
			_ = c.Error(api.ErrInvalidParams)
			return
		}
		file, err := ah.FileService.GetFileByHash(req.FileContentHash)
		if err != nil {
			_ = c.Error(api.ErrFileNotFound)
			return
		}
		title := req.Question
		runes := []rune(title)
		if len(runes) > 20 {
			title = string(runes[:20]) + "..."
		}
		threadObj, err = ah.ThreadService.CreateThread(userID, file.ID, title)
		if err != nil {
			_ = c.Error(err)
			return
		}
	}

	threadUUID = threadObj.UUID
	threadID = threadObj.ID
	fileHash = req.FileContentHash

	// Save User Message
	_, err := ah.MessageService.CreateMessage(threadID, "user", req.Question)
	if err != nil {
		_ = c.Error(err)
		return
	}

	// Call AI Service
	stream, err := ah.AiService.Ask(req.Question, fileHash, threadUUID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")
	// Send thread_id in header for frontend to know
	c.Writer.Header().Set("X-Thread-ID", threadUUID)

	clientGone := c.Request.Context().Done()

	var fullAnswer string // Accumulate answer outside the loop

	c.Stream(func(w io.Writer) bool {
		select {
		case <-clientGone:
			return false
		case token, ok := <-stream:
			if !ok {
				// Stream finished, save assistant message
				if fullAnswer != "" {
					ah.MessageService.CreateMessage(threadID, "assistant", fullAnswer)
				}
				return false
			}
			c.SSEvent("message", token)
			fullAnswer += token
			return true
		}
	})
}

func (ah *AiHandler) AddFileIndexTask(c *gin.Context) {
	hash := c.Query("content_hash")
	taskID, err := ah.AiService.CreateIndexTask(hash)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.Set(api.ResponseDataKey, taskID)
}

func (ah *AiHandler) GetFileIndexTaskResult(c *gin.Context) {
	taskID := c.Query("task_id")
	result, err := ah.AiService.GetIndexTaskResult(taskID)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.Set(api.ResponseDataKey, result)
}
