// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package router

import (
	"github.com/danielgtaylor/huma/v2"
	"github.com/gin-gonic/gin"
	"github.com/lin-snow/ech0/internal/handler"
	"github.com/lin-snow/ech0/internal/handler/humares"
	authService "github.com/lin-snow/ech0/internal/service/auth"
)

const (
	humaAPITitle   = "Ech0 API 文档"
	humaAPIVersion = "1.0"
	humaAPIBase    = "/api"
)

// setupHumaAPI 在全局中间件挂载之后，于一个**无组级鉴权**的 /api 组上创建统一的 Huma API。
// auth/scope 下沉为各 register*Huma 里的 per-operation 中间件，故 public/auth/optional
// 三种姿态可共存于同一份 OpenAPI 文档。docs: /api/docs，spec: /api/openapi.json|.yaml。
func setupHumaAPI(r *gin.Engine) huma.API {
	humaGroup := r.Group(humaAPIBase)
	return humares.NewAPI(r, humaGroup, humaAPITitle, humaAPIVersion, humaAPIBase)
}

// revokerFromCtx 从路由上下文取出 token 吊销器（供 Bridge 复用 RequireAuth）。
func revokerFromCtx(ctx *RouterContext) authService.TokenRevoker {
	if ctx.MWDeps != nil {
		return ctx.MWDeps.TokenRevoker
	}
	return nil
}

// RegisterHumaOperations 注册所有已迁移到 Huma 的域的 operation。
// 迁移新域时在此追加对应的 register*Huma 调用——这是唯一的注册清单，
// 运行时服务器与离线 spec 生成器（GenerateOpenAPIYAML）共用，保证两者一致。
func RegisterHumaOperations(api huma.API, h *handler.Bundle, revoker authService.TokenRevoker) {
	registerEmbeddingHuma(api, h, revoker)
}

// GenerateOpenAPIYAML 构造一个一次性的 Huma API、注册全部 operation 并导出 OpenAPI YAML。
// 它仅反射 operation 的输入/输出类型，不会调用任何 handler，故可用零 Bundle / nil revoker —
// 供 `make openapi` 门禁离线生成提交到仓库的 spec。
func GenerateOpenAPIYAML() ([]byte, error) {
	gin.SetMode(gin.TestMode)
	api := setupHumaAPI(gin.New())
	RegisterHumaOperations(api, &handler.Bundle{}, nil)
	return api.OpenAPI().YAML()
}
