package auth

/*
This file is what the interceptor will call to get the jwt from the req, verify it, and refresh it if needed
*/
import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"connectrpc.com/connect"
	"github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/application/services"
	"github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/domain"
	"github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/utils"
	"github.com/google/uuid"
)

const (
	apiCookieName = "jwt"
	cookieMaxAge  = 7 * 24 * time.Hour
)

type HTTPAuthAdapter struct {
	authService *services.AuthService
}

type UserContextKey struct{}

func User(ctx context.Context) (*domain.User, error) {
	user, ok := ctx.Value(UserContextKey{}).(*domain.User)
	if !ok || user == nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, domain.ErrUnauthenticated)
	}
	return user, nil
}

func NewHTTPAuthAdapater(authService *services.AuthService) *HTTPAuthAdapter {
	return &HTTPAuthAdapter{
		authService: authService,
	}
}

func (h *HTTPAuthAdapter) AuthenticateWithJWT(ctx context.Context, req connect.AnyRequest) (*domain.User, bool) {
	token := h.GetJWTCookie(ctx, req)
	if token == "" {
		return nil, false
	}

	user, err := h.authService.ValidateSession(ctx, token)
	if errors.Is(err, services.ErrAuthTokenExpiresSoon) {
		return user, true
	}

	if err != nil {
		return nil, false
	}

	return user, false
}

func (h *HTTPAuthAdapter) GetJWTCookie(ctx context.Context, req connect.AnyRequest) string {
	cookies, err := http.ParseCookie(req.Header().Get("Cookie"))
	if err != nil {
		return ""
	}

	var token string
	for _, cookie := range cookies {
		if cookie.Name == apiCookieName {
			token = cookie.Value
		} else {
			token = ""
		}
	}
	return token
}

func (h *HTTPAuthAdapter) RefreshJWTCookie(ctx context.Context, res connect.AnyResponse, userID uuid.UUID) {
	token, err := h.authService.CreateSession(ctx, userID)
	if err != nil {
		utils.Logger.Error("unable to refresh jwt", "error", err)
		return
	}

	h.SetJWTCookie(ctx, res, token)
}

// TODO MUST UPDATE DOMAIN TO WHAT DOMAIN WILL BE WHEN I REGISTER ONE
func (h *HTTPAuthAdapter) SetJWTCookie(ctx context.Context, res connect.AnyResponse, token string) {
	now := time.Now().Add(cookieMaxAge).UTC().Format(time.RFC1123)

	res.Header().Add(
		"Set-Cookie", fmt.Sprintf(
			"%s=%s; Expires=%s; HttpOnly; Secure; SameSite=Lax; Domain=.vulx.ai; Path=/", apiCookieName, token, now,
		),
	)
}

func (h *HTTPAuthAdapter) ClearJWTCookie(ctx context.Context, res connect.AnyResponse) {
	zero := time.Unix(0, 0).UTC().Format(time.RFC1123)

	res.Header().Add(
		"Set-Cookie", fmt.Sprintf(
			"%s=; Expires=%s; HttpOnly; Secure; SameSite=Lax; Domain=.vulx.ai; Path=/", apiCookieName, zero,
		),
	)
}
