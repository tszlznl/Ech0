// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	logUtil "github.com/lin-snow/ech0/internal/util/log"
	versionPkg "github.com/lin-snow/ech0/internal/version"
	"github.com/lin-snow/ech0/pkg/viewer"
	"go.uber.org/zap"
)

const toolTimeout = 10 * time.Second

type ctxKey int

const (
	ctxKeyRawToken ctxKey = iota
	ctxKeyBaseURL
)

func RawTokenFromContext(ctx context.Context) string {
	v, _ := ctx.Value(ctxKeyRawToken).(string)
	return v
}

func BaseURLFromContext(ctx context.Context) string {
	v, _ := ctx.Value(ctxKeyBaseURL).(string)
	return v
}

type Server struct {
	registry *Registry
}

func NewServer(registry *Registry) *Server {
	return &Server{registry: registry}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		s.handlePost(w, r)
	case http.MethodGet:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"status":  "ok",
			"name":    ServerName,
			"version": versionPkg.Version,
		})
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handlePost(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	body, err := io.ReadAll(io.LimitReader(r.Body, 256*1024))
	if err != nil {
		writeRPCError(w, nil, ErrCodeParse, "failed to read request body")
		return
	}

	var req Request
	if err := json.Unmarshal(body, &req); err != nil {
		writeRPCError(w, nil, ErrCodeParse, "invalid JSON")
		return
	}
	if req.JSONRPC != "2.0" {
		writeRPCError(w, req.ID, ErrCodeInvalidRequest, "jsonrpc must be 2.0")
		return
	}

	ctx := r.Context()
	if auth := r.Header.Get("Authorization"); strings.HasPrefix(auth, "Bearer ") {
		ctx = context.WithValue(ctx, ctxKeyRawToken, strings.TrimPrefix(auth, "Bearer "))
	}
	scheme := "http"
	if r.TLS != nil || strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https") {
		scheme = "https"
	}
	ctx = context.WithValue(ctx, ctxKeyBaseURL, scheme+"://"+r.Host)
	r = r.WithContext(ctx)

	v := viewer.MustFromContext(ctx)

	result, rpcErr := s.dispatch(r, &req, v)

	logUtil.GetLogger().Info("mcp_request",
		zap.String("method", req.Method),
		zap.String("user_id", v.UserID()),
		zap.String("token_id", v.TokenID()),
		zap.Duration("latency", time.Since(start)),
		zap.Bool("error", rpcErr != nil),
	)

	if rpcErr != nil {
		writeRPCError(w, req.ID, rpcErr.Code, rpcErr.Message)
		return
	}
	writeRPCResult(w, req.ID, result)
}

func (s *Server) dispatch(r *http.Request, req *Request, v viewer.Context) (any, *RPCError) {
	switch req.Method {
	case "initialize":
		return s.handleInitialize()
	case "notifications/initialized":
		return map[string]any{}, nil
	case "tools/list":
		return s.handleToolsList()
	case "tools/call":
		return s.handleToolsCall(r, req, v)
	case "resources/list":
		return s.handleResourcesList()
	case "resources/read":
		return s.handleResourcesRead(r, req, v)
	default:
		return nil, &RPCError{Code: ErrCodeMethodNotFound, Message: fmt.Sprintf("method %q not found", req.Method)}
	}
}

func (s *Server) handleInitialize() (*InitializeResult, *RPCError) {
	return &InitializeResult{
		ProtocolVersion: ProtocolVersion,
		Capabilities: ServerCapabilities{
			Tools:     &ToolsCapability{ListChanged: false},
			Resources: &ResourcesCapability{Subscribe: false, ListChanged: false},
		},
		ServerInfo: ServerInfo{
			Name:    ServerName,
			Version: versionPkg.Version,
		},
	}, nil
}

func (s *Server) handleToolsList() (*ToolsListResult, *RPCError) {
	return &ToolsListResult{Tools: s.registry.ToolDefinitions()}, nil
}

func (s *Server) handleToolsCall(r *http.Request, req *Request, v viewer.Context) (*ToolCallResult, *RPCError) {
	var params ToolCallParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return nil, &RPCError{Code: ErrCodeInvalidParams, Message: "invalid tool call params"}
	}

	handler, requiredScopes, ok := s.registry.LookupTool(params.Name)
	if !ok {
		return nil, &RPCError{Code: ErrCodeInvalidParams, Message: fmt.Sprintf("tool %q not found", params.Name)}
	}

	if !checkScopes(v.Scopes(), requiredScopes) {
		return &ToolCallResult{
			Content: []ContentItem{{Type: "text", Text: "permission denied: insufficient scopes"}},
			IsError: true,
		}, nil
	}

	ctx, cancel := context.WithTimeout(r.Context(), toolTimeout)
	defer cancel()

	result, err := handler(ctx, params.Arguments)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return &ToolCallResult{
				Content: []ContentItem{{Type: "text", Text: "tool execution timed out"}},
				IsError: true,
			}, nil
		}
		return &ToolCallResult{
			Content: []ContentItem{{Type: "text", Text: err.Error()}},
			IsError: true,
		}, nil
	}
	return result, nil
}

func (s *Server) handleResourcesList() (*ResourcesListResult, *RPCError) {
	return &ResourcesListResult{Resources: s.registry.ResourceDefinitions()}, nil
}

func (s *Server) handleResourcesRead(r *http.Request, req *Request, v viewer.Context) (*ResourceReadResult, *RPCError) {
	var params ResourceReadParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return nil, &RPCError{Code: ErrCodeInvalidParams, Message: "invalid resource read params"}
	}

	handler, requiredScopes, ok := s.registry.LookupResource(params.URI)
	if !ok {
		return nil, &RPCError{Code: ErrCodeInvalidParams, Message: fmt.Sprintf("resource %q not found", params.URI)}
	}

	if !checkScopes(v.Scopes(), requiredScopes) {
		return nil, &RPCError{Code: ErrCodeInternal, Message: "permission denied: insufficient scopes"}
	}

	result, err := handler(r.Context(), params.URI)
	if err != nil {
		return nil, &RPCError{Code: ErrCodeInternal, Message: err.Error()}
	}
	return result, nil
}

func checkScopes(actual, required []string) bool {
	if len(required) == 0 {
		return true
	}
	set := make(map[string]struct{}, len(actual))
	for _, s := range actual {
		set[s] = struct{}{}
	}
	for _, s := range required {
		if _, ok := set[s]; !ok {
			return false
		}
	}
	return true
}

func writeRPCResult(w http.ResponseWriter, id json.RawMessage, result any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(Response{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	})
}

func writeRPCError(w http.ResponseWriter, id json.RawMessage, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(Response{
		JSONRPC: "2.0",
		ID:      id,
		Error:   &RPCError{Code: code, Message: message},
	})
}
