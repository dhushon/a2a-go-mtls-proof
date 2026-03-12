# Developer Guidelines: A2A mTLS Zero-Trust

When extending this project or implementing new agents, adhere to the following mandates:

## 1. Directory Structure
- **Shared Utilities**: Place logic used by both requesters and responders in root `/pkg/`.
- **Server Middleware**: Place common HTTP security logic in `/server/middleware/`.
- **Agent Entry Points**: Place each agent's `main.go` and its exclusive logic in dedicated subdirectories under `/client/` or `/server/`.

## 2. Security First
- **Always use `pkg/auth`**: Do not manually configure `tls.Config`. Use the centralized functions to ensure proper CA verification and cert loading.
- **Enforce mTLS Binding**: All responder agents MUST include `middleware.MTLSBindingMiddleware` (from `server/middleware`) to validate the OBO token's `cnf` claim against the presented peer certificate.
- **Use the OBO Interface**: Clients MUST perform token exchanges via the `auth.OBOExchange` interface to support both mock and live OAuth environments.

## 3. Context Propagation
- **Session Identity**: Every request MUST carry a `SessionID` and `TraceID`.
- **Proper Flow**: 
    1. Initialize `agentcontext.Metadata` at the start of a session.
    2. Create a session context via `agentcontext.New(context.Background(), md)`.
    3. Use this context for all subsequent calls (`ExchangeToken`, `http.NewRequestWithContext`).
- **Use `pkg/agentcontext`**: Use the middleware on servers and `md.InjectIntoRequest(r)` on clients to automate propagation.

## 4. Observability & Logging
- **Initialize Observability**: Call `observability.Init(serviceName)` in `main.go`.
- **Context-Aware Logging**: Use `pkg/logger` and pass the session context to ensure logs are correlated with the active session.

## 5. Development Workflow
- **Test Before Acting**: Use the "Hello-World" pattern to verify your mTLS and context plumbing before attaching complex business logic.
- **Coverage**: Maintain high unit test coverage for all shared packages in `/pkg/`.
