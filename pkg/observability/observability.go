package observability

import (
	"context"
	"os"
	"strconv"
	"sync"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.27.0"
)

// Enablement levels
const (
	LevelOff   = 0
	LevelBasic = 1
	LevelCost  = 2
	LevelDebug = 3
)

var (
	level int
	once  sync.Once
	
	// Metrics
	meter        metric.Meter
	stepLatency  metric.Float64Histogram
	tokenCounter metric.Int64Counter
	costCounter  metric.Float64Counter
	stepCounter  metric.Int64Counter
)

// Init initializes the observability stack based on environment variables.
func Init(serviceName string) error {
	var err error
	once.Do(func() {
		levelStr := os.Getenv("AGENT_OBSERVABILITY_LEVEL")
		if levelStr == "" {
			level = LevelBasic // Default
		} else {
			level, _ = strconv.Atoi(levelStr)
		}

		if level == LevelOff {
			return
		}

		// Initialize Resource without explicit schema URL to avoid conflicts
		res := resource.NewWithAttributes(
			"", // Empty schema URL
			semconv.ServiceNameKey.String(serviceName),
		)
		
		// Initialize Meter
		meter = otel.GetMeterProvider().Meter(serviceName)

		// Define Base Metrics
		if level >= LevelBasic {
			stepLatency, _ = meter.Float64Histogram(
				"agent.step.latency",
				metric.WithDescription("Latency of a single agent step"),
				metric.WithUnit("ms"),
			)
			stepCounter, _ = meter.Int64Counter(
				"agent.steps.count",
				metric.WithDescription("Number of iterations performed"),
			)
		}

		if level >= LevelCost {
			tokenCounter, _ = meter.Int64Counter(
				"agent.tokens.usage",
				metric.WithDescription("Cumulative token usage"),
			)
			costCounter, _ = meter.Float64Counter(
				"agent.session.cost",
				metric.WithDescription("Calculated financial cost"),
				metric.WithUnit("USD"),
			)
		}
		
		// Dummy use of res to satisfy compiler if needed
		_ = res
	})
	return err
}

// Level returns the current enablement level.
func Level() int {
	return level
}

// RecordStep records the completion of an agent step.
func RecordStep(ctx context.Context, latencyMs float64, attrs ...attribute.KeyValue) {
	if level >= LevelBasic {
		if stepLatency != nil {
			stepLatency.Record(ctx, latencyMs, metric.WithAttributes(attrs...))
		}
		if stepCounter != nil {
			stepCounter.Add(ctx, 1, metric.WithAttributes(attrs...))
		}
	}
}

// RecordUsage records token and cost information.
func RecordUsage(ctx context.Context, tokens int64, cost float64, attrs ...attribute.KeyValue) {
	if level >= LevelCost {
		if tokenCounter != nil {
			tokenCounter.Add(ctx, tokens, metric.WithAttributes(attrs...))
		}
		if costCounter != nil {
			costCounter.Add(ctx, cost, metric.WithAttributes(attrs...))
		}
	}
}

// StartSpan starts a new trace span if debug level is enabled.
func StartSpan(ctx context.Context, name string) (context.Context, trace.Span) {
	if level >= LevelDebug {
		return otel.Tracer("agent").Start(ctx, name)
	}
	// Return a no-op span if not enabled
	return trace.NewNoopTracerProvider().Tracer("").Start(ctx, "")
}
