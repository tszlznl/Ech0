package mcp

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"

	authModel "github.com/lin-snow/ech0/internal/model/auth"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	echoModel "github.com/lin-snow/ech0/internal/model/echo"
)

func (a *Adapter) registerEchoTools(reg *Registry) {
	reg.RegisterTool(ToolDefinition{
		Name:        "search_posts",
		Title:       "Search Posts",
		Description: "Search posts with optional filters (query text, tags, pagination).",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"query":     map[string]any{"type": "string", "description": "Full-text search keyword"},
				"tag_ids":   map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "Filter by tag IDs"},
				"page":      map[string]any{"type": "integer", "description": "Page number (default 1)"},
				"page_size": map[string]any{"type": "integer", "description": "Items per page (default 20, max 100)"},
			},
		},
	}, a.searchPosts, authModel.ScopeEchoRead)

	reg.RegisterTool(ToolDefinition{
		Name:        "get_post",
		Title:       "Get Post",
		Description: "Get a single post by its ID.",
		InputSchema: map[string]any{
			"type":     "object",
			"required": []string{"id"},
			"properties": map[string]any{
				"id": map[string]any{"type": "string", "description": "Post UUID"},
			},
		},
	}, a.getPost, authModel.ScopeEchoRead)

	reg.RegisterTool(ToolDefinition{
		Name:        "list_tags",
		Title:       "List Tags",
		Description: "List all available tags.",
		InputSchema: map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		},
	}, a.listTags, authModel.ScopeEchoRead)

	reg.RegisterTool(ToolDefinition{
		Name:        "create_post",
		Title:       "Create Post",
		Description: "Create a new post.",
		InputSchema: map[string]any{
			"type":     "object",
			"required": []string{"content"},
			"properties": map[string]any{
				"content": map[string]any{"type": "string", "description": "Post content (Markdown supported)"},
				"tags":    map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "Tag names to attach"},
				"private": map[string]any{"type": "boolean", "description": "Whether the post is private (default false)"},
			},
		},
	}, a.createPost, authModel.ScopeEchoWrite)

	reg.RegisterTool(ToolDefinition{
		Name:        "update_post",
		Title:       "Update Post",
		Description: "Update an existing post by ID.",
		InputSchema: map[string]any{
			"type":     "object",
			"required": []string{"id"},
			"properties": map[string]any{
				"id":      map[string]any{"type": "string", "description": "Post UUID"},
				"content": map[string]any{"type": "string", "description": "New content"},
				"tags":    map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "Replacement tag names"},
				"private": map[string]any{"type": "boolean", "description": "Whether the post is private"},
			},
		},
	}, a.updatePost, authModel.ScopeEchoWrite)

	reg.RegisterTool(ToolDefinition{
		Name:        "delete_post",
		Title:       "Delete Post",
		Description: "Delete a post by ID.",
		InputSchema: map[string]any{
			"type":     "object",
			"required": []string{"id"},
			"properties": map[string]any{
				"id": map[string]any{"type": "string", "description": "Post UUID"},
			},
		},
	}, a.deletePost, authModel.ScopeEchoWrite)
}

func (a *Adapter) registerEchoResources(reg *Registry) {
	reg.RegisterResource(ResourceDefinition{
		URI:         "ech0://tags",
		Name:        "tags",
		Title:       "All Tags",
		Description: "List of all tags with usage counts.",
		MimeType:    "application/json",
	}, a.resourceTags, authModel.ScopeEchoRead)

	reg.RegisterResource(ResourceDefinition{
		URI:         "ech0://posts/recent",
		Name:        "recent_posts",
		Title:       "Recent Posts",
		Description: "Most recent posts (default 20).",
		MimeType:    "application/json",
	}, a.resourceRecentPosts, authModel.ScopeEchoRead)
}

// --- Tool handlers ---

