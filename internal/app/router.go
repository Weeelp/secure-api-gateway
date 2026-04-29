package app

import (
	"net/http"
	"time"

	"secure-api-gateway/internal/cache"
	"secure-api-gateway/internal/config"
	"secure-api-gateway/internal/logger"
	"secure-api-gateway/internal/middleware"
	"secure-api-gateway/internal/proxy"
)

func NewRouter(targetURL string, cfg *config.Config, rds *cache.Redis) http.Handler {
	mux := http.NewServeMux()

	logMW := middleware.LoggerMiddleware
	blMW := middleware.IPBlacklistMiddleware(rds)
	bsMW := middleware.MaxBodySizeMiddleware(1024 * 1024)
	rlMW := middleware.RateLimitMiddleware(rds, 10, time.Minute)
	jwtMW := middleware.JWTAuthMiddleware([]byte(cfg.JWTS), rds)
	gwProxy, err := proxy.NewProxy(targetURL)
	if err != nil {
		logger.Log.Error("Failed to initialize proxy", "err", err)
	}

	homeHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log := logger.FromContext(r.Context())
		log.Debug("/: запрос на главную", "path", r.URL.Path)

		gwProxy.ServeHTTP(w, r)
	})

	healthHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		alive := gwProxy.IsAlive()
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
	mux.Handle("/health", logMW(blMW(healthHandler)))
	mux.Handle("/form", logMW(blMW(bsMW(formHandler))))
	mux.Handle("/favicon.ico", favicoHandler)

	return mux
}
