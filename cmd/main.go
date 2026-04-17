package main

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"secure-api-gateway/internal/config"
	"secure-api-gateway/internal/logger"
)

var proxy *httputil.ReverseProxy

func homeHandler(resp http.ResponseWriter, req *http.Request) {
	logger.Log.Info("/: запрос на гланвую", "path", req.URL.Path)
	proxy.ServeHTTP(resp, req)
}

func healthHandler(resp http.ResponseWriter, req *http.Request) {
	logger.Log.Info("OK")
}

func formHandler(resp http.ResponseWriter, req *http.Request) {
	proxy.ServeHTTP(resp, req)

	switch req.Method {
	case http.MethodGet:
		logger.Log.Info(`
				<form method="POST">
					<input type="text" name="name" placeholder="Enter your name">
					<button type="submit">Submit</button>
				</form>
			`)
	case http.MethodPost:
		err := req.ParseForm()
		if err != nil {
			logger.Log.Warn("Error parsing form", "error", err)
			return
		}

		name := req.FormValue("name")
		logger.Log.Info("form", "name", name)
	}
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

	http.HandleFunc("/", homeHandler)

	http.HandleFunc("/health", healthHandler)

	http.HandleFunc("/form", formHandler)

	server := &http.Server{
		Addr:         cfg.Port,
		Handler:      nil,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	logger.Log.Info("Server is running http://localhost", "port", cfg.Port)

	err = server.ListenAndServe()
	if err != nil {
		logger.Log.Fatal("500: Error fatal", "fatal err", err)
	}
}
