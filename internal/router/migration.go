// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package router

import (
	"github.com/lin-snow/ech0/internal/handler"
	"github.com/lin-snow/ech0/internal/middleware"
	authModel "github.com/lin-snow/ech0/internal/model/auth"
)

func setupMigrationRoutes(appRouterGroup *AppRouterGroup, h *handler.Bundle) {
	appRouterGroup.AuthRouterGroup.POST(
		"/migration/upload",
		middleware.RequireScopes(authModel.ScopeAdminSettings),
		h.MigrationHandler.UploadSourceZip(),
	)
	appRouterGroup.AuthRouterGroup.POST(
		"/migration/start",
		middleware.RequireScopes(authModel.ScopeAdminSettings),
		h.MigrationHandler.StartMigration(),
	)
	appRouterGroup.AuthRouterGroup.GET(
		"/migration/status",
		middleware.RequireScopes(authModel.ScopeAdminSettings),
		h.MigrationHandler.GetMigrationStatus(),
	)
	appRouterGroup.AuthRouterGroup.POST(
		"/migration/cancel",
		middleware.RequireScopes(authModel.ScopeAdminSettings),
		h.MigrationHandler.CancelMigration(),
	)
	appRouterGroup.AuthRouterGroup.POST(
		"/migration/cleanup",
		middleware.RequireScopes(authModel.ScopeAdminSettings),
		h.MigrationHandler.CleanupMigration(),
	)
	// 导出（手动快照异步出口）：与导入对称，统一收敛到 Migrator 域。
	appRouterGroup.AuthRouterGroup.POST(
		"/migration/export",
		middleware.RequireScopes(authModel.ScopeAdminSettings),
		h.MigrationHandler.StartExport(),
	)
	appRouterGroup.AuthRouterGroup.GET(
		"/migration/export/status",
		middleware.RequireScopes(authModel.ScopeAdminSettings),
		h.MigrationHandler.GetExportStatus(),
	)
	appRouterGroup.AuthRouterGroup.POST(
		"/migration/export/cancel",
		middleware.RequireScopes(authModel.ScopeAdminSettings),
		h.MigrationHandler.CancelExport(),
	)
	appRouterGroup.AuthRouterGroup.GET(
		"/migration/export/download",
		middleware.RequireScopes(authModel.ScopeAdminSettings),
		h.MigrationHandler.DownloadExport(),
	)
}
