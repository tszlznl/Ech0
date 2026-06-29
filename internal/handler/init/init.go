// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

// Package handler 暴露系统初始化的 HTTP 接口（Huma type-first，公开无鉴权）。
package handler

import (
	"context"

	"github.com/lin-snow/ech0/internal/handler/humares"
	authModel "github.com/lin-snow/ech0/internal/model/auth"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	initModel "github.com/lin-snow/ech0/internal/model/init"
	service "github.com/lin-snow/ech0/internal/service/init"
)

type InitHandler struct {
	initService service.Service
}

func NewInitHandler(initService service.Service) *InitHandler {
	return &InitHandler{initService: initService}
}

type (
	GetInitStatusInput struct{}
	InitOwnerInput     struct {
		Body authModel.RegisterDto
	}
)

// GetInitStatus 返回站点是否已初始化、是否已存在 Owner（前端引导首次部署）。
func (h *InitHandler) GetInitStatus(ctx context.Context, _ *GetInitStatusInput) (*humares.Envelope[initModel.Status], error) {
	status, err := h.initService.GetStatus()
	if err != nil {
		return nil, humares.Err(ctx, err)
	}
	return humares.OK(ctx, status), nil
}

// InitOwner 创建首个 Owner 账号（仅在未初始化时可用）。
func (h *InitHandler) InitOwner(ctx context.Context, in *InitOwnerInput) (*humares.Envelope[any], error) {
	if err := h.initService.InitOwner(&in.Body); err != nil {
		return nil, humares.Err(ctx, err)
	}
	return humares.OK[any](ctx, nil, commonModel.INIT_OWNER_SUCCESS), nil
}
