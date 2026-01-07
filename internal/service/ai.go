package service

type AiService interface {
	// HTTP
	Ask(question string, fileContentHash string, threadID string) (<-chan string, error)

	// MQ
	CreateIndexTask(fileContentHash string) (string, error)
	GetIndexTaskResult(taskID string) (string, error)
}
