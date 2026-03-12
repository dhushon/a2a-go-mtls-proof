package main

import (
	"fmt"
	"a2a-go-mtls-proof/pkg/weather"
)

// Recommendation represents the packing advice.
type Recommendation struct {
	Shorts int
	Pants  int
	Advice string
}

// SuggestPacking analyzes the weather forecast and suggests what to pack.
func SuggestPacking(forecasts []weather.Forecast) Recommendation {
	shorts := 0
	pants := 0
	
	for _, f := range forecasts {
		// Rule: If temp > 75F, pack shorts, else pants.
		if f.Temp > 75.0 {
			shorts++
		} else {
			pants++
		}
	}

	advice := ""
	if shorts > pants {
		advice = "Mostly warm weather expected. Pack more shorts!"
	} else if pants > shorts {
		advice = "Cooler weather ahead. Bring your pants."
	} else {
		advice = "Mixed weather. Bring a balance of both."
	}

	return Recommendation{
		Shorts: shorts,
		Pants:  pants,
		Advice: advice,
	}
}

// FormatRecommendation returns a string representation of the packing advice.
func (r Recommendation) FormatRecommendation() string {
	return fmt.Sprintf("Packing Suggestion:\n- Shorts: %d\n- Pants: %d\n- Summary: %s\n", r.Shorts, r.Pants, r.Advice)
}
