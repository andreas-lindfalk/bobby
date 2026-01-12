package flows

import (
	"context"
	"testing"
	"time"
)

func TestRunVehicleImportFlow(t *testing.T) {
	tests := []struct {
		name          string
		input         VehicleImportInput
		wantHotLead   bool
		wantPlanSteps int
	}{
		{
			name: "September arrival with car details - standard flow",
			input: VehicleImportInput{
				ArrivalDate:  time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC),
				CarValue:     25000,
				CO2Emissions: 140,
			},
			wantHotLead:   false,
			wantPlanSteps: 6, // Padrón + NIE + ITV + DGT + Tax + Sweden
		},
		{
			name: "Electric car - no tax step",
			input: VehicleImportInput{
				ArrivalDate:  time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC),
				CarValue:     40000,
				CO2Emissions: 0,
				IsElectric:   true,
			},
			wantHotLead:   false,
			wantPlanSteps: 5, // No tax payment step
		},
		{
			name: "Very late arrival - hot lead",
			input: VehicleImportInput{
				ArrivalDate: time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC),
			},
			wantHotLead:   true,
			wantPlanSteps: 5, // No tax step (no car details)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := RunVehicleImportFlow(context.Background(), tt.input)
			if err != nil {
				t.Fatalf("RunVehicleImportFlow() error = %v", err)
			}

			if output.IsHotLead != tt.wantHotLead {
				t.Errorf("IsHotLead = %v, want %v", output.IsHotLead, tt.wantHotLead)
			}

			if len(output.ActionPlan) != tt.wantPlanSteps {
				t.Errorf("ActionPlan steps = %v, want %v", len(output.ActionPlan), tt.wantPlanSteps)
			}
		})
	}
}

func TestGenerateActionPlan(t *testing.T) {
	// Basic test - ensure we get steps in order
	output, _ := RunVehicleImportFlow(context.Background(), VehicleImportInput{
		ArrivalDate:  time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC),
		CarValue:     25000,
		CO2Emissions: 140,
	})

	for i, item := range output.ActionPlan {
		if item.Order != i+1 {
			t.Errorf("Step %d has Order = %d, want %d", i, item.Order, i+1)
		}
		if item.Title == "" {
			t.Errorf("Step %d has empty Title", i)
		}
	}
}
