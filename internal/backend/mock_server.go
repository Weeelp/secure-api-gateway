package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func homeHandler(resp http.ResponseWriter, req *http.Request) {
	log.Printf("/: запрос на главную, path: %s", req.URL.Path)
	log.Println(req.RemoteAddr)
	resp.Write([]byte("get\n"))
}

func formHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, `
			<form method="POST">
				<input type="text" name="name" placeholder="Enter your name">
				<button type="submit">Submit</button>
			</form>
		`)
	case http.MethodPost:
		if err := r.ParseForm(); err != nil {
			log.Printf("Ошибка парсинга: %v", err)
			return
		}
		name := r.FormValue("name")
		log.Printf("Получена форма: name=%s", name)
		fmt.Fprintf(w, "Привет, %s! Бэкенд получил твои данные.", name)
	}
}

func main() {
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/form", formHandler)

	server := &http.Server{
		Addr:         ":9090",
		Handler:      nil,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Printf("Server is running on http://localhost:9090")

	err := server.ListenAndServe()
	if err != nil {
		log.Fatalf("Ошибка сервера: %v", err)
	}
}
