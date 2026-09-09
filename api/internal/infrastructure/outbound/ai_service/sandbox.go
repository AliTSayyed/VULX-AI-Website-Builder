package aiservice

import (
	"context"
	"errors"
	"net/http"

	"github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/domain"
	"github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/utils"
)

func (a *AIService) CreateSandbox(ctx context.Context) (*domain.Sandbox, error) {
	ctx, cancel := context.WithTimeout(ctx, createSandboxTimeout)
	defer cancel()

	var sandbox domain.Sandbox
	if err := a.do(ctx, http.MethodPost, "/sandbox/", nil, &sandbox); err != nil {
		return nil, domain.WrapError("ai service create sandbox", err)
	}
	if sandbox.ID == "" || sandbox.URL == "" {
		return nil, domain.NewError(domain.ErrorTypeInternal,
			errors.New("ai service returned an empty sandbox id or url"))
	}

	utils.Logger.Info("sandbox created", "sandbox_id", sandbox.ID, "url", sandbox.URL)
	return &sandbox, nil
}
