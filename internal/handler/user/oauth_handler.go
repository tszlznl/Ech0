package handler

import (
	"strings"

	"github.com/gin-gonic/gin"
	res "github.com/lin-snow/ech0/internal/handler/response"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
)

func normalizeOAuthProvider(provider string) (string, bool) {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case string(commonModel.OAuth2GITHUB):
		return string(commonModel.OAuth2GITHUB), true
	case string(commonModel.OAuth2GOOGLE):
		return string(commonModel.OAuth2GOOGLE), true
	case string(commonModel.OAuth2QQ):
		return string(commonModel.OAuth2QQ), true
	case string(commonModel.OAuth2CUSTOM):
		return string(commonModel.OAuth2CUSTOM), true
	default:
		return "", false
	}
}

func (userHandler *UserHandler) OAuthLogin() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		provider, ok := normalizeOAuthProvider(ctx.Param("provider"))
		if !ok {
			return res.Response{Msg: commonModel.INVALID_PARAMS}
		}

		redirectURI := ctx.Query("redirect_uri")
		redirectURL, err := userHandler.userService.GetOAuthLoginURL(provider, redirectURI)
		if err != nil {
			return res.Response{Err: err}
		}
		ctx.Redirect(302, redirectURL)
		return res.Response{}
	})
}

func (userHandler *UserHandler) OAuthCallback() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		provider, ok := normalizeOAuthProvider(ctx.Param("provider"))
		if !ok {
			return res.Response{Msg: commonModel.INVALID_PARAMS}
		}
		code := ctx.Query("code")
		state := ctx.Query("state")
		if code == "" || state == "" {
			return res.Response{Msg: commonModel.INVALID_PARAMS}
		}

		redirectURL, err := userHandler.userService.HandleOAuthCallback(provider, code, state)
		if err != nil || redirectURL == "" {
			return res.Response{Msg: commonModel.INVALID_PARAMS, Err: err}
		}
		ctx.Redirect(302, redirectURL)
		return res.Response{}
	})
}

func (userHandler *UserHandler) OAuthBind() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		provider, ok := normalizeOAuthProvider(ctx.Param("provider"))
		if !ok {
			return res.Response{Msg: commonModel.INVALID_PARAMS}
		}
		type reqBody struct {
			RedirectURI string `json:"redirect_uri"`
		}
		var req reqBody
		if err := ctx.ShouldBindJSON(&req); err != nil {
			return res.Response{Msg: commonModel.INVALID_REQUEST_BODY, Err: err}
		}

		bindURL, err := userHandler.userService.BindOAuth(ctx.Request.Context(), provider, req.RedirectURI)
		if err != nil {
			return res.Response{Err: err}
		}
		return res.Response{Data: bindURL, Msg: commonModel.GET_OAUTH_BINGURL_SUCCESS}
	})
}
