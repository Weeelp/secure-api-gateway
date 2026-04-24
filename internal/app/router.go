package app

import (
	"net/http"

	"secure-api-gateway/internal/logger"
	"secure-api-gateway/internal/middleware"
	"secure-api-gateway/internal/proxy"
)

func NewRouter(targetURL string) http.Handler {
	mux := http.NewServeMux()

	gwProxy, err := proxy.NewProxy(targetURL)
	if err != nil {
		logger.Log.Error("Failed to initialize proxy", "err", err)
	}

	logMW := middleware.StructuredLogger

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

	mux.Handle("/", logMW(homeHandler))
	mux.Handle("/health", logMW(healthHandler))
	mux.Handle("/form", logMW(formHandler))
	mux.Handle("/favicon.ico", favicoHandler)

	return mux
}
