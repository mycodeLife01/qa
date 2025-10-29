package impl

import (
	"fmt"
	"log"

	"github.com/mycodeLife01/qa/config"
	"github.com/mycodeLife01/qa/internal/model"
	"github.com/mycodeLife01/qa/internal/pkg/client"
	"github.com/mycodeLife01/qa/internal/service"
	"github.com/tencentyun/cos-go-sdk-v5"
	"gorm.io/gorm"
)

type aiService struct {
	DB           *gorm.DB
	PythonClient *client.PythonServiceClient
	CosClient    *cos.Client
}

func NewAiService(db *gorm.DB, pythonClient *client.PythonServiceClient, cosClient *cos.Client) service.AiService {
	return &aiService{
		DB:           db,
		PythonClient: pythonClient,
		CosClient:    cosClient,
	}
}

func (as *aiService) Ask(question string, fileContentHash string) (string, error) {
	// 1. 和index worker通信，worker查询向量数据库中有没有对应文档
	log.Printf("[AiService] 开始处理问题，文档hash: %s", fileContentHash)

	checkResp, err := as.PythonClient.CheckDocument(
		config.C.Services.QAIndexWorkerURL,
		fileContentHash,
	)
	if err != nil {
		return "", fmt.Errorf("检查文档索引状态失败: %w", err)
	}

	log.Printf("[AiService] 文档索引状态: exists=%v", checkResp.Exists)

	// 2. 如果没有，worker执行index操作，完成后返回结果
	if !checkResp.Exists {
		log.Printf("[AiService] 文档未索引，开始执行索引任务")

		// 从数据库查询文件信息
		var file model.File
		if err := as.DB.Where("content_hash = ?", fileContentHash).First(&file).Error; err != nil {
			return "", fmt.Errorf("查询文件信息失败: %w", err)
		}

		// 构造文件URL（这里假设使用COS URL格式，根据实际情况调整）
		// fileURL := fmt.Sprintf("https://%s.cos.ap-guangzhou.myqcloud.com/%s",
		// 	file.BucketName, file.ObjectKey)
		ourl := as.CosClient.Object.GetObjectURL(file.ObjectKey)

		// 推断文件类型（根据文件扩展名）
		fileType := file.FileType // 默认为pdf，实际应该根据文件扩展名判断

		// 调用index worker进行索引
		indexReq := &client.IndexRequest{
			JobID:       fmt.Sprintf("job-%s", fileContentHash),
			ContentHash: fileContentHash,
			FileURL:     ourl.String(),
			FileType:    fileType,
		}

		indexResp, err := as.PythonClient.IndexDocument(
			config.C.Services.QAIndexWorkerURL,
			indexReq,
		)
		if err != nil {
			return "", fmt.Errorf("索引文档失败: %w", err)
		}

		if indexResp.Status != "success" {
			return "", fmt.Errorf("索引文档失败: %s", indexResp.Message)
		}

		log.Printf("[AiService] 文档索引完成: %s", indexResp.Message)
	}

	// 3. 如果有，调用QA Agent, Agent查询向量数据库，完成后生成答案
	log.Printf("[AiService] 调用QA Agent生成答案")

	askResp, err := as.PythonClient.AskQuestion(
		config.C.Services.QAAgentURL,
		question,
		fileContentHash,
	)
	if err != nil {
		return "", fmt.Errorf("调用QA Agent失败: %w", err)
	}

	log.Printf("[AiService] 答案生成完成，上下文文档数: %d", len(askResp.Context))

	// 4. 返回答案
	return askResp.Answer, nil
}
