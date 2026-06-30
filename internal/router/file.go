// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package router

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/lin-snow/ech0/internal/handler"
	"github.com/lin-snow/ech0/internal/middleware"
	authModel "github.com/lin-snow/ech0/internal/model/auth"
	authService "github.com/lin-snow/ech0/internal/service/auth"
)

// setupFileRoutes 仅保留非 JSON 端点走裸 gin：二进制流式下载 + multipart 上传。
// JSON 端点（列表/树/元信息/删除/外链/预签名）由 registerFileHuma 注册。
func setupFileRoutes(appRouterGroup *AppRouterGroup, h *handler.Bundle) {
	appRouterGroup.AuthRouterGroup.GET(
		"/file/stream",
		middleware.RequireScopes(authModel.ScopeFileRead),
		h.FileHandler.StreamFileByPath,
	)
	appRouterGroup.AuthRouterGroup.GET(
		"/file/:id/stream",
		middleware.RequireScopes(authModel.ScopeFileRead),
		h.FileHandler.StreamFileByID,
	)
	appRouterGroup.AuthRouterGroup.POST(
		"/files/upload",
		middleware.RequireScopes(authModel.ScopeFileWrite),
		h.FileHandler.UploadFile(),
	)
}

// registerFileHuma 注册文件的 JSON 端点。
func registerFileHuma(api huma.API, h *handler.Bundle, revoker authService.TokenRevoker) {
	register(api, secured(revoker, authModel.ScopeFileRead), huma.Operation{
		OperationID: "file-list",
		Method:      http.MethodGet,
		Path:        "/files",
		Summary:     "分页获取文件列表",
		Tags:        []string{"File"},
	}, h.FileHandler.ListFiles)

	register(api, secured(revoker, authModel.ScopeFileRead), huma.Operation{
		OperationID: "file-tree",
		Method:      http.MethodGet,
		Path:        "/file/tree",
		Summary:     "获取文件树",
		Tags:        []string{"File"},
	}, h.FileHandler.ListFileTree)

	register(api, secured(revoker, authModel.ScopeFileRead), huma.Operation{
		OperationID: "file-get",
		Method:      http.MethodGet,
		Path:        "/file/{id}",
		Summary:     "获取文件元信息",
		Tags:        []string{"File"},
	}, h.FileHandler.GetFileByID)

	register(api, secured(revoker, authModel.ScopeFileWrite), huma.Operation{
		OperationID: "file-update-meta",
		Method:      http.MethodPut,
		Path:        "/file/{id}/meta",
		Summary:     "更新对象存储文件元信息",
		Tags:        []string{"File"},
	}, h.FileHandler.UpdateFileMeta)

	register(api, secured(revoker, authModel.ScopeFileWrite), huma.Operation{
		OperationID: "file-external",
		Method:      http.MethodPost,
		Path:        "/files/external",
		Summary:     "登记外链文件",
		Tags:        []string{"File"},
	}, h.FileHandler.CreateExternalFile)

	register(api, secured(revoker, authModel.ScopeFileWrite), huma.Operation{
		OperationID: "file-delete",
		Method:      http.MethodDelete,
		Path:        "/file/{id}",
		Summary:     "删除文件",
		Tags:        []string{"File"},
	}, h.FileHandler.DeleteFile)

	register(api, secured(revoker, authModel.ScopeFileWrite), huma.Operation{
		OperationID: "file-presign",
		Method:      http.MethodPut,
		Path:        "/files/presign",
		Summary:     "获取对象存储直传预签名 URL",
		Tags:        []string{"File"},
	}, h.FileHandler.GetFilePresignURL)
}
