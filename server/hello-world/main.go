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
)

func taskHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	md, _ := agentcontext.From(r.Context())
	logger.Info("Processing task", "session_id", md.SessionID, "trace_id", md.TraceID)
	
	time.Sleep(100 * time.Millisecond)
	
	observability.RecordStep(r.Context(), float64(time.Since(start).Milliseconds()))
	observability.RecordUsage(r.Context(), 500, 0.005)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"status":     "task completed",
		"session_id": md.SessionID,
	})
}

func main() {
	if err := observability.Init("responder-agent"); err != nil {
		logger.Error("Failed to initialize observability", "error", err)
	}

	tlsConfig, err := auth.GetServerTLSConfig()
	if err != nil {
		logger.Error("Failed to get TLS config", "error", err)
		os.Exit(1)
	}

	mux := http.NewServeMux()
	finalHandler := agentcontext.Middleware(middleware.MTLSBindingMiddleware(http.HandlerFunc(taskHandler)))
	mux.Handle("/task", finalHandler)

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
