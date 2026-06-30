// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package router

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/lin-snow/ech0/internal/handler"
	authService "github.com/lin-snow/ech0/internal/service/auth"
)

// registerCommon 注册通用 JSON 路由。
// 注意：RSS / robots.txt / sitemap.xml / healthz 是非 JSON（XML/纯文本）输出，
// 仍由 setupResourceRoutes 走裸 gin，不在此迁移。
func registerCommon(api huma.API, h *handler.Bundle, revoker authService.TokenRevoker) {
	route(api, public(), huma.Operation{
		OperationID: "common-heatmap",
		Method:      http.MethodGet,
		Path:        "/heatmap",
		Summary:     "获取发布热力图",
		Tags:        []string{"Common"},
	}, h.CommonHandler.GetHeatMap)

	route(api, public(), huma.Operation{
		OperationID: "common-hello",
		Method:      http.MethodGet,
		Path:        "/hello",
		Summary:     "Hello / 版本信息",
		Tags:        []string{"Common"},
	}, h.CommonHandler.HelloEch0)

	route(api, secured(revoker), huma.Operation{
		OperationID: "common-website-title",
		Method:      http.MethodGet,
		Path:        "/website/title",
		Summary:     "获取目标网站标题",
		Tags:        []string{"Common"},
	}, h.CommonHandler.GetWebsiteTitle)
}
