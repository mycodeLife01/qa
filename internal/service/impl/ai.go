package impl

import (
	"fmt"
	"log"

	"github.com/mycodeLife01/qa/config"
	"github.com/mycodeLife01/qa/internal/model"
	"github.com/mycodeLife01/qa/internal/pkg/api"
	"github.com/mycodeLife01/qa/internal/pkg/client"
	"github.com/mycodeLife01/qa/internal/service"
	"github.com/tencentyun/cos-go-sdk-v5"
	"gorm.io/gorm"
)

type aiService struct {
	DB           *gorm.DB
	PythonClient *client.PythonServiceClient
	CosClient    *cos.Client
	TaskClient   *client.CeleryClient
}

func NewAiService(db *gorm.DB, pythonClient *client.PythonServiceClient, cosClient *cos.Client, taskClient *client.CeleryClient) service.AiService {
	return &aiService{
		DB:           db,
		PythonClient: pythonClient,
		CosClient:    cosClient,
		TaskClient:   taskClient,
	}
}

func (as *aiService) Ask(question string, fileContentHash string, threadID string) (<-chan string, error) {
	log.Printf("[AiService] 开始处理问题，文档hash: %s", fileContentHash)

	// 1. Check file status in DB
	var file model.File
	if err := as.DB.Where("content_hash = ?", fileContentHash).First(&file).Error; err != nil {
		return nil, fmt.Errorf("查询文件失败: %w", err)
	}

	if file.Status == "PENDING" || file.Status == "PROCESSING" {
		return nil, fmt.Errorf("文件正在索引中，请稍后再试")
	}
	if file.Status == "FAILURE" {
		return nil, fmt.Errorf("文件索引失败，无法回答问题")
	}
	if file.Status != "SUCCESS" {
		return nil, fmt.Errorf("文件状态未知: %s", file.Status)
	}

	// 2. Call QA Agent
	log.Printf("[AiService] 调用QA Agent生成答案")

	askStream, err := as.PythonClient.AskQuestionStream(
		config.C.Services.QAAgentURL,
		question,
		fileContentHash,
		threadID,
	)
	if err != nil {
		return nil, fmt.Errorf("调用QA Agent失败: %w", err)
	}

	return askStream, nil
}

func (as *aiService) getFileInfoByContentHash(contentHash string) (map[string]string, error) {
	var file model.File
	if err := as.DB.Where("content_hash = ?", contentHash).First(&file).Error; err != nil {
		return nil, fmt.Errorf("查询文件信息失败: %w", err)
	}

	ourl := as.CosClient.Object.GetObjectURL(file.ObjectKey)
	fileType := file.FileType

	return map[string]string{
		"fileUrl":  ourl.String(),
		"fileType": fileType,
	}, nil
}

func (as *aiService) CreateIndexTask(fileContentHash string) (string, error) {
	fileInfo, err := as.getFileInfoByContentHash(fileContentHash)
	if err != nil {
		return "", fmt.Errorf("获取文件信息失败: %w", err)
	}

	taskID, err := as.TaskClient.CreateIndexTask(fileContentHash, fileInfo["fileUrl"], fileInfo["fileType"])
	if err != nil {
		return "", fmt.Errorf("创建Celery索引任务失败: %w", err)
	}
	return taskID, nil
}

func (as *aiService) GetIndexTaskResult(taskID string) (string, error) {
	result, err := as.TaskClient.GetResult(taskID)
	if err != nil {
		return "", fmt.Errorf("获取Celery索引任务结果失败: %w", err)
	}
	if result == nil {
		return "PENDING", nil
	}
	if result.Status == "FAILURE" {
		return "FAILURE", api.ErrHandleIndexTask
	}
	return "SUCCESS", nil
}
