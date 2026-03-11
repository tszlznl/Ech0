package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	res "github.com/lin-snow/ech0/internal/handler/response"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	service "github.com/lin-snow/ech0/internal/service/inbox"
)

// InboxHandler 负责处理收件箱相关 HTTP 请求
type InboxHandler struct {
	inboxService service.Service
}

// NewInboxHandler 创建新的 InboxHandler 实例
func NewInboxHandler(inboxService service.Service) *InboxHandler {
	return &InboxHandler{inboxService: inboxService}
}

// GetInboxList 获取收件箱消息列表
//
//	@Summary		获取收件箱列表
//	@Description	根据分页条件获取系统收件箱
//	@Tags			收件箱
//	@Accept			json
//	@Produce		json
//	@Param			page		query		int				false	"页码"
//	@Param			pageSize	query		int				false	"每页数量"
//	@Param			search		query		string			false	"搜索关键词"
//	@Success		200			{object}	res.Response	"获取成功"
//	@Failure		200			{object}	res.Response	"获取失败"
//	@Router			/inbox [get]
func (inboxHandler *InboxHandler) GetInboxList() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		var pageQuery commonModel.PageQueryDto
		if err := ctx.ShouldBindQuery(&pageQuery); err != nil {
			return res.Response{
				Msg: commonModel.INVALID_QUERY_PARAMS,
				Err: err,
			}
		}

		result, err := inboxHandler.inboxService.GetInboxList(ctx.Request.Context(), pageQuery)
		if err != nil {
			return res.Response{Err: err}
		}

		return res.Response{
			Data: result,
			Msg:  commonModel.GET_INBOX_LIST_SUCCESS,
		}
	})
}

// GetUnreadInbox 获取所有未读消息
//
//	@Summary		获取未读消息
//	@Description	获取所有未读收件箱消息
//	@Tags			收件箱
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	res.Response	"获取成功"
//	@Failure		200	{object}	res.Response	"获取失败"
//	@Router			/inbox/unread [get]
func (inboxHandler *InboxHandler) GetUnreadInbox() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		inboxes, err := inboxHandler.inboxService.GetUnreadInbox(ctx.Request.Context())
		if err != nil {
			return res.Response{Err: err}
		}

		return res.Response{
			Data: inboxes,
			Msg:  commonModel.GET_UNREAD_INBOX_SUCCESS,
		}
	})
}

// MarkInboxAsRead 将消息标记为已读
//
//	@Summary		标记消息为已读
//	@Description	根据 ID 将消息标记为已读
//	@Tags			收件箱
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int				true	"收件箱ID"
//	@Success		200	{object}	res.Response	"标记成功"
//	@Failure		200	{object}	res.Response	"标记失败"
//	@Router			/inbox/{id}/read [put]
func (inboxHandler *InboxHandler) MarkInboxAsRead() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		inboxID, err := parseUUIDParam(ctx.Param("id"))
		if err != nil {
			return res.Response{
				Msg: commonModel.INVALID_PARAMS_BODY,
				Err: err,
			}
		}

		if err := inboxHandler.inboxService.MarkAsRead(ctx.Request.Context(), inboxID); err != nil {
			return res.Response{Err: err}
		}

		return res.Response{Msg: commonModel.MARK_INBOX_READ_SUCCESS}
	})
}

// DeleteInbox 删除指定的收件箱消息
//
//	@Summary		删除收件箱消息
//	@Description	根据 ID 删除收件箱消息
//	@Tags			收件箱
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int				true	"收件箱ID"
//	@Success		200	{object}	res.Response	"删除成功"
//	@Failure		200	{object}	res.Response	"删除失败"
//	@Router			/inbox/{id} [delete]
func (inboxHandler *InboxHandler) DeleteInbox() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		inboxID, err := parseUUIDParam(ctx.Param("id"))
		if err != nil {
			return res.Response{
				Msg: commonModel.INVALID_PARAMS_BODY,
				Err: err,
			}
		}

		if err := inboxHandler.inboxService.DeleteInbox(ctx.Request.Context(), inboxID); err != nil {
			return res.Response{Err: err}
		}

		return res.Response{Msg: commonModel.DELETE_INBOX_SUCCESS}
	})
}

// ClearInbox 清空收件箱
//
//	@Summary		清空收件箱
//	@Description	删除所有收件箱消息
//	@Tags			收件箱
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	res.Response	"清空成功"
//	@Failure		200	{object}	res.Response	"清空失败"
//	@Router			/inbox [delete]
func (inboxHandler *InboxHandler) ClearInbox() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		if err := inboxHandler.inboxService.ClearInbox(ctx.Request.Context()); err != nil {
			return res.Response{Err: err}
		}

		return res.Response{Msg: commonModel.CLEAR_INBOX_SUCCESS}
	})
}

func parseUUIDParam(raw string) (string, error) {
	if _, err := uuid.Parse(raw); err != nil {
		return "", err
	}
	return raw, nil
}
