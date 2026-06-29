// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package humares

import (
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humagin"
	"github.com/gin-gonic/gin"
)

// SecuritySchemeBearer 是 OpenAPI 安全方案名，operation 用它声明所需 scope（见 Secured）。
const SecuritySchemeBearer = "bearerAuth"

// NewAPI 在给定的 gin engine + 无组级鉴权的 group 上创建统一的 Huma API 实例。
// auth/scope 全部下沉为 per-operation 中间件（见 router 层 + Bridge），
// 这样 public / auth / optional 三种姿态能共存于同一份 OpenAPI 文档。
//
// basePath 写入 OpenAPI.Servers，使 spec 中的相对路径（如 /echo/query）拼上 /api 前缀。
// docs 在 group 前缀下：/api/docs、/api/openapi.json、/api/openapi.yaml。
func NewAPI(engine *gin.Engine, group *gin.RouterGroup, title, version, basePath string) huma.API {
	installErrorModel()

	cfg := huma.DefaultConfig(title, version)
	cfg.Servers = []*huma.Server{{URL: basePath}}
	cfg.Components.SecuritySchemes = map[string]*huma.SecurityScheme{
		SecuritySchemeBearer: {Type: "http", Scheme: "bearer", BearerFormat: "JWT"},
	}

	// 关掉默认的 schema-link transformer：它会往响应体注入 $schema 字段、加 Link 头，
	// 破坏「信封不变」。清空 CreateHooks/Transformers 并停掉 /schemas 路由。
	cfg.CreateHooks = nil
	cfg.Transformers = nil
	cfg.SchemasPath = ""

	api := humagin.NewWithGroup(engine, group, cfg)
	api.UseMiddleware(injectLocalizer)
	return api
}

// Secured 声明 operation 需要 bearer 鉴权 + 指定 scope（仅用于 OpenAPI 文档展示；
// 实际拦截由 router 层用 Bridge 包裹的 RequireAuth/RequireScopes 完成）。
func Secured(scopes ...string) []map[string][]string {
	if scopes == nil {
		scopes = []string{}
	}
	return []map[string][]string{{SecuritySchemeBearer: scopes}}
}
