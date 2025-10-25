package services

/*
This Oauth service expects oauth flows for loging in a user,
currently not using oauth for "using" some other applicatons
would break out the OauthProvider interface into seperate
purposes if morre than login was needed.
*/

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/domain"
	"golang.org/x/oauth2"
)

var ErrUnverifiedEmail = domain.NewError(domain.ErrorTypeInvalid, errors.New("unverified email"))

type OauthProviderWithID struct {
	ProviderName   string
	ProviderUserID string
}

type OauthLoginResult struct {
	FirstName   string
	LastName    string
	Email       string
	Credentials OauthProviderWithID
}

type OauthOptions struct {
	ExtraParams map[string]string
}

// tailored for login only right now
// split out Profile() in the future if there are needs for oauth other than login
type OauthProvider interface {
	Name() string
	Config() *oauth2.Config
	AuthURL(state string, options *OauthOptions) (string, error)
	Exchange(ctx context.Context, code string, options *OauthOptions) (*oauth2.Token, error)
	Profile(ctx context.Context, accessToken string) (*OauthLoginResult, error)
}

type OauthProviderRegistry interface {
	Provider(name string) (OauthProvider, error)
}

type OauthService struct {
	providerRegistry OauthProviderRegistry
	cache            Cache
}

// takes in interfaces not pointers to the implementations
func NewOauthService(providerRegistry OauthProviderRegistry, cache Cache) *OauthService {
	return &OauthService{
		providerRegistry: providerRegistry,
		cache:            cache,
	}
}

func (o *OauthService) BeginLoginFlow(ctx context.Context, provider domain.LoginProvider, options *OauthOptions) (string, error) {
	if options == nil {
		options = &OauthOptions{ExtraParams: make(map[string]string)}
	}

	// get specific provider
	oauthprovider, err := o.providerRegistry.Provider(provider.String())
	if err != nil {
		return "", domain.WrapError("oauth service begin login flow", err)
	}

	// create a state to verify Oauth callback
	state, err := o.GenerateState(32)
	if err != nil {
		return "", domain.WrapError("oauth service begin login flow", err)
	}

	// cache state and options
	if err := o.cache.Set(ctx, fmt.Sprintf("provider:%s", state), provider.String(), 10*time.Minute); err != nil {
		return "", domain.WrapError("oauth service begin login flow", err)
	}

	if len(options.ExtraParams) > 0 {
		extraParamsJSON, err := json.Marshal(options.ExtraParams)
		if err != nil {
			return "", domain.NewError(domain.ErrorTypeInternal, fmt.Errorf("failed to marshal extra params: %w", err))
		}

		if err := o.cache.Set(ctx, fmt.Sprintf("options:%s", state), string(extraParamsJSON), 10*time.Minute); err != nil {
			return "", domain.WrapError("oauth service begin login flow", err)
		}
	}

	// get a auth url to show to user
	authURL, err := oauthprovider.AuthURL(state, options)
	if err != nil {
		return "", domain.WrapError("oauth service begin login flow", err)
	}

	return authURL, nil
}

func (o *OauthService) CompleteLoginFlow(ctx context.Context, code string, state string, options *OauthOptions) (*OauthLoginResult, error) {
	if options == nil {
		options = &OauthOptions{ExtraParams: make(map[string]string)}
	}

	// after user allows auhtorization get the code from front end url
	if code == "" {
		return nil, domain.NewError(domain.ErrorTypeInvalid, errors.New("code is required"))
	}

	if state == "" {
		return nil, domain.NewError(domain.ErrorTypeInvalid, errors.New("state is required"))
	}

	provider, err := o.cache.Get(ctx, fmt.Sprintf("provider:%s", state))
	if err != nil {
		return nil, domain.NewError(domain.ErrorTypeInvalid, fmt.Errorf("invalid or expired oauth provider:%w", err))
	}

	oauthprovider, err := o.providerRegistry.Provider(provider)
	if err != nil {
		return nil, domain.WrapError("oauth service complete login flow", err)
	}

	// exchange code for token
	token, err := oauthprovider.Exchange(ctx, code, options)
	if err != nil {
		return nil, domain.WrapError("oauth service complete login flow", err)
	}

	// use token to get user profile information
	loginResult, err := oauthprovider.Profile(ctx, token.AccessToken)
	if err != nil {
		return nil, domain.WrapError("oauth service complete login flow", err)
	}

	email := strings.ToLower(strings.TrimSpace(loginResult.Email))
	if email == "alitsayyed@gmail.com" {
		// set some special perms here for me like infinite credits
	}
	return loginResult, nil
}

func (o *OauthService) GenerateState(length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", domain.NewError(domain.ErrorTypeInternal, fmt.Errorf("failed to generate random bytes for state: %w", err))
	}
	state := base64.URLEncoding.EncodeToString(b)
	return state, nil
}
