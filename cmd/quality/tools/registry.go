// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package tools

import "sync"

// DefaultRegistry is the default registry implementation
type DefaultRegistry struct {
	mu    sync.RWMutex
	tools map[string]QualityTool
}

// NewRegistry creates a new tool registry
func NewRegistry() *DefaultRegistry {
	return &DefaultRegistry{
		tools: make(map[string]QualityTool),
	}
}

// Register adds a tool to the registry
func (r *DefaultRegistry) Register(tool QualityTool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tools[tool.Name()] = tool
}

// GetTools returns all registered tools
func (r *DefaultRegistry) GetTools() []QualityTool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tools := make([]QualityTool, 0, len(r.tools))
	for _, tool := range r.tools {
		tools = append(tools, tool)
	}
	return tools
}

// GetToolsByLanguage returns tools for a specific language
func (r *DefaultRegistry) GetToolsByLanguage(language string) []QualityTool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var tools []QualityTool
	for _, tool := range r.tools {
		if tool.Language() == language {
			tools = append(tools, tool)
		}
	}
	return tools
}

// GetToolsByType returns tools of a specific type
func (r *DefaultRegistry) GetToolsByType(toolType ToolType) []QualityTool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var tools []QualityTool
	for _, tool := range r.tools {
		if tool.Type() == toolType || tool.Type() == BOTH {
			tools = append(tools, tool)
		}
	}
	return tools
}

// FindTool finds a tool by name
func (r *DefaultRegistry) FindTool(name string) QualityTool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.tools[name]
}

// Ensure DefaultRegistry implements ToolRegistry
var _ ToolRegistry = (*DefaultRegistry)(nil)
