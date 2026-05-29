// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	i18nUtil "github.com/lin-snow/ech0/internal/i18n"
	authModel "github.com/lin-snow/ech0/internal/model/auth"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	authService "github.com/lin-snow/ech0/internal/service/auth"
	errUtil "github.com/lin-snow/ech0/internal/util/err"
	jwtUtil "github.com/lin-snow/ech0/internal/util/jwt"
	"github.com/lin-snow/ech0/pkg/viewer"
)

// rejection 描述一次鉴权失败需要返回给客户端的内容（状态码 + 业务错误码 + i18n key + 回退文案）。
type rejection struct {
	status  int
	errCode string
	msgKey  string
	msg     string
}

// resolveViewer 从请求中解析 JWT 并按需将用户 viewer 写入 request context。
//
// 返回值语义：
//   - rej == nil：鉴权成功，用户 viewer 已挂载，调用方直接放行；
//   - rej != nil && !hard：鉴权失败但「可降级为匿名」（token 缺失 / 格式错误 / 解析失败 / 已吊销）——
//     RequireAuth 会拒绝，OptionalAuth 会挂匿名 viewer 后放行；
//   - rej != nil && hard：「不可降级」的硬拒绝（如带 admin scope 的 token 经 query 串传入），
//     两种模式都必须拒绝——避免高权限 token 在 URL 中泄漏。
//
// 鉴权策略（强制 / 可匿名）由路由注册时选择 RequireAuth 还是 OptionalAuth 决定，
// 中间件内部不再维护任何「公开路由 path 名单」。
func resolveViewer(ctx *gin.Context, tokenBlacklist authService.TokenRevoker) (rej *rejection, hard bool) {
	// 优先取 Authorization 头；缺失时回退到 query ?token=（用于 <audio>/<video> 直链等无法设置头部的场景）
	auth := strings.TrimSpace(ctx.Request.Header.Get("Authorization"))
	tokenFromQuery := false
	if auth == "" {
		queryToken := strings.TrimSpace(ctx.Query("token"))
		queryToken = strings.Trim(queryToken, `"`)
		if queryToken != "" && queryToken != "null" && queryToken != "undefined" {
			auth = "Bearer " + queryToken
			tokenFromQuery = true
		}
	}

	parts := strings.SplitN(auth, " ", 2)

	// token 缺失 / 整体格式不对 / value 为空
	if auth == "" || len(parts) != 2 || len(parts[1]) == 0 || parts[1] == "null" || parts[1] == "undefined" {
		return &rejection{http.StatusUnauthorized, commonModel.ErrCodeTokenMissing, commonModel.MsgKeyAuthTokenMissing, commonModel.TOKEN_NOT_FOUND}, false
	}
	// scheme 必须是 Bearer
	if parts[0] != "Bearer" {
		return &rejection{http.StatusUnauthorized, commonModel.ErrCodeTokenInvalid, commonModel.MsgKeyAuthTokenInvalid, commonModel.TOKEN_NOT_VALID}, false
	}

	// 验证签名与过期（仅接受 typ=session / typ=access）
	mc, err := jwtUtil.ParseToken(parts[1])
	if err != nil {
		return &rejection{http.StatusUnauthorized, commonModel.ErrCodeTokenParse, commonModel.MsgKeyAuthTokenParse, commonModel.TOKEN_PARSE_ERROR}, false
	}

	// 黑名单检查：已登出 / 已吊销的 token 即使签名有效也拒绝（mc.ID 即 jti claim）
	if tokenBlacklist != nil && mc.ID != "" && tokenBlacklist.IsTokenRevoked(mc.ID) {
		return &rejection{http.StatusUnauthorized, commonModel.ErrCodeTokenRevoked, commonModel.MsgKeyAuthTokenRevoked, commonModel.TOKEN_REVOKED}, false
	}

	// 传输安全硬拒绝：禁止 admin scope token 经 query 串传入（URL 易被日志 / Referer 泄漏）。
	// 该拒绝不可降级——即使是公开路由也必须返回 403，而非静默降级为匿名。
	if tokenFromQuery && authModel.HasAdminScope(mc.Scopes) {
		return &rejection{http.StatusForbidden, commonModel.ErrCodeTokenTransportForbidden, commonModel.MsgKeyAuthTokenTransportForbidden, commonModel.NO_PERMISSION_DENIED}, true
	}

	// 鉴权成功：挂载用户 viewer，并在请求未显式指定语言时按用户偏好覆盖语言上下文
	viewer.AttachToRequest(
		&ctx.Request,
		viewer.NewUserViewerWithToken(mc.Userid, mc.Type, mc.Scopes, []string(mc.Audience), mc.ID),
	)
	i18nUtil.ApplyUserLocaleFromUserID(ctx, mc.Userid)
	return nil, false
}

// writeRejection 按 rejection 写出本地化错误响应并中断请求链。
func writeRejection(ctx *gin.Context, rej *rejection) {
	msg := i18nUtil.Localize(
		i18nUtil.LocalizerFromGin(ctx),
		rej.msgKey,
		errUtil.HandleError(&commonModel.ServerError{Msg: rej.msg, Err: nil}),
		nil,
	)
	ctx.JSON(rej.status, commonModel.FailWithLocalized[any](msg, rej.errCode, rej.msgKey, nil))
	ctx.Abort()
}

// RequireAuth 强制鉴权中间件：token 缺失 / 无效 / 已吊销一律拒绝。用于所有需要登录身份的路由。
func RequireAuth(tokenBlacklist authService.TokenRevoker) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if rej, _ := resolveViewer(ctx, tokenBlacklist); rej != nil {
			writeRejection(ctx, rej)
			return
		}
		ctx.Next()
	}
}

// OptionalAuth 可匿名降级中间件：携带有效 token 时按用户身份处理（管理员可见私密内容），
// 否则挂载匿名 viewer 继续放行。用于「公开可读、但对管理员展示更多」的路由（如 echo 列表 / 详情）。
// 注意：不可降级的硬拒绝（admin token 经 query 串传入）即使在此模式下也会被拒绝。
func OptionalAuth(tokenBlacklist authService.TokenRevoker) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		rej, hard := resolveViewer(ctx, tokenBlacklist)
		if rej != nil {
			if hard {
				writeRejection(ctx, rej)
				return
			}
			viewer.AttachToRequest(&ctx.Request, viewer.NewNoopViewer())
		}
		ctx.Next()
	}
}
