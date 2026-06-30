// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

// Package handler 暴露系统初始化的 HTTP 接口（Huma type-first，公开无鉴权）。
package handler

import (
	"context"

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

type ( // 输入
	GetInitStatusInput struct{}
	InitOwnerInput     struct {
		Body authModel.RegisterDto
	}
)

type ( // 输出
	StatusOutput = commonModel.Result[initModel.Status]
	EmptyOutput  = commonModel.Result[any]
)

// GetInitStatus 返回站点是否已初始化、是否已存在 Owner（前端引导首次部署）。
func (h *InitHandler) GetInitStatus(ctx context.Context, _ *GetInitStatusInput) (StatusOutput, error) {
	status, err := h.initService.GetStatus()
	if err != nil {
		return StatusOutput{}, err
	}
	return commonModel.OK(status), nil
}

// InitOwner 创建首个 Owner 账号（仅在未初始化时可用）。
func (h *InitHandler) InitOwner(ctx context.Context, in *InitOwnerInput) (EmptyOutput, error) {
	if err := h.initService.InitOwner(&in.Body); err != nil {
		return EmptyOutput{}, err
	}
	return commonModel.OK[any](nil, commonModel.INIT_OWNER_SUCCESS), nil
}
