// Package agent provides the Genkit-based AI agent for car import assistance.
package agent

import (
	"fmt"
	"strings"
	"time"

	"github.com/andreas-lindfalk/bobby/internal/embeddings"
	"github.com/andreas-lindfalk/bobby/internal/logic/carimport"
	"github.com/andreas-lindfalk/bobby/internal/store"
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
)

// Config holds configuration for the agent.
type Config struct {
	Store       *store.Store
	Embedder    embeddings.Embedder
	ModelName   string // e.g., "googleai/gemini-3-flash"
	RAGTopK     int    // Number of documents to retrieve (default: 5)
	Temperature float32
}

// Agent is the car import assistant powered by Genkit.
type Agent struct {
	g        *genkit.Genkit
	config   Config
	store    *store.Store
	embedder embeddings.Embedder
	tools    []ai.ToolRef
}

// New creates a new car import agent with the given Genkit instance and config.
func New(g *genkit.Genkit, cfg Config) *Agent {
	if cfg.RAGTopK == 0 {
		cfg.RAGTopK = 5
	}
	if cfg.Temperature == 0 {
		cfg.Temperature = 0.7
	}

	agent := &Agent{
		g:        g,
		config:   cfg,
		store:    cfg.Store,
		embedder: cfg.Embedder,
		tools:    make([]ai.ToolRef, 0),
	}

	// Register tools
	agent.registerTools()

	return agent
}

// registerTools registers all MCP-style tools with Genkit.
func (a *Agent) registerTools() {
	// Tool: Calculate matriculation deadlines
	deadlineTool := genkit.DefineTool(a.g, "calculate_deadlines",
		"Calculate car matriculation deadlines based on arrival date and optional padrón registration date. Returns deadline dates and urgency level (green/yellow/red).",
		func(ctx *ai.ToolContext, input DeadlineInput) (*DeadlineOutput, error) {
			arrivalDate, err := time.Parse("2006-01-02", input.ArrivalDate)
			if err != nil {
				return nil, fmt.Errorf("invalid arrival_date format, use YYYY-MM-DD: %w", err)
			}

			var padronDate *time.Time
			if input.PadronDate != "" {
				pd, err := time.Parse("2006-01-02", input.PadronDate)
				if err != nil {
					return nil, fmt.Errorf("invalid padron_date format, use YYYY-MM-DD: %w", err)
				}
				padronDate = &pd
			}

			userCtx := carimport.UserContext{
				ArrivalDate: arrivalDate,
				PadronDate:  padronDate,
				CurrentDate: time.Now(),
			}

			result := carimport.CalculateDeadlines(userCtx)
			return &DeadlineOutput{
				TaxResidenceDeadline: result.TaxResidenceDeadline.Format("2006-01-02"),
				TaxExemptionDeadline: result.TaxExemptionDeadline.Format("2006-01-02"),
				PadronDeadline:       formatOptionalDate(result.PadronDeadline),
				EarliestDeadline:     result.EarliestDeadline.Format("2006-01-02"),
				DaysRemaining:        result.DaysRemaining,
				Urgency:              string(result.Urgency),
				TaxExemptionMissed:   result.TaxExemptionMissed,
				IsHotLead:            result.Urgency == carimport.UrgencyRed,
			}, nil
		},
	)
	a.tools = append(a.tools, deadlineTool)

	// Tool: Estimate registration tax
	taxTool := genkit.DefineTool(a.g, "estimate_tax",
		"Estimate Spanish car registration tax (Impuesto de Matriculación) based on car value, CO2 emissions, and vehicle type. Returns tax amount in EUR.",
		func(ctx *ai.ToolContext, input TaxInput) (*TaxOutput, error) {
			arrivalDate, err := time.Parse("2006-01-02", input.ArrivalDate)
			if err != nil {
				return nil, fmt.Errorf("invalid arrival_date format, use YYYY-MM-DD: %w", err)
			}

			carDetails := carimport.CarDetails{
				Value:        input.CarValue,
				CO2Emissions: input.CO2Emissions,
				IsElectric:   input.IsElectric,
				IsHybrid:     input.IsHybrid,
			}

			result := carimport.CalculateTax(carDetails, arrivalDate, time.Now())

			return &TaxOutput{
				Amount:       result.Amount,
				TaxBand:      result.TaxBand,
				CO2Category:  result.CO2Category,
				Exempt:       result.Exempt,
				ExemptReason: result.ExemptReason,
			}, nil
		},
	)
	a.tools = append(a.tools, taxTool)

	// Tool: Search knowledge base (RAG)
	if a.embedder != nil && a.store != nil {
		searchTool := genkit.DefineTool(a.g, "search_knowledge_base",
			"Search the knowledge base for relevant information about Swedish car import to Spain. Use this to find details about deadlines, ITV inspections, taxes, required documents, and local services.",
			func(ctx *ai.ToolContext, input SearchInput) (*SearchOutput, error) {
				embedding, err := a.embedder.Embed(ctx, input.Query)
				if err != nil {
					return nil, fmt.Errorf("embedding query: %w", err)
				}

				docs, err := a.store.SearchByEmbedding(ctx, embedding, a.config.RAGTopK)
				if err != nil {
					return nil, fmt.Errorf("searching documents: %w", err)
				}

				results := make([]DocumentResult, len(docs))
				for i, doc := range docs {
					results[i] = DocumentResult{
						Title:   doc.Title,
						Content: doc.Content,
						Source:  doc.Source,
						URL:     doc.URL,
					}
				}

				return &SearchOutput{
					Documents: results,
					Query:     input.Query,
				}, nil
			},
		)
		a.tools = append(a.tools, searchTool)
	}
}

