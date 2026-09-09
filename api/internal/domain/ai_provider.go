package domain

// AI provider selectable for a project. String() values are AI-service URL path segments —
// openai / google / anthropic — matched exactly by the FastAPI routers. Never "gemini", never "claude".

import "strings"

type AIProvider int

const (
	AIProviderUnspecified AIProvider = iota
	AIProviderOpenAI
	AIProviderGoogle
	AIProviderAnthropic
)

func (a AIProvider) String() string {
	switch a {
	case AIProviderOpenAI:
		return "openai"
	case AIProviderGoogle:
		return "google"
	case AIProviderAnthropic:
		return "anthropic"
	default:
		return "unspecified"
	}
}

func ParseAIProvider(s string) AIProvider {
	switch strings.ToLower(s) {
	case "openai":
		return AIProviderOpenAI
	case "google":
		return AIProviderGoogle
	case "anthropic":
		return AIProviderAnthropic
	default:
		return AIProviderUnspecified
	}
}
