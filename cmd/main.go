package main

import (
	"fmt"
	"log/slog"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"time"

	"secure-api-gateway/internal/cache"
	"secure-api-gateway/internal/config"
	"secure-api-gateway/internal/logger"
	"secure-api-gateway/internal/middleware"
)

var proxy *httputil.ReverseProxy
var challengeAnswer string

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

func generateChallenge(resp http.ResponseWriter, req *http.Request) {
	// Генерируем простой вопрос
	num1 := rand.Intn(10) + 1
	num2 := rand.Intn(10) + 1
	challengeAnswer = strconv.Itoa(num1 + num2)

	resp.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(resp, `
		<form method="POST">
			<p>Решите задачу: %d + %d = ?</p>
			<input type="text" name="answer" required>
			<button type="submit">Проверить</button>
		</form>
	`, num1, num2)
}

func verifyChallenge(resp http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Redirect(resp, req, "/challenge", http.StatusSeeOther)
		return
	}

	req.ParseForm()
	answer := req.FormValue("answer")

	if answer == challengeAnswer {
		resp.WriteHeader(http.StatusOK)
		fmt.Fprint(resp, "Вы прошли проверку.")
	} else {
		resp.WriteHeader(http.StatusForbidden)
		fmt.Fprint(resp, "Попробуйте снова.")
	}
}

func main() {
	logger.Init()
	defer logger.Close()

	if err := logger.InitAuditLogger("blocked_requests.log"); err != nil {
		slog.Error("Ошибка инициализации Audit Logger", "error", err)
		os.Exit(1)
	}
	defer logger.CloseAuditLogger()

	cfg := config.New()

	cache.InitRedis()
	defer cache.CloseRedis()

	// Загрузка секретного ключа из окружения
	secretKey := []byte(os.Getenv("JWT_SECRET"))
	if secretKey == nil || len(secretKey) == 0 {
		slog.Error("JWT_SECRET not set")
		os.Exit(1)
	}

	backServer, err := url.Parse("http://localhost:9090")
	if err != nil {
		logger.Log.Fatal("Error server connection", "err", err)
	}

	proxy = httputil.NewSingleHostReverseProxy(backServer)

	mux := http.NewServeMux()

	mux.HandleFunc("/", homeHandler)
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/form", formHandler)

	// Роуты для Challenge
	mux.HandleFunc("/challenge", generateChallenge)
	mux.HandleFunc("/challenge/verify", verifyChallenge)

	//Сборка цепочки Middleware
	// Снаружи -> Внутрь: Bot Check -> Secure Headers -> JWT Auth -> Logger -> Handler

	handler := http.Handler(mux)

	handler = middleware.BotDetectionMiddleware(mux)
	handler = middleware.SecureHeadersMiddleware(handler)
	handler = middleware.JWTAuthMiddleware(secretKey)(handler)
	handler = middleware.StructuredLogger(handler)

	server := &http.Server{
		Addr:         cfg.Port,
		Handler:      handler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	logger.Log.Info(" Сервер запущен", "port", cfg.Port)

	if err := server.ListenAndServe(); err != nil {
		logger.Log.Fatal("Ошибка при запуске сервера", "fatal_err", err)
	}
}