// Tools returns the registered tools for this agent.
func (a *Agent) Tools() []ai.ToolRef {
	return a.tools
}

// DeadlineInput is the input schema for the deadline calculator tool.
type DeadlineInput struct {
	ArrivalDate string `json:"arrival_date" jsonschema:"description=Date of arrival in Spain (YYYY-MM-DD)"`
	PadronDate  string `json:"padron_date,omitempty" jsonschema:"description=Date of padrón registration (YYYY-MM-DD) if already registered"`
}

// DeadlineOutput is the output schema for the deadline calculator tool.
type DeadlineOutput struct {
	TaxResidenceDeadline string `json:"tax_residence_deadline"`
	TaxExemptionDeadline string `json:"tax_exemption_deadline"`
	PadronDeadline       string `json:"padron_deadline,omitempty"`
	EarliestDeadline     string `json:"earliest_deadline"`
	DaysRemaining        int    `json:"days_remaining"`
	Urgency              string `json:"urgency"`
	TaxExemptionMissed   bool   `json:"tax_exemption_missed"`
	IsHotLead            bool   `json:"is_hot_lead"`
}

// TaxInput is the input schema for the tax estimator tool.
type TaxInput struct {
	CarValue     float64 `json:"car_value" jsonschema:"description=Car market value in EUR"`
	CO2Emissions int     `json:"co2_emissions" jsonschema:"description=CO2 emissions in g/km"`
	IsElectric   bool    `json:"is_electric,omitempty" jsonschema:"description=True if the car is fully electric (BEV)"`
	IsHybrid     bool    `json:"is_hybrid,omitempty" jsonschema:"description=True if the car is a plug-in hybrid (PHEV)"`
	ArrivalDate  string  `json:"arrival_date" jsonschema:"description=Date of arrival in Spain (YYYY-MM-DD)"`
}

// TaxOutput is the output schema for the tax estimator tool.
type TaxOutput struct {
	Amount       float64 `json:"amount"`
	TaxBand      float64 `json:"tax_band"`
	CO2Category  string  `json:"co2_category"`
	Exempt       bool    `json:"exempt"`
	ExemptReason string  `json:"exempt_reason,omitempty"`
}

// SearchInput is the input schema for the knowledge base search tool.
type SearchInput struct {
	Query string `json:"query" jsonschema:"description=Search query for finding relevant information"`
}

// SearchOutput is the output schema for the knowledge base search tool.
type SearchOutput struct {
	Documents []DocumentResult `json:"documents"`
	Query     string           `json:"query"`
}

// DocumentResult represents a single document from the knowledge base.
type DocumentResult struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	Source  string `json:"source"`
	URL     string `json:"url,omitempty"`
}

// SystemPrompt returns the system prompt for the car import agent.
func SystemPrompt() string {
	return strings.TrimSpace(`
You are Bobby, an expert AI assistant specializing in helping Swedish expats import and register their vehicles in Spain (Orihuela Costa / Costa Blanca region).

## Your Expertise
- Spanish vehicle matriculation process (registering foreign cars)
- Legal deadlines: 183-day rule after arrival, 30-day rule after padrón registration
- ITV homologation inspections and requirements
- Spanish taxes: Impuesto de Matriculación, IVTM
- Swedish-specific considerations: Transportstyrelsen deregistration, COC certificates
- Local services: ITV stations, DGT offices, gestorías in Alicante province

## Guidelines
1. **Always check deadlines first** - Use the calculate_deadlines tool to assess urgency
2. **Use the knowledge base** - Search for specific procedures, requirements, and local information
3. **Be proactive about risks** - Warn about common pitfalls (missing COC, headlight adjustments, etc.)
4. **Provide actionable steps** - Give concrete next steps, not just general advice
5. **Recommend professional help when appropriate** - For complex cases, suggest a gestoría

## Conversation Style
- Friendly but professional
- Concise answers, but thorough when needed
- Use Swedish/English terminology when helpful (e.g., "padrón", "matriculación")
- If the user's situation is urgent (red/yellow urgency), emphasize this clearly

## Tool Usage
- calculate_deadlines: Use when the user mentions arrival dates or asks about timing
- estimate_tax: Use when discussing car value, CO2 emissions, or registration costs
- search_knowledge_base: Use to find specific procedural details, local services, or requirements

Remember: Many users are stressed about bureaucracy in a foreign country. Be reassuring but honest about complexity.
`)
}

func formatOptionalDate(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("2006-01-02")
}
