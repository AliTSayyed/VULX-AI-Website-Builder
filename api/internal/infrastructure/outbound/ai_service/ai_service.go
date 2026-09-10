package aiservice

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/domain"
	"github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/utils"
)

const (
	createSandboxTimeout = 90 * time.Second
	generateTitleTimeout = 10 * time.Second
	// RunCodeAgent deliberately has none — it inherits the caller's context.
)

type AIService struct {
	baseURL string
	client  *http.Client
}

func NewAIService(baseUrl string) *AIService {
	return &AIService{
		baseURL: baseUrl,
		// Backstop only. Real deadlines are per-call, set with context.WithTimeout by
		// each method, so a caller (or a Temporal activity) governs how long its own
		// call may run. A hard client Timeout would override those and cap every call,
		// which is the bug this replaces: 30s is shorter than every code-agent run.
		client: &http.Client{Timeout: 15 * time.Minute},
	}
}

func (a *AIService) do(ctx context.Context, method, path string, in any, out any) error {
	var body io.Reader
	if in != nil {
		b, err := json.Marshal(in)
		if err != nil {
			return domain.NewError(domain.ErrorTypeInternal,
				fmt.Errorf("ai service marshal %s: %w", path, err))
		}
		body = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, a.baseURL+path, body)
	if err != nil {
		return domain.NewError(domain.ErrorTypeInternal,
			fmt.Errorf("ai service new request %s: %w", path, err))
	}
	if in != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := a.client.Do(req)
	if err != nil {
		var netErr net.Error
		if errors.Is(err, context.DeadlineExceeded) || (errors.As(err, &netErr) && netErr.Timeout()) {
			return domain.NewError(domain.ErrorTypeTimeout,
				fmt.Errorf("ai service %s timed out: %w", path, err))
		}
		return domain.NewError(domain.ErrorTypeUnavailable,
			fmt.Errorf("ai service %s unreachable: %w", path, err))
	}
	defer resp.Body.Close() // AFTER the error check — see below

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var e struct {
			Detail string `json:"detail"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&e)
		utils.Logger.Error("ai service returned an error",
			"path", path, "status", resp.StatusCode, "detail", e.Detail)
		if resp.StatusCode == http.StatusUnprocessableEntity {
			// 422 means WE sent something malformed. A bug in Go, not an outage.
			return domain.NewError(domain.ErrorTypeInternal,
				fmt.Errorf("ai service rejected our request to %s (422)", path))
		}
		return domain.NewError(domain.ErrorTypeUnavailable,
			fmt.Errorf("ai service %s failed with status %d", path, resp.StatusCode))
	}

	if out != nil {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			return domain.NewError(domain.ErrorTypeInternal,
				fmt.Errorf("ai service decode %s: %w", path, err))
		}
	}
	return nil
}
