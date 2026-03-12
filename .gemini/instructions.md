# Developer Guidelines: A2A mTLS Zero-Trust

When extending this project or implementing new agents, adhere to the following mandates:

## 1. Security First
- **Always use `pkg/auth`**: Do not manually configure `tls.Config`. Use the centralized functions to ensure proper CA verification and cert loading.
- **Enforce mTLS Binding**: All responder agents MUST include middleware that validates the OBO token's `cnf` claim against the presented peer certificate.

## 2. Context Propagation
- **Mandatory Metadata**: Every request MUST carry `SessionID` and `TraceID`.
- **Use `pkg/agentcontext`**:
    - Use `agentcontext.Middleware` on servers to hydrate the context.
    - Use `md.InjectIntoRequest(r)` on clients to cascade metadata.
- **Cascading**: If Agent A calls Agent B, it MUST propagate the incoming `SessionID` and `TraceID` to maintain the session chain.

## 3. Observability & Logging
- **Initialize Observability**: Call `observability.Init(serviceName)` in `main.go`.
- **Structured Recording**: Record steps and token usage in your handlers using `observability.RecordStep()` and `observability.RecordUsage()`.
- **Context-Aware Logging**: Use `pkg/logger` and pass relevant context to ensure logs are correlated with the active session.

## 4. Development Workflow
- **Test Before Acting**: Use the "Hello-World" pattern to verify your mTLS and context plumbing before attaching complex LLM or business logic.
- **Selective Enablement**: Test your agents with `AGENT_OBSERVABILITY_LEVEL=3` to ensure distributed traces are being correctly initiated.
