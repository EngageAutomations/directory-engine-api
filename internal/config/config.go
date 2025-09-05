package config

import (
	"os"
	"strconv"
)

type Config struct {
	DatabaseURL     string
	RedisURL        string
	NangoSecretKey  string
	NangoPublicKey  string
	NangoServerURL  string
	JWTSecret       string
	RateLimitRPS    int
	CacheExpiration int // in minutes
	LogLevel        string
}

func Load() *Config {
	rateLimitRPS, _ := strconv.Atoi(getEnv("RATE_LIMIT_RPS", "100"))
	cacheExpiration, _ := strconv.Atoi(getEnv("CACHE_EXPIRATION", "60"))

	return &Config{
		DatabaseURL:     getEnv("DATABASE_URL", "postgres://user:password@localhost/marketplace?sslmode=disable"),
		RedisURL:        getEnv("REDIS_URL", "redis://localhost:6379"),
		NangoSecretKey:  getEnv("NANGO_SECRET_KEY", ""),
		NangoPublicKey:  getEnv("NANGO_PUBLIC_KEY", ""),
		NangoServerURL:  getEnv("NANGO_SERVER_URL", "https://api.nango.dev"),
		JWTSecret:       getEnv("JWT_SECRET", "your-secret-key"),
		RateLimitRPS:    rateLimitRPS,
		CacheExpiration: cacheExpiration,
		LogLevel:        getEnv("LOG_LEVEL", "info"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}