// Package main provides an MCP server that exposes Bobby's car import tools
// for use in Claude Desktop, VS Code, and other MCP clients.
//
// This uses Genkit's built-in MCP plugin - no need for external MCP libraries.
//
// Usage:
//
//	go build -o bobby-mcp ./cmd/mcp
//	./bobby-mcp
package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/andreas-lindfalk/bobby/internal/logic/carimport"
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/mcp"
)

func main() {
	ctx := context.Background()

	// Initialize Genkit (no LLM plugin needed - we're just exposing tools)
	g := genkit.Init(ctx)

	// Define car import tools as Genkit tools
	defineCarImportTools(g)

	// Create MCP server that automatically exposes all defined tools
	server := mcp.NewMCPServer(g, mcp.MCPServerOptions{
		Name:    "bobby",
		Version: "1.0.0",
	})

	log.SetOutput(os.Stderr) // MCP uses stdout for protocol, logs go to stderr
	log.Println("Starting Bobby MCP Server...")
	log.Printf("Registered tools: %v", server.ListRegisteredTools())

	// Start the stdio server (for Claude Desktop / VS Code integration)
	if err := server.ServeStdio(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

// defineCarImportTools registers car import tools with Genkit.
func defineCarImportTools(g *genkit.Genkit) {
	// Tool 1: Calculate Deadlines
	genkit.DefineTool(g, "calculate_deadlines",
		"Calculate car matriculation deadlines for Spain based on arrival date and padrón registration. Returns key deadlines like DGT inspection, tax payment, and residency requirements with urgency indicators.",
		func(ctx *ai.ToolContext, input *DeadlineInput) (*carimport.DeadlineResult, error) {
			arrivalDate, err := time.Parse("2006-01-02", input.ArrivalDate)
			if err != nil {
				return nil, err
			}

			var padronDate *time.Time
			if input.PadronDate != "" {
				pd, err := time.Parse("2006-01-02", input.PadronDate)
				if err != nil {
					return nil, err
				}
				padronDate = &pd
			}

			userCtx := carimport.UserContext{
				ArrivalDate: arrivalDate,
				PadronDate:  padronDate,
				CurrentDate: time.Now(),
			}

			result := carimport.CalculateDeadlines(userCtx)
			return &result, nil
		},
	)

	// Tool 2: Estimate Tax
	genkit.DefineTool(g, "estimate_tax",
		"Estimate Spanish car registration tax (Impuesto de Matriculación) based on car value and CO2 emissions. Includes electric/hybrid exemptions.",
		func(ctx *ai.ToolContext, input *TaxInput) (*carimport.TaxEstimate, error) {
			arrivalDate, err := time.Parse("2006-01-02", input.ArrivalDate)
			if err != nil {
				return nil, err
			}

			car := carimport.CarDetails{
				Value:        input.CarValue,
				CO2Emissions: input.CO2Emissions,
				IsElectric:   input.IsElectric,
				IsHybrid:     input.IsHybrid,
			}

			result := carimport.CalculateTax(car, arrivalDate, time.Now())
			return &result, nil
		},
	)
}

// DeadlineInput is the input schema for the calculate_deadlines tool.
type DeadlineInput struct {
	// Date of arrival in Spain (YYYY-MM-DD format)
	ArrivalDate string `json:"arrival_date"`
	// Date of padrón registration (YYYY-MM-DD format). Optional.
	PadronDate string `json:"padron_date,omitempty"`
}

// TaxInput is the input schema for the estimate_tax tool.
type TaxInput struct {
	// Car value in EUR
	CarValue float64 `json:"car_value"`
	// CO2 emissions in g/km
	CO2Emissions int `json:"co2_emissions"`
	// Whether the car is fully electric
	IsElectric bool `json:"is_electric,omitempty"`
	// Whether the car is a plug-in hybrid
	IsHybrid bool `json:"is_hybrid,omitempty"`
	// Date of arrival in Spain (YYYY-MM-DD format)
	ArrivalDate string `json:"arrival_date"`
}
