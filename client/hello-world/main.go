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
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	if err := observability.Init("requester-agent"); err != nil {
		logger.Error("Failed to initialize observability", "error", err)
	}

	// 1. Setup Session Context
	md := agentcontext.Metadata{
		SessionID: "session-abc-123",
		TraceID:   "trace-xyz-789",
		ParentID:  "requester-agent",
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

	// 2. Setup OBO Exchange
	exchange := auth.NewOBOExchange(
		cfg.OAuthServerType,
		cfg.OAuthDomain,
		cfg.OAuthClientID,
		cfg.OAuthPrivateKeyPath,
		cfg.OAuthMTLSDomain,
		thumbprint,
	)

	// 3. Exchange Token using Context
	oboToken, err := exchange.ExchangeToken(ctx, "initial-token", "api://responder-agent/access")
	if err != nil {
		logger.Error("Failed to exchange token", "error", err)
		os.Exit(1)
	}

	client := &http.Client{
		Transport: &http.Transport{TLSClientConfig: tlsConfig},
	}

	reqBody, _ := json.Marshal(map[string]string{"task": "hello-world-round-trip"})
	
	// 4. Create request with the Session Context
	req, err := http.NewRequestWithContext(ctx, "POST", "https://localhost:8443/task", bytes.NewBuffer(reqBody))
	if err != nil {
		logger.Error("Failed to create request", "error", err)
		os.Exit(1)
	}

	req.Header.Set("Authorization", "Bearer "+oboToken)
	req.Header.Set("Content-Type", "application/json")
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
