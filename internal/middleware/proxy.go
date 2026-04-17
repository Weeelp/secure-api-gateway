package middleware

import (
	"net/http"
	"net/http/httputil"
	"net/url"
)

// Proxy создает middleware, который пересылает запросы на targetURL
func Proxy(targetURL string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			target, err := url.Parse(targetURL)
			if err != nil {
				http.Error(w, "Invalid target URL", http.StatusInternalServerError)
				return
			}

			proxy := httputil.NewSingleHostReverseProxy(target)

			// Важно: меняем хост, чтобы внешний сервер понял, куда мы стучимся
			r.Host = target.Host

			// Вызываем прокси
			proxy.ServeHTTP(w, r)
		})
	}
}
