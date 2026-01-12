// Package flows contains Genkit flow definitions for the agentic workflows.
package flows

import (
	"context"
	"fmt"
	"time"

	"github.com/andreas-lindfalk/bobby/internal/logic/carimport"
	"github.com/andreas-lindfalk/bobby/internal/store"
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
	Languages    []string   `json:"languages,omitempty"` // User's preferred languages (e.g., ["sv", "en"])
}

// VehicleImportOutput contains the result of the vehicle import flow.
type VehicleImportOutput struct {
	State              VehicleImportState       `json:"state"`
	Deadlines          carimport.DeadlineResult `json:"deadlines"`
	TaxEstimate        *carimport.TaxEstimate   `json:"tax_estimate,omitempty"`
	ActionPlan         []ActionItem             `json:"action_plan,omitempty"`
	RAGContext         []RAGDocument            `json:"rag_context,omitempty"` // Relevant docs from knowledge base
	IsHotLead          bool                     `json:"is_hot_lead"`
	RecommendedNext    string                   `json:"recommended_next"`
	RecommendedPartner *PartnerRecommendation   `json:"recommended_partner,omitempty"`
}

// RAGDocument represents a relevant document from the knowledge base.
type RAGDocument struct {
	Title   string `json:"title"`
	Summary string `json:"summary,omitempty"`
	Source  string `json:"source"`
	URL     string `json:"url,omitempty"`
}

// PartnerRecommendation represents a recommended gestoría or service provider.
type PartnerRecommendation struct {
	Name     string  `json:"name"`
	City     string  `json:"city"`
	Phone    string  `json:"phone,omitempty"`
	Email    string  `json:"email,omitempty"`
	Rating   float64 `json:"rating"`
	Discount string  `json:"discount,omitempty"`
}

// ActionItem represents a single step in the action plan.
type ActionItem struct {
	Order          int    `json:"order"`
	Title          string `json:"title"`
	Description    string `json:"description"`
	Location       string `json:"location,omitempty"`
	Address        string `json:"address,omitempty"`
	Phone          string `json:"phone,omitempty"`
	AppointmentURL string `json:"appointment_url,omitempty"`
	Deadline       string `json:"deadline,omitempty"`
	Cost           string `json:"cost,omitempty"`
	WaitDays       int    `json:"wait_days,omitempty"` // Expected wait time for appointments
}

// Embedder generates vector embeddings for RAG queries.
type Embedder interface {
	Embed(ctx context.Context, text string) ([]float32, error)
}

// FlowDependencies contains external dependencies for the flow.
type FlowDependencies struct {
	Store    *store.Store
	Embedder Embedder // Optional: enables RAG context retrieval
}

// RunVehicleImportFlow executes the vehicle import workflow.
// If deps is nil, the flow runs without store-backed recommendations.
func RunVehicleImportFlow(ctx context.Context, input VehicleImportInput, deps *FlowDependencies) (*VehicleImportOutput, error) {
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

	// Step 2: Query RAG for relevant context
	if deps != nil && deps.Store != nil && deps.Embedder != nil {
		query := "Swedish car import to Spain matriculation registration tax ITV"
		if input.IsElectric {
			query += " electric vehicle EV exemption"
		}
		if embedding, err := deps.Embedder.Embed(ctx, query); err == nil {
			if docs, err := deps.Store.SearchByEmbedding(ctx, embedding, 3); err == nil {
				for _, doc := range docs {
					output.RAGContext = append(output.RAGContext, RAGDocument{
						Title:   doc.Title,
						Summary: doc.Summary,
						Source:  doc.Source,
						URL:     doc.URL,
					})
				}
			}
		}
	}

	// Step 3: Calculate tax if car details provided
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

	// Step 4: Generate action plan
	var itv *store.ITVStation
	if deps != nil && deps.Store != nil {
		itv, _ = deps.Store.GetBestITVForImports(ctx) // Ignore error, just use default if not available
	}
	output.ActionPlan = generateActionPlan(output.Deadlines, output.TaxEstimate, itv)
	output.State = StateActionPlan

	// Step 5: Check for hot lead and get partner recommendation
	output.IsHotLead = carimport.IsHotLead(output.Deadlines)
	if output.IsHotLead {
		output.State = StateLeadHandoff
		output.RecommendedNext = "Connect with verified gestoría in Orihuela Costa"

		// Get partner recommendation from store
		if deps != nil && deps.Store != nil {
			languages := input.Languages
			if len(languages) == 0 {
				languages = []string{"en"} // Default to English
			}
			gestorias, err := deps.Store.GetVerifiedGestoriaForService(ctx, "matriculation", languages)
			if err == nil && len(gestorias) > 0 {
				g := gestorias[0] // First is partner (sorted by partner DESC, rating DESC)
				output.RecommendedPartner = &PartnerRecommendation{
					Name:   g.Name,
					City:   g.City,
					Phone:  g.Phone,
					Email:  g.Email,
					Rating: g.Rating,
				}
				if g.Partner {
					output.RecommendedPartner.Discount = "15% platform discount"
				}
			}
		}
	} else {
		output.State = StateComplete
		output.RecommendedNext = fmt.Sprintf("Start process within %d days", output.Deadlines.DaysRemaining)
	}

	return output, nil
}

