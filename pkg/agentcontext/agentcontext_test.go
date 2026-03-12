package agentcontext

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMetadataContext(t *testing.T) {
	md := Metadata{
		SessionID: "s1",
		TraceID:   "t1",
		ParentID:  "p1",
		Params:    map[string]string{"foo": "bar"},
	}

	ctx := New(context.Background(), md)
	retrieved, ok := From(ctx)

	if !ok {
		t.Fatal("expected metadata to be in context")
	}
	if retrieved.SessionID != md.SessionID {
		t.Errorf("expected SessionID %s, got %s", md.SessionID, retrieved.SessionID)
	}
	if retrieved.Params["foo"] != "bar" {
		t.Errorf("expected param foo=bar, got %s", retrieved.Params["foo"])
	}
}

func TestHTTPPropagation(t *testing.T) {
	md := Metadata{
		SessionID: "s-123",
		TraceID:   "t-456",
		ParentID:  "p-789",
	}

	// Test Injection
	req, _ := http.NewRequest("GET", "/", nil)
	md.InjectIntoRequest(req)

	if req.Header.Get("X-Session-ID") != md.SessionID {
		t.Errorf("header SessionID mismatch")
	}

	// Test Extraction
	extracted := ExtractFromRequest(req)
	if extracted.SessionID != md.SessionID || extracted.TraceID != md.TraceID {
		t.Errorf("extracted metadata mismatch")
	}
}

func TestMiddleware(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		md, ok := From(r.Context())
		if !ok {
			t.Error("middleware failed to hydrate context")
		}
		if md.SessionID != "mw-session" {
			t.Errorf("expected session mw-session, got %s", md.SessionID)
		}
	})

	mw := Middleware(handler)
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("X-Session-ID", "mw-session")
	
	rr := httptest.NewRecorder()
	mw.ServeHTTP(rr, req)
}
