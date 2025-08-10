/*
* This file sets up all the connections and variables needed to run the server
* LoadConfig will create the var that holds the values needed to make the connections
 */

package config

import "os"

type Config struct {
	ServerPort string
	DB         Db
	Origins    []string
}

type Db struct {
	Host     string
	Name     string
	User     string
	Password string
}

func LoadConfig() Config {
	cfg := Config{
		ServerPort: getEnvOrDefault("SERVER_PORT", "8080"),
		DB: Db{
			Host:     getEnvOrDefault("DB_HOST", "sql"), // sql is name of service in docker compose
			Name:     getEnvOrDefault("DB_NAME", "local"),
			User:     getEnvOrDefault("DB_USER", "postgres"),
			Password: getEnvOrDefault("DB_PASSWORD", "password"),
		},
		Origins: []string{"http://localhost:5173"},
	}
	return cfg
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
