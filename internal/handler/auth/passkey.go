// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package handler

import (
	"context"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lin-snow/ech0/internal/config"
	res "github.com/lin-snow/ech0/internal/handler/response"
	authModel "github.com/lin-snow/ech0/internal/model/auth"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	cookieUtil "github.com/lin-snow/ech0/internal/util/cookie"
)

type (
	ListPasskeysInput struct{}
	PasskeyIDInput    struct {
		ID string `path:"id" format:"uuid" doc:"Passkey 设备 ID（UUID）"`
	}
	UpdatePasskeyNameInput struct {
		ID   string `path:"id" format:"uuid" doc:"Passkey 设备 ID（UUID）"`
		Body authModel.PasskeyUpdateDeviceNameReq
	}
)

type (
	PasskeyListOutput = commonModel.Result[[]authModel.PasskeyDeviceDto]
	EmptyOutput       = commonModel.Result[any]
)

func (h *AuthHandler) getPasskeyOriginAndRPID(ctx *gin.Context) (origin string, rpID string) {
	// 管理员在面板配置的固定 RP（passkey_setting）优先；未配置则回退到请求来源。
	rpID, origins := h.authService.PasskeyBoundary(ctx.Request.Context())
	if rpID != "" && len(origins) > 0 {
		return strings.TrimSpace(origins[0]), rpID
	}
	origin = strings.TrimSpace(ctx.GetHeader("Origin"))
	if origin == "" {
		return "", ""
	}
	if u, err := url.Parse(origin); err == nil && u.Hostname() != "" {
		return origin, u.Hostname()
	}
	return "", ""
}

func getOriginAndRPID(ctx *gin.Context) (origin string, rpID string) {
	if configured := strings.TrimSpace(config.Config().Setting.Serverurl); configured != "" {
		if u, err := url.Parse(configured); err == nil && u.Scheme != "" && u.Hostname() != "" {
			return u.Scheme + "://" + u.Host, u.Hostname()
		}
	}

	origin = strings.TrimSpace(ctx.GetHeader("Origin"))
	if origin == "" {
		ref := strings.TrimSpace(ctx.GetHeader("Referer"))
		if ref != "" {
			if u, err := url.Parse(ref); err == nil && u.Scheme != "" && u.Host != "" {
				origin = u.Scheme + "://" + u.Host
			}
		}
		if origin == "" {
			scheme := "http"
			if ctx.Request.TLS != nil {
				scheme = "https"
			}
			origin = scheme + "://" + ctx.Request.Host
		}
	}

	if u, err := url.Parse(origin); err == nil {
		if h := u.Hostname(); h != "" {
			rpID = h
		}
	}
	if rpID == "" {
		host := ctx.Request.Host
		if strings.Contains(host, ":") {
			host = strings.Split(host, ":")[0]
		}
		rpID = host
	}
	return origin, rpID
}

func (h *AuthHandler) PasskeyLoginBeginV2() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		origin, rpID := h.getPasskeyOriginAndRPID(ctx)
		if origin == "" || rpID == "" {
			return res.Response{Msg: commonModel.INVALID_PARAMS}
		}
		data, err := h.authService.PasskeyLoginBegin(rpID, origin)
		if err != nil {
			return res.Response{Err: err}
		}
		return res.Response{Data: data}
	})
}

func (h *AuthHandler) PasskeyLoginFinishV2() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		var req authModel.PasskeyFinishReq
		if err := ctx.ShouldBindJSON(&req); err != nil {
			return res.Response{Msg: commonModel.INVALID_REQUEST_BODY, Err: err}
		}
		origin, rpID := h.getPasskeyOriginAndRPID(ctx)
		if origin == "" || rpID == "" {
			return res.Response{Msg: commonModel.INVALID_PARAMS}
		}
		tokenPair, err := h.authService.PasskeyLoginFinish(rpID, origin, req.Nonce, req.Credential)
		if err != nil {
			return res.Response{Err: err}
		}
		cookieUtil.SetRefreshTokenCookie(ctx, tokenPair.RefreshToken, config.Config().Auth.Jwt.RefreshExpires)
		return res.Response{Data: tokenPair}
	})
}

