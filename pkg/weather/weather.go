package weather

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

// Forecast represents a single day's weather prediction.
type Forecast struct {
	Date        time.Time
	Temp        float64
	Probability float64 // Probability of being "normal" for this date
}

// Result carries the 10-day prediction and historical analysis.
type Result struct {
	ZipCode    string
	Predictions []Forecast
}

// Get10DayProbability simulates fetching 10-day and 50-year data.
func Get10DayProbability(ctx context.Context, zipCode string) (*Result, error) {
	// Simulate API latency
	time.Sleep(200 * time.Millisecond)

	result := &Result{
		ZipCode: zipCode,
	}

	now := time.Now()
	for i := 0; i < 10; i++ {
		date := now.AddDate(0, 0, i)
		
		temp := 65.0 + rand.Float64()*20.0
		prob := 0.6 + rand.Float64()*0.4 

		result.Predictions = append(result.Predictions, Forecast{
			Date:        date,
			Temp:        temp,
			Probability: prob,
		})
	}

	return result, nil
}

// GenerateProbabilityChart returns a simple ASCII chart representing the probabilities.
func (r *Result) GenerateProbabilityChart() string {
	chart := fmt.Sprintf("10-Day Weather Probability Chart for %s\n", r.ZipCode)
	chart += "Date       | Prob | Visual\n"
	chart += "-----------|------|--------------------\n"
	for _, f := range r.Predictions {
		bar := ""
		for j := 0; j < int(f.Probability*20); j++ {
			bar += "█"
		}
		chart += fmt.Sprintf("%s | %2d%%  | %s\n", f.Date.Format("2006-01-02"), int(f.Probability*100), bar)
	}
	return chart
}
