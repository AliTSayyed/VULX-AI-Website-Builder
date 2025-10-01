package authToken

import (
	"crypto/ed25519"
	"encoding/base64"
	"strings"

	"github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/config"
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

func (t *TokenService) CreateJWT() {
}

func (t *TokenService) ValidateJWT() {
}

func (t *TokenService) ParseJWT() {
}
