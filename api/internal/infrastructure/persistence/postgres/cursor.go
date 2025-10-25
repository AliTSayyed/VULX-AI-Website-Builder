package postgres

import (
	"encoding/base64"
	"errors"
	"time"

	"github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/domain"
)

var ErrTokenInvalid = domain.NewError(domain.ErrorTypeInvalid, errors.New("invalid token"))

func encodeToken(createdAt time.Time) string {
	return base64.URLEncoding.EncodeToString([]byte(createdAt.Format(time.RFC3339Nano)))
}

func decodeToken(token string) (time.Time, error) {
	if token == "" {
		return time.Time{}, nil
	}

	data, err := base64.URLEncoding.DecodeString(token)
	if err != nil {
		return time.Time{}, ErrTokenInvalid
	}

	createdAt, err := time.Parse(time.RFC3339Nano, string(data))
	if err != nil {
		return time.Time{}, ErrTokenInvalid
	}

	return createdAt, nil
}
