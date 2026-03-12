package logger

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
)

// mockLogger is a simple implementation of the Logger interface for testing.
type mockLogger struct {
	buffer *bytes.Buffer
}

func (m *mockLogger) Debug(msg string, args ...any) { m.buffer.WriteString("DEBUG: " + msg + "\n") }
func (m *mockLogger) Info(msg string, args ...any)  { m.buffer.WriteString("INFO: " + msg + "\n") }
func (m *mockLogger) Warn(msg string, args ...any)  { m.buffer.WriteString("WARN: " + msg + "\n") }
func (m *mockLogger) Error(msg string, args ...any) { m.buffer.WriteString("ERROR: " + msg + "\n") }
func (m *mockLogger) With(args ...any) Logger       { return m }

func TestUpstreamBinding(t *testing.T) {
	buffer := &bytes.Buffer{}
	mock := &mockLogger{buffer: buffer}

	// Bind the mock logger
	Set(mock)

	Info("test message")

	if !strings.Contains(buffer.String(), "INFO: test message") {
		t.Errorf("expected log to contain 'INFO: test message', got %q", buffer.String())
	}
}

func TestLocalLoggerInstance(t *testing.T) {
	buffer := &bytes.Buffer{}
	handler := slog.NewTextHandler(buffer, &slog.HandlerOptions{Level: slog.LevelInfo})
	localLogger := New(handler)

	localLogger.Info("local message")

	// slog.TextHandler output format is key=value
	if !strings.Contains(buffer.String(), "level=INFO") || !strings.Contains(buffer.String(), "msg=\"local message\"") {
		t.Errorf("expected log to contain level=INFO and msg=\"local message\", got %q", buffer.String())
	}
}
