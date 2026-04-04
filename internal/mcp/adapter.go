package mcp

import (
	"encoding/json"
	"fmt"

	echoModel "github.com/lin-snow/ech0/internal/model/echo"
	echoService "github.com/lin-snow/ech0/internal/service/echo"
	userService "github.com/lin-snow/ech0/internal/service/user"
)

type Adapter struct {
	echoSvc echoService.Service
	userSvc userService.Service
}

func NewAdapter(echoSvc echoService.Service, userSvc userService.Service) *Adapter {
	return &Adapter{echoSvc: echoSvc, userSvc: userSvc}
}

func (a *Adapter) RegisterAll(reg *Registry) {
	a.registerEchoTools(reg)
	a.registerEchoResources(reg)
	a.registerUserResources(reg)
}

// --- Argument helpers ---

func stringArg(args map[string]any, key string) string {
	if v, ok := args[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func intArg(args map[string]any, key string, fallback int) int {
	if v, ok := args[key]; ok {
		switch n := v.(type) {
		case float64:
			return int(n)
		case json.Number:
			if i, err := n.Int64(); err == nil {
				return int(i)
			}
		}
	}
	return fallback
}

func boolArg(args map[string]any, key string) bool {
	if v, ok := args[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}

func buildTags(args map[string]any) []echoModel.Tag {
	raw, ok := args["tags"]
	if !ok {
		return nil
	}
	arr, ok := raw.([]any)
	if !ok {
		return nil
	}
	var tags []echoModel.Tag
	for _, v := range arr {
		if s, ok := v.(string); ok && s != "" {
			tags = append(tags, echoModel.Tag{Name: s})
		}
	}
	return tags
}

// --- Result helpers ---

func jsonResult(v any) (*ToolCallResult, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("marshal result: %w", err)
	}
	return &ToolCallResult{
		Content: []ContentItem{{Type: "text", Text: string(data)}},
	}, nil
}

func textResult(msg string) *ToolCallResult {
	return &ToolCallResult{
		Content: []ContentItem{{Type: "text", Text: msg}},
	}
}

func textError(msg string) *ToolCallResult {
	return &ToolCallResult{
		Content: []ContentItem{{Type: "text", Text: msg}},
		IsError: true,
	}
}
