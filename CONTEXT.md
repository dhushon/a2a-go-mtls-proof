# Agentic Session Context Management

This document outlines the approach for managing agentic sessions in Go, addressing context propagation, configuration, early termination, and re-entrance.

## 1. Architectural Strategy: `pkg/agentcontext`

We recommend centralizing session metadata in a dedicated package to ensure consistency across the distributed system. This package follows an **interceptor-based pattern**, allowing both "well-known" and "developer-provided" handlers to participate in metadata propagation.

### The Metadata Carrier
```go
package agentcontext

type Metadata struct {
    SessionID string            `json:"session_id"`
    TraceID   string            `json:"trace_id"`
    ParentID  string            `json:"parent_id"`
    Params    map[string]string `json:"params,omitempty"`
}
```

## 2. Extensible Metadata Handlers

To support various transport mechanisms (HTTP, gRPC, NATS) and custom metadata types, we define a standard interface for metadata extraction and injection.

### Interface Definitions
```go
// Extractor pulls metadata from a transport carrier (e.g., *http.Request)
type Extractor interface {
    Extract(carrier any) (Metadata, error)
}

// Injector pushes metadata into a transport carrier (e.g., *http.Request)
type Injector interface {
    Inject(md Metadata, carrier any) error
}
```

### Well-Known Handlers
The package should include default implementations for:
- **`HTTPHeaderHandler`**: Handles standard `X-Session-ID`, `X-Trace-ID` headers.
- **`OTelBaggageHandler`**: Bridges agent metadata to/from OpenTelemetry Baggage for industry-standard observability.

## 3. Middleware & RoundTripper

The `pkg/agentcontext` package leverages these handlers to automate propagation.

### Server Middleware
The middleware uses a chain of `Extractors` to hydrate the `context.Context`. If no metadata is found, it may initialize a new `SessionID` and `TraceID`.

### Client RoundTripper
A custom `http.RoundTripper` uses a chain of `Injectors` to ensure that every outgoing request carries the active session context.

## 4. Configuration & Re-entrance

### Functional Options (Static Config)
Use functional options during agent instantiation for immutable settings (e.g., `WithMaxIterations(n)`).

### Context-Scoped Params (Dynamic Config)
Use the `Params` map within `Metadata` for dynamic, request-scoped configuration. This map is propagated automatically across agent boundaries.

### Re-entrance Pattern
Re-entrance is supported by "re-hydrating" a context from a persistent store:
1.  System retrieves the `Metadata` snapshot.
2.  `agentcontext.New(ctx, md)` creates a valid context for the resumed agent.
3.  The agent continues its execution, and the `RoundTripper` ensures any new downstream calls remain linked to the original `SessionID`.

## 5. Auditability & Logging

All log entries emitted via `pkg/logger` should automatically extract `SessionID` and `TraceID` from the context, providing a unified view of an agentic session across multiple hops and re-entrance events.

## 6. Observability & Metrics (OpenTelemetry)

Agentic workflows require specialized observability to manage costs, latency, and reliability. We recommend integrating **OpenTelemetry (OTel)** to capture traces, metrics, and logs in a unified format.

### Key Metrics to Establish
- **Cost & Token Usage**: Record input/output LLM tokens and calculate estimated cost per session.
- **Latency**: Measure total session duration and individual "Think-Act" cycle latency.
- **Iteration Count**: Track the number of cycles per session to detect runaway agents.
- **Success Rate**: Monitor terminal statuses (Complete, Failed, Timed Out).

### Propagation via Baggage
While `agentcontext.Metadata` handles session identity, **OTel Baggage** should be used for *distributed accumulators*.
- **Example**: If Agent A calls Agent B, Agent A passes the current "total session cost" in Baggage. Agent B adds its own cost and returns/cascades the new total.

### Trace Correlation
Every log entry emitted by `pkg/logger` MUST be correlated with the active OTel Trace ID. This allows for a "single pane of glass" view where logs from multiple agents and re-entrance events are perfectly interleaved in a trace timeline.
