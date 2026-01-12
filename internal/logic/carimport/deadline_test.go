package carimport

import (
	"testing"
	"time"
)

func TestCalculateDeadlines(t *testing.T) {
	now := time.Date(2026, 1, 12, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name                   string
		arrival                time.Time
		padron                 *time.Time
		wantUrgency            Urgency
		wantDaysRemaining      int
		wantTaxExemptionMissed bool
		wantPadronRegistered   bool
	}{
		{
			name:                   "September arrival, no padrón - green urgency",
			arrival:                time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC),
			padron:                 nil,
			wantUrgency:            UrgencyGreen,
			wantDaysRemaining:      50, // Sept 1 + 183 = March 3, Jan 12 to March 3
			wantTaxExemptionMissed: true,
			wantPadronRegistered:   false,
		},
		{
			name:                   "Recent arrival - green urgency",
			arrival:                time.Date(2025, 12, 1, 0, 0, 0, 0, time.UTC),
			padron:                 nil,
			wantUrgency:            UrgencyGreen,
			wantDaysRemaining:      141, // Dec 1 + 183 = June 2
			wantTaxExemptionMissed: false,
			wantPadronRegistered:   false,
		},
		{
			name:                   "Very close to 183-day deadline - red urgency",
			arrival:                time.Date(2025, 7, 20, 0, 0, 0, 0, time.UTC),
			padron:                 nil,
			wantUrgency:            UrgencyRed,
			wantDaysRemaining:      7,
			wantTaxExemptionMissed: true,
			wantPadronRegistered:   false,
		},
		{
			name:                   "Padrón registered recently - padrón deadline applies",
			arrival:                time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC),
			padron:                 timePtr(time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC)),
			wantUrgency:            UrgencyYellow,
			wantDaysRemaining:      23,
			wantTaxExemptionMissed: true,
			wantPadronRegistered:   true,
		},
		{
			name:                   "Padrón deadline imminent - red urgency",
			arrival:                time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC),
			padron:                 timePtr(time.Date(2025, 12, 25, 0, 0, 0, 0, time.UTC)),
			wantUrgency:            UrgencyRed,
			wantDaysRemaining:      12,
			wantTaxExemptionMissed: true,
			wantPadronRegistered:   true,
		},
		{
			name:                   "Deadline already passed",
			arrival:                time.Date(2025, 5, 1, 0, 0, 0, 0, time.UTC),
			padron:                 nil,
			wantUrgency:            UrgencyRed,
			wantDaysRemaining:      -73, // May 1 + 183 = Oct 31, now Jan 12
			wantTaxExemptionMissed: true,
			wantPadronRegistered:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := UserContext{
				ArrivalDate: tt.arrival,
				PadronDate:  tt.padron,
				CurrentDate: now,
			}

			got := CalculateDeadlines(ctx)

			if got.Urgency != tt.wantUrgency {
				t.Errorf("Urgency = %v, want %v", got.Urgency, tt.wantUrgency)
			}
			if got.DaysRemaining != tt.wantDaysRemaining {
				t.Errorf("DaysRemaining = %v, want %v", got.DaysRemaining, tt.wantDaysRemaining)
			}
			if got.TaxExemptionMissed != tt.wantTaxExemptionMissed {
				t.Errorf("TaxExemptionMissed = %v, want %v", got.TaxExemptionMissed, tt.wantTaxExemptionMissed)
			}
			if got.PadronRegistered != tt.wantPadronRegistered {
				t.Errorf("PadronRegistered = %v, want %v", got.PadronRegistered, tt.wantPadronRegistered)
			}
		})
	}
}

func TestIsHotLead(t *testing.T) {
	tests := []struct {
		urgency Urgency
		want    bool
	}{
		{UrgencyRed, true},
		{UrgencyYellow, false},
		{UrgencyGreen, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.urgency), func(t *testing.T) {
			result := DeadlineResult{Urgency: tt.urgency}
			if got := IsHotLead(result); got != tt.want {
				t.Errorf("IsHotLead() = %v, want %v", got, tt.want)
			}
		})
	}
}

func timePtr(t time.Time) *time.Time {
	return &t
}
