package router

import "github.com/lin-snow/ech0/internal/handler"

func setupFileRoutes(appRouterGroup *AppRouterGroup, h *handler.Bundle) {
	// Auth
	appRouterGroup.AuthRouterGroup.POST("/files/upload", h.FileHandler.UploadFile())
	appRouterGroup.AuthRouterGroup.GET("/files", h.FileHandler.ListFiles())
	appRouterGroup.AuthRouterGroup.GET("/file/tree", h.FileHandler.ListFileTree())
	appRouterGroup.AuthRouterGroup.GET("/file/:id", h.FileHandler.GetFileByID())
	appRouterGroup.AuthRouterGroup.GET("/file/:id/stream", h.FileHandler.StreamFileByID)
	appRouterGroup.AuthRouterGroup.PUT("/file/:id/meta", h.FileHandler.UpdateFileMeta())
	appRouterGroup.AuthRouterGroup.POST("/files/external", h.FileHandler.CreateExternalFile())
	appRouterGroup.AuthRouterGroup.DELETE("/file/:id", h.FileHandler.DeleteFile())
	appRouterGroup.AuthRouterGroup.PUT("/files/presign", h.FileHandler.GetFilePresignURL())
}
