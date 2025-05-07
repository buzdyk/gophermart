package config

import (
	"os"
	"strings"
)

type Config struct {
	DatabaseURI string

	ServerAddress     string
	AccrualSystemAddr string

	JWTSecretKey string
}

func New() *Config {
	return &Config{
		DatabaseURI:       getEnv("DATABASE_URI", "postgres://postgres:secret@localhost:5432/gophermart?sslmode=disable"),
		ServerAddress:     getEnv("RUN_ADDRESS", ":9090"),
		AccrualSystemAddr: getEnv("ACCRUAL_SYSTEM_ADDRESS", "http://localhost:8080"),
		JWTSecretKey:      getEnv("JWT_SECRET_KEY", "your-secret-key"),
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists && strings.TrimSpace(value) != "" {
		return value
	}
	return fallback
}
