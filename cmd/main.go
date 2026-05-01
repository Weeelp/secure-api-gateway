package main

import (
	"net/http"
	"strings"
	"time"

	"secure-api-gateway/internal/app"
	"secure-api-gateway/internal/cache"
	"secure-api-gateway/internal/config"
	"secure-api-gateway/internal/logger"
)

func main() {
	cfg := config.New()

	logger.Init()
	defer logger.Close()

	rds := cache.NewRedis(cfg)
	defer rds.CloseRedis()

	targets := strings.Split(cfg.BackendURL, ",")
	router := app.NewRouter(targets, cfg, rds)
	server := &http.Server{
		Addr:         cfg.Port,
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	logger.Log.Info("Gateway is running", "addr", cfg.Port, "target", targets)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Log.Fatal("Proxy stopped with error", "err", err)
	}
}
