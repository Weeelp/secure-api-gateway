package middleware

import (
	"net/http"

	"secure-api-gateway/internal/logger"

	"github.com/google/uuid"
)

// StructuredLogger - middleware для логирования
func StructuredLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := uuid.New().String()

		// Добавляем ID в контекст (полезно для отслеживания)
		l := logger.Log.With("request_id", requestID)
		ctx := logger.ToContext(r.Context(), l)
		r = r.WithContext(ctx)

		wrappedWriter := &responseWriterWithStatus{w, http.StatusOK}

		next.ServeHTTP(wrappedWriter, r)
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
