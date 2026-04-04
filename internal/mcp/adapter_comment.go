package mcp

import (
	"context"
	"encoding/json"

	authModel "github.com/lin-snow/ech0/internal/model/auth"
	model "github.com/lin-snow/ech0/internal/model/comment"
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

	reg.RegisterTool(ToolDefinition{
		Name:        "create_integration_comment",
		Title:       "Create Integration Comment",
		Description: "Post a comment on behalf of an integration/AI agent. Bypasses captcha and form-token verification. Requires comment:write scope.",
		InputSchema: map[string]any{
			"type":     "object",
			"required": []string{"echo_id", "content"},
			"properties": map[string]any{
				"echo_id":  map[string]any{"type": "string", "format": "uuid", "description": "Post UUID to comment on"},
				"content":  map[string]any{"type": "string", "description": "Comment text (max 200 characters)"},
				"nickname": map[string]any{"type": "string", "description": "Display name for the comment (defaults to 'Integration')"},
				"metadata": map[string]any{"type": "string", "description": "Optional metadata (e.g. model name, provider)"},
			},
		},
	}, a.createIntegrationComment, authModel.ScopeCommentWrite)
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

func (a *Adapter) createIntegrationComment(ctx context.Context, args map[string]any) (*ToolCallResult, error) {
	echoID := stringArg(args, "echo_id")
	content := stringArg(args, "content")
	if echoID == "" || content == "" {
		return textError("echo_id and content are required"), nil
	}
	dto := &model.CreateIntegrationCommentDto{
		EchoID:   echoID,
		Content:  content,
		Nickname: stringArg(args, "nickname"),
		Metadata: stringArg(args, "metadata"),
	}
	result, err := a.commentSvc.CreateIntegrationComment(ctx, "", "MCP", dto)
	if err != nil {
		return nil, err
	}
	return jsonResult(result)
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
