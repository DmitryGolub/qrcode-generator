package config

import (
	"os"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"
)

type Config struct {
	GRPCServerAddress string
	DatabaseURL       string
	JWTSecret         string
	JWTTTL            time.Duration
	AppBaseURL        string

	IPInfoToken       string
	IPInfoHTTPTimeout time.Duration
}

func New() *Config {
	jwtSecret := getEnv("JWT_SECRET", "default_super_secret_key_change_me")
	if jwtSecret == "default_super_secret_key_change_me" {
		log.Warn().Msg("Using default JWT secret. Please set JWT_SECRET environment variable for production.")
	}

	ttlMinutesStr := getEnv("JWT_TTL", "60")
	ttlMinutes, err := strconv.Atoi(ttlMinutesStr)
	if err != nil {
		log.Warn().Msgf("Invalid JWT_TTL value, using default 60 minutes. Error: %v", err)
		ttlMinutes = 60
	}

	return &Config{
		GRPCServerAddress: getEnv("GRPC_SERVER_ADDRESS", ":8080"),
		DatabaseURL:       getEnv("DATABASE_URL", ""),
		JWTSecret:         jwtSecret,
		JWTTTL:            time.Duration(ttlMinutes) * time.Minute,
		AppBaseURL:        getEnv("APP_BASE_URL", "http://localhost:8080"),

		IPInfoToken:       getEnv("IPINFO_TOKEN", ""),
		IPInfoHTTPTimeout: time.Duration(2) * time.Second,
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	if defaultValue == "" {
		log.Warn().Msgf("Important environment variable %s is not set", key)
	}
	return defaultValue
}
