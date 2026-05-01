package middleware

import (
	"net/http"
	"strconv"
)

func MaxBodySizeMiddleware(maxSize int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			bodyLength := r.Header.Get("Content-Length")

			if bodyLength != "" {
				size, err := strconv.ParseInt(bodyLength, 10, 64)

				if err != nil {
					http.Error(w, "Invalid Content-Length", http.StatusBadRequest)
					return
				}

				if size > maxSize {
					w.Header().Set("Connection", "close")
					http.Error(w, "Request body too large", http.StatusRequestEntityTooLarge)
					return
				}
			}

			r.Body = http.MaxBytesReader(w, r.Body, maxSize)

			next.ServeHTTP(w, r)
		})
	}
}
