package core

import (
	"fmt"

	"github.com/narcilee7/dew/pkg/tools"
)

// InMemoryToolRegistry is a simple in-memory tool registry.
type InMemoryToolRegistry struct {
	tools map[string]tools.Tool
}

// NewToolRegistry creates a new in-memory tool registry.
func NewToolRegistry() tools.ToolRegistry {
	return &InMemoryToolRegistry{tools: make(map[string]tools.Tool)}
}

// Register adds a tool to the registry.
func (r *InMemoryToolRegistry) Register(t tools.Tool) error {
	if t == nil {
		return fmt.Errorf("tool is nil")
	}
	name := t.Name()
	if name == "" {
		return fmt.Errorf("tool name is empty")
	}
	if _, exists := r.tools[name]; exists {
		return fmt.Errorf("tool %q already registered", name)
	}
	r.tools[name] = t
	return nil
}

// Get retrieves a tool by name.
func (r *InMemoryToolRegistry) Get(name string) (tools.Tool, bool) {
	t, ok := r.tools[name]
	return t, ok
}

// List returns all registered tools.
func (r *InMemoryToolRegistry) List() []tools.Tool {
	list := make([]tools.Tool, 0, len(r.tools))
	for _, t := range r.tools {
		list = append(list, t)
	}
	return list
}

// Filter returns a new registry containing only the named tools.
func (r *InMemoryToolRegistry) Filter(names []string) tools.ToolRegistry {
	filtered := NewToolRegistry()
	for _, name := range names {
		if t, ok := r.tools[name]; ok {
			_ = filtered.Register(t)
		}
	}
	return filtered
}
