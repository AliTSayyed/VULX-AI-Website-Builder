package oauth

/*
This file will configure the oauth flow to google specifcally.
Just need the token from google to get the users profile.
Not storing google token, will use jwt after user creation
*/

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/application/services"
	"github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/config"
	"github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/domain"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type GoogleLoginProvider struct {
	config *oauth2.Config
}

type googleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
}

func NewGoogleLoginProvider(cfg config.OauthProvider) *GoogleLoginProvider {
	return &GoogleLoginProvider{
		config: &oauth2.Config{
			ClientID:     strings.TrimSpace(cfg.ClientID),
			ClientSecret: strings.TrimSpace(cfg.ClientSecret),
			Endpoint:     google.Endpoint,
			RedirectURL:  cfg.RedirectURL,
			Scopes: []string{
				"https://www.googleapis.com/auth/userinfo.email",
				"https://www.googleapis.com/auth/userinfo.profile",
				"openid",
			},
		},
	}
}

func (g *GoogleLoginProvider) Name() string {
	return domain.LoginProviderGoogle.String()
}

func (g *GoogleLoginProvider) Config() *oauth2.Config {
	return g.config
}

func (g *GoogleLoginProvider) AuthURL(state string, options *services.OauthOptions) (string, error) {
	return g.config.AuthCodeURL(state), nil
}

func (g *GoogleLoginProvider) Exchange(ctx context.Context, code string, options *services.OauthOptions) (*oauth2.Token, error) {
	token, err := g.config.Exchange(ctx, code)
	if err != nil {
		return nil, domain.NewError(domain.ErrorTypeUnavailable, fmt.Errorf("failed to exchange google code: %w", err))
	}
	return token, nil
}

func (g *GoogleLoginProvider) Profile(ctx context.Context, accessToken string) (*services.OauthLoginResult, error) {
	// create google request for user info
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("https://www.googleapis.com/oauth2/v2/userinfo?access_token=%s", accessToken), nil)
	if err != nil {
		return nil, domain.NewError(domain.ErrorTypeUnavailable, fmt.Errorf("failed to create google oauth request: %w", err))
	}
	// send request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, domain.NewError(domain.ErrorTypeUnavailable, fmt.Errorf("failed to get google user info: %w", err))
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, domain.NewError(domain.ErrorTypeUnavailable, fmt.Errorf("failed to get google oauth user info, bad status: %d", resp.StatusCode))
	}

	// decode response
	var userInfo googleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, domain.NewError(domain.ErrorTypeUnavailable, fmt.Errorf("failed to decode google oauth use info: %w", err))
	}

	if !userInfo.VerifiedEmail {
		return nil, services.ErrUnverifiedEmail
	}

	// return to the oauth service
	return &services.OauthLoginResult{
		FirstName: userInfo.GivenName,
		LastName:  userInfo.FamilyName,
		Email:     userInfo.Email,
		Credentials: services.OauthProviderWithID{
			ProviderName:   "google",
			ProviderUserID: userInfo.ID,
		},
	}, nil
}
