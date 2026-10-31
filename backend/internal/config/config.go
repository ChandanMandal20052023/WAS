package config

import (
	"os"
	"strconv"
)

type Config struct {
	AppName          string
	AppEnv           string
	Port             string
	DatabaseURL      string
	RedisURL         string
	JWTSecret        string
	EvolutionURL     string
	EvolutionKey     string
	WebhookBaseURL   string
	UseMockEvolution bool
}

func Load() *Config {
	useMock, _ := strconv.ParseBool(getEnv("USE_MOCK_EVOLUTION_ADAPTER", "false"))

	return &Config{
		AppName:          getEnv("APP_NAME", "WAS"),
		AppEnv:           getEnv("APP_ENV", "development"),
		Port:             getEnv("PORT", "8080"),
		DatabaseURL:      getEnv("DATABASE_URL", "postgres://was_user:was_password@localhost:5432/was_crm?sslmode=disable"),
		RedisURL:         getEnv("REDIS_URL", "redis://localhost:6379"),
		JWTSecret:        getEnv("JWT_SECRET", "was_jwt_super_secret_signing_key_32bytes"),
		EvolutionURL:     getEnv("EVOLUTION_API_URL", "http://localhost:8081"),
		EvolutionKey:     getEnv("EVOLUTION_GLOBAL_KEY", "was_global_secret_key_12345"),
		WebhookBaseURL:   getEnv("WEBHOOK_BASE_URL", "http://localhost:8080"),
		UseMockEvolution: useMock,
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

