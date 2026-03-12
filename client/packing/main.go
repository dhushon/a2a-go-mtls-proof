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
	"a2a-go-mtls-proof/pkg/weather"

	"github.com/golang-jwt/jwt/v5"
)

func getOBOToken(userSubject string, tlsConfig *tls.Config) (string, error) {
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

	claims := auth.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "simulated-idp",
			Subject:   userSubject,
			Audience:  jwt.ClaimStrings{"api://weather-agent/access"},
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

	logger.Info("Generated OBO token for Weather Agent", "cnf", encodedThumbprint)
	return tokenString, nil
}

func main() {
	if err := observability.Init("packing-agent"); err != nil {
		logger.Error("Failed to initialize observability", "error", err)
	}

	tlsConfig, err := auth.GetClientTLSConfig()
	if err != nil {
		logger.Error("Failed to get TLS config", "error", err)
		os.Exit(1)
	}

	// 1. Get OBO Token
	userSubject := "traveler-5678"
	oboToken, err := getOBOToken(userSubject, tlsConfig)
	if err != nil {
		logger.Error("Failed to get OBO token", "error", err)
		os.Exit(1)
	}

	// 2. Prepare Session Metadata
	md := agentcontext.Metadata{
		SessionID: "trip-session-999",
		TraceID:   "packing-trace-111",
		ParentID:  "packing-agent",
	}

	// 3. Call the Weather Agent
	httpClient := &http.Client{
		Transport: &http.Transport{TLSClientConfig: tlsConfig},
	}

	zipCode := "90210"
	reqBody, _ := json.Marshal(map[string]string{"zip_code": zipCode})
	req, err := http.NewRequest("POST", "https://localhost:8443/weather", bytes.NewBuffer(reqBody))
	if err != nil {
		logger.Error("Failed to create request", "error", err)
		os.Exit(1)
	}

	req.Header.Set("Authorization", "Bearer "+oboToken)
	req.Header.Set("Content-Type", "application/json")
	md.InjectIntoRequest(req)

	logger.Info("Requesting weather research", "zip_code", zipCode, "session_id", md.SessionID)
	resp, err := httpClient.Do(req)
	if err != nil {
		logger.Error("Failed to call weather agent", "error", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logger.Error("Weather agent returned error", "status", resp.Status, "body", string(body))
		os.Exit(1)
	}

	// 4. Process Weather Result
	// Simulate receiving structured data for the packing logic
	forecasts, _ := weather.Get10DayProbability(req.Context(), zipCode)
	
	// 5. Suggest Packing (Logic moved into main package)
	recommendation := SuggestPacking(forecasts.Predictions)
	
	fmt.Println("--- Weather Research Result ---")
	fmt.Println(forecasts.GenerateProbabilityChart())
	
	fmt.Println("--- Packing Decision ---")
	fmt.Println(recommendation.FormatRecommendation())
	
	logger.Info("Packing analysis complete", "session_id", md.SessionID)
}
