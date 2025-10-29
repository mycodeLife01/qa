package service

type AiService interface {
	Ask(question string, fileContentHash string) (string, error)
}
