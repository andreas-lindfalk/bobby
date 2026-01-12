// Package flows contains Genkit flow definitions for the agentic workflows.
package flows

import (
	"context"
	"fmt"
	"time"

	"github.com/andreas-lindfalk/bobby/internal/logic/carimport"
)

// VehicleImportState represents the current state of the workflow.
type VehicleImportState string

const (
	StateInquiry       VehicleImportState = "inquiry"
	StateDeadlineCheck VehicleImportState = "deadline_check"
	StateTaxAssessment VehicleImportState = "tax_assessment"
	StateActionPlan    VehicleImportState = "action_plan"
	StateLeadHandoff   VehicleImportState = "lead_handoff"
	StateComplete      VehicleImportState = "complete"
)

// VehicleImportInput contains the user's input for the vehicle import flow.
type VehicleImportInput struct {
	ArrivalDate  time.Time  `json:"arrival_date"`
	PadronDate   *time.Time `json:"padron_date,omitempty"`
	CarValue     float64    `json:"car_value,omitempty"`
	CO2Emissions int        `json:"co2_emissions,omitempty"`
	IsElectric   bool       `json:"is_electric,omitempty"`
	IsHybrid     bool       `json:"is_hybrid,omitempty"`
}

// VehicleImportOutput contains the result of the vehicle import flow.
type VehicleImportOutput struct {
	State           VehicleImportState       `json:"state"`
	Deadlines       carimport.DeadlineResult `json:"deadlines"`
	TaxEstimate     *carimport.TaxEstimate   `json:"tax_estimate,omitempty"`
	ActionPlan      []ActionItem             `json:"action_plan,omitempty"`
	IsHotLead       bool                     `json:"is_hot_lead"`
	RecommendedNext string                   `json:"recommended_next"`
}

// ActionItem represents a single step in the action plan.
type ActionItem struct {
	Order       int    `json:"order"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Location    string `json:"location,omitempty"`
	Deadline    string `json:"deadline,omitempty"`
	Cost        string `json:"cost,omitempty"`
}

// RunVehicleImportFlow executes the vehicle import workflow.
func RunVehicleImportFlow(ctx context.Context, input VehicleImportInput) (*VehicleImportOutput, error) {
	output := &VehicleImportOutput{
		State: StateInquiry,
	}

	// Step 1: Calculate deadlines
	userCtx := carimport.UserContext{
		ArrivalDate: input.ArrivalDate,
		PadronDate:  input.PadronDate,
		CurrentDate: time.Now(),
	}
	output.Deadlines = carimport.CalculateDeadlines(userCtx)
	output.State = StateDeadlineCheck

	// Step 2: Calculate tax if car details provided
	if input.CarValue > 0 {
		car := carimport.CarDetails{
			Value:        input.CarValue,
			CO2Emissions: input.CO2Emissions,
			IsElectric:   input.IsElectric,
			IsHybrid:     input.IsHybrid,
		}
		tax := carimport.CalculateTax(car, input.ArrivalDate, time.Now())
		output.TaxEstimate = &tax
		output.State = StateTaxAssessment
	}

	// Step 3: Generate action plan
	output.ActionPlan = generateActionPlan(output.Deadlines, output.TaxEstimate)
	output.State = StateActionPlan

	// Step 4: Check for hot lead
	output.IsHotLead = carimport.IsHotLead(output.Deadlines)
	if output.IsHotLead {
		output.State = StateLeadHandoff
		output.RecommendedNext = "Connect with verified gestoría in Orihuela Costa"
	} else {
		output.State = StateComplete
		output.RecommendedNext = fmt.Sprintf("Start process within %d days", output.Deadlines.DaysRemaining)
	}

	return output, nil
}

// generateActionPlan creates a step-by-step action plan based on the user's situation.
func generateActionPlan(deadlines carimport.DeadlineResult, tax *carimport.TaxEstimate) []ActionItem {
	var plan []ActionItem
	order := 1

	// Step 1: Padrón registration (if not done)
	if !deadlines.PadronRegistered {
		plan = append(plan, ActionItem{
			Order:       order,
			Title:       "Register on Padrón",
			Description: "Register at your local town hall. This starts a 30-day countdown.",
			Location:    "Ayuntamiento de Orihuela (Playa Flamenca office)",
			Cost:        "Free",
		})
		order++
	}

	// Step 2: Get NIE if needed
	plan = append(plan, ActionItem{
		Order:       order,
		Title:       "Verify NIE/TIE Status",
		Description: "Ensure your NIE is valid for vehicle registration.",
		Location:    "Oficina de Extranjería or Comisaría",
	})
	order++

	// Step 3: ITV Homologation
	plan = append(plan, ActionItem{
		Order:       order,
		Title:       "ITV Homologation Inspection",
		Description: "Book import inspection. Tip: San Javier (Murcia) often has shorter waits than Torrevieja.",
		Location:    "ITV San Javier or ITV Torrevieja",
		Cost:        "~€150-200",
	})
	order++

	// Step 4: DGT Registration
	plan = append(plan, ActionItem{
		Order:       order,
		Title:       "DGT Vehicle Registration",
		Description: "Submit application for Spanish plates at traffic authority.",
		Location:    "DGT Alicante",
		Cost:        "~€100 (tasas)",
	})
	order++

	// Step 5: SUMA Tax Payment
	if tax != nil && !tax.Exempt && tax.Amount > 0 {
		plan = append(plan, ActionItem{
			Order:       order,
			Title:       "Pay Registration Tax (SUMA)",
			Description: fmt.Sprintf("Pay Impuesto de Matriculación. Estimated: €%.2f", tax.Amount),
			Location:    "SUMA Alicante or online",
			Cost:        fmt.Sprintf("€%.2f", tax.Amount),
		})
		order++
	}

	// Step 6: Swedish Deregistration
	plan = append(plan, ActionItem{
		Order:       order,
		Title:       "Deregister in Sweden",
		Description: "Notify Transportstyrelsen that the vehicle has been exported.",
		Location:    "Online at transportstyrelsen.se",
		Cost:        "Free",
	})

	return plan
}
