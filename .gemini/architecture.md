# System Architecture: A2A Zero-Trust Proof of Concept

This document describes the design of the agent-to-agent (A2A) zero-trust architecture.

## 1. Security Overview

The system uses a layered approach to security:
- **Transport Security (mTLS)**: Verifies the identity of the service presenting the certificate.
- **Application Security (OBO Token)**: Verifies the identity of the user on whose behalf the service is acting.
- **Binding (RFC 8705)**: Bridges these layers by ensuring the `oboToken` can only be used by the service possessing the specific certificate identified in the token's `cnf` claim.

## 2. Metadata & Context Propagation (`pkg/agentcontext`)

The `Metadata` struct carries session identity through the system:
- `SessionID`: Uniquely identifies the end-to-end user session.
- `TraceID`: Correlates distributed operations across multiple agents.
- `ParentID`: Identifies the immediate caller in the agentic chain.

### Lifecycle of a Request
1.  **Requester (Client)**: 
    - Sets metadata in a `Metadata` struct.
    - Uses `InjectIntoRequest(r)` to set `X-Session-ID`, `X-Trace-ID`, etc.
2.  **Responder (Server)**:
    - `agentcontext.Middleware` extracts headers.
    - Hydrates the Go `context` with a structured `Metadata` object.
    - Business logic accesses metadata via `agentcontext.From(ctx)`.

## 3. Observability Strategy (`pkg/observability`)

The system implements a tiered monitoring approach based on the `AGENT_OBSERVABILITY_LEVEL` environment variable.

### Tiered Levels
- **Level 1 (Basic)**: Records latency and iteration counts.
- **Level 2 (Cost)**: Level 1 + records token usage and financial cost.
- **Level 3 (Debug)**: Level 2 + starts OpenTelemetry trace spans for every task.

### Automatic Correlation
Every log entry emitted by `pkg/logger` is designed to be correlating with the active session and trace IDs extracted from the context, ensuring a "single pane of glass" view during analysis.

## 4. Architectural Patterns

### Middleware-Based Enforcement
Security and context extraction are handled as decorators (middleware) around the core business logic. This keeps the application code focused and ensures consistent enforcement.

### Stateless Re-entrance
By externalizing `Metadata` and session state into a structured context, the system is designed for "re-entrance." An agent session can be paused, persisted to a store, and re-hydrated later by restoring the `Metadata` object into a new context.
