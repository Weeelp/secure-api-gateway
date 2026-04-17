package main

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"secure-api-gateway/internal/config"
	"secure-api-gateway/internal/logger"
	"secure-api-gateway/internal/middleware"
)

var proxy *httputil.ReverseProxy

func homeHandler(resp http.ResponseWriter, req *http.Request) {
	log := logger.FromContext(req.Context())
	log.Info("/: запрос на гланвую", "path", req.URL.Path)
	proxy.ServeHTTP(resp, req)
}

func healthHandler(resp http.ResponseWriter, req *http.Request) {
	log := logger.FromContext(req.Context())
	log.Info("OK")
}

func formHandler(resp http.ResponseWriter, req *http.Request) {
	log := logger.FromContext(req.Context())
	log.Info("Form request")
	proxy.ServeHTTP(resp, req)
}

func favicoHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func main() {
	logger.Init()
	defer logger.Close()
	cfg := config.New()

	backServer, err := url.Parse("http://localhost:9090")
	if err != nil {
		logger.Log.Fatal("Error server connection", "err", err)
	}

	proxy = httputil.NewSingleHostReverseProxy(backServer)

	mux := http.NewServeMux()

	mux.HandleFunc("/", homeHandler)
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/form", formHandler)
	mux.HandleFunc("/favicon.ico", favicoHandler)

	server := &http.Server{
		Addr:         cfg.Port,
		Handler:      middleware.StructuredLogger(mux),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	logger.Log.Info("Server is running http://localhost", "port", cfg.Port)

	err = server.ListenAndServe()
	if err != nil {
		logger.Log.Fatal("500: Error fatal", "fatal err", err)
	}
}
