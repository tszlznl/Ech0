package router

import "github.com/lin-snow/ech0/internal/handler"

func setupMigrationRoutes(appRouterGroup *AppRouterGroup, h *handler.Bundle) {
	appRouterGroup.AuthRouterGroup.POST("/migration/upload", h.MigrationHandler.UploadSourceZip())
	appRouterGroup.AuthRouterGroup.POST("/migration/start", h.MigrationHandler.StartMigration())
	appRouterGroup.AuthRouterGroup.GET("/migration/status", h.MigrationHandler.GetMigrationStatus())
	appRouterGroup.AuthRouterGroup.POST("/migration/cancel", h.MigrationHandler.CancelMigration())
	appRouterGroup.AuthRouterGroup.POST("/migration/cleanup", h.MigrationHandler.CleanupMigration())
}
