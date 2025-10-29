package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// PythonServiceClient 用于调用Python服务的HTTP客户端
type PythonServiceClient struct {
	httpClient *http.Client
}

// NewPythonServiceClient 创建新的Python服务客户端
func NewPythonServiceClient() *PythonServiceClient {
	return &PythonServiceClient{
		httpClient: &http.Client{
			Timeout: 120 * time.Second, // 设置较长超时，因为LLM调用可能需要时间
		},
	}
}

// CheckRequest 检查文档是否已索引的请求
type CheckRequest struct {
	ContentHash string `json:"content_hash"`
}

// CheckResponse 检查文档的响应
type CheckResponse struct {
	ContentHash string  `json:"content_hash"`
	Exists      bool    `json:"exists"`
	ChunkCount  *int    `json:"chunk_count,omitempty"`
	IndexedAt   *string `json:"indexed_at,omitempty"`
}

// IndexRequest 索引文档的请求
type IndexRequest struct {
	JobID       string `json:"job_id"`
	ContentHash string `json:"content_hash"`
	FileURL     string `json:"file_url"`
	FileType    string `json:"file_type"`
}

// IndexResponse 索引文档的响应
type IndexResponse struct {
	JobID      string `json:"job_id"`
	Status     string `json:"status"` // "success" or "failed"
	Message    string `json:"message"`
	ChunkCount *int   `json:"chunk_count,omitempty"`
}

// AskRequest 问答请求
type AskRequest struct {
	Question    string `json:"question"`
	ContentHash string `json:"content_hash"`
}

// ContextItem 上下文文档项
type ContextItem struct {
	Content  string                 `json:"content"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// AskResponse 问答响应
type AskResponse struct {
	Question  string        `json:"question"`
	Answer    string        `json:"answer"`
	Context   []ContextItem `json:"context"`
	Timestamp string        `json:"timestamp"`
}

// CheckDocument 检查文档是否已索引
func (c *PythonServiceClient) CheckDocument(workerURL string, contentHash string) (*CheckResponse, error) {
	reqBody := CheckRequest{
		ContentHash: contentHash,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	url := fmt.Sprintf("%s/check", workerURL)
	resp, err := c.httpClient.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("调用检查接口失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("检查接口返回错误状态码 %d: %s", resp.StatusCode, string(body))
	}

	var checkResp CheckResponse
	if err := json.Unmarshal(body, &checkResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &checkResp, nil
}

// IndexDocument 索引文档
func (c *PythonServiceClient) IndexDocument(workerURL string, req *IndexRequest) (*IndexResponse, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	url := fmt.Sprintf("%s/index", workerURL)
	resp, err := c.httpClient.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("调用索引接口失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("索引接口返回错误状态码 %d: %s", resp.StatusCode, string(body))
	}

	var indexResp IndexResponse
	if err := json.Unmarshal(body, &indexResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &indexResp, nil
}

// AskQuestion 调用QA Agent回答问题
func (c *PythonServiceClient) AskQuestion(agentURL string, question string, contentHash string) (*AskResponse, error) {
	reqBody := AskRequest{
		Question:    question,
		ContentHash: contentHash,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	url := fmt.Sprintf("%s/ask", agentURL)
	resp, err := c.httpClient.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("调用问答接口失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("问答接口返回错误状态码 %d: %s", resp.StatusCode, string(body))
	}

	var askResp AskResponse
	if err := json.Unmarshal(body, &askResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &askResp, nil
}
