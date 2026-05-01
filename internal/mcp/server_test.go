// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lin-snow/ech0/pkg/viewer"
)

func testViewer() viewer.Context {
	return viewer.NewUserViewerWithToken("test-user", "access", []string{"echo:read", "echo:write", "profile:read"}, []string{"mcp-remote"}, "test-jti")
}

func testRequest(t *testing.T, method, body string) *http.Request {
	t.Helper()
	req := httptest.NewRequest(method, "/mcp", bytes.NewBufferString(body))
	req = req.WithContext(viewer.WithContext(req.Context(), testViewer()))
	return req
}

func setupTestServer() *Server {
	reg := NewRegistry()
	reg.RegisterTool(ToolDefinition{
		Name:        "echo_tool",
		Description: "test tool",
		InputSchema: map[string]any{"type": "object", "properties": map[string]any{}},
	}, func(_ context.Context, _ map[string]any) (*ToolCallResult, error) {
		return &ToolCallResult{Content: []ContentItem{{Type: "text", Text: "hello"}}}, nil
	}, "echo:read")

	reg.RegisterTool(ToolDefinition{
		Name:        "write_tool",
		Description: "test write tool",
		InputSchema: map[string]any{"type": "object", "properties": map[string]any{}},
	}, func(_ context.Context, _ map[string]any) (*ToolCallResult, error) {
		return &ToolCallResult{Content: []ContentItem{{Type: "text", Text: "written"}}}, nil
	}, "admin:settings")

	reg.RegisterResource(ResourceDefinition{
		URI:      "ech0://test",
		Name:     "test",
		MimeType: "text/plain",
	}, func(_ context.Context, _ string) (*ResourceReadResult, error) {
		return &ResourceReadResult{
			Contents: []ResourceContent{{URI: "ech0://test", MimeType: "text/plain", Text: "test data"}},
		}, nil
	}, "echo:read")

	reg.RegisterResource(ResourceDefinition{
		URI:      "ech0://items/{id}",
		Name:     "item",
		MimeType: "application/json",
	}, func(_ context.Context, uri string) (*ResourceReadResult, error) {
		return &ResourceReadResult{
			Contents: []ResourceContent{{URI: uri, MimeType: "application/json", Text: `{"uri":"` + uri + `"}`}},
		}, nil
	}, "echo:read")

	return NewServer(reg)
}

func doRPC(t *testing.T, srv *Server, method string, params any) Response {
	t.Helper()
	var paramsJSON json.RawMessage
	if params != nil {
		b, _ := json.Marshal(params)
		paramsJSON = b
	}
	body, _ := json.Marshal(Request{JSONRPC: "2.0", ID: json.RawMessage(`1`), Method: method, Params: paramsJSON})
	req := testRequest(t, http.MethodPost, string(body))
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	var resp Response
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	return resp
}

func TestInitialize(t *testing.T) {
	srv := setupTestServer()
	resp := doRPC(t, srv, "initialize", nil)
	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}
	b, _ := json.Marshal(resp.Result)
	var result InitializeResult
	if err := json.Unmarshal(b, &result); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}
	if result.ProtocolVersion != ProtocolVersion {
		t.Errorf("protocol version = %q, want %q", result.ProtocolVersion, ProtocolVersion)
	}
	if result.ServerInfo.Name != ServerName {
		t.Errorf("server name = %q, want %q", result.ServerInfo.Name, ServerName)
	}
}

func TestToolsList(t *testing.T) {
	srv := setupTestServer()
	resp := doRPC(t, srv, "tools/list", nil)
	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}
	b, _ := json.Marshal(resp.Result)
	var result ToolsListResult
	if err := json.Unmarshal(b, &result); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}
	if len(result.Tools) != 2 {
		t.Errorf("tools count = %d, want 2", len(result.Tools))
	}
}

