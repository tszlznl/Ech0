// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

// Package handler 暴露 Echo（动态）与 Tag（标签）相关的 HTTP 接口（Huma type-first）。
//
// 读接口走「可匿名降级」中间件（无 token 仅公开内容，管理员可见私密）；写接口需 echo:write。
// 点赞接口匿名可访问，但叠加 IP 限速 + (IP, echoID) 去重窗口（经桥接的 RateLimitWithIdempotency）。
package handler

import (
	"context"

	"github.com/lin-snow/ech0/internal/handler/humares"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	model "github.com/lin-snow/ech0/internal/model/echo"
	service "github.com/lin-snow/ech0/internal/service/echo"
	timezoneUtil "github.com/lin-snow/ech0/internal/util/timezone"
)

type EchoHandler struct {
	echoService service.Service
}

// NewEchoHandler EchoHandler 的构造函数
func NewEchoHandler(echoService service.Service) *EchoHandler {
	return &EchoHandler{
		echoService: echoService,
	}
}

type (
	EchoUpsertInput struct {
		Body model.EchoUpsertDto
	}
	EchoIDInput struct {
		ID string `path:"id" format:"uuid" doc:"Echo ID"`
	}
	QueryEchosInput struct {
		Body commonModel.EchoQueryDto
	}
	EchoPageGetInput struct {
		Page     int    `query:"page"`
		PageSize int    `query:"pageSize"`
		Search   string `query:"search"`
	}
	EchoPagePostInput struct {
		Body commonModel.PageQueryDto
	}
	GetEchosByTagIDInput struct {
		TagID    string `path:"tagid" format:"uuid" doc:"标签 ID"`
		Page     int    `query:"page"`
		PageSize int    `query:"pageSize"`
		Search   string `query:"search"`
	}
	TimezoneInput struct {
		Timezone string `header:"X-Timezone" doc:"客户端时区（IANA 名）"`
	}
	GetHotEchosInput struct {
		Limit int `query:"limit" default:"5" doc:"返回条数，默认 5，最大 20"`
	}
	GetRandomEchoInput struct{}
	GetAllTagsInput    struct{}
	CreateTagInput     struct {
		Body model.CreateTagDto
	}
	TagIDInput struct {
		ID string `path:"id" format:"uuid" doc:"标签 ID"`
	}
	LikeEchoInput struct {
		ID string `path:"id" format:"uuid" doc:"Echo ID"`
	}
)

type echoPage = commonModel.PageQueryResult[[]model.Echo]

// PostEcho 创建一条新的 Echo（echo:write）。
func (echoHandler *EchoHandler) PostEcho(ctx context.Context, in *EchoUpsertInput) (*humares.Envelope[any], error) {
	if err := echoHandler.echoService.PostEcho(ctx, in.Body.ToModel()); err != nil {
		return nil, humares.Err(ctx, err)
	}
	return humares.OK[any](ctx, nil, commonModel.POST_ECHO_SUCCESS), nil
}

// UpdateEcho 更新指定 Echo 内容（echo:write）。
func (echoHandler *EchoHandler) UpdateEcho(ctx context.Context, in *EchoUpsertInput) (*humares.Envelope[any], error) {
	if err := echoHandler.echoService.UpdateEcho(ctx, in.Body.ToModel()); err != nil {
		return nil, humares.Err(ctx, err)
	}
	return humares.OK[any](ctx, nil, commonModel.UPDATE_ECHO_SUCCESS), nil
}

// DeleteEcho 根据 ID 删除 Echo（echo:write）。
func (echoHandler *EchoHandler) DeleteEcho(ctx context.Context, in *EchoIDInput) (*humares.Envelope[any], error) {
	if err := echoHandler.echoService.DeleteEchoById(ctx, in.ID); err != nil {
		return nil, humares.Err(ctx, err)
	}
	return humares.OK[any](ctx, nil, commonModel.DELETE_ECHO_SUCCESS), nil
}

// LikeEcho 为指定 Echo 点赞（匿名可访问，限速 + 去重）。
func (echoHandler *EchoHandler) LikeEcho(ctx context.Context, in *LikeEchoInput) (*humares.Envelope[any], error) {
	if err := echoHandler.echoService.LikeEcho(ctx, in.ID); err != nil {
		return nil, humares.Err(ctx, err)
	}
	return humares.OK[any](ctx, nil, commonModel.LIKE_ECHO_SUCCESS), nil
}

// GetEchoById 获取指定 ID 的 Echo 详情（可匿名降级）。
func (echoHandler *EchoHandler) GetEchoById(ctx context.Context, in *EchoIDInput) (*humares.Envelope[*model.Echo], error) {
	echo, err := echoHandler.echoService.GetEchoById(ctx, in.ID)
	if err != nil {
		return nil, humares.Err(ctx, err)
	}
	return humares.OK(ctx, echo, commonModel.GET_ECHO_BY_ID_SUCCESS), nil
}

// QueryEchos 统一查询 Echo 列表（可匿名降级）。
func (echoHandler *EchoHandler) QueryEchos(ctx context.Context, in *QueryEchosInput) (*humares.Envelope[echoPage], error) {
	result, err := echoHandler.echoService.QueryEchos(ctx, in.Body)
	if err != nil {
		return nil, humares.Err(ctx, err)
	}
	return humares.OK(ctx, result, commonModel.QUERY_ECHOS_SUCCESS), nil
}

// GetEchosByPageGet 分页获取 Echo 列表（GET query，Deprecated，请用 POST /echo/query）。
func (echoHandler *EchoHandler) GetEchosByPageGet(ctx context.Context, in *EchoPageGetInput) (*humares.Envelope[echoPage], error) {
	return echoHandler.getEchosByPage(ctx, commonModel.PageQueryDto{Page: in.Page, PageSize: in.PageSize, Search: in.Search})
}

// GetEchosByPagePost 分页获取 Echo 列表（POST body，Deprecated，请用 POST /echo/query）。
func (echoHandler *EchoHandler) GetEchosByPagePost(ctx context.Context, in *EchoPagePostInput) (*humares.Envelope[echoPage], error) {
	return echoHandler.getEchosByPage(ctx, in.Body)
}

func (echoHandler *EchoHandler) getEchosByPage(ctx context.Context, page commonModel.PageQueryDto) (*humares.Envelope[echoPage], error) {
	result, err := echoHandler.echoService.GetEchosByPage(ctx, page)
	if err != nil {
		return nil, humares.Err(ctx, err)
	}
	return humares.OK(ctx, result, commonModel.GET_ECHOS_BY_PAGE_SUCCESS), nil
}

// GetEchosByTagId 按标签 ID 获取 Echo 列表（可匿名降级，Deprecated）。
func (echoHandler *EchoHandler) GetEchosByTagId(ctx context.Context, in *GetEchosByTagIDInput) (*humares.Envelope[echoPage], error) {
	result, err := echoHandler.echoService.GetEchosByTagId(ctx, in.TagID, commonModel.PageQueryDto{Page: in.Page, PageSize: in.PageSize, Search: in.Search})
	if err != nil {
		return nil, humares.Err(ctx, err)
	}
	return humares.OK(ctx, result, commonModel.GET_ECHOS_BY_TAG_ID_SUCCESS), nil
}

// GetTodayEchos 获取今天发布的 Echo 列表（可匿名降级）。
func (echoHandler *EchoHandler) GetTodayEchos(ctx context.Context, in *TimezoneInput) (*humares.Envelope[[]model.Echo], error) {
	result, err := echoHandler.echoService.GetTodayEchos(ctx, timezoneUtil.NormalizeTimezone(in.Timezone))
	if err != nil {
		return nil, humares.Err(ctx, err)
	}
	return humares.OK(ctx, result, commonModel.GET_TODAY_ECHOS_SUCCESS), nil
}

// GetHotEchos 获取热门 Echo 列表（可匿名降级）。
func (echoHandler *EchoHandler) GetHotEchos(ctx context.Context, in *GetHotEchosInput) (*humares.Envelope[[]model.Echo], error) {
	limit := in.Limit
	if limit <= 0 {
		limit = 5
	}
	result, err := echoHandler.echoService.GetHotEchos(ctx, limit)
	if err != nil {
		return nil, humares.Err(ctx, err)
	}
	return humares.OK(ctx, result, commonModel.GET_HOT_ECHOS_SUCCESS), nil
}

// GetRandomEcho 随机返回一篇可见 Echo（可匿名降级；无可见内容时 data 为 null）。
func (echoHandler *EchoHandler) GetRandomEcho(ctx context.Context, _ *GetRandomEchoInput) (*humares.Envelope[*model.Echo], error) {
	echo, err := echoHandler.echoService.GetRandomEcho(ctx)
	if err != nil {
		return nil, humares.Err(ctx, err)
	}
	return humares.OK(ctx, echo, commonModel.GET_RANDOM_ECHO_SUCCESS), nil
}

// GetOnThisDayEchos 那年今日：过去年份中与今天同「月-日」的 Echo（可匿名降级）。
func (echoHandler *EchoHandler) GetOnThisDayEchos(ctx context.Context, in *TimezoneInput) (*humares.Envelope[[]model.Echo], error) {
	result, err := echoHandler.echoService.GetOnThisDayEchos(ctx, timezoneUtil.NormalizeTimezone(in.Timezone))
	if err != nil {
		return nil, humares.Err(ctx, err)
	}
	return humares.OK(ctx, result, commonModel.GET_ON_THIS_DAY_ECHOS_SUCCESS), nil
}

// GetAllTags 获取所有标签及其使用次数（公开）。
func (echoHandler *EchoHandler) GetAllTags(ctx context.Context, _ *GetAllTagsInput) (*humares.Envelope[[]model.Tag], error) {
	tags, err := echoHandler.echoService.GetAllTags()
	if err != nil {
		return nil, humares.Err(ctx, err)
	}
	return humares.OK(ctx, tags, commonModel.GET_ALL_TAGS_SUCCESS), nil
}

// CreateTag 管理员显式创建一个标签（echo:write）。
func (echoHandler *EchoHandler) CreateTag(ctx context.Context, in *CreateTagInput) (*humares.Envelope[*model.Tag], error) {
	tag, err := echoHandler.echoService.CreateTag(ctx, in.Body.Name)
	if err != nil {
		return nil, humares.Err(ctx, err)
	}
	return humares.OK(ctx, tag, commonModel.CREATE_TAG_SUCCESS), nil
}

// DeleteTag 根据 ID 删除标签（echo:write）。
func (echoHandler *EchoHandler) DeleteTag(ctx context.Context, in *TagIDInput) (*humares.Envelope[any], error) {
	if err := echoHandler.echoService.DeleteTag(ctx, in.ID); err != nil {
		return nil, humares.Err(ctx, err)
	}
	return humares.OK[any](ctx, nil, commonModel.DELETE_TAG_SUCCESS), nil
}