// generateActionPlan creates a step-by-step action plan based on the user's situation.
func generateActionPlan(deadlines carimport.DeadlineResult, tax *carimport.TaxEstimate, itv *store.ITVStation) []ActionItem {
	var plan []ActionItem
	order := 1

	// Step 1: Padrón registration (if not done)
	if !deadlines.PadronRegistered {
		plan = append(plan, ActionItem{
			Order:          order,
			Title:          "Register on Padrón",
			Description:    "Register at your local town hall. This starts a 30-day countdown.",
			Location:       "Ayuntamiento de Orihuela (Playa Flamenca office)",
			Address:        "C/ Maestro Albéniz 2, La Zenia",
			Phone:          "+34 966 760 000",
			AppointmentURL: "https://orihuela.es/padron",
			Cost:           "Free",
		})
		order++
	}

	// Step 2: Get NIE if needed
	plan = append(plan, ActionItem{
		Order:          order,
		Title:          "Verify NIE/TIE Status",
		Description:    "Ensure your NIE is valid for vehicle registration.",
		Location:       "Oficina de Extranjeros Alicante",
		Address:        "C/ Ebanistería 4, 03008 Alicante",
		Phone:          "+34 965 936 840",
		AppointmentURL: "https://sede.administracionespublicas.gob.es",
	})
	order++

	// Step 3: ITV Homologation - use real data if available
	itvItem := ActionItem{
		Order:       order,
		Title:       "ITV Homologation Inspection",
		Description: "Book import inspection at ITV station.",
		Cost:        "~€150-200",
	}
	if itv != nil {
		itvItem.Location = itv.Name
		itvItem.Address = itv.Address
		itvItem.Phone = itv.Phone
		itvItem.AppointmentURL = itv.AppointmentURL
		itvItem.WaitDays = itv.AvgWaitDays
		itvItem.Description = fmt.Sprintf("Book import inspection. Current avg wait: %d days.", itv.AvgWaitDays)
	} else {
		itvItem.Location = "ITV Torrevieja"
		itvItem.Description = "Book import inspection. Tip: Check wait times online first."
	}
	plan = append(plan, itvItem)
	order++

	// Step 4: DGT Registration
	plan = append(plan, ActionItem{
		Order:          order,
		Title:          "DGT Vehicle Registration",
		Description:    "Submit application for Spanish plates at traffic authority.",
		Location:       "Jefatura Provincial de Tráfico Alicante",
		Address:        "C/ Churruca 24, 03003 Alicante",
		Phone:          "+34 965 935 800",
		AppointmentURL: "https://sedeapl.dgt.gob.es/WEB_NCIT_CONSULTA/sol498.faces",
		Cost:           "~€100 (tasas)",
	})
	order++

	// Step 5: SUMA Tax Payment
	if tax != nil && !tax.Exempt && tax.Amount > 0 {
		plan = append(plan, ActionItem{
			Order:          order,
			Title:          "Pay Registration Tax (SUMA)",
			Description:    fmt.Sprintf("Pay Impuesto de Matriculación. Estimated: €%.2f", tax.Amount),
			Location:       "Agencia Tributaria Torrevieja",
			Address:        "Av. de las Habaneras 78",
			Phone:          "+34 966 920 700",
			AppointmentURL: "https://www.agenciatributaria.es",
			Cost:           fmt.Sprintf("€%.2f", tax.Amount),
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
