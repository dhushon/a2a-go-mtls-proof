package weather

import (
	"context"
	"strings"
	"testing"
)

func TestGet10DayProbability(t *testing.T) {
	ctx := context.Background()
	zip := "12345"
	result, err := Get10DayProbability(ctx, zip)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ZipCode != zip {
		t.Errorf("expected zip %s, got %s", zip, result.ZipCode)
	}
	if len(result.Predictions) != 10 {
		t.Errorf("expected 10 predictions, got %d", len(result.Predictions))
	}
}

func TestGenerateProbabilityChart(t *testing.T) {
	result := &Result{
		ZipCode: "90210",
		Predictions: []Forecast{
			{Temp: 80.0, Probability: 0.9},
			{Temp: 60.0, Probability: 0.5},
		},
	}

	chart := result.GenerateProbabilityChart()

	if !strings.Contains(chart, "90210") {
		t.Error("chart missing zipcode")
	}
	if !strings.Contains(chart, "90%") {
		t.Error("chart missing probability percentage")
	}
	if !strings.Contains(chart, "█") {
		t.Error("chart missing visual bar")
	}
}
