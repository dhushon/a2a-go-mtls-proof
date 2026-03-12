package main

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"

	"a2a-go-mtls-proof/pkg/agentcontext"
	"a2a-go-mtls-proof/pkg/auth"
	"a2a-go-mtls-proof/pkg/logger"
	"a2a-go-mtls-proof/pkg/observability"

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

		// 3. Perform Binding Check: Compare cert thumbprint to the 'cnf' claim
		fingerprint := sha256.Sum256(clientCert.Raw)
		encodedThumbprint := base64.RawURLEncoding.EncodeToString(fingerprint[:])

		if claims.Confirmation.X5tS256 != encodedThumbprint {
			logger.Warn("Certificate-Token binding mismatch", 
				"cert_thumbprint", encodedThumbprint,
				"token_cnf", claims.Confirmation.X5tS256)
			http.Error(w, "Certificate-Token binding mismatch", http.StatusForbidden)
			return
		}

		// Success: Identity is confirmed for both the service (mTLS) and the user (OBO Token)
		logger.Info("Successfully validated token", "subject", claims.Subject)
		next.ServeHTTP(w, r)
	})
}

func taskHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	
	// Extract metadata from context (hydrated by middleware)
	md, _ := agentcontext.From(r.Context())
	
	logger.Info("Processing task", "session_id", md.SessionID, "trace_id", md.TraceID)
	
	// Simulate "agentic" work
	time.Sleep(100 * time.Millisecond)
	
	// Record metrics
	observability.RecordStep(r.Context(), float64(time.Since(start).Milliseconds()))
	observability.RecordUsage(r.Context(), 500, 0.005) // Fake 500 tokens, $0.005 cost

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"status":     "task completed",
		"session_id": md.SessionID,
	})
}

func main() {
	// Initialize Observability
	if err := observability.Init("responder-agent"); err != nil {
		logger.Error("Failed to initialize observability", "error", err)
	}

	tlsConfig, err := auth.GetServerTLSConfig()
	if err != nil {
		logger.Error("Failed to get TLS config", "error", err)
		os.Exit(1)
	}

	// Create a new ServeMux and apply the middleware
	mux := http.NewServeMux()
	
	// Chain: agentcontext (metadata extraction) -> MTLSBinding (auth) -> taskHandler
	finalHandler := agentcontext.Middleware(MTLSBindingMiddleware(http.HandlerFunc(taskHandler)))
	mux.Handle("/task", finalHandler)

	// Create the HTTPS server
	server := &http.Server{
		Addr:      ":8443",
		Handler:   mux,
		TLSConfig: tlsConfig,
	}

	logger.Info("Starting server on https://localhost:8443")
	if err := server.ListenAndServeTLS("", ""); err != nil {
		logger.Error("Server failed to start", "error", err)
		os.Exit(1)
	}
}
