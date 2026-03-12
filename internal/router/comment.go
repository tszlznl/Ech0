package router

import "github.com/lin-snow/ech0/internal/handler"

func setupCommentRoutes(appRouterGroup *AppRouterGroup, h *handler.Bundle) {
	// Public
	appRouterGroup.PublicRouterGroup.GET("/comments/form", h.CommentHandler.GetFormMeta())
	appRouterGroup.PublicRouterGroup.GET("/comments", h.CommentHandler.ListCommentsByEchoID())
	appRouterGroup.PublicRouterGroup.POST("/comments", h.CommentHandler.CreateComment())

	// Admin Panel
	appRouterGroup.AuthRouterGroup.GET("/panel/comments", h.CommentHandler.ListPanelComments())
	appRouterGroup.AuthRouterGroup.GET("/panel/comments/:id", h.CommentHandler.GetCommentByID())
	appRouterGroup.AuthRouterGroup.PATCH("/panel/comments/:id/status", h.CommentHandler.UpdateCommentStatus())
	appRouterGroup.AuthRouterGroup.PATCH("/panel/comments/:id/hot", h.CommentHandler.UpdateCommentHot())
	appRouterGroup.AuthRouterGroup.DELETE("/panel/comments/:id", h.CommentHandler.DeleteComment())
	appRouterGroup.AuthRouterGroup.POST("/panel/comments/batch", h.CommentHandler.BatchAction())
	appRouterGroup.AuthRouterGroup.GET("/panel/comments/settings", h.CommentHandler.GetCommentSetting())
	appRouterGroup.AuthRouterGroup.PUT("/panel/comments/settings", h.CommentHandler.UpdateCommentSetting())
}
