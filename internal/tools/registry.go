package tools

import "context"

// Registry holds all registered tools and their handlers.
type Registry struct {
	tools    map[string]Tool
	handlers map[string]ToolHandler
}

// NewRegistry creates a new tool registry.
func NewRegistry() *Registry {
	return &Registry{
		tools:    make(map[string]Tool),
		handlers: make(map[string]ToolHandler),
	}
}

// Register adds a tool and its handler to the registry.
func (r *Registry) Register(tool Tool, handler ToolHandler) {
	r.tools[tool.Name] = tool
	r.handlers[tool.Name] = handler
}

// GetTools returns all registered tools.
func (r *Registry) GetTools() []Tool {
	tools := make([]Tool, 0, len(r.tools))
	for _, t := range r.tools {
		tools = append(tools, t)
	}
	return tools
}

// Invoke executes a tool by name with the given parameters.
func (r *Registry) Invoke(ctx context.Context, name string, paramsJSON []byte) ToolResult {
	handler, ok := r.handlers[name]
	if !ok {
		return ErrorResult("unknown tool: " + name)
	}
	return handler(paramsJSON)
}

// DefaultRegistry is the global tool registry.
var DefaultRegistry = NewRegistry()

// Register adds a tool to the default registry.
func Register(tool Tool, handler ToolHandler) {
	DefaultRegistry.Register(tool, handler)
}

// GetAllTools returns all tools from the default registry.
func GetAllTools() []Tool {
	return DefaultRegistry.GetTools()
}

// Invoke executes a tool from the default registry.
func Invoke(ctx context.Context, name string, paramsJSON []byte) ToolResult {
	return DefaultRegistry.Invoke(ctx, name, paramsJSON)
}
