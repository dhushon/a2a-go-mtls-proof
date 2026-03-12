package main

import (
	"strings"
	"testing"
	"a2a-go-mtls-proof/pkg/weather"
)

func TestSuggestPacking(t *testing.T) {
	tests := []struct {
		name      string
		forecasts []weather.Forecast
		wantAdvice string
	}{
		{
			name: "mostly hot",
			forecasts: []weather.Forecast{
				{Temp: 80.0}, {Temp: 85.0}, {Temp: 70.0},
			},
			wantAdvice: "Mostly warm weather expected",
		},
		{
			name: "mostly cold",
			forecasts: []weather.Forecast{
				{Temp: 60.0}, {Temp: 65.0}, {Temp: 70.0},
			},
			wantAdvice: "Cooler weather ahead",
		},
		{
			name: "balanced",
			forecasts: []weather.Forecast{
				{Temp: 80.0}, {Temp: 60.0},
			},
			wantAdvice: "Mixed weather",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := SuggestPacking(tt.forecasts)
			if !strings.Contains(rec.Advice, tt.wantAdvice) {
				t.Errorf("expected advice to contain %q, got %q", tt.wantAdvice, rec.Advice)
			}
		})
	}
}