func TestToolsCallSuccess(t *testing.T) {
	srv := setupTestServer()
	resp := doRPC(t, srv, "tools/call", ToolCallParams{Name: "echo_tool", Arguments: map[string]any{}})
	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}
	b, _ := json.Marshal(resp.Result)
	var result ToolCallResult
	if err := json.Unmarshal(b, &result); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}
	if result.IsError {
		t.Error("expected success but got isError=true")
	}
	if len(result.Content) == 0 || result.Content[0].Text != "hello" {
		t.Errorf("unexpected content: %v", result.Content)
	}
}

func TestToolsCallInsufficientScopes(t *testing.T) {
	srv := setupTestServer()
	resp := doRPC(t, srv, "tools/call", ToolCallParams{Name: "write_tool", Arguments: map[string]any{}})
	if resp.Error != nil {
		t.Fatalf("unexpected rpc error: %v", resp.Error)
	}
	b, _ := json.Marshal(resp.Result)
	var result ToolCallResult
	if err := json.Unmarshal(b, &result); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}
	if !result.IsError {
		t.Error("expected isError=true for insufficient scopes")
	}
}

func TestToolsCallNotFound(t *testing.T) {
	srv := setupTestServer()
	resp := doRPC(t, srv, "tools/call", ToolCallParams{Name: "nonexistent", Arguments: map[string]any{}})
	if resp.Error == nil {
		t.Fatal("expected error for nonexistent tool")
	}
	if resp.Error.Code != ErrCodeInvalidParams {
		t.Errorf("error code = %d, want %d", resp.Error.Code, ErrCodeInvalidParams)
	}
}

func TestResourcesList(t *testing.T) {
	srv := setupTestServer()
	resp := doRPC(t, srv, "resources/list", nil)
	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}
	b, _ := json.Marshal(resp.Result)
	var result ResourcesListResult
	if err := json.Unmarshal(b, &result); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}
	if len(result.Resources) != 2 {
		t.Errorf("resources count = %d, want 2", len(result.Resources))
	}
}

func TestResourcesReadSuccess(t *testing.T) {
	srv := setupTestServer()
	resp := doRPC(t, srv, "resources/read", ResourceReadParams{URI: "ech0://test"})
	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}
	b, _ := json.Marshal(resp.Result)
	var result ResourceReadResult
	if err := json.Unmarshal(b, &result); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}
	if len(result.Contents) == 0 || result.Contents[0].Text != "test data" {
		t.Errorf("unexpected content: %v", result.Contents)
	}
}

func TestResourcesReadPrefixMatch(t *testing.T) {
	srv := setupTestServer()
	resp := doRPC(t, srv, "resources/read", ResourceReadParams{URI: "ech0://items/abc-123"})
	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}
	b, _ := json.Marshal(resp.Result)
	var result ResourceReadResult
	if err := json.Unmarshal(b, &result); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}
	if len(result.Contents) == 0 {
		t.Fatal("expected content from prefix-matched resource")
	}
	if result.Contents[0].URI != "ech0://items/abc-123" {
		t.Errorf("URI = %q, want %q", result.Contents[0].URI, "ech0://items/abc-123")
	}
}

func TestResourcesReadNotFound(t *testing.T) {
	srv := setupTestServer()
	resp := doRPC(t, srv, "resources/read", ResourceReadParams{URI: "ech0://missing"})
	if resp.Error == nil {
		t.Fatal("expected error for nonexistent resource")
	}
}

func TestMethodNotFound(t *testing.T) {
	srv := setupTestServer()
	resp := doRPC(t, srv, "unknown/method", nil)
	if resp.Error == nil {
		t.Fatal("expected error for unknown method")
	}
	if resp.Error.Code != ErrCodeMethodNotFound {
		t.Errorf("error code = %d, want %d", resp.Error.Code, ErrCodeMethodNotFound)
	}
}

func TestGetEndpoint(t *testing.T) {
	srv := setupTestServer()
	req := httptest.NewRequest(http.MethodGet, "/mcp", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("GET status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestInvalidJSON(t *testing.T) {
	srv := setupTestServer()
	req := testRequest(t, http.MethodPost, "not json")
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	var resp Response
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.Error == nil || resp.Error.Code != ErrCodeParse {
		t.Errorf("expected parse error, got %v", resp.Error)
	}
}
