package llm

import (
	"github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/config"
	"github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/utils"
	"github.com/tmc/langchaingo/llms/openai"
)

type LLM struct {
	OpenaiClient *openai.LLM
}

func New(cfg config.LLM) *LLM {
	llm, err := openai.New(openai.WithToken(cfg.OpenaiApiKey), openai.WithModel("gpt-4.1"))
	if err != nil {
		utils.Logger.Error("Error creating OpenAI client", "error", err)
		return nil
	}
	return &LLM{
		OpenaiClient: llm,
	}
}
