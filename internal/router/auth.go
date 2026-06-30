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

// setupAuthRoutes 保留**留在裸 gin** 的认证端点：OAuth2 重定向、cookie/token 签发流程、
// WebAuthn 注册/登录仪式。这些依赖 302 重定向、HttpOnly cookie 读写、或 WebAuthn 协议 blob，
// 不是干净的 JSON-REST，强行迁 Huma 既别扭又有安全回归风险，故保留现状。
func setupAuthRoutes(appRouterGroup *AppRouterGroup, h *handler.Bundle) {
	// OAuth2/OIDC 重定向（统一 provider 路由）
	appRouterGroup.ResourceGroup.GET("/oauth/:provider/login", middleware.NoCache(), h.AuthHandler.OAuthLogin())
	appRouterGroup.ResourceGroup.GET("/oauth/:provider/callback", middleware.NoCache(), h.AuthHandler.OAuthCallback())

	// 公开：登录 / WebAuthn 登录仪式 / token 生命周期（均读写 cookie）
	appRouterGroup.PublicRouterGroup.POST("/login", middleware.NoCache(), h.AuthHandler.Login())
	appRouterGroup.PublicRouterGroup.POST("/passkey/login/begin", middleware.NoCache(), h.AuthHandler.PasskeyLoginBeginV2())
	appRouterGroup.PublicRouterGroup.POST("/passkey/login/finish", middleware.NoCache(), h.AuthHandler.PasskeyLoginFinishV2())
	appRouterGroup.PublicRouterGroup.POST("/auth/refresh", middleware.NoCache(), h.AuthHandler.Refresh())
	appRouterGroup.PublicRouterGroup.POST("/auth/logout", middleware.NoCache(), h.AuthHandler.Logout())
	appRouterGroup.PublicRouterGroup.POST("/auth/exchange", middleware.NoCache(), h.AuthHandler.Exchange())

	// 鉴权：WebAuthn 注册仪式（profile:write）
	appRouterGroup.AuthRouterGroup.POST(
		"/passkey/register/begin",
		middleware.RequireScopes(authModel.ScopeProfileWrite),
		h.AuthHandler.PasskeyRegisterBeginV2(),
	)
	appRouterGroup.AuthRouterGroup.POST(
		"/passkey/register/finish",
		middleware.RequireScopes(authModel.ScopeProfileWrite),
		h.AuthHandler.PasskeyRegisterFinishV2(),
	)
}

// registerAuthHuma 注册**干净 JSON** 的认证端点（无 cookie / 无重定向 / 无 WebAuthn blob）。
func registerAuthHuma(api huma.API, h *handler.Bundle, revoker authService.TokenRevoker) {
	write := humares.Secured(authModel.ScopeProfileWrite)
	writeMW := securedMW(revoker, authModel.ScopeProfileWrite)
	read := humares.Secured(authModel.ScopeProfileRead)
	readMW := securedMW(revoker, authModel.ScopeProfileRead)

	reg(api, huma.Operation{
		OperationID: "oauth-bind",
		Method:      http.MethodPost,
		Path:        "/oauth/{provider}/bind",
		Summary:     "绑定 OAuth2 账号到当前用户",
		Tags:        []string{"Auth"},
		Security:    write,
		Middlewares: writeMW,
	}, h.AuthHandler.OAuthBind)

	reg(api, huma.Operation{
		OperationID: "oauth-info",
		Method:      http.MethodGet,
		Path:        "/oauth/info",
		Summary:     "获取当前用户的 OAuth2 绑定信息",
		Tags:        []string{"Auth"},
		Security:    read,
		Middlewares: readMW,
	}, h.AuthHandler.GetOAuthInfo)

	reg(api, huma.Operation{
		OperationID: "passkey-list",
		Method:      http.MethodGet,
		Path:        "/passkeys",
		Summary:     "列出当前用户的 Passkey 设备",
		Tags:        []string{"Auth"},
		Security:    read,
		Middlewares: readMW,
	}, h.AuthHandler.ListPasskeys)

	reg(api, huma.Operation{
		OperationID: "passkey-delete",
		Method:      http.MethodDelete,
		Path:        "/passkeys/{id}",
		Summary:     "删除 Passkey 设备",
		Tags:        []string{"Auth"},
		Security:    write,
		Middlewares: writeMW,
	}, h.AuthHandler.DeletePasskey)

	reg(api, huma.Operation{
		OperationID: "passkey-update-name",
		Method:      http.MethodPut,
		Path:        "/passkeys/{id}",
		Summary:     "更新 Passkey 设备名称",
		Tags:        []string{"Auth"},
		Security:    write,
		Middlewares: writeMW,
	}, h.AuthHandler.UpdatePasskeyDeviceName)
}
