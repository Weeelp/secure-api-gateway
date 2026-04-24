package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port       string
	BackendURL string
}

func New() *Config {
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("WARN: .env файл не найден, используем системные переменные")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	BackendURL := os.Getenv("BACKEND_URL")
	if BackendURL == "" {
		BackendURL = "http://localhost:9090"
	}

	return &Config{
		Port:       ":" + port,
		BackendURL: BackendURL,
	}
}
