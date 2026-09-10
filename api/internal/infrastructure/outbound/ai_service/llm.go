package aiservice

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/domain"
	"github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/utils"
)

type codeAgentRequest struct {
	Message string `json:"message"`
}

type queryRequest struct {
	Message string `json:"message"`
}

type queryResponse struct {
	Content string `json:"content"`
}

func (a *AIService) RunCodeAgent(
	ctx context.Context, provider domain.AIProvider, sandboxID, message string,
) (*domain.CodeAgentResult, error) {
	if provider == domain.AIProviderUnspecified {
		return nil, domain.NewError(domain.ErrorTypeInvalid,
			errors.New("ai provider cannot be unspecified"))
	}
	if sandboxID == "" {
		return nil, domain.NewError(domain.ErrorTypeInvalid,
			errors.New("sandbox id cannot be empty"))
	}

	// No timeout of its own: an agent run is minutes and the CALLER owns the deadline.
	// The 15-minute client backstop in ai_service.go is the only hard limit.
	path := fmt.Sprintf("/%s/%s/code", provider.String(), url.PathEscape(sandboxID))

	var out domain.CodeAgentResult
	if err := a.do(ctx, http.MethodPost, path, codeAgentRequest{Message: message}, &out); err != nil {
		return nil, domain.WrapError("ai service run code agent", err)
	}

	utils.Logger.Info("code agent finished",
		"sandbox_id", sandboxID, "provider", provider.String(),
		"files", len(out.Files), "commands", len(out.Commands))

	return &out, nil
}

const titlePromptTemplate = "Summarize the following website-build request into a short, accurate " +
	"project title of 3 to 7 words. No trailing punctuation, no quotation marks, no preamble — " +
	"respond with only the title.\n\nRequest: %s"

// GenerateTitle always hits OpenAI via /openai/query, independent of the project's chosen
// provider — title generation is a cosmetic side task, not part of the code-gen path.
// ProjectService.Create falls back to provisionalTitle on any error, so this is never
// load-bearing for whether a project can be created.
func (a *AIService) GenerateTitle(ctx context.Context, firstPrompt string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, generateTitleTimeout)
	defer cancel()

	message := fmt.Sprintf(titlePromptTemplate, firstPrompt)

	var out queryResponse
	if err := a.do(ctx, http.MethodPost, "/openai/query", queryRequest{Message: message}, &out); err != nil {
		return "", domain.WrapError("ai service generate title", err)
	}

	return out.Content, nil
}
