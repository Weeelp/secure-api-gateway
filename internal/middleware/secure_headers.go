// проверка заголовков, является ли ботом
package middleware

import (
	"net/http"
)

// SecureHeadersMiddleware добавляет безопасные заголовки
func SecureHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// HSTS
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

		// CSP
		w.Header().Set("Content-Security-Policy", "default-src 'self'")

		// X-Frame-Options
		w.Header().Set("X-Frame-Options", "DENY")

		// X-Content-Type-Options
		w.Header().Set("X-Content-Type-Options", "nosniff")

		// X-XSS-Protection
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		// Referrer-Policy
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		next.ServeHTTP(w, r)
	})
}
