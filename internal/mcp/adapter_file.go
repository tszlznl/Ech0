package mcp

import (
	"context"

	authModel "github.com/lin-snow/ech0/internal/model/auth"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
)

func (a *Adapter) registerFileTools(reg *Registry) {
	reg.RegisterTool(ToolDefinition{
		Name:        "list_files",
		Title:       "List Files",
		Description: "List uploaded file metadata with optional search and storage type filter. Returns paginated results: {items, total, page, page_size}. Each item's id can be used as echo_files[].file_id in create_post/update_post. Does not return file contents.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"page":         map[string]any{"type": "integer", "description": "Page number, 1-based", "default": 1},
				"page_size":    map[string]any{"type": "integer", "description": "Results per page (1–100)", "default": 20},
				"search":       map[string]any{"type": "string", "description": "Search by file name"},
				"storage_type": map[string]any{"type": "string", "enum": []string{"local", "s3"}, "description": "Filter by storage backend"},
			},
		},
	}, a.listFiles, authModel.ScopeFileRead)

	reg.RegisterTool(ToolDefinition{
		Name:        "get_file",
		Title:       "Get File",
		Description: "Get metadata for a single file (name, size, mime type, storage type, URL, timestamps). The id can be used as echo_files[].file_id in create_post/update_post. Does not return file contents.",
		InputSchema: map[string]any{
			"type":     "object",
			"required": []string{"id"},
			"properties": map[string]any{
				"id": map[string]any{"type": "string", "format": "uuid", "description": "File UUID"},
			},
		},
	}, a.getFile, authModel.ScopeFileRead)

	reg.RegisterTool(ToolDefinition{
		Name:        "delete_file",
		Title:       "Delete File",
		Description: "Permanently delete a file from storage. Returns {id, message}. This action cannot be undone.",
		InputSchema: map[string]any{
			"type":     "object",
			"required": []string{"id"},
			"properties": map[string]any{
				"id": map[string]any{"type": "string", "format": "uuid", "description": "File UUID"},
			},
		},
	}, a.deleteFile, authModel.ScopeFileWrite)

	reg.RegisterTool(ToolDefinition{
		Name:  "create_external_file",
		Title: "Create External File",
		Description: "Register an external URL as an Ech0 file record (no upload needed). " +
			"Returns the created file metadata including id, which can be used as echo_files[].file_id in create_post/update_post.",
		InputSchema: map[string]any{
			"type":     "object",
			"required": []string{"url"},
			"properties": map[string]any{
				"url":          map[string]any{"type": "string", "format": "uri", "description": "External file URL (e.g. image hosted on a CDN)"},
				"name":         map[string]any{"type": "string", "description": "Display name for the file"},
				"content_type": map[string]any{"type": "string", "description": "MIME type (e.g. image/png)"},
				"category":     map[string]any{"type": "string", "enum": []string{"image", "video", "audio", "document", "file"}, "description": "File category"},
				"width":        map[string]any{"type": "integer", "description": "Image/video width in pixels"},
				"height":       map[string]any{"type": "integer", "description": "Image/video height in pixels"},
			},
		},
	}, a.createExternalFile, authModel.ScopeFileWrite)
}

func (a *Adapter) listFiles(ctx context.Context, args map[string]any) (*ToolCallResult, error) {
	page := intArg(args, "page", 1)
	pageSize := intArg(args, "page_size", 20)
	if pageSize > 100 {
		pageSize = 100
	}
	search := stringArg(args, "search")
	storageType := stringArg(args, "storage_type")

	result, err := a.fileSvc.ListFiles(ctx, commonModel.FileListQueryDto{
		Page:        page,
		PageSize:    pageSize,
		Search:      search,
		StorageType: storageType,
	})
	if err != nil {
		return nil, err
	}
	return jsonResult(result)
}

func (a *Adapter) getFile(ctx context.Context, args map[string]any) (*ToolCallResult, error) {
	id := stringArg(args, "id")
	if id == "" {
		return textError("id is required"), nil
	}
	file, err := a.fileSvc.GetFileByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return jsonResult(file)
}

func (a *Adapter) createExternalFile(ctx context.Context, args map[string]any) (*ToolCallResult, error) {
	url := stringArg(args, "url")
	if url == "" {
		return textError("url is required"), nil
	}
	dto := commonModel.CreateExternalFileDto{
		URL:         url,
		Name:        stringArg(args, "name"),
		ContentType: stringArg(args, "content_type"),
		Category:    stringArg(args, "category"),
		Width:       intArg(args, "width", 0),
		Height:      intArg(args, "height", 0),
	}
	file, err := a.fileSvc.CreateExternalFile(ctx, dto)
	if err != nil {
		return nil, err
	}
	return jsonResult(file)
}

func (a *Adapter) deleteFile(ctx context.Context, args map[string]any) (*ToolCallResult, error) {
	id := stringArg(args, "id")
	if id == "" {
		return textError("id is required"), nil
	}
	if err := a.fileSvc.DeleteFile(ctx, id); err != nil {
		return nil, err
	}
	return jsonResult(map[string]string{"id": id, "message": "file deleted successfully"})
}