func (h *AuthHandler) PasskeyRegisterBeginV2() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		var req authModel.PasskeyRegisterBeginReq
		_ = ctx.ShouldBindJSON(&req)
		origin, rpID := h.getPasskeyOriginAndRPID(ctx)
		if origin == "" || rpID == "" {
			return res.Response{Msg: commonModel.INVALID_PARAMS}
		}
		data, err := h.authService.PasskeyRegisterBegin(ctx.Request.Context(), rpID, origin, req.DeviceName)
		if err != nil {
			return res.Response{Err: err}
		}
		return res.Response{Data: data}
	})
}

func (h *AuthHandler) PasskeyRegisterFinishV2() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		var req authModel.PasskeyFinishReq
		if err := ctx.ShouldBindJSON(&req); err != nil {
			return res.Response{Msg: commonModel.INVALID_REQUEST_BODY, Err: err}
		}
		origin, rpID := h.getPasskeyOriginAndRPID(ctx)
		if origin == "" || rpID == "" {
			return res.Response{Msg: commonModel.INVALID_PARAMS}
		}
		if err := h.authService.PasskeyRegisterFinish(ctx.Request.Context(), rpID, origin, req.Nonce, req.Credential); err != nil {
			return res.Response{Err: err}
		}
		return res.Response{}
	})
}

func (h *AuthHandler) PasskeyLoginBegin() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		origin, rpID := getOriginAndRPID(ctx)
		data, err := h.authService.PasskeyLoginBegin(rpID, origin)
		if err != nil {
			return res.Response{Err: err}
		}
		return res.Response{Data: data}
	})
}

func (h *AuthHandler) PasskeyLoginFinish() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		var req authModel.PasskeyFinishReq
		if err := ctx.ShouldBindJSON(&req); err != nil {
			return res.Response{Msg: commonModel.INVALID_REQUEST_BODY, Err: err}
		}
		origin, rpID := getOriginAndRPID(ctx)
		token, err := h.authService.PasskeyLoginFinish(rpID, origin, req.Nonce, req.Credential)
		if err != nil {
			return res.Response{Err: err}
		}
		return res.Response{Data: token}
	})
}

func (h *AuthHandler) PasskeyRegisterBegin() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		var req authModel.PasskeyRegisterBeginReq
		_ = ctx.ShouldBindJSON(&req)

		origin, rpID := getOriginAndRPID(ctx)
		data, err := h.authService.PasskeyRegisterBegin(
			ctx.Request.Context(),
			rpID,
			origin,
			req.DeviceName,
		)
		if err != nil {
			return res.Response{Err: err}
		}
		return res.Response{Data: data}
	})
}

func (h *AuthHandler) PasskeyRegisterFinish() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		var req authModel.PasskeyFinishReq
		if err := ctx.ShouldBindJSON(&req); err != nil {
			return res.Response{Msg: commonModel.INVALID_REQUEST_BODY, Err: err}
		}

		origin, rpID := getOriginAndRPID(ctx)
		if err := h.authService.PasskeyRegisterFinish(ctx.Request.Context(), rpID, origin, req.Nonce, req.Credential); err != nil {
			return res.Response{Err: err}
		}
		return res.Response{}
	})
}

func (h *AuthHandler) ListPasskeys(ctx context.Context, _ *ListPasskeysInput) (PasskeyListOutput, error) {
	devs, err := h.authService.ListPasskeys(ctx)
	if err != nil {
		return PasskeyListOutput{}, err
	}
	return commonModel.OK(devs), nil
}

func (h *AuthHandler) DeletePasskey(ctx context.Context, in *PasskeyIDInput) (EmptyOutput, error) {
	if err := h.authService.DeletePasskey(ctx, in.ID); err != nil {
		return EmptyOutput{}, err
	}
	return commonModel.OK[any](nil), nil
}

func (h *AuthHandler) UpdatePasskeyDeviceName(ctx context.Context, in *UpdatePasskeyNameInput) (EmptyOutput, error) {
	if err := h.authService.UpdatePasskeyDeviceName(ctx, in.ID, in.Body.DeviceName); err != nil {
		return EmptyOutput{}, err
	}
	return commonModel.OK[any](nil), nil
}
