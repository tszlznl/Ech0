package router

import "github.com/lin-snow/ech0/internal/handler"

func setupFileRoutes(appRouterGroup *AppRouterGroup, h *handler.Bundle) {
	// Public
	appRouterGroup.PublicRouterGroup.GET("/audio/current", h.FileHandler.GetCurrentAudio())
	appRouterGroup.PublicRouterGroup.GET("/audio/stream", h.FileHandler.StreamCurrentAudio)

	// Auth
	appRouterGroup.AuthRouterGroup.POST("/files/upload", h.FileHandler.UploadFile())
	appRouterGroup.AuthRouterGroup.POST("/files/external", h.FileHandler.CreateExternalFile())
	appRouterGroup.AuthRouterGroup.DELETE("/files/delete", h.FileHandler.DeleteFile())
	appRouterGroup.AuthRouterGroup.POST("/files/audio/upload", h.FileHandler.UploadAudioFile())
	appRouterGroup.AuthRouterGroup.DELETE("/files/audio", h.FileHandler.DeleteAudioFile())
	appRouterGroup.AuthRouterGroup.PUT("/files/presign", h.FileHandler.GetFilePresignURL())
}
