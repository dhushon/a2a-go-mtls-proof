# A2A Go mTLS Proof of Concept (OBO Token Binding)

This project demonstrates a zero-trust, agent-to-agent (A2A) communication pattern in Go. It combines **mTLS (Mutual TLS)** at the transport layer with **OAuth2 "On-Behalf-Of" (OBO) token exchange** at the application layer, implementing certificate-bound tokens as defined in **RFC 8705**.

## Key Features

- **mTLS Authentication**: Both client and server authenticate using certificates signed by a shared Test CA.
- **OBO Token Binding**: Tokens are cryptographically bound to the client's mTLS certificate thumbprint (`cnf` claim with `x5t#S256`).
- **Context Propagation**: Automatic metadata cascading (SessionID, TraceID, ParentID) via `pkg/agentcontext`.
- **Tiered Observability**: Metrics for latency, token usage, and cost via `pkg/observability` (OpenTelemetry).
- **Structured Logging**: Context-aware, JSON-based logging via `pkg/logger`.

## Project Structure

- `client/`: Requester agent implementation.
- `server/`: Responder agent implementation.
- `pkg/agentcontext/`: Metadata extraction/injection and HTTP middleware.
- `pkg/auth/`: Centralized TLS and certificate handling logic.
- `pkg/logger/`: Structured logging wrapper.
- `pkg/observability/`: OpenTelemetry metrics and tracing stack.
- `certs/`: (Generated) Test certificates and keys.

## Quick Start

### 1. Prerequisites
- Go 1.26+
- OpenSSL (for certificate generation)

### 2. Generate Test Certificates
```bash
# This script creates a Test CA and signed certificates for server and client
./scripts/generate_certs.sh  # (Or use the openssl commands in the history)
```

### 3. Run the Responder Agent (Server)
```bash
export AGENT_OBSERVABILITY_LEVEL=1  # 0:Off, 1:Basic, 2:Cost, 3:Debug
go run server/main.go
```

### 4. Run the Requester Agent (Client)
```bash
export AGENT_OBSERVABILITY_LEVEL=1
go run client/main.go
```

## Architectural Design

The project follows a **Middleware-first** approach to security and metadata. The Responder (Server) uses a chain of handlers:
1. `agentcontext.Middleware`: Extracts metadata from headers and hydrates the Go context.
2. `MTLSBindingMiddleware`: Validates the mTLS certificate against the OBO token's `cnf` claim.
3. `taskHandler`: Executes business logic and records observability metrics.

For a detailed design overview, see [.gemini/architecture.md](.gemini/architecture.md).
