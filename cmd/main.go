package main

import (
	"log"
	"net/http"
	"time"

	"secure-api-gateway/internal/config"
)

func homeHandler(resp http.ResponseWriter, req *http.Request) {
	log.Printf("/: %s", req.URL.Path)
}

func healthHandler(resp http.ResponseWriter, req *http.Request) {
	log.Printf("OK")
}

func formHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		log.Printf(`
				<form method="POST">
					<input type="text" name="name" placeholder="Enter your name">
					<button type="submit">Submit</button>
				</form>
			`)
	case http.MethodPost:
		err := r.ParseForm()
		if err != nil {
			log.Printf("Error parsing form: %v", err)
			return
		}

		name := r.FormValue("name")
		log.Printf("form: %s", name)
	}
}

func main() {
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

	log.Printf("Сервер запущен на http://localhost%s", cfg.Port)

	err := server.ListenAndServe()
	if err != nil {
		log.Fatalf("500: Error fatal %v", err)
	}
}
