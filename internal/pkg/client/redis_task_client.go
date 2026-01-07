package client

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type CeleryClient struct {
	businessClient *redis.Client
	brokerClient   *redis.Client
	resultClient   *redis.Client
	queueName      string
}

// --- 补充的数据结构 ---

// CeleryTaskResult 对应 Celery 在 Redis 中存储的外层 JSON 结构
type CeleryTaskResult struct {
	Status    string          `json:"status"`    // Celery 状态: SUCCESS, FAILURE, PENDING, RETRY 等
	Result    json.RawMessage `json:"result"`    // 具体的返回值或错误信息（根据 Status 不同而不同）
	Traceback interface{}     `json:"traceback"` // 失败时的堆栈信息
	DateDone  string          `json:"date_done"` // 完成时间
	TaskID    string          `json:"task_id"`   // 任务ID
	Children  []interface{}   `json:"children"`  // 子任务
}

// IndexTaskResponse 对应业务逻辑成功时返回的内层数据
// 即 Python 代码中 return {...} 的部分
type IndexTaskResponse struct {
	Status      string `json:"status"`       // 业务状态 "success"
	ContentHash string `json:"content_hash"` // 文档哈希
	// ChunkCount  int    `json:"chunk_count"`  // 切片数量
	Message string `json:"message"` // 提示信息
}

// --- 补充的方法 ---

// GetResult 查询任务状态和结果
// 如果任务不存在或尚未完成（Pending），返回 (nil, nil)
// 如果任务已完成（无论成功失败），返回 (*CeleryTaskResult, nil)
func (c *CeleryClient) GetResult(taskID string) (*CeleryTaskResult, error) {
	ctx := context.Background()

	// 1. 构造 Celery 默认的结果 Key 格式
	// Celery 默认使用 "celery-task-meta-" 前缀
	key := fmt.Sprintf("celery-task-meta-%s", taskID)

	// 2. 从 resultClient (DB 2) 获取数据
	val, err := c.resultClient.Get(ctx, key).Result()
	if err == redis.Nil {
		// Key 不存在，说明任务还在队列中排队，或者正在处理但未更新状态
		// 此时视为 Pending
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("查询 Redis 失败: %w", err)
	}

	// 3. 反序列化 JSON
	var taskResult CeleryTaskResult
	if err := json.Unmarshal([]byte(val), &taskResult); err != nil {
		return nil, fmt.Errorf("解析 Celery 结果失败: %w", err)
	}

	return &taskResult, nil
}

// NewCeleryClient 初始化
func NewCeleryClient(businessRedisURL string, brokerRedisURL string, resultRedisURL string, queueName string) (*CeleryClient, error) {
	opts1, err := redis.ParseURL(businessRedisURL)
	if err != nil {
		return nil, err
	}
	opts2, err := redis.ParseURL(brokerRedisURL)
	if err != nil {
		return nil, err
	}
	opts3, err := redis.ParseURL(resultRedisURL)
	if err != nil {
		return nil, err
	}
	return &CeleryClient{
		businessClient: redis.NewClient(opts1),
		brokerClient:   redis.NewClient(opts2),
		resultClient:   redis.NewClient(opts3),
		queueName:      queueName,
	}, nil
}

// CreateIndexTask 发布索引任务（含去重锁 + Celery Protocol v2）
func (c *CeleryClient) CreateIndexTask(contentHash, fileURL, fileType string) (string, error) {
	ctx := context.Background()

	// --- 1. 业务去重锁 (Deduplication) ---
	taskID := uuid.New().String()
	lockKey := fmt.Sprintf("index_lock:%s", contentHash)

	// 尝试获取锁，设置 1 小时过期 (防止 Worker 崩溃导致死锁)
	// value 设置为 taskID，以便后续查询
	locked, err := c.businessClient.SetNX(ctx, lockKey, taskID, time.Hour).Result()
	if err != nil {
		return "", fmt.Errorf("Redis 错误: %w", err)
	}
	if !locked {
		// 如果已锁定，获取当前的 taskID
		existingTaskID, err := c.businessClient.Get(ctx, lockKey).Result()
		if err != nil {
			return "", fmt.Errorf("该文档正在索引中，但获取 TaskID 失败: %w", err)
		}
		// 返回已存在的 TaskID，视为成功（幂等）
		return existingTaskID, nil
	}

	// --- 2. 构造 Celery Protocol v2 Payload ---


	// 2.1 构造参数列表 [args, kwargs, callbacks]
	// Python 函数签名: def index_document_task(self, content_hash, file_url, file_type)
	args := []interface{}{contentHash, fileURL, fileType}
	kwargs := map[string]interface{}{}

	// Celery 消息体结构
	bodyData := []interface{}{args, kwargs, nil}
	bodyBytes, err := json.Marshal(bodyData)
	if err != nil {
		c.ReleaseLock(contentHash) // 失败回滚锁
		return "", fmt.Errorf("序列化 Body 失败: %w", err)
	}

	// 2.2 构造完整消息 (Protocol v2)
	payload := map[string]interface{}{
		"body":             base64.StdEncoding.EncodeToString(bodyBytes), // Body 必须 Base64
		"content-encoding": "utf-8",
		"content-type":     "application/json",
		"headers": map[string]interface{}{
			"lang":    "go",
			"task":    "index_document", // 必须与 Python @task(name=...) 一致
			"id":      taskID,
			"root_id": taskID,
		},
		"properties": map[string]interface{}{
			"correlation_id": taskID,
			"reply_to":       uuid.New().String(),
			"delivery_mode":  2, // 2 = 持久化消息
			"delivery_info": map[string]interface{}{
				"exchange":    "",
				"routing_key": c.queueName,
			},
			"body_encoding": "base64",
			"delivery_tag":  uuid.New().String(),
		},
	}

	taskJSON, err := json.Marshal(payload)
	if err != nil {
		c.ReleaseLock(contentHash)
		return "", fmt.Errorf("序列化 Payload 失败: %w", err)
	}

	// --- 3. 推送到 Redis ---
	err = c.brokerClient.LPush(ctx, c.queueName, taskJSON).Err()
	if err != nil {
		c.ReleaseLock(contentHash) // 发送失败，必须释放锁
		return "", fmt.Errorf("发送任务到 Redis 失败: %w", err)
	}

	return taskID, nil
}

// ReleaseLock 释放去重锁 (通常由 Go 在异常时调用，或用于管理)
func (c *CeleryClient) ReleaseLock(contentHash string) {
	lockKey := fmt.Sprintf("index_lock:%s", contentHash)
	c.businessClient.Del(context.Background(), lockKey)
}
