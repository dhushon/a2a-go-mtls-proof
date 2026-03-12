# System Architecture: Zero-Trust Multi-Agent Flow

This document describes the design and logic distribution of the agentic system.

## 1. Directory Strategy: Shared vs. Exclusive

To keep the codebase modular and scalable, logic is partitioned based on its scope of use:

-   **`pkg/` (Shared)**: Contains core utilities used by every component in the system.
    -   `auth/`: Certificate loading, thumbprint calculation, and OBO exchange interfaces.
    -   `config/`: Viper-based configuration management.
    -   `agentcontext/`: Session identity and automated header propagation.
    -   `observability/`: Tiered metrics and OTel SDK initialization.
    -   `weather/`: Weather research logic used by both providers and consumers.
-   **`server/middleware/`**: Common HTTP security logic shared by all responders (e.g., mTLS Binding).
-   **`client/packing/`**: Entry point and domain-specific analysis logic for the packing agent.

## 2. Multi-Agent Workflow

The system demonstrates a chain of autonomous agents:

1.  **Requester Agent (`client/packing`)**:
    -   Acts as the entry point for a user session.
    -   Performs a Token Exchange (OBO) via `auth.OBOExchange`, binding the token to its mTLS certificate.
    -   Calls the Weather Agent via mTLS using a session-aware context.
2.  **Responder Agent (`server/weather`)**:
    -   Validates the incoming requester's identity via `middleware.MTLSBindingMiddleware`.
    -   Performs weather research using the `weather` package.
    -   Returns findings (forecast + probability chart).
3.  **Analysis**:
    -   The Packing Agent consumes the weather data and makes an autonomous decision based on temperature thresholds.

## 3. Security Implementation

The system enforces zero-trust at every hop:
-   **Transport**: mTLS verifies service-level identity.
-   **Application**: OBO Token verifies user-level identity.
-   **Binding**: The `x5t#S256` thumbprint ensures that even if an OBO token is intercepted, it is useless without the requester's private key.

## 4. Observability

Every agent hop records:
-   **Latency**: Time spent in research and analysis.
-   **Cost**: Simulated token counts and USD costs per agentic step.
-   **Tracing**: (Level 3) Full distributed trace linking multiple agents into a single session timeline.
