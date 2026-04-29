package middleware

import (
	"fmt"
	"net"
	"net/http"

	"secure-api-gateway/internal/cache"
)

func IPBlacklistMiddleware(rds *cache.Redis) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			rdsClinet := rds.GetIPEngine()

			host, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				host = r.RemoteAddr
			}

			key := fmt.Sprintf("bl:%s", host)
			exists, err := rdsClinet.Exists(ctx, key).Result()
			
			if exists > 0 {
				http.Error(w, "Access denied: Banned IP", http.StatusForbidden) // 403
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
