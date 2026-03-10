package router

import "github.com/lin-snow/ech0/internal/handler"

func setupMigrationRoutes(appRouterGroup *AppRouterGroup, h *handler.Bundle) {
	appRouterGroup.AuthRouterGroup.POST("/migration/jobs", h.MigrationHandler.CreateJob())
	appRouterGroup.AuthRouterGroup.GET("/migration/jobs/:id", h.MigrationHandler.GetJob())
	appRouterGroup.AuthRouterGroup.POST("/migration/jobs/:id/cancel", h.MigrationHandler.CancelJob())
	appRouterGroup.AuthRouterGroup.POST("/migration/jobs/:id/retry-failed", h.MigrationHandler.RetryFailed())
}
