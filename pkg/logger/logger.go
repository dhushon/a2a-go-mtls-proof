package logger

import (
	"context"
	"log/slog"
	"os"
)

// Logger defines the interface for logging within the project.
// This allows upstream applications to provide their own logger.
type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
	With(args ...any) Logger
}

// slogLogger is a default implementation using slog.
type slogLogger struct {
	*slog.Logger
}

func (l *slogLogger) With(args ...any) Logger {
	return &slogLogger{l.Logger.With(args...)}
}

var current Logger = &slogLogger{slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))}

// Set sets the global logger. This can be used to bind an upstream logger.
func Set(l Logger) {
	current = l
}

// Get returns the current logger.
func Get() Logger {
	return current
}

// New returns a new slog-based logger.
func New(handler slog.Handler) Logger {
	if handler == nil {
		handler = slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug})
	}
	return &slogLogger{slog.New(handler)}
}

// Convenience package-level logging functions
func Debug(msg string, args ...any) { current.Debug(msg, args...) }
func Info(msg string, args ...any)  { current.Info(msg, args...) }
func Warn(msg string, args ...any)  { current.Warn(msg, args...) }
func Error(msg string, args ...any) { current.Error(msg, args...) }

// Log is a more flexible logging method
func Log(ctx context.Context, level slog.Level, msg string, args ...any) {
	if sl, ok := current.(*slogLogger); ok {
		sl.Logger.Log(ctx, level, msg, args...)
	} else {
		// Fallback for non-slog loggers
		switch level {
		case slog.LevelDebug:
			current.Debug(msg, args...)
		case slog.LevelInfo:
			current.Info(msg, args...)
		case slog.LevelWarn:
			current.Warn(msg, args...)
		case slog.LevelError:
			current.Error(msg, args...)
		}
	}
}
