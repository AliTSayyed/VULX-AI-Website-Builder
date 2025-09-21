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
	return nil
}
