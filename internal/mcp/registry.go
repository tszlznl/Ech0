package mcp

import (
	"context"
	"strings"
)

type ToolHandler func(ctx context.Context, args map[string]any) (*ToolCallResult, error)

type registeredTool struct {
	definition ToolDefinition
	handler    ToolHandler
	scopes     []string
}

type ResourceHandler func(ctx context.Context, uri string) (*ResourceReadResult, error)

type registeredResource struct {
	definition ResourceDefinition
	handler    ResourceHandler
	scopes     []string
	uriPrefix  string
}

type Registry struct {
	tools     []registeredTool
	resources []registeredResource
	toolIndex map[string]int
}

func NewRegistry() *Registry {
	return &Registry{
		toolIndex: make(map[string]int),
	}
}

func (r *Registry) RegisterTool(def ToolDefinition, handler ToolHandler, scopes ...string) {
	r.toolIndex[def.Name] = len(r.tools)
	r.tools = append(r.tools, registeredTool{
		definition: def,
		handler:    handler,
		scopes:     scopes,
	})
}

func (r *Registry) RegisterResource(def ResourceDefinition, handler ResourceHandler, scopes ...string) {
	var prefix string
	if idx := strings.Index(def.URI, "{"); idx > 0 {
		prefix = def.URI[:idx]
	}
	r.resources = append(r.resources, registeredResource{
		definition: def,
		handler:    handler,
		scopes:     scopes,
		uriPrefix:  prefix,
	})
}

func (r *Registry) ToolDefinitions() []ToolDefinition {
	defs := make([]ToolDefinition, len(r.tools))
	for i, t := range r.tools {
		defs[i] = t.definition
	}
	return defs
}

func (r *Registry) ResourceDefinitions() []ResourceDefinition {
	defs := make([]ResourceDefinition, len(r.resources))
	for i, res := range r.resources {
		defs[i] = res.definition
	}
	return defs
}

func (r *Registry) LookupTool(name string) (ToolHandler, []string, bool) {
	idx, ok := r.toolIndex[name]
	if !ok {
		return nil, nil, false
	}
	t := r.tools[idx]
	return t.handler, t.scopes, true
}

func (r *Registry) LookupResource(uri string) (ResourceHandler, []string, bool) {
	for _, res := range r.resources {
		if res.definition.URI == uri {
			return res.handler, res.scopes, true
		}
	}
	for _, res := range r.resources {
		if res.uriPrefix != "" && strings.HasPrefix(uri, res.uriPrefix) {
			return res.handler, res.scopes, true
		}
	}
	return nil, nil, false
}
