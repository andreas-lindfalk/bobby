package carimport

import (
	"math"
	"testing"
	"time"
)

func TestCalculateTax(t *testing.T) {
	arrival := time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name       string
		car        CarDetails
		currentDay time.Time
		wantExempt bool
		wantAmount float64
		wantBand   float64
	}{
		{
			name: "Electric car - always exempt",
			car: CarDetails{
				Value:        30000,
				CO2Emissions: 0,
				IsElectric:   true,
			},
			currentDay: time.Date(2026, 1, 12, 0, 0, 0, 0, time.UTC),
			wantExempt: true,
			wantAmount: 0,
			wantBand:   0,
		},
		{
			name: "Within 60-day window - exempt",
			car: CarDetails{
				Value:        25000,
				CO2Emissions: 140,
			},
			currentDay: time.Date(2025, 10, 15, 0, 0, 0, 0, time.UTC),
			wantExempt: true,
			wantAmount: 0,
			wantBand:   CO2Band1Pct,
		},
		{
			name: "Low CO2 after exemption window - still 0% band",
			car: CarDetails{
				Value:        25000,
				CO2Emissions: 100,
			},
			currentDay: time.Date(2026, 1, 12, 0, 0, 0, 0, time.UTC),
			wantExempt: false,
			wantAmount: 0,
			wantBand:   0,
		},
		{
			name: "Medium-low CO2 after exemption - 4.75%",
			car: CarDetails{
				Value:        25000,
				CO2Emissions: 140,
			},
			currentDay: time.Date(2026, 1, 12, 0, 0, 0, 0, time.UTC),
			wantExempt: false,
			wantAmount: 1187.50,
			wantBand:   CO2Band1Pct,
		},
		{
			name: "Medium-high CO2 after exemption - 9.75%",
			car: CarDetails{
				Value:        25000,
				CO2Emissions: 180,
			},
			currentDay: time.Date(2026, 1, 12, 0, 0, 0, 0, time.UTC),
			wantExempt: false,
			wantAmount: 2437.50,
			wantBand:   CO2Band2Pct,
		},
		{
			name: "High CO2 after exemption - 14.75%",
			car: CarDetails{
				Value:        25000,
				CO2Emissions: 220,
			},
			currentDay: time.Date(2026, 1, 12, 0, 0, 0, 0, time.UTC),
			wantExempt: false,
			wantAmount: 3687.50,
			wantBand:   CO2Band3Pct,
		},
		{
			name: "Hybrid with higher CO2 - 75% reduction",
			car: CarDetails{
				Value:        30000,
				CO2Emissions: 140,
				IsHybrid:     true,
			},
			currentDay: time.Date(2026, 1, 12, 0, 0, 0, 0, time.UTC),
			wantExempt: false,
			wantAmount: 356.25,
			wantBand:   CO2Band1Pct,
		},
		{
			name: "Exactly on 60-day deadline - still exempt",
			car: CarDetails{
				Value:        25000,
				CO2Emissions: 140,
			},
			currentDay: time.Date(2025, 10, 31, 0, 0, 0, 0, time.UTC),
			wantExempt: true,
			wantAmount: 0,
			wantBand:   CO2Band1Pct,
		},
		{
			name: "Day 61 - exemption lost",
			car: CarDetails{
				Value:        25000,
				CO2Emissions: 140,
			},
			currentDay: time.Date(2025, 11, 1, 0, 0, 0, 0, time.UTC),
			wantExempt: false,
			wantAmount: 1187.50,
			wantBand:   CO2Band1Pct,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateTax(tt.car, arrival, tt.currentDay)

			if got.Exempt != tt.wantExempt {
				t.Errorf("Exempt = %v, want %v", got.Exempt, tt.wantExempt)
			}
			if !floatEquals(got.Amount, tt.wantAmount) {
				t.Errorf("Amount = %v, want %v", got.Amount, tt.wantAmount)
			}
			if got.TaxBand != tt.wantBand {
				t.Errorf("TaxBand = %v, want %v", got.TaxBand, tt.wantBand)
			}
		})
	}
}

func floatEquals(a, b float64) bool {
	return math.Abs(a-b) < 0.01
}
