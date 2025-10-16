/*
* This file sets up all the connections and variables needed to run the server
* LoadConfig will create the var that holds the values needed to make the connections
 */

package config

import (
	"errors"
	"os"
)

type Config struct {
	DB           Db
	AppUrl       string
	ApiUrl       string
	Temporal     Temporal
	AIServiceUrl string
	Oauth        Oauth
	Redis        Redis
	Crypto       Crypto
}

type Db struct {
	Host     string
	Name     string
	User     string
	Password string
}

type Temporal struct {
	HostPort string
}

type Oauth struct {
	Google OauthProvider
}

type OauthProvider struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

type Redis struct {
	Host string
	Name string
}

// generate seed with openssl rand -base64 32
type Crypto struct {
	Seed string
}

func LoadConfig() (*Config, error) {
	cfg := &Config{
		DB: Db{
			Host:     getEnvOrDefault("DB_HOST", "sql"),
			Name:     getEnvOrDefault("DB_NAME", "local"),
			User:     getEnvOrDefault("DB_USER", "postgres"),
			Password: getEnvOrDefault("DB_PASSWORD", "password"),
		},
		ApiUrl: getEnvOrDefault("API_URL", "http://localhost:8080"),
		AppUrl: getEnvOrDefault("APP_URL", "http://localhost:3000"),
		Temporal: Temporal{
			HostPort: getEnvOrDefault("TEMPORAL_ADDRESS", "temporal:7233"),
		},
		AIServiceUrl: getEnvOrDefault("AI_SERVICE_URL", "http://ai-service:9999/ai-service/v1"),
		Oauth: Oauth{
			Google: OauthProvider{
				getEnvOrDefault("GOOGLE_CLIENT_ID", ""),
				getEnvOrDefault("GOOGLE_CLIENT_SECRET", ""),
				getEnvOrDefault("REDIRECT_URL", "localhost:8080/auth/callback"),
			},
		},
		Redis: Redis{
			Host: getEnvOrDefault("REDIS_HOST", ""),
			Name: getEnvOrDefault("REDIS_NAME", "0"),
		},
		Crypto: Crypto{
			Seed: getEnvOrDefault("JWT_SEED", ""),
		},
	}
	if err := cfg.validate(); err != nil {
		return &Config{}, err
	}
	return cfg, nil
}

func getEnvOrDefault(key, defaultValue string) string {
	var configValue string

	if value, exists := os.LookupEnv(key); exists {
		configValue = value
	} else {
		configValue = defaultValue
	}
	return configValue
}

func (c *Config) validate() error {
	if c.DB.Host == "" {
		return errors.New("DB_HOST cannot be empty")
	}
	if c.DB.Name == "" {
		return errors.New("DB_NAME cannot be empty")
	}
	if c.DB.User == "" {
		return errors.New("DB_USER cannot be empty")
	}
	if c.DB.Password == "" {
		return errors.New("DB_PASSWORD cannot be empty")
	}
	if c.ApiUrl == "" {
		return errors.New("API_URL cannot be empty")
	}
	if c.AppUrl == "" {
		return errors.New("APP_URL cannot be empty")
	}
	if c.Temporal.HostPort == "" {
		return errors.New("TEMPORAL_ADDRESS cannot be empty")
	}
	if c.AIServiceUrl == "" {
		return errors.New("AI_SERVICE_URL cannot be empty")
	}
	if c.Oauth.Google.ClientID == "" {
		return errors.New("GOOGLE_CLIENT_ID cannot be empty")
	}
	if c.Oauth.Google.ClientSecret == "" {
		return errors.New("GOOGLE_SECRET cannot be empty")
	}
	if c.Oauth.Google.RedirectURL == "" {
		return errors.New("REDIRECT_URL cannot be empty")
	}
	if c.Redis.Host == "" {
		return errors.New("REDIS_HOST cannot be empty")
	}
	if c.Redis.Name == "" {
		return errors.New("REDIS_NAME cannot be empty")
	}
	if c.Crypto.Seed == "" {
		return errors.New("JWT_SEED cannot be empty")
	}
	return nil
}
