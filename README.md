# A2A Go mTLS Proof of Concept (Zero-Trust Agents)

This project demonstrates a zero-trust, multi-agent communication pattern in Go. It implements certificate-bound tokens (RFC 8705) to ensure that agent-to-agent calls are cryptographically linked to the transport layer identity.

## Key Capabilities

- **Zero-Trust Handshake**: Combines mTLS with OBO (On-Behalf-Of) token binding.
- **Multi-Agent Workflows**:
    - **Weather Agent**: Performs 10-day research using 50 years of historical data (simulated).
    - **Packing Agent**: Consumes weather research to make autonomous "what to pack" decisions.
- **Rationalized Architecture**: 
    - Shared core logic in `./pkg`.
    - Server-exclusive middleware in `./server/pkg`.
    - Client-exclusive analysis in `./client/pkg`.
- **Tiered Observability**: Cost, token, and latency metrics via OpenTelemetry.

## Project Structure

- `client/`: Requester agents (e.g., `packing_main.go`).
    - `pkg/packing/`: Client-exclusive analysis logic.
- `server/`: Responder agents (e.g., `weather_main.go`).
    - `pkg/middleware/`: Server-exclusive HTTP security wrappers.
- `pkg/`: Shared packages (Auth, Context, Logger, Observability, Weather).
- `certs/`: (Generated) Test certificates and keys.

## Quick Start

### 1. Prerequisites
- Go 1.26+
- OpenSSL

### 2. Generate Test Certificates
```bash
# Generates CA, server, and client certs with proper SANs
./scripts/generate_certs.sh
```

### 3. Run the Multi-Agent Flow
Start the Weather Agent:
```bash
AGENT_OBSERVABILITY_LEVEL=1 go run server/weather_main.go
```

In a new terminal, run the Packing Agent:
```bash
AGENT_OBSERVABILITY_LEVEL=1 go run client/packing_main.go
```

## Security & Metadata
All agents utilize `pkg/agentcontext` for automatic metadata cascading (Session/Trace IDs) and `pkg/auth` for RFC 8705 compliant thumbprint verification.
