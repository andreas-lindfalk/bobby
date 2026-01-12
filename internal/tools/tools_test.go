package tools

import (
	"context"
	"encoding/json"
	"testing"
)

func TestRegistryGetTools(t *testing.T) {
	tools := GetAllTools()
	if len(tools) < 2 {
		t.Errorf("Expected at least 2 tools, got %d", len(tools))
	}

	names := make(map[string]bool)
	for _, tool := range tools {
		names[tool.Name] = true
	}

	if !names["deadline_calculator"] {
		t.Error("Missing deadline_calculator tool")
	}
	if !names["tax_estimator"] {
		t.Error("Missing tax_estimator tool")
	}
}

func TestInvokeDeadlineCalculator(t *testing.T) {
	params := DeadlineCalculatorParams{ArrivalDate: "2025-09-01"}
	paramsJSON, _ := json.Marshal(params)
	result := Invoke(context.Background(), "deadline_calculator", paramsJSON)

	if !result.Success {
		t.Fatalf("Expected success, got error: %s", result.Error)
	}
	if result.Data == nil {
		t.Error("Expected data, got nil")
	}
}

func TestInvokeTaxEstimator(t *testing.T) {
	params := TaxEstimatorParams{
		CarValue:     25000,
		CO2Emissions: 140,
		ArrivalDate:  "2025-09-01",
	}
	paramsJSON, _ := json.Marshal(params)
	result := Invoke(context.Background(), "tax_estimator", paramsJSON)

	if !result.Success {
		t.Fatalf("Expected success, got error: %s", result.Error)
	}
	if result.Data == nil {
		t.Error("Expected data, got nil")
	}
}

func TestInvokeUnknownTool(t *testing.T) {
	result := Invoke(context.Background(), "unknown_tool", []byte("{}"))
	if result.Success {
		t.Error("Expected failure for unknown tool")
	}
}

func TestResultHelpers(t *testing.T) {
	errResult := ErrorResult("test error")
	if errResult.Success || errResult.Error != "test error" {
		t.Error("ErrorResult failed")
	}

	successResult := SuccessResult(42)
	if !successResult.Success || successResult.Data != 42 {
		t.Error("SuccessResult failed")
	}
}
