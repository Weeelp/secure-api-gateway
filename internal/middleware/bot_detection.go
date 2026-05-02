package middleware

import (
	"net/http"
	"strings"
)

// BotDetectionMiddleware проверяет заголовки на подозрительные признаки ботов
func BotDetectionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userAgent := r.Header.Get("User-Agent")

		// Список подозрительных User-Agent
		suspiciousAgents := []string{
			"", // Пустой User-Agent
			"python-requests",
			"curl/",
			"wget/",
			"libwww-perl",
			"java/",
			"go-http-client",
		}

		for _, agent := range suspiciousAgents {
			if strings.Contains(strings.ToLower(userAgent), strings.ToLower(agent)) {
				http.Error(w, "Bot detected", http.StatusForbidden)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}
