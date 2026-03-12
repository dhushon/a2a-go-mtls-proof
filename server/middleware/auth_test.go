package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMTLSBindingMiddleware(t *testing.T) {
	// 1. Setup a dummy handler
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mw := MTLSBindingMiddleware(nextHandler)

	// 2. Test Case: Missing mTLS
	req, _ := http.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()
	mw.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for missing mTLS, got %d", rr.Code)
	}
}
