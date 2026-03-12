package agentcontext

import (
	"context"
	"net/http"
)

type contextKey string

const SessionMetadataKey contextKey = "agent-session-metadata"

// Metadata carries session and trace identity across agentic hops.
type Metadata struct {
	SessionID string            `json:"session_id"`
	TraceID   string            `json:"trace_id"`
	ParentID  string            `json:"parent_id"`
	Params    map[string]string `json:"params,omitempty"`
}

// New creates a new context with the provided metadata.
func New(ctx context.Context, md Metadata) context.Context {
	return context.WithValue(ctx, SessionMetadataKey, md)
}

// From retrieves Metadata from the context.
func From(ctx context.Context) (Metadata, bool) {
	md, ok := ctx.Value(SessionMetadataKey).(Metadata)
	return md, ok
}

// ExtractFromRequest builds a Metadata object from HTTP headers.
func ExtractFromRequest(r *http.Request) Metadata {
	return Metadata{
		SessionID: r.Header.Get("X-Session-ID"),
		TraceID:   r.Header.Get("X-Trace-ID"),
		ParentID:  r.Header.Get("X-Agent-ID"),
	}
}

// InjectIntoRequest adds Metadata to HTTP headers for downstream calls.
func (md Metadata) InjectIntoRequest(r *http.Request) {
	if md.SessionID != "" {
		r.Header.Set("X-Session-ID", md.SessionID)
	}
	if md.TraceID != "" {
		r.Header.Set("X-Trace-ID", md.TraceID)
	}
	if md.ParentID != "" {
		r.Header.Set("X-Agent-ID", md.ParentID)
	}
}

// Middleware is a helper for servers to automatically hydrate the context.
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		md := ExtractFromRequest(r)
		ctx := New(r.Context(), md)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
