package main

import (
	"net/http"
	"time"

	"secure-api-gateway/internal/app"
	"secure-api-gateway/internal/config"
	"secure-api-gateway/internal/logger"
)

func main() {
	logger.Init()
	defer logger.Close()
	cfg := config.New()

	target := cfg.BackendURL
	router := app.NewRouter(target)
	server := &http.Server{
		Addr:         cfg.Port,
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	logger.Log.Info("Gateway is running", "addr", cfg.Port, "target", target)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Log.Fatal("Server stopped with error", "err", err)
	}
}
