package authToken

import (
	"crypto/ed25519"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/application/services"
	"github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/config"
	"github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/domain"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

/*
implementation of JWT creation, validations, and parsing
token created here is stored in cookies on browser
*/

type TokenService struct {
	private ed25519.PrivateKey
	public  ed25519.PublicKey
}

// seed will always give the same public and private key
func NewTokenService(crypto config.Crypto) *TokenService {
	seed, err := base64.StdEncoding.DecodeString(strings.TrimSpace(crypto.Seed))
	if err != nil {
		panic(err)
	}

	private := ed25519.NewKeyFromSeed(seed)
	public, ok := private.Public().(ed25519.PublicKey)
	if !ok {
		panic("failed to convert public key to ed25519.PublicKey")
	}

	return &TokenService{
		private: private,
		public:  public,
	}
}

func (t *TokenService) CreateJWT(userID uuid.UUID) (string, error) {
	expirestAt := time.Now().Add(24 * time.Hour * 7)
	j := jwt.NewWithClaims(&jwt.SigningMethodEd25519{}, jwt.RegisteredClaims{
		Issuer:    "api.vulx.ai",
		Subject:   userID.String(),
		Audience:  []string{"api.vulx.ai"},
		ExpiresAt: jwt.NewNumericDate(expirestAt),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ID:        uuid.NewString(),
	})

	token, err := j.SignedString(t.private)
	if err != nil {
		return "", domain.NewError(domain.ErrorTypeUnauthenticated, fmt.Errorf("failed to sign JWT: %w", err))
	}

	return token, nil
}

func (t *TokenService) ValidateJWT(token string) (uuid.UUID, error) {
	parsedToken, err := jwt.ParseWithClaims(token, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		return t.public, nil
	},
		jwt.WithValidMethods([]string{"EdDSA"}),
		jwt.WithIssuer("api.vulx.ai"),
		jwt.WithAudience("api.vulx.ai"),
		jwt.WithExpirationRequired(),
		jwt.WithLeeway(5*time.Minute),
		jwt.WithIssuedAt())
	if err != nil {
		return uuid.Nil, domain.NewError(domain.ErrorTypeUnauthenticated, fmt.Errorf("failed ot parse JWT: %w", err))
	}

	if !parsedToken.Valid {
		return uuid.Nil, domain.ErrUnauthenticated
	}

	sub, err := parsedToken.Claims.GetSubject()
	if err != nil {
		return uuid.Nil, domain.NewError(domain.ErrorTypeUnauthenticated, fmt.Errorf("failed to get JWT subject: %w", err))
	}

	userID, err := uuid.Parse(sub)
	if err != nil {
		return uuid.Nil, domain.NewError(domain.ErrorTypeUnauthenticated, fmt.Errorf("failed to parse user ID: %w", err))
	}

	if duration, err := parsedToken.Claims.GetExpirationTime(); err == nil && time.Until(duration.Time) < 42*time.Hour {
		return userID, services.ErrAuthTokenExpiresSoon
	}

	return userID, nil
}

func (t *TokenService) ParseJWT(token string) (uuid.UUID, time.Time, error) {
	parsedToken, err := jwt.ParseWithClaims(token, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		return t.public, nil
	},
		jwt.WithValidMethods([]string{"EdDSA"}),
		jwt.WithIssuer("api.vulx.ai"),
		jwt.WithAudience("api.vulx.ai"))
	if err != nil {
		return uuid.Nil, time.Time{}, domain.NewError(domain.ErrorTypeUnauthenticated, fmt.Errorf("failed to parse JWT: %w", err))
	}

	if !parsedToken.Valid {
		return uuid.Nil, time.Time{}, domain.ErrUnauthenticated
	}

	sub, err := parsedToken.Claims.GetSubject()
	if err != nil {
		return uuid.Nil, time.Time{}, domain.NewError(domain.ErrorTypeUnauthenticated, fmt.Errorf("failed to get JWT subject: %w", err))
	}

	userID, err := uuid.Parse(sub)
	if err != nil {
		return uuid.Nil, time.Time{}, domain.NewError(domain.ErrorTypeUnauthenticated, fmt.Errorf("faield to parse user ID: %w", err))
	}

	expiresAt, err := parsedToken.Claims.GetExpirationTime()
	if err != nil {
		return uuid.Nil, time.Time{}, domain.NewError(domain.ErrorTypeUnauthenticated, fmt.Errorf("failed to get expiry time: %w", err))
	}

	return userID, expiresAt.Time, nil
}
