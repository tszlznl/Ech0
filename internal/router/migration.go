// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package router

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/lin-snow/ech0/internal/handler"
	"github.com/lin-snow/ech0/internal/handler/humares"
	"github.com/lin-snow/ech0/internal/middleware"
	authModel "github.com/lin-snow/ech0/internal/model/auth"
	authService "github.com/lin-snow/ech0/internal/service/auth"
)

// setupMigrationRoutes 仅保留非 JSON 端点走裸 gin：multipart 上传源 zip + 二进制快照下载。
func setupMigrationRoutes(appRouterGroup *AppRouterGroup, h *handler.Bundle) {
	appRouterGroup.AuthRouterGroup.POST(
		"/migration/upload",
		middleware.RequireScopes(authModel.ScopeAdminSettings),
		h.MigrationHandler.UploadSourceZip(),
	)
	appRouterGroup.AuthRouterGroup.GET(
		"/migration/export/download",
		middleware.RequireScopes(authModel.ScopeAdminSettings),
		h.MigrationHandler.DownloadExport(),
	)
}

// registerMigrationHuma 注册数据迁移控制面的 JSON 端点（admin:settings）。
func registerMigrationHuma(api huma.API, h *handler.Bundle, revoker authService.TokenRevoker) {
	sec := humares.Secured(authModel.ScopeAdminSettings)
	mw := securedMW(revoker, authModel.ScopeAdminSettings)
	meta := func(id, summary string) huma.Operation {
		return huma.Operation{OperationID: id, Summary: summary, Tags: []string{"Migration"}, Security: sec, Middlewares: mw}
	}
	register := func(id, method, path, summary string) huma.Operation {
		o := meta(id, summary)
		o.Method = method
		o.Path = path
		return o
	}

	huma.Register(api, register("migration-start", http.MethodPost, "/migration/start", "启动全局数据迁移"), h.MigrationHandler.StartMigration)
	huma.Register(api, register("migration-status", http.MethodGet, "/migration/status", "查询全局迁移状态"), h.MigrationHandler.GetMigrationStatus)
	huma.Register(api, register("migration-cancel", http.MethodPost, "/migration/cancel", "取消进行中的全局迁移"), h.MigrationHandler.CancelMigration)
	huma.Register(api, register("migration-cleanup", http.MethodPost, "/migration/cleanup", "清理迁移中间产物"), h.MigrationHandler.CleanupMigration)
	huma.Register(api, register("migration-export", http.MethodPost, "/migration/export", "提交导出作业"), h.MigrationHandler.StartExport)
	huma.Register(api, register("migration-export-status", http.MethodGet, "/migration/export/status", "查询导出作业状态"), h.MigrationHandler.GetExportStatus)
	huma.Register(api, register("migration-export-cancel", http.MethodPost, "/migration/export/cancel", "取消导出作业"), h.MigrationHandler.CancelExport)
}
