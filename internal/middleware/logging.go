package middleware

import (
	"net/http"
	"time"

	"secure-api-gateway/internal/logger"

	"github.com/google/uuid"
)

// StructuredLogger - middleware для логирования
func StructuredLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := uuid.New().String()

		w.Header().Set("X-Request-ID", requestID)
		r.Header.Set("X-Request-ID", requestID)

		// Добавляем ID в контекст (полезно для отслеживания)
		l := logger.Log.With("request_id", requestID)
		ctx := logger.ToContext(r.Context(), l)
		r = r.WithContext(ctx)

		wrappedWriter := &responseWriterWithStatus{w, http.StatusOK}

		start := time.Now()

		next.ServeHTTP(wrappedWriter, r)

		l.Info("HTTP Request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", wrappedWriter.statusCode,
			"duration", time.Since(start).String(),
			"ip", r.RemoteAddr,
			"size", r.Header.Get("Content-Length"),
		)
	})
}

type responseWriterWithStatus struct {
	http.ResponseWriter
	statusCode int
}

func (w *responseWriterWithStatus) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}
