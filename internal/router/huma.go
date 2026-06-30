// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package router

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/gin-gonic/gin"
	"github.com/lin-snow/ech0/internal/config"
	"github.com/lin-snow/ech0/internal/handler"
	"github.com/lin-snow/ech0/internal/handler/humares"
	"github.com/lin-snow/ech0/internal/middleware"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	authService "github.com/lin-snow/ech0/internal/service/auth"
)

// 本文件是「Huma 装配层」：建统一 API、定义端点的鉴权姿态(posture)与注册器(route)、
// 汇总所有域的注册清单(registerOperations)、离线导出 OpenAPI(GenerateOpenAPIYAML)。

const (
	humaAPITitle   = "Ech0 API 文档"
	humaAPIVersion = "1.0"
	humaAPIBase    = "/api"
)

// setupHumaAPI 在全局中间件挂载之后，于一个**无组级鉴权**的 /api 组上创建统一的 Huma API。
// auth/scope 下沉为各 operation 自带的 posture 中间件，故 public/auth/optional 三种姿态
// 可共存于同一份 OpenAPI 文档。docs: /api/docs，spec: /api/openapi.json|.yaml。
func setupHumaAPI(r *gin.Engine) huma.API {
	humaGroup := r.Group(humaAPIBase)
	docs := humares.ParseDocsRenderer(config.Config().OpenAPI.DocsRenderer)
	return humares.NewAPI(r, humaGroup, humaAPITitle, humaAPIVersion, humaAPIBase, docs)
}

// posture 是一个 JSON 端点的「认证(Authn) + 授权(Authz)」姿态。它一处同时产出：
//   - OpenAPI 的 Security 声明（给文档）
//   - 运行时中间件链（给拦截，复用现有 gin 中间件经 humares.Bridge 适配）
//
// 二者由同一个构造函数生成，scope 不可能再写得对不上。authn 维度 = public/optional/secured，
// authz 维度 = secured 的 scopes 参数与 .audience()。
type posture struct {
	security    []map[string][]string
	middlewares huma.Middlewares
}

// public：不认证、不授权——任何人可访问。
func public() posture { return posture{} }

// optional：可选认证(Authn)——无 token 按匿名继续，携带有效 token 则识别用户身份；不做授权。
func optional(revoker authService.TokenRevoker) posture {
	return posture{middlewares: huma.Middlewares{
		humares.Bridge(middleware.NoCache()),
		humares.Bridge(middleware.OptionalAuth(revoker)),
	}}
}

// secured：必须认证(Authn)；附带 scopes 时再要求 scope 授权(Authz)，不带即「仅认证」。
func secured(revoker authService.TokenRevoker, scopes ...string) posture {
	mws := huma.Middlewares{
		humares.Bridge(middleware.NoCache()),
		humares.Bridge(middleware.RequireAuth(revoker)),
	}
	if len(scopes) > 0 {
		mws = append(mws, humares.Bridge(middleware.RequireScopes(scopes...)))
	}
	return posture{security: humares.Secured(scopes...), middlewares: mws}
}

// audience：在 secured 之上叠加 audience 授权(Authz)，用于集成令牌端点。
func (p posture) audience(auds ...string) posture {
	p.middlewares = append(p.middlewares, humares.Bridge(middleware.RequireAudience(auds...)))
	return p
}

// noCache 返回「仅 NoCache」的中间件，给公开但敏感的端点（如注册）在 op 字面量里显式声明。
func noCache() huma.Middlewares {
	return huma.Middlewares{humares.Bridge(middleware.NoCache())}
}

// route 注册一个 JSON 端点：套用 posture（Security + 鉴权中间件），并把中立 handler 经
// humares.Wrap 折成统一信封后交给 huma.Register。
//
// 中间件顺序：posture 的（认证/授权）在前，op 字面量自带的（如限速、评论的 StashMeta 等
// 非鉴权中间件）在后——与按组挂中间件的旧行为一致。
func route[I, T any](api huma.API, p posture, op huma.Operation, h func(context.Context, *I) (commonModel.Result[T], error)) {
	op.Security = p.security
	if len(p.middlewares) > 0 || len(op.Middlewares) > 0 {
		mws := make(huma.Middlewares, 0, len(p.middlewares)+len(op.Middlewares))
		mws = append(mws, p.middlewares...)
		mws = append(mws, op.Middlewares...)
		op.Middlewares = mws
	}
	huma.Register(api, op, humares.Wrap(h))
}

// registerOperations 注册所有已迁移到 Huma 的域的 operation。
// 迁移新域时在此追加对应的 register* 调用——这是唯一的注册清单，
// 运行时服务器与离线 spec 生成器（GenerateOpenAPIYAML）共用，保证两者一致。
func registerOperations(api huma.API, h *handler.Bundle, revoker authService.TokenRevoker) {
	registerInit(api, h)
	registerAuth(api, h, revoker)
	registerCommon(api, h, revoker)
	registerEcho(api, h, revoker)
	registerConnect(api, h, revoker)
	registerUser(api, h, revoker)
	registerSetting(api, h, revoker)
	registerFile(api, h, revoker)
	registerDashboard(api, h, revoker)
	registerCopilot(api, h, revoker)
	registerComment(api, h, revoker)
	registerMigration(api, h, revoker)
	registerEmbedding(api, h, revoker)
}

// GenerateOpenAPIYAML 构造一个一次性的 Huma API、注册全部 operation 并导出 OpenAPI YAML。
// 它仅反射 operation 的输入/输出类型，不会调用任何 handler，故可用零 Bundle / nil revoker —
// 供 `make openapi` 门禁离线生成提交到仓库的 spec。
func GenerateOpenAPIYAML() ([]byte, error) {
	gin.SetMode(gin.TestMode)
	api := setupHumaAPI(gin.New())
	registerOperations(api, &handler.Bundle{}, nil)
	return api.OpenAPI().YAML()
}
