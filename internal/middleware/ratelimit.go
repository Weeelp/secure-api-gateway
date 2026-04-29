package middleware

import (
	"fmt"
	"net/http"
	"time"

	"secure-api-gateway/internal/cache"
)

func RateLimitMiddleware(rds *cache.Redis, limit int, window time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			rdsClinet := rds.GetRlEngine()

			userID, ok := ctx.Value(UserIDKey).(float64)
			if !ok {
				userID = 0
			}

			key := fmt.Sprintf("lim:%v", userID)
			count, err := rdsClinet.Incr(ctx, key).Result()
			if err != nil {
				http.Error(w, "Internal error", http.StatusInternalServerError)
				return
			}

			if count == 1 {
				rdsClinet.Expire(ctx, key, window)
			}

			if count > int64(limit) {
				w.Header().Set("X-Rate-Limit", "exceeded")
				http.Error(w, "Too many requests. Slow down!", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
