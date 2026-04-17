package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// StructuredLogger - middleware для логирования
func StructuredLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		requestID := uuid.New().String()

		// Добавляем ID в контекст (полезно для отслеживания)
		ctx := context.WithValue(r.Context(), "requestID", requestID)
		r = r.WithContext(ctx)

		wrappedWriter := &responseWriterWithStatus{w, http.StatusOK}

		next.ServeHTTP(wrappedWriter, r)

		slog.Info("request completed",
			"request_id", requestID,
			"method", r.Method,
			"path", r.URL.Path,
			"status", wrappedWriter.statusCode,
			"duration", time.Since(start),
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
