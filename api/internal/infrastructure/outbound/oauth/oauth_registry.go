package oauth

/*
Generic oauth that contains the reference to all providers
Pass this "general" provider into the New OAuth Service func
This abstracts the logic from the service and it is implemented
in the specific provider.go files.
*/

import (
	"fmt"

	"github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/application/services"
	"github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/config"
	"github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/domain"
)

type OauthProviderRegistry struct {
	google *GoogleLoginProvider
}

func NewOauthRegistry(oauth config.Oauth) *OauthProviderRegistry {
	google := NewGoogleLoginProvider(oauth.Google)
	return &OauthProviderRegistry{
		google: google,
	}
}

func (o *OauthProviderRegistry) Provider(name string) (services.OauthProvider, error) {
	switch name {
	case "google":
		return o.google, nil
	default:
		return nil, domain.NewError(domain.ErrorTypeInvalid, fmt.Errorf("oauth provider not found"))
	}
}
