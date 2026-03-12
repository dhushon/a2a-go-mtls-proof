package main

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"a2a-go-mtls-proof/pkg/agentcontext"
	"a2a-go-mtls-proof/pkg/auth"
	"a2a-go-mtls-proof/pkg/logger"
	"a2a-go-mtls-proof/server/middleware"
	"a2a-go-mtls-proof/pkg/observability"
	"a2a-go-mtls-proof/pkg/weather"
)

func weatherHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	md, _ := agentcontext.From(r.Context())
	
	var reqBody struct {
		ZipCode string `json:"zip_code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	logger.Info("Researching weather probability", "zip_code", reqBody.ZipCode, "session_id", md.SessionID)
	
	result, err := weather.Get10DayProbability(r.Context(), reqBody.ZipCode)
	if err != nil {
		http.Error(w, "Weather research failed", http.StatusInternalServerError)
		return
	}

	chart := result.GenerateProbabilityChart()
	
	observability.RecordStep(r.Context(), float64(time.Since(start).Milliseconds()))
	observability.RecordUsage(r.Context(), 1000, 0.01)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"zip_code": reqBody.ZipCode,
		"chart":    chart,
		"status":   "research complete",
	})
}

func main() {
	if err := observability.Init("weather-probability-agent"); err != nil {
		logger.Error("Failed to initialize observability", "error", err)
	}

	tlsConfig, err := auth.GetServerTLSConfig()
	if err != nil {
		logger.Error("Failed to get TLS config", "error", err)
		os.Exit(1)
	}

	mux := http.NewServeMux()
	finalHandler := agentcontext.Middleware(middleware.MTLSBindingMiddleware(http.HandlerFunc(weatherHandler)))
	mux.Handle("/weather", finalHandler)

	server := &http.Server{
		Addr:      ":8443",
		Handler:   mux,
		TLSConfig: tlsConfig,
	}

	logger.Info("Weather Agent starting on https://localhost:8443")
	if err := server.ListenAndServeTLS("", ""); err != nil {
		logger.Error("Server failed to start", "error", err)
		os.Exit(1)
	}
}
