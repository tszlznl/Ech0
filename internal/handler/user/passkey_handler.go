package handler

import (
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lin-snow/ech0/internal/config"
	res "github.com/lin-snow/ech0/internal/handler/response"
	authModel "github.com/lin-snow/ech0/internal/model/auth"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
)

func getPasskeyOriginAndRPID(ctx *gin.Context) (origin string, rpID string) {
	cfg := config.Config().Auth.WebAuthn
	if strings.TrimSpace(cfg.RPID) != "" && len(cfg.Origins) > 0 {
		return strings.TrimSpace(cfg.Origins[0]), strings.TrimSpace(cfg.RPID)
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

func (userHandler *UserHandler) PasskeyLoginBeginV2() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		origin, rpID := getPasskeyOriginAndRPID(ctx)
		if origin == "" || rpID == "" {
			return res.Response{Msg: commonModel.INVALID_PARAMS}
		}
		data, err := userHandler.userService.PasskeyLoginBegin(rpID, origin)
		if err != nil {
			return res.Response{Err: err}
		}
		return res.Response{Data: data}
	})
}

func (userHandler *UserHandler) PasskeyLoginFinishV2() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		var req authModel.PasskeyFinishReq
		if err := ctx.ShouldBindJSON(&req); err != nil {
			return res.Response{Msg: commonModel.INVALID_REQUEST_BODY, Err: err}
		}
		origin, rpID := getPasskeyOriginAndRPID(ctx)
		if origin == "" || rpID == "" {
			return res.Response{Msg: commonModel.INVALID_PARAMS}
		}
		token, err := userHandler.userService.PasskeyLoginFinish(rpID, origin, req.Nonce, req.Credential)
		if err != nil {
			return res.Response{Err: err}
		}
		return res.Response{Data: token}
	})
}

func (userHandler *UserHandler) PasskeyRegisterBeginV2() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		var req authModel.PasskeyRegisterBeginReq
		_ = ctx.ShouldBindJSON(&req)
		origin, rpID := getPasskeyOriginAndRPID(ctx)
		if origin == "" || rpID == "" {
			return res.Response{Msg: commonModel.INVALID_PARAMS}
		}
		data, err := userHandler.userService.PasskeyRegisterBegin(ctx.Request.Context(), rpID, origin, req.DeviceName)
		if err != nil {
			return res.Response{Err: err}
		}
		return res.Response{Data: data}
	})
}

func (userHandler *UserHandler) PasskeyRegisterFinishV2() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		var req authModel.PasskeyFinishReq
		if err := ctx.ShouldBindJSON(&req); err != nil {
			return res.Response{Msg: commonModel.INVALID_REQUEST_BODY, Err: err}
		}
		origin, rpID := getPasskeyOriginAndRPID(ctx)
		if origin == "" || rpID == "" {
			return res.Response{Msg: commonModel.INVALID_PARAMS}
		}
		if err := userHandler.userService.PasskeyRegisterFinish(ctx.Request.Context(), rpID, origin, req.Nonce, req.Credential); err != nil {
			return res.Response{Err: err}
		}
		return res.Response{}
	})
}
