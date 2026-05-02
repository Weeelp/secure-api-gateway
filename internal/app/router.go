package app

import (
	"context"
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

func NewRouter(targetURLs []string, cfg *config.Config, rds *cache.Redis) http.Handler {
	mux := http.NewServeMux()
	ctx := context.Background()

	rlClient := rds.GetRlEngine()

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
		if gwProxy.IsAnyAlive() {
			w.Write([]byte("OK"))
		} else {
			w.WriteHeader(503)
		}
	})

	formHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log := logger.FromContext(r.Context())
		log.Debug("Form request received")
		gwProxy.ServeHTTP(w, r)
	})

	challengeHandler := func(w http.ResponseWriter, r *http.Request) {
		num1, num2 := rand.Intn(10)+1, rand.Intn(10)+1
		answer := num1 + num2
		key := "challenge:" + r.RemoteAddr

		err := rlClient.Set(ctx, key, strconv.Itoa(answer), 5*time.Minute).Err()
		if err != nil {
			logger.Log.Error("Redis challenge error", "err", err)
			http.Error(w, "Service temp unavailable", 500)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, "<h2>Решите задачу: %d + %d</h2><form method='POST' action='/challenge/verify'><input name='answer' type='number'><button>Проверить</button></form>", num1, num2)
	}

	verifyChallenge := func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Bad request", 400)
			return
		}

		userIP := r.RemoteAddr

		challengeKey := "challenge:" + userIP
		correctAnswer, err := rds.GetRlEngine().Get(ctx, challengeKey).Result()

		if err != nil || correctAnswer == "" {
			http.Error(w, "Challenge expired", http.StatusForbidden)
			return
		}

		if r.FormValue("answer") == correctAnswer {
			rds.GetRlEngine().Del(ctx, challengeKey)

			err := rds.GetIPEngine().Del(ctx, userIP).Err()
			if err != nil {
				logger.Log.Error("Failed to unblock IP", "ip", userIP, "err", err)
			} else {
				logger.Log.Info("IP unblocked via challenge", "ip", userIP)
			}

			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			fmt.Fprint(w, "<h1>✅ Верно! Вы разблокированы.</h1><p><a href='/'>Перейти на главную</a></p>")
		} else {
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprint(w, "<h1>❌ Неверно. Попробуйте еще раз.</h1>")
		}
	}

	mux.Handle("/", logMW(blMW(jwtMW(rlMW(homeHandler)))))
	mux.Handle("/app", logMW(blMW(homeHandler)))
	mux.Handle("/health", logMW(blMW(healthHandler)))
	mux.Handle("/form", logMW(blMW(bsMW(formHandler))))
	mux.Handle("/favicon.ico", http.NotFoundHandler())
	mux.HandleFunc("/challenge", challengeHandler)
	mux.HandleFunc("/challenge/verify", verifyChallenge)

	return mux
}
