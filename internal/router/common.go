package router

import "github.com/lin-snow/ech0/internal/handler"

// setupCommonRoutes 设置普通路由
func setupCommonRoutes(appRouterGroup *AppRouterGroup, h *handler.Bundle) {
	// Public
	appRouterGroup.PublicRouterGroup.GET("/status", h.CommonHandler.GetStatus())
	appRouterGroup.PublicRouterGroup.GET("/heatmap", h.CommonHandler.GetHeatMap())
	appRouterGroup.PublicRouterGroup.GET("/audio/current", h.CommonHandler.GetCurrentAudio())
	appRouterGroup.PublicRouterGroup.GET("/audio/stream", h.CommonHandler.StreamCurrentAudio)
	appRouterGroup.PublicRouterGroup.GET("/hello", h.CommonHandler.HelloEch0())
	appRouterGroup.PublicRouterGroup.GET("/backup/export", h.BackupHandler.ExportBackup())
	appRouterGroup.PublicRouterGroup.GET("/website/title", h.CommonHandler.GetWebsiteTitle())

	// Auth
	appRouterGroup.AuthRouterGroup.POST("/files/upload", h.CommonHandler.UploadFile())
	appRouterGroup.AuthRouterGroup.POST("/files/external", h.CommonHandler.CreateExternalFile())
	appRouterGroup.AuthRouterGroup.DELETE("/files/delete", h.CommonHandler.DeleteFile())
	appRouterGroup.AuthRouterGroup.POST("/files/audio/upload", h.CommonHandler.UploadAudioFile())
	appRouterGroup.AuthRouterGroup.DELETE("/files/audio", h.CommonHandler.DeleteAudioFile())
	appRouterGroup.AuthRouterGroup.PUT("/files/presign", h.CommonHandler.GetFilePresignURL())
	appRouterGroup.AuthRouterGroup.GET("/backup", h.BackupHandler.Backup())
	appRouterGroup.AuthRouterGroup.POST("/backup/import", h.BackupHandler.ImportBackup())
}
