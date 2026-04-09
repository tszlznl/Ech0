package router

import (
	"github.com/lin-snow/ech0/internal/handler"
	"github.com/lin-snow/ech0/internal/middleware"
	authModel "github.com/lin-snow/ech0/internal/model/auth"
)

// setupInboxRoutes 配置收件箱相关路由
func setupInboxRoutes(appRouterGroup *AppRouterGroup, h *handler.Bundle) {
	appRouterGroup.AuthRouterGroup.GET(
		"/inbox",
		middleware.RequireScopes(authModel.ScopeAdminSettings),
		h.InboxHandler.GetInboxList(),
	)
	appRouterGroup.AuthRouterGroup.GET(
		"/inbox/unread",
		middleware.RequireScopes(authModel.ScopeAdminSettings),
		h.InboxHandler.GetUnreadInbox(),
	)
	appRouterGroup.AuthRouterGroup.PUT(
		"/inbox/:id/read",
		middleware.RequireScopes(authModel.ScopeAdminSettings),
		h.InboxHandler.MarkInboxAsRead(),
	)
	appRouterGroup.AuthRouterGroup.DELETE(
		"/inbox/:id",
		middleware.RequireScopes(authModel.ScopeAdminSettings),
		h.InboxHandler.DeleteInbox(),
	)
	appRouterGroup.AuthRouterGroup.DELETE(
		"/inbox",
		middleware.RequireScopes(authModel.ScopeAdminSettings),
		h.InboxHandler.ClearInbox(),
	)
}
