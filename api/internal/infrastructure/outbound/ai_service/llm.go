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
