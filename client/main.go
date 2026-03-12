package main

import (
	"bytes"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
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

// In a real OBO flow, this function would make a request to an identity provider.
func getOBOToken(userSubject string, tlsConfig *tls.Config) (string, error) {
	// 1. Calculate the certificate thumbprint
	if len(tlsConfig.Certificates) == 0 || len(tlsConfig.Certificates[0].Certificate) == 0 {
		return "", fmt.Errorf("no certificate found in the client credentials")
	}
	clientCert := tlsConfig.Certificates[0]
	cert, err := x509.ParseCertificate(clientCert.Certificate[0])
	if err != nil {
		return "", fmt.Errorf("failed to parse certificate: %w", err)
	}
	fingerprint := sha256.Sum256(cert.Raw)
	encodedThumbprint := base64.RawURLEncoding.EncodeToString(fingerprint[:])

	// 2. Create the claims for the JWT
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "simulated-idp",
			Subject:   userSubject,
			Audience:  jwt.ClaimStrings{"api://responder-agent/access"},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(5 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		Confirmation: struct {
			X5tS256 string `json:"x5t#S256"`
		}{
			X5tS256: encodedThumbprint,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	tokenString, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	logger.Info("Generated OBO token", "cnf", encodedThumbprint)
	return tokenString, nil
}

func main() {
	// Initialize Observability
	if err := observability.Init("requester-agent"); err != nil {
		logger.Error("Failed to initialize observability", "error", err)
	}

	tlsConfig, err := auth.GetClientTLSConfig()
	if err != nil {
		logger.Error("Failed to get TLS config", "error", err)
		os.Exit(1)
	}

	// 2. The "On-Behalf-Of" Token Exchange
	userSubject := "user-12345"
	oboToken, err := getOBOToken(userSubject, tlsConfig)
	if err != nil {
		logger.Error("Failed to get OBO token", "error", err)
		os.Exit(1)
	}

	// 3. Prepare Metadata for context propagation
	md := agentcontext.Metadata{
		SessionID: "session-abc-123",
		TraceID:   "trace-xyz-789",
		ParentID:  "requester-agent",
	}

	// 4. Call the Responder Agent
	client := &http.Client{
		Transport: &http.Transport{TLSClientConfig: tlsConfig},
	}

	reqBody, _ := json.Marshal(map[string]string{"task": "hello-world-round-trip"})
	req, err := http.NewRequest("POST", "https://localhost:8443/task", bytes.NewBuffer(reqBody))
	if err != nil {
		logger.Error("Failed to create request", "error", err)
		os.Exit(1)
	}

	req.Header.Set("Authorization", "Bearer "+oboToken)
	req.Header.Set("Content-Type", "application/json")
	
	// Inject Metadata into outgoing headers
	md.InjectIntoRequest(req)

	logger.Info("Sending request to https://localhost:8443/task", "session_id", md.SessionID)
	resp, err := client.Do(req)
	if err != nil {
		logger.Error("Failed to send request", "error", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("Failed to read response body", "error", err)
		os.Exit(1)
	}

	logger.Info("Server Response", "body", string(body), "status", resp.Status)
}
