package client

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
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

// AskRequest 问答请求
type AskRequest struct {
	Input          string `json:"input"`
	DocContentHash string `json:"doc_content_hash"`
	ThreadID       string `json:"thread_id"`
}

// AskResponse 问答响应
type AskResponse struct {
	Answer string `json:"answer"`
}

// AskQuestionStream 调用QA Agent回答问题（流式）
func (c *PythonServiceClient) AskQuestionStream(agentURL string, input string, contentHash string, threadID string) (<-chan string, error) {
	reqBody := AskRequest{
		Input:          input,
		DocContentHash: contentHash,
		ThreadID:       threadID,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	url := fmt.Sprintf("%s/chat", agentURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")

	// 使用 Do 发送请求
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("调用问答接口失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("问答接口返回错误状态码 %d", resp.StatusCode)
	}

	// 创建通道
	stream := make(chan string)

	// 启动协程读取流
	go func() {
		defer resp.Body.Close()
		defer close(stream)

		reader := bufio.NewReader(resp.Body)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err != io.EOF {
					// 可以在这里记录日志
					fmt.Printf("Error reading stream: %v\n", err)
				}
				break
			}

			// Note: strings.TrimSpace will remove trailing newlines which are critical separators in SSE
			// But since we are reading line by line with ReadString('\n'), the line ends with \n.
			// SSE data lines format: "data: <content>\n\n" or just "\n"
			// The python code yields: f"data: {message.text}\n\n"
			// So each meaningful line starts with "data: "

			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			if strings.HasPrefix(line, "data: ") {
				data := strings.TrimPrefix(line, "data: ")

				// Handle specific python agent error format if present
				if strings.HasPrefix(data, "[ERROR]:") {
					fmt.Printf("Agent returned error: %s\n", data)
					break
				}
				stream <- data
			}
		}
	}()

	return stream, nil
}
