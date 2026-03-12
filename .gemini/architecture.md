# System Architecture: Zero-Trust Multi-Agent Flow

This document describes the design and logic distribution of the agentic system.

## 1. Directory Strategy: Shared vs. Exclusive

To keep the codebase modular and scalable, logic is partitioned based on its scope of use:

-   **`pkg/` (Shared)**: Contains core utilities used by every component in the system.
    -   `auth/`: Certificate loading and RFC 8705 thumbprint calculation.
    -   `agentcontext/`: Session identity and automated header propagation.
    -   `observability/`: Tiered metrics and OTel SDK initialization.
    -   `weather/`: Weather research logic used by both the provider and consumers.
-   **`server/pkg/` (Server-Only)**: Logic exclusive to responders.
    -   `middleware/`: Standardized security enforcement (mTLS Binding).
-   **`client/pkg/` (Client-Only)**: Logic exclusive to requesters.
    -   `packing/`: Domain-specific analysis (e.g., shorts vs. pants decision engine).

## 2. Multi-Agent Workflow

The system demonstrates a chain of autonomous agents:

1.  **Requester Agent (`packing_main.go`)**:
    -   Acts as the entry point for a user session.
    -   Generates an OBO token bound to its mTLS certificate.
    -   Calls the Weather Agent via mTLS.
2.  **Responder Agent (`weather_main.go`)**:
    -   Validates the incoming requester's identity via `middleware.MTLSBindingMiddleware`.
    -   Performs weather research using the `weather` package.
    -   Returns findings (forecast + probability chart).
3.  **Analysis**:
    -   The Requester Agent consumes the weather data.
    -   Utilizes the `packing` package to make an autonomous decision based on temperature thresholds.

## 3. Security Implementation

The system enforces zero-trust at every hop:
-   **Transport**: mTLS verifies service-level identity.
-   **Application**: OBO Token verifies user-level identity.
-   **Binding**: The `x5t#S256` thumbprint ensures that even if an OBO token is intercepted, it is useless without the requester's private key.

## 4. Observability

Every agent hop records:
-   **Latency**: Time spent in the research and analysis phases.
-   **Cost**: Simulated token counts and USD costs per agentic step.
-   **Tracing**: (Level 3) Full distributed trace showing the relationship between the Packing Agent and the Weather Agent.
