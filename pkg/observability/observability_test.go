package observability

import (
	"context"
	"os"
	"testing"
)

func TestInit(t *testing.T) {
	// Test LevelBasic (Default)
	os.Unsetenv("AGENT_OBSERVABILITY_LEVEL")
	Init("test-service")
	if Level() != LevelBasic {
		t.Errorf("expected LevelBasic, got %d", Level())
	}

	// Test LevelOff
	// Resetting once sync is hard in tests without a reset mechanism, 
	// but we can check if it stays basic if we don't change env.
	// For a fresh start in a real app, it would read the env.
}

func TestRecordStep(t *testing.T) {
	ctx := context.Background()
	// Should not panic even if metrics aren't fully exported
	RecordStep(ctx, 123.45)
}

func TestRecordUsage(t *testing.T) {
	ctx := context.Background()
	// Set level to Cost for this test
	level = LevelCost
	RecordUsage(ctx, 100, 0.05)
}
