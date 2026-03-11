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
	"log"
	"net/http"
	"time"

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

// In a real OBO flow, this function would make a request to an identity provider.
// For this example, we'll simulate the IdP's behavior by generating a self-signed,
// certificate-bound token.
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

	// 3. Create and sign the token.
	// For this example, we'll just create a token without a signature, as the server
	// is configured to not verify it (for simplicity). In a real-world scenario,
	// the IdP would sign this with its private key.
	token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	tokenString, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	log.Printf("Generated OBO token with cnf: %s", encodedThumbprint)
	return tokenString, nil
}

func main() {
	tlsConfig, err := auth.GetClientTLSConfig()
	if err != nil {
		log.Fatalf("Failed to get TLS config: %v", err)
	}

	// 2. The "On-Behalf-Of" Token Exchange
	// In a real scenario, the user's identity would come from an incoming request.
	userSubject := "user-12345"
	oboToken, err := getOBOToken(userSubject, tlsConfig)
	if err != nil {
		log.Fatalf("Failed to get OBO token: %v", err)
	}

	// 3. Call the Responder Agent
	client := &http.Client{
		Transport: &http.Transport{TLSClientConfig: tlsConfig},
	}

	reqBody, _ := json.Marshal(map[string]string{"task": "analyze"})
	req, err := http.NewRequest("POST", "https://localhost:8443/task", bytes.NewBuffer(reqBody))
	if err != nil {
		log.Fatalf("Failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+oboToken)
	req.Header.Set("Content-Type", "application/json")

	log.Println("Sending request to https://localhost:8443/task")
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to read response body: %v", err)
	}

	log.Printf("Server Response: %s (Status: %s)", string(body), resp.Status)
}
