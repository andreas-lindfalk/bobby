package tools

import (
	"encoding/json"
	"time"

	"github.com/andreas-lindfalk/bobby/internal/logic/carimport"
)

// DeadlineCalculatorParams are the parameters for the deadline calculator tool.
type DeadlineCalculatorParams struct {
	ArrivalDate string  `json:"arrival_date"`
	PadronDate  *string `json:"padron_date"`
}

// TaxEstimatorParams are the parameters for the tax estimator tool.
type TaxEstimatorParams struct {
	CarValue     float64 `json:"car_value"`
	CO2Emissions int     `json:"co2_emissions"`
	IsElectric   bool    `json:"is_electric"`
	IsHybrid     bool    `json:"is_hybrid"`
	ArrivalDate  string  `json:"arrival_date"`
}

func init() {
	RegisterCarImportTools(DefaultRegistry)
}

// RegisterCarImportTools registers all car import tools with the given registry.
func RegisterCarImportTools(r *Registry) {
	r.Register(
		Tool{
			Name:        "deadline_calculator",
			Description: "Calculate car matriculation deadlines based on arrival date and padrón registration status.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"arrival_date": map[string]string{
						"type":        "string",
						"description": "Date of arrival in Spain (YYYY-MM-DD)",
					},
					"padron_date": map[string]string{
						"type":        "string",
						"description": "Date of padrón registration (YYYY-MM-DD), optional",
					},
				},
				"required": []string{"arrival_date"},
			},
		},
		handleDeadlineCalculator,
	)

	r.Register(
		Tool{
			Name:        "tax_estimator",
			Description: "Estimate Spanish car registration tax based on car value and CO2 emissions.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"car_value": map[string]string{
						"type":        "number",
						"description": "Car value in EUR",
					},
					"co2_emissions": map[string]string{
						"type":        "integer",
						"description": "CO2 emissions in g/km",
					},
					"arrival_date": map[string]string{
						"type":        "string",
						"description": "Date of arrival in Spain (YYYY-MM-DD)",
					},
				},
				"required": []string{"car_value", "co2_emissions", "arrival_date"},
			},
		},
		handleTaxEstimator,
	)
}

func handleDeadlineCalculator(paramsJSON []byte) ToolResult {
	var params DeadlineCalculatorParams
	if err := json.Unmarshal(paramsJSON, &params); err != nil {
		return ErrorResult("invalid parameters: " + err.Error())
	}

	arrivalDate, err := time.Parse("2006-01-02", params.ArrivalDate)
	if err != nil {
		return ErrorResult("invalid arrival_date: " + err.Error())
	}

	var padronDate *time.Time
	if params.PadronDate != nil && *params.PadronDate != "" {
		pd, err := time.Parse("2006-01-02", *params.PadronDate)
		if err != nil {
			return ErrorResult("invalid padron_date: " + err.Error())
		}
		padronDate = &pd
	}

	userCtx := carimport.UserContext{
		ArrivalDate: arrivalDate,
		PadronDate:  padronDate,
		CurrentDate: time.Now(),
	}

	result := carimport.CalculateDeadlines(userCtx)
	return SuccessResult(result)
}

func handleTaxEstimator(paramsJSON []byte) ToolResult {
	var params TaxEstimatorParams
	if err := json.Unmarshal(paramsJSON, &params); err != nil {
		return ErrorResult("invalid parameters: " + err.Error())
	}

	arrivalDate, err := time.Parse("2006-01-02", params.ArrivalDate)
	if err != nil {
		return ErrorResult("invalid arrival_date: " + err.Error())
	}

	car := carimport.CarDetails{
		Value:        params.CarValue,
		CO2Emissions: params.CO2Emissions,
		IsElectric:   params.IsElectric,
		IsHybrid:     params.IsHybrid,
	}

	result := carimport.CalculateTax(car, arrivalDate, time.Now())
	return SuccessResult(result)
}
