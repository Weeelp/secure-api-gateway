package config

import (
	"os"
	"strconv"

	"secure-api-gateway/internal/logger"

	"github.com/joho/godotenv"
)

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type Config struct {
	Port       string
	BackendURL string
	RedisURL   string
	JWTS       string
	Redis      RedisConfig
}

func New() *Config {
	_ = godotenv.Load(".env")

	getEnv := func(key, defaultValue string) string {
		if value := os.Getenv(key); value != "" {
			return value
		}
		return defaultValue
	}

	dbStr := getEnv("REDIS_DB", "0")
	dbInt, err := strconv.Atoi(dbStr)
	if err != nil {
		dbInt = 0
	}

	jwtS := os.Getenv("JWT_SECRET")
	if jwtS == "" {
		logger.Log.Fatal("Fatal JWT_SECRET undefined!")
	}

	return &Config{
		Port:       ":" + getEnv("PORT", "8080"),
		BackendURL: getEnv("BACKEND_URL", "http://localhost:9090"),
		RedisURL:   getEnv("REDIS_URL", "redis://localhost:6379"),
		JWTS:       jwtS,
		Redis: RedisConfig{
			Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
			Password: getEnv("REDIS_PASS", ""),
			DB:       dbInt,
		},
	}
}
