// Package tools provides MCP tool infrastructure and registry.
package tools

// Tool represents an MCP tool definition.
type Tool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Parameters  interface{} `json:"parameters"`
}

// ToolResult represents the result of a tool invocation.
type ToolResult struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// ToolHandler is a function that handles a tool invocation.
type ToolHandler func(paramsJSON []byte) ToolResult

// ErrorResult creates a failed ToolResult with an error message.
func ErrorResult(err string) ToolResult {
	return ToolResult{Success: false, Error: err}
}

// SuccessResult creates a successful ToolResult with data.
func SuccessResult(data interface{}) ToolResult {
	return ToolResult{Success: true, Data: data}
}
