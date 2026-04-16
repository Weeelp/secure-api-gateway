package main

import (
	"net/http"
	"time"

	"secure-api-gateway/internal/config"
	"secure-api-gateway/internal/logger"
)

func homeHandler(resp http.ResponseWriter, req *http.Request) {
	logger.Log.Info("/: запрос на гланвую", "path", req.URL.Path)
}

func healthHandler(resp http.ResponseWriter, req *http.Request) {
	logger.Log.Info("OK")
}

func formHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		logger.Log.Info(`
				<form method="POST">
					<input type="text" name="name" placeholder="Enter your name">
					<button type="submit">Submit</button>
				</form>
			`)
	case http.MethodPost:
		err := r.ParseForm()
		if err != nil {
			logger.Log.Warn("Error parsing form", "error", err)
			return
		}

		name := r.FormValue("name")
		logger.Log.Info("form", "name", name)
	}
}

func main() {
	logger.Init()
	defer logger.Close()

	cfg := config.New()

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

	err := server.ListenAndServe()
	if err != nil {
		logger.Log.Fatal("500: Error fatal", "fatal err", err)
	}
}
