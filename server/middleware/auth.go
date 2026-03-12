package middleware

import (
	"net/http"
	"strings"

	"a2a-go-mtls-proof/pkg/auth"
	"a2a-go-mtls-proof/pkg/logger"

	"github.com/golang-jwt/jwt/v5"
)

// MTLSBindingMiddleware ensures the token is constrained to the sender's certificate (RFC 8705).
func MTLSBindingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. Verify mTLS Connection
		if r.TLS == nil || len(r.TLS.PeerCertificates) == 0 {
			http.Error(w, "mTLS certificate required", http.StatusUnauthorized)
			return
		}
		clientCert := r.TLS.PeerCertificates[0]

		// 2. Parse OAuth2 Token
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "Bearer token required", http.StatusUnauthorized)
			return
		}
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

		claims := &auth.Claims{}
		// Note: In production, use a proper key provider to verify the signature.
		_, _, err := new(jwt.Parser).ParseUnverified(tokenStr, claims)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// 3. Perform Binding Check using shared utility
		encodedThumbprint := auth.GetCertificateThumbprint(clientCert)

		if claims.Confirmation.X5tS256 != encodedThumbprint {
			logger.Warn("Certificate-Token binding mismatch", 
				"cert_thumbprint", encodedThumbprint,
				"token_cnf", claims.Confirmation.X5tS256)
			http.Error(w, "Certificate-Token binding mismatch", http.StatusForbidden)
			return
		}

		// Success: Identity is confirmed
		logger.Info("Successfully validated bound token", "subject", claims.Subject)
		next.ServeHTTP(w, r)
	})
}
