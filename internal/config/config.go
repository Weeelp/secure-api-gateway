package config

import (
	"os"
	"strconv"

	"secure-api-gateway/internal/logger"

	"github.com/joho/godotenv"
)

type RedisConfig struct {
	Addr        string
	Password    string
	JwtDB       int
	RateLimitDB int
	BlacklistDB int
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

	getEnvStr := func(key, defaultValue string) string {
		if value := os.Getenv(key); value != "" {
			return value
		}
		return defaultValue
	}

	getEnvInt := func(key, defaultValue string) int {
		valueStr := getEnvStr(key, defaultValue)
		valueInt, err := strconv.Atoi(valueStr)
		if err != nil {
			valueInt = 0
		}
		return valueInt
	}

	jwtS := os.Getenv("JWT_SECRET")
	if jwtS == "" {
		logger.Log.Fatal("Fatal JWT_SECRET undefined!")
	}

	return &Config{
		Port:       ":" + getEnvStr("PORT", "8080"),
		BackendURL: getEnvStr("BACKEND_URL", "http://localhost:9090"),
		RedisURL:   getEnvStr("REDIS_URL", "redis://localhost:6379"),
		JWTS:       jwtS,
		Redis: RedisConfig{
			Addr:        getEnvStr("REDIS_ADDR", "localhost:6379"),
			Password:    getEnvStr("REDIS_PASS", ""),
			JwtDB:       getEnvInt("JWT_DB", "0"),
			RateLimitDB: getEnvInt("RATE_LIMIT_DB", "1"),
			BlacklistDB: getEnvInt("BLACKLIST_DB", "2"),
		},
	}
}
