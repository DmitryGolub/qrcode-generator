package config

import (
	"os"
)

type Config struct {
	AppBaseURL  string
	JWTSecret   string
	GRPCTarget  string
	HTTPPort    string
}

func New() *Config {
	return &Config{
		AppBaseURL: getEnv("APP_BASE_URL", "http://localhost:8080"),
		JWTSecret:  getEnv("JWT_SECRET", "default_secret"),
		GRPCTarget: getEnv("GRPC_TARGET", "backend:9000"),
		HTTPPort:   getEnv("HTTP_PORT", ":8080"),
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}