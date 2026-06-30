// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

// Package handler 暴露实例互联（Connect）的 HTTP 接口（Huma type-first）。
package handler

import (
	"context"

	commonModel "github.com/lin-snow/ech0/internal/model/common"
	connectModel "github.com/lin-snow/ech0/internal/model/connect"
	service "github.com/lin-snow/ech0/internal/service/connect"
)

type ConnectHandler struct {
	connectService service.Service
}

// NewConnectHandler ConnectHandler 的构造函数
func NewConnectHandler(connectService service.Service) *ConnectHandler {
	return &ConnectHandler{
		connectService: connectService,
	}
}

type ( // 输入
	GetConnectInput      struct{}
	GetConnectsInput     struct{}
	GetConnectsInfoInput struct{}
	GetConnectsHealthIn  struct{}
	AddConnectInput      struct {
		Body connectModel.Connected
	}
	DeleteConnectInput struct {
		ID string `path:"id" format:"uuid" doc:"连接 ID（UUID）"`
	}
)

type ( // 输出
	ConnectOutput       = commonModel.Result[connectModel.Connect]
	ConnectedListOutput = commonModel.Result[[]connectModel.Connected]
	ConnectListOutput   = commonModel.Result[[]connectModel.Connect]
	ConnectHealthOutput = commonModel.Result[[]connectModel.ConnectedHealth]
	EmptyOutput         = commonModel.Result[any]
)

// GetConnect 提供当前实例的连接信息（公开）。
func (connectHandler *ConnectHandler) GetConnect(ctx context.Context, _ *GetConnectInput) (ConnectOutput, error) {
	connect, err := connectHandler.connectService.GetConnect()
	if err != nil {
		return ConnectOutput{}, err
	}
	return commonModel.OK(connect, commonModel.CONNECT_SUCCESS), nil
}

// GetConnects 获取当前实例添加的所有连接（公开）。
func (connectHandler *ConnectHandler) GetConnects(ctx context.Context, _ *GetConnectsInput) (ConnectedListOutput, error) {
	connects, err := connectHandler.connectService.GetConnects()
	if err != nil {
		return ConnectedListOutput{}, err
	}
	return commonModel.OK(connects, commonModel.GET_CONNECTED_LIST_SUCCESS), nil
}

// GetConnectsInfo 获取所有已添加连接的详细信息（公开）。
func (connectHandler *ConnectHandler) GetConnectsInfo(ctx context.Context, _ *GetConnectsInfoInput) (ConnectListOutput, error) {
	connects, err := connectHandler.connectService.GetConnectsInfo()
	if err != nil {
		return ConnectListOutput{}, err
	}
	return commonModel.OK(connects, commonModel.GET_CONNECT_INFO_SUCCESS), nil
}

// GetConnectsHealth 探测已保存互联地址的可达性及远端版本（connect:read）。
func (connectHandler *ConnectHandler) GetConnectsHealth(ctx context.Context, _ *GetConnectsHealthIn) (ConnectHealthOutput, error) {
	rows, err := connectHandler.connectService.GetConnectsHealth()
	if err != nil {
		return ConnectHealthOutput{}, err
	}
	return commonModel.OK(rows, commonModel.GET_CONNECT_HEALTH_SUCCESS), nil
}

// AddConnect 添加一个新的连接（connect:write）。
func (connectHandler *ConnectHandler) AddConnect(ctx context.Context, in *AddConnectInput) (EmptyOutput, error) {
	if err := connectHandler.connectService.AddConnect(ctx, in.Body); err != nil {
		return EmptyOutput{}, err
	}
	return commonModel.OK[any](nil, commonModel.ADD_CONNECT_SUCCESS), nil
}

// DeleteConnect 根据 ID 删除一个已添加的连接（connect:write）。
func (connectHandler *ConnectHandler) DeleteConnect(ctx context.Context, in *DeleteConnectInput) (EmptyOutput, error) {
	if err := connectHandler.connectService.DeleteConnect(ctx, in.ID); err != nil {
		return EmptyOutput{}, err
	}
	return commonModel.OK[any](nil, commonModel.DELETE_CONNECT_SUCCESS), nil
}
