package router

import "github.com/lin-snow/ech0/internal/handler"

// setupUserRoutes 设置用户路由
func setupUserRoutes(appRouterGroup *AppRouterGroup, h *handler.Bundle) {
	// OAuth2/OIDC (统一 provider 路由)
	appRouterGroup.ResourceGroup.GET("/oauth/:provider/login", h.UserHandler.OAuthLogin())
	appRouterGroup.ResourceGroup.GET("/oauth/:provider/callback", h.UserHandler.OAuthCallback())

	// Public
	appRouterGroup.PublicRouterGroup.POST("/login", h.UserHandler.Login())
	appRouterGroup.PublicRouterGroup.POST("/register", h.UserHandler.Register())
	appRouterGroup.PublicRouterGroup.GET("/allusers", h.UserHandler.GetAllUsers())
	appRouterGroup.PublicRouterGroup.POST("/passkey/login/begin", h.UserHandler.PasskeyLoginBeginV2())
	appRouterGroup.PublicRouterGroup.POST(
		"/passkey/login/finish",
		h.UserHandler.PasskeyLoginFinishV2(),
	)

	// Auth
	appRouterGroup.AuthRouterGroup.GET("/user", h.UserHandler.GetUserInfo())
	appRouterGroup.AuthRouterGroup.PUT("/user", h.UserHandler.UpdateUser())
	appRouterGroup.AuthRouterGroup.DELETE("/user/:id", h.UserHandler.DeleteUser())
	appRouterGroup.AuthRouterGroup.PUT("/user/admin/:id", h.UserHandler.UpdateUserAdmin())
	appRouterGroup.AuthRouterGroup.POST("/oauth/:provider/bind", h.UserHandler.OAuthBind())
	appRouterGroup.AuthRouterGroup.GET("/oauth/info", h.UserHandler.GetOAuthInfo())
	appRouterGroup.AuthRouterGroup.POST(
		"/passkey/register/begin",
		h.UserHandler.PasskeyRegisterBeginV2(),
	)
	appRouterGroup.AuthRouterGroup.POST(
		"/passkey/register/finish",
		h.UserHandler.PasskeyRegisterFinishV2(),
	)
	appRouterGroup.AuthRouterGroup.GET("/passkeys", h.UserHandler.ListPasskeys())
	appRouterGroup.AuthRouterGroup.DELETE("/passkeys/:id", h.UserHandler.DeletePasskey())
	appRouterGroup.AuthRouterGroup.PUT("/passkeys/:id", h.UserHandler.UpdatePasskeyDeviceName())
}