func (a *Adapter) searchPosts(ctx context.Context, args map[string]any) (*ToolCallResult, error) {
	query := stringArg(args, "query")
	page := intArg(args, "page", 1)
	pageSize := intArg(args, "page_size", 20)
	if pageSize > 100 {
		pageSize = 100
	}

	var tagIDs []string
	if raw, ok := args["tag_ids"]; ok {
		if arr, ok := raw.([]any); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					tagIDs = append(tagIDs, s)
				}
			}
		}
	}

	result, err := a.echoSvc.QueryEchos(ctx, commonModel.EchoQueryDto{
		Page:     page,
		PageSize: pageSize,
		Search:   query,
		TagIDs:   tagIDs,
	})
	if err != nil {
		return nil, err
	}
	return jsonResult(result)
}

func (a *Adapter) getPost(ctx context.Context, args map[string]any) (*ToolCallResult, error) {
	id := stringArg(args, "id")
	if id == "" {
		return textError("id is required"), nil
	}
	echo, err := a.echoSvc.GetEchoById(ctx, id)
	if err != nil {
		return nil, err
	}
	return jsonResult(echo)
}

func (a *Adapter) listTags(_ context.Context, _ map[string]any) (*ToolCallResult, error) {
	tags, err := a.echoSvc.GetAllTags()
	if err != nil {
		return nil, err
	}
	return jsonResult(tags)
}

func (a *Adapter) createPost(ctx context.Context, args map[string]any) (*ToolCallResult, error) {
	content := stringArg(args, "content")
	if content == "" {
		return textError("content is required"), nil
	}
	private := boolArg(args, "private")
	tags := buildTags(args)

	echo := &echoModel.Echo{
		Content: content,
		Private: private,
		Tags:    tags,
	}
	if err := a.echoSvc.PostEcho(ctx, echo); err != nil {
		return nil, err
	}
	return textResult("post created successfully"), nil
}

func (a *Adapter) updatePost(ctx context.Context, args map[string]any) (*ToolCallResult, error) {
	id := stringArg(args, "id")
	if id == "" {
		return textError("id is required"), nil
	}

	echo := &echoModel.Echo{ID: id}
	if content, ok := args["content"]; ok {
		if s, ok := content.(string); ok {
			echo.Content = s
		}
	}
	if p, ok := args["private"]; ok {
		if b, ok := p.(bool); ok {
			echo.Private = b
		}
	}
	echo.Tags = buildTags(args)

	if err := a.echoSvc.UpdateEcho(ctx, echo); err != nil {
		return nil, err
	}
	return textResult("post updated successfully"), nil
}

func (a *Adapter) deletePost(ctx context.Context, args map[string]any) (*ToolCallResult, error) {
	id := stringArg(args, "id")
	if id == "" {
		return textError("id is required"), nil
	}
	if err := a.echoSvc.DeleteEchoById(ctx, id); err != nil {
		return nil, err
	}
	return textResult("post deleted successfully"), nil
}

// --- Resource handlers ---

func (a *Adapter) resourceTags(_ context.Context, _ string) (*ResourceReadResult, error) {
	tags, err := a.echoSvc.GetAllTags()
	if err != nil {
		return nil, err
	}
	data, _ := json.Marshal(tags)
	return &ResourceReadResult{
		Contents: []ResourceContent{{URI: "ech0://tags", MimeType: "application/json", Text: string(data)}},
	}, nil
}

func (a *Adapter) resourceRecentPosts(ctx context.Context, uri string) (*ResourceReadResult, error) {
	limit := 20
	if parts := strings.SplitN(uri, "?limit=", 2); len(parts) == 2 {
		if n, err := strconv.Atoi(parts[1]); err == nil && n > 0 && n <= 100 {
			limit = n
		}
	}
	result, err := a.echoSvc.QueryEchos(ctx, commonModel.EchoQueryDto{
		Page:     1,
		PageSize: limit,
	})
	if err != nil {
		return nil, err
	}
	data, _ := json.Marshal(result)
	return &ResourceReadResult{
		Contents: []ResourceContent{{URI: "ech0://posts/recent", MimeType: "application/json", Text: string(data)}},
	}, nil
}
