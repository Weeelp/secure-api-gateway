package app

import (
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"secure-api-gateway/internal/cache"
	"secure-api-gateway/internal/config"
	"secure-api-gateway/internal/logger"
	"secure-api-gateway/internal/middleware"
	"secure-api-gateway/internal/proxy"
)

var challengeAnswer string

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
		fmt.Fprint(resp, `<h1>✅ Верно! Вы прошли проверку безопасности.</h1>`)
	} else {
		resp.WriteHeader(http.StatusForbidden)
		resp.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(resp, `<h1>❌ Неверно. Попробуйте снова.</h1>`)
	}
}

func NewRouter(targetURLs []string, cfg *config.Config, rds *cache.Redis) http.Handler {
	mux := http.NewServeMux()

	logMW := middleware.LoggerMiddleware
	blMW := middleware.IPBlacklistMiddleware(rds)
	bsMW := middleware.MaxBodySizeMiddleware(1024 * 1024)
	rlMW := middleware.RateLimitMiddleware(rds, 10, time.Minute)
	jwtMW := middleware.JWTAuthMiddleware([]byte(cfg.JWTS), rds)

	gwProxy, err := proxy.NewProxy(targetURLs)
	if err != nil {
		logger.Log.Error("Failed to initialize proxy", "err", err)
	}

	homeHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log := logger.FromContext(r.Context())
		log.Debug("/: запрос на главную", "path", r.URL.Path)

		gwProxy.ServeHTTP(w, r)
	})

	healthHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		alive := gwProxy.IsAnyAlive()
		log := logger.FromContext(r.Context())
		log.Debug("Health check performed", "is_alive", alive)

		w.Header().Set("Content-Type", "text/plain")

		if !alive {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("ERROR: Backend is down"))
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	formHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log := logger.FromContext(r.Context())
		log.Debug("Form request received")
		gwProxy.ServeHTTP(w, r)
	})

	favicoHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	mux.Handle("/", logMW(blMW(jwtMW(rlMW(homeHandler)))))
	mux.Handle("/app", logMW(blMW(homeHandler)))
	mux.Handle("/health", logMW(blMW(healthHandler)))
	mux.Handle("/form", logMW(blMW(bsMW(formHandler))))
	mux.Handle("/favicon.ico", favicoHandler)
	mux.HandleFunc("/challenge", generateChallenge)
	mux.HandleFunc("/challenge/verify", verifyChallenge)

	return mux
}
