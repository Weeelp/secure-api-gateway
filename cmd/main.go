// cmd/main.go
package main

import (
	"fmt"
	"log/slog"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"os"
	"strconv"
	"time"

	"secure-api-gateway/internal/app"
	"secure-api-gateway/internal/cache"
	"secure-api-gateway/internal/config"
	"secure-api-gateway/internal/logger"
)

var proxy *httputil.ReverseProxy
var challengeAnswer string

// --- Обработчики страниц (Твои функции) ---

func homeHandler(resp http.ResponseWriter, req *http.Request) {
	logger.Log.Info("Запрос на главную", "path", req.URL.Path)
	proxy.ServeHTTP(resp, req)
}

func healthHandler(resp http.ResponseWriter, req *http.Request) {
	logger.Log.Info("Проверка здоровья: OK")
	resp.WriteHeader(http.StatusOK)
	fmt.Fprint(resp, "OK")
}

func formHandler(resp http.ResponseWriter, req *http.Request) {
	proxy.ServeHTTP(resp, req)

	switch req.Method {
	case http.MethodGet:
		logger.Log.Info("GET запрос формы")
		resp.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(resp, `
			<form method="POST">
				<input type="text" name="name" placeholder="Enter your name">
				<button type="submit">Submit</button>
			</form>
		`)
	case http.MethodPost:
		err := req.ParseForm()
		if err != nil {
			logger.Log.Warn("Ошибка парсинга формы", "error", err)
			return
		}
		name := req.FormValue("name")
		logger.Log.Info("Форма отправлена", "name", name)
		resp.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(resp, "<h1>Привет, %s! Данные получены.</h1>", name)
	}
}

func generateChallenge(resp http.ResponseWriter, req *http.Request) {
	num1 := rand.Intn(10) + 1
	num2 := rand.Intn(10) + 1
	challengeAnswer = strconv.Itoa(num1 + num2)

	resp.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(resp, `
		<!DOCTYPE html>
		<html>
		<head><title>Security Challenge</title></head>
		<body>
			<h2>Безопасность: Решите задачу</h2>
			<p>Сколько будет: <strong>%d + %d</strong>?</p>
			<form method="POST">
				<input type="number" name="answer" required placeholder="Ответ">
				<button type="submit">Проверить</button>
			</form>
		</body>
		</html>
	`, num1, num2)
}

func verifyChallenge(resp http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Redirect(resp, req, "/challenge", http.StatusSeeOther)
		return
	}
	req.ParseForm()
	userAnswer := req.FormValue("answer")

	if userAnswer == challengeAnswer {
		resp.WriteHeader(http.StatusOK)
		resp.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(resp, `<h1>Вы прошли проверку безопасности.</h1>`)
	} else {
		resp.WriteHeader(http.StatusForbidden)
		resp.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(resp, `<h1>Попробуйте снова.</h1>`)
	}
}

// --- Основная функция ---

func main() {
	logger.Init()
	defer logger.Close()

	if err := logger.InitAuditLogger("blocked_requests.log"); err != nil {
		slog.Error("Ошибка инициализации Audit Logger", "error", err)
		os.Exit(1)
	}
	defer logger.CloseAuditLogger()

	cfg := config.New()

	rds := cache.NewRedis(cfg)
	defer rds.CloseRedis()

	router := app.NewRouter([]string{cfg.BackendURL}, cfg, rds)

	server := &http.Server{
		Addr:         cfg.Port,
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	logger.Log.Info("Сервер запущен", "port", cfg.Port)

	if err := server.ListenAndServe(); err != nil {
		logger.Log.Fatal("Ошибка при запуске сервера", "fatal_err", err)
	}
}
