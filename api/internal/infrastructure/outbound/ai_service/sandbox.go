package aiservice

import (
	"context"
	"encoding/json"

	"github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/utils"
)

type SandboxResponse struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}

func (a *AIService) CreateSandbox(ctx context.Context) (*SandboxResponse, error) {
	resp, err := a.client.Get(a.baseURL + "/sandbox/create")
	if err != nil {
		// TODO domain wrap this error
	}

	defer resp.Body.Close()

	var sandbox SandboxResponse
	if err := json.NewDecoder(resp.Body).Decode(&sandbox); err != nil {
		return nil, err
	}

	utils.Logger.Info("Sandbox created", "sandbox_id", sandbox.ID)
	utils.Logger.Info("Sandbox url", "sandbox_url", sandbox.URL)

	return &sandbox, nil
}
