package mcp

import (
	"context"
	"encoding/json"

	authModel "github.com/lin-snow/ech0/internal/model/auth"
)

func (a *Adapter) registerCommentTools(reg *Registry) {
	reg.RegisterTool(ToolDefinition{
		Name:        "list_comments",
		Title:       "List Comments",
		Description: "List all approved public comments for a specific post. Returns an array of comment objects (author, content, created_at).",
		InputSchema: map[string]any{
			"type":     "object",
			"required": []string{"echo_id"},
			"properties": map[string]any{
				"echo_id": map[string]any{"type": "string", "format": "uuid", "description": "Post UUID whose comments to retrieve"},
			},
		},
	}, a.listComments, authModel.ScopeCommentRead)
}

func (a *Adapter) registerCommentResources(reg *Registry) {
	reg.RegisterResource(ResourceDefinition{
		URI:         "ech0://comments/recent",
		Name:        "recent_comments",
		Title:       "Recent Comments",
		Description: "The 20 most recent approved public comments across all posts, newest first.",
		MimeType:    "application/json",
	}, a.resourceRecentComments, authModel.ScopeCommentRead)
}

func (a *Adapter) listComments(ctx context.Context, args map[string]any) (*ToolCallResult, error) {
	echoID := stringArg(args, "echo_id")
	if echoID == "" {
		return textError("echo_id is required"), nil
	}
	comments, err := a.commentSvc.ListPublicByEchoID(ctx, echoID)
	if err != nil {
		return nil, err
	}
	return jsonResult(comments)
}

func (a *Adapter) resourceRecentComments(ctx context.Context, _ string) (*ResourceReadResult, error) {
	comments, err := a.commentSvc.ListPublicComments(ctx, 20)
	if err != nil {
		return nil, err
	}
	data, _ := json.Marshal(comments)
	return &ResourceReadResult{
		Contents: []ResourceContent{{URI: "ech0://comments/recent", MimeType: "application/json", Text: string(data)}},
	}, nil
}
