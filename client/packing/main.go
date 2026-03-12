package main

import (
	"bytes"
	"context"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"a2a-go-mtls-proof/pkg/agentcontext"
	"a2a-go-mtls-proof/pkg/auth"
	"a2a-go-mtls-proof/pkg/config"
	"a2a-go-mtls-proof/pkg/logger"
	"a2a-go-mtls-proof/pkg/observability"
	"a2a-go-mtls-proof/pkg/weather"
)

func main() {
	// 1. Load Configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// 2. Initialize Observability
	if err := observability.Init("packing-agent"); err != nil {
		logger.Error("Failed to initialize observability", "error", err)
	}

	// 3. Prepare Session Context FIRST
	md := agentcontext.Metadata{
		SessionID: "trip-session-999",
		TraceID:   "packing-trace-111",
		ParentID:  "packing-agent",
	}
	ctx := agentcontext.New(context.Background(), md)

	tlsConfig, err := auth.GetClientTLSConfig(cfg.A2AServerName)
	if err != nil {
		logger.Error("Failed to get TLS config", "error", err)
		os.Exit(1)
	}

	// Calculate thumbprint for mock binding
	cert, _ := x509.ParseCertificate(tlsConfig.Certificates[0].Certificate[0])
	thumbprint := auth.GetCertificateThumbprint(cert)

	// 4. Setup OBO Exchange
	exchange := auth.NewOBOExchange(
		cfg.OAuthServerType,
		cfg.OAuthDomain,
		cfg.OAuthClientID,
		cfg.OAuthPrivateKeyPath,
		cfg.OAuthMTLSDomain,
		thumbprint,
	)

	// 5. Perform Token Exchange using the Session Context
	subjectToken := "initial-user-token" 
	oboToken, err := exchange.ExchangeToken(ctx, subjectToken, "api://weather-agent/access")
	if err != nil {
		logger.Error("Failed to exchange token", "error", err)
		os.Exit(1)
	}

	// 6. Call the Weather Agent using the same context
	httpClient := &http.Client{
		Transport: &http.Transport{TLSClientConfig: tlsConfig},
	}

	zipCode := "90210"
	reqBody, _ := json.Marshal(map[string]string{"zip_code": zipCode})
	
	req, err := http.NewRequestWithContext(ctx, "POST", "https://localhost:8443/weather", bytes.NewBuffer(reqBody))
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

	// 7. Process Weather Result
	forecasts, _ := weather.Get10DayProbability(ctx, zipCode)
	recommendation := SuggestPacking(forecasts.Predictions)
	
	fmt.Println("--- Weather Research Result ---")
	fmt.Println(forecasts.GenerateProbabilityChart())
	fmt.Println("--- Packing Decision ---")
	fmt.Println(recommendation.FormatRecommendation())
	
	logger.Info("Packing analysis complete", "session_id", md.SessionID)
}
