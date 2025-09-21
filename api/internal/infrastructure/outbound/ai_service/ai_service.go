package aiservice

import (
	"net/http"
	"time"
)

type AIService struct {
	baseURL string
	client  *http.Client
}

func NewAIService(baseUrl string) *AIService {
	return &AIService{
		baseURL: baseUrl,
		client:  &http.Client{Timeout: 30 * time.Second},
	}
}
