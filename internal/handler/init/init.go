package handler

import (
	"github.com/gin-gonic/gin"
	res "github.com/lin-snow/ech0/internal/handler/response"
	authModel "github.com/lin-snow/ech0/internal/model/auth"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	service "github.com/lin-snow/ech0/internal/service/init"
)

type InitHandler struct {
	initService service.Service
}

func NewInitHandler(initService service.Service) *InitHandler {
	return &InitHandler{initService: initService}
}

func (h *InitHandler) GetInitStatus() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		status, err := h.initService.GetStatus()
		if err != nil {
			return res.Response{
				Msg: "",
				Err: err,
			}
		}
		return res.Response{
			Data: status,
			Msg:  commonModel.SUCCESS_MESSAGE,
		}
	})
}

func (h *InitHandler) InitOwner() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		var dto authModel.RegisterDto
		if err := ctx.ShouldBindJSON(&dto); err != nil {
			return res.Response{
				Msg: commonModel.INVALID_REQUEST_BODY,
				Err: err,
			}
		}

		if err := h.initService.InitOwner(&dto); err != nil {
			return res.Response{
				Msg: "",
				Err: err,
			}
		}

		return res.Response{
			Msg: commonModel.INIT_OWNER_SUCCESS,
		}
	})
}
