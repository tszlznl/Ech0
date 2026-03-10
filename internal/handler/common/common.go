package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	res "github.com/lin-snow/ech0/internal/handler/response"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	service "github.com/lin-snow/ech0/internal/service/common"
	errorUtil "github.com/lin-snow/ech0/internal/util/err"
	timezoneUtil "github.com/lin-snow/ech0/internal/util/timezone"
)

type CommonHandler struct {
	commonService service.Service
}

func NewCommonHandler(commonService service.Service) *CommonHandler {
	return &CommonHandler{
		commonService: commonService,
	}
}

func (commonHandler *CommonHandler) GetHeatMap() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		timezone := timezoneUtil.NormalizeTimezone(ctx.GetHeader(timezoneUtil.DefaultTimezoneHeader))
		heatMap, err := commonHandler.commonService.GetHeatMap(timezone)
		if err != nil {
			return res.Response{
				Msg: "",
				Err: err,
			}
		}

		return res.Response{
			Data: heatMap,
			Msg:  commonModel.GET_HEATMAP_SUCCESS,
		}
	})
}

func (commonHandler *CommonHandler) GetRss(ctx *gin.Context) {
	atom, err := commonHandler.commonService.GenerateRSS(ctx)
	if err != nil {
		ctx.JSON(
			http.StatusOK,
			commonModel.Fail[string](errorUtil.HandleError(&commonModel.ServerError{
				Msg: "",
				Err: err,
			})),
		)
		return
	}

	ctx.Data(http.StatusOK, "application/rss+xml; charset=utf-8", []byte(atom))
}

func (commonHandler *CommonHandler) HelloEch0() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		hello := struct {
			Hello   string `json:"hello"`
			Version string `json:"version"`
			Github  string `json:"github"`
		}{
			Hello:   "Hello, Ech0! 👋",
			Version: commonModel.Version,
			Github:  "https://github.com/lin-snow/Ech0",
		}

		return res.Response{
			Msg:  commonModel.GET_HELLO_SUCCESS,
			Data: hello,
		}
	})
}

func (commonHandler *CommonHandler) Healthz() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		return res.Response{
			Msg: commonModel.GET_HEALTHZ_SUCCESS,
			Data: struct {
				Status string `json:"status"`
			}{
				Status: "ok",
			},
		}
	})
}

func (commonHandler *CommonHandler) GetWebsiteTitle() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		var dto commonModel.GetWebsiteTitleDto
		if err := ctx.ShouldBindQuery(&dto); err != nil {
			return res.Response{
				Msg: commonModel.INVALID_QUERY_PARAMS,
				Err: err,
			}
		}
		title, err := commonHandler.commonService.GetWebsiteTitle(dto.WebSiteURL)
		if err != nil {
			return res.Response{
				Msg: "",
				Err: err,
			}
		}
		return res.Response{
			Data: title,
			Msg:  commonModel.GET_WEBSITE_TITLE_SUCCESS,
		}
	})
}
