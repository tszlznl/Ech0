// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

// Package handler 暴露用户相关的 HTTP 接口（Huma type-first）。
package handler

import (
	"context"

	authModel "github.com/lin-snow/ech0/internal/model/auth"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	model "github.com/lin-snow/ech0/internal/model/user"
	service "github.com/lin-snow/ech0/internal/service/user"
	"github.com/lin-snow/ech0/pkg/viewer"
)

type UserHandler struct {
	userService service.Service
}

func NewUserHandler(userService service.Service) *UserHandler {
	return &UserHandler{userService: userService}
}

type ( // 输入
	RegisterInput struct {
		Body authModel.RegisterDto
	}
	UpdateUserInput struct {
		Body model.UserInfoDto
	}
	UpdateUserAdminInput struct {
		ID string `path:"id" format:"uuid" doc:"用户 ID（UUID）"`
	}
	GetAllUsersInput struct{}
	DeleteUserInput  struct {
		ID string `path:"id" format:"uuid" doc:"用户 ID（UUID）"`
	}
	GetUserInfoInput struct{}
)

type ( // 输出
	UserListOutput = commonModel.Result[[]model.User]
	UserOutput     = commonModel.Result[model.User]
	EmptyOutput    = commonModel.Result[any]
)

// Register 注册新用户账号（公开）。
func (userHandler *UserHandler) Register(ctx context.Context, in *RegisterInput) (EmptyOutput, error) {
	if err := userHandler.userService.Register(&in.Body); err != nil {
		return EmptyOutput{}, err
	}
	return commonModel.OK[any](nil, commonModel.REGISTER_SUCCESS), nil
}

// UpdateUser 更新当前已认证用户的个人信息（profile:write）。
func (userHandler *UserHandler) UpdateUser(ctx context.Context, in *UpdateUserInput) (EmptyOutput, error) {
	if err := userHandler.userService.UpdateUser(ctx, in.Body); err != nil {
		return EmptyOutput{}, err
	}
	return commonModel.OK[any](nil, commonModel.UPDATE_USER_SUCCESS), nil
}

// UpdateUserAdmin 由管理员切换指定用户的管理员状态（admin:user）。
func (userHandler *UserHandler) UpdateUserAdmin(ctx context.Context, in *UpdateUserAdminInput) (EmptyOutput, error) {
	if err := userHandler.userService.UpdateUserAdmin(ctx, in.ID); err != nil {
		return EmptyOutput{}, err
	}
	return commonModel.OK[any](nil, commonModel.UPDATE_USER_SUCCESS), nil
}

// GetAllUsers 管理员获取系统中所有用户的列表（admin:user）。
func (userHandler *UserHandler) GetAllUsers(ctx context.Context, _ *GetAllUsersInput) (UserListOutput, error) {
	allusers, err := userHandler.userService.GetAllUsers(ctx)
	if err != nil {
		return UserListOutput{}, err
	}
	return commonModel.OK(allusers, commonModel.GET_USER_SUCCESS), nil
}

// DeleteUser 管理员根据 ID 删除指定用户（admin:user）。
func (userHandler *UserHandler) DeleteUser(ctx context.Context, in *DeleteUserInput) (EmptyOutput, error) {
	if err := userHandler.userService.DeleteUser(ctx, in.ID); err != nil {
		return EmptyOutput{}, err
	}
	return commonModel.OK[any](nil, commonModel.DELETE_USER_SUCCESS), nil
}

// GetUserInfo 获取当前已认证用户的详细信息（profile:read）。
func (userHandler *UserHandler) GetUserInfo(ctx context.Context, _ *GetUserInfoInput) (UserOutput, error) {
	userid := viewer.MustFromContext(ctx).UserID()
	user, err := userHandler.userService.GetUserByID(userid)
	user.Password = ""
	if err != nil {
		return UserOutput{}, err
	}
	return commonModel.OK(user, commonModel.GET_USER_INFO_SUCCESS), nil
}
