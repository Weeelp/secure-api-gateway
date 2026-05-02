package test

import (
	"net/http"
	"net/http/httptest"
	"secure-api-gateway/internal/middleware"
	"testing"
)

func TestSecureHeaders(t *testing.T) {
	handler := middleware.SecureHeadersMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	headers := recorder.Header()
	if headers.Get("Strict-Transport-Security") == "" {
		t.Error("Missing HSTS header")
	}
	if headers.Get("X-Frame-Options") == "" {
		t.Error("Missing X-Frame-Options header")
	}
}
