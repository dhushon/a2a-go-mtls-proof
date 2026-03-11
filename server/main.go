package main

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"a2a-go-mtls-proof/pkg/auth"

	"github.com/golang-jwt/jwt/v5"
)

// Claims defines the JWT structure including the 'cnf' (confirmation) claim.
type Claims struct {
	jwt.RegisteredClaims
	Confirmation struct {
		X5tS256 string `json:"x5t#S256"` // SHA-256 thumbprint of the cert
	} `json:"cnf"`
}

// MTLSBindingMiddleware ensures the token is constrained to the sender's certificate.
func MTLSBindingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. Verify mTLS Connection: Extract the client certificate from the TLS state
		if r.TLS == nil || len(r.TLS.PeerCertificates) == 0 {
			http.Error(w, "mTLS certificate required", http.StatusUnauthorized)
			return
		}
		clientCert := r.TLS.PeerCertificates[0]

		// 2. Parse OAuth2 Token: Extract from the Authorization header
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "Bearer token required", http.StatusUnauthorized)
			return
		}
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

		claims := &Claims{}
		// In a real scenario, you would use a keyset to validate the signature
		// from the identity provider. For this example, we'll skip signature
		// verification by using jwt.UnsafeAllowNoneSignatureType.
		_, _, err := new(jwt.Parser).ParseUnverified(tokenStr, claims)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// It's still a good idea to validate the standard claims, like expiry and issuer
		// ...

		// 3. Perform Binding Check: Compare cert thumbprint to the 'cnf' claim
		fingerprint := sha256.Sum256(clientCert.Raw)
		encodedThumbprint := base64.RawURLEncoding.EncodeToString(fingerprint[:])

		if claims.Confirmation.X5tS256 != encodedThumbprint {
			log.Printf("Cert Thumbprint: %s", encodedThumbprint)
			log.Printf("Token cnf: %s", claims.Confirmation.X5tS256)
			http.Error(w, "Certificate-Token binding mismatch", http.StatusForbidden)
			return
		}

		// Success: Identity is confirmed for both the service (mTLS) and the user (OBO Token)
		// We can add the user's identity to the request context for the handler
		log.Printf("Successfully validated token for subject: %s", claims.Subject)
		next.ServeHTTP(w, r)
	})
}

func taskHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "task completed"})
}

func main() {
	tlsConfig, err := auth.GetServerTLSConfig()
	if err != nil {
		log.Fatalf("Failed to get TLS config: %v", err)
	}

	// Create a new ServeMux and apply the middleware
	mux := http.NewServeMux()
	finalHandler := http.HandlerFunc(taskHandler)
	mux.Handle("/task", MTLSBindingMiddleware(finalHandler))

	// Create the HTTPS server
	server := &http.Server{
		Addr:      ":8443",
		Handler:   mux,
		TLSConfig: tlsConfig,
	}

	log.Println("Starting server on https://localhost:8443")
	if err := server.ListenAndServeTLS("", ""); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
