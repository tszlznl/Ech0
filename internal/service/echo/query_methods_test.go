// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service_test

import (
	"context"
	"errors"
	"testing"

	commonModel "github.com/lin-snow/ech0/internal/model/common"
	echoModel "github.com/lin-snow/ech0/internal/model/echo"
	echoService "github.com/lin-snow/ech0/internal/service/echo"
	"github.com/lin-snow/ech0/internal/test/helpers"
	commonmock "github.com/lin-snow/ech0/internal/test/mocks/commonmock"
	echomock "github.com/lin-snow/ech0/internal/test/mocks/echomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// readMethod 描述一个「按 viewer 解析 showPrivate 后转发给仓储」的只读方法：
// expectRepo 用解析出的 showPrivate 设置仓储期望，invoke 触发该方法。
type readMethod struct {
	name       string
	expectRepo func(repo *echomock.MockRepository, showPrivate bool)
	invoke     func(svc *echoService.EchoService, ctx context.Context) error
}

func readMethods() []readMethod {
	return []readMethod{
		{
			name: "GetTodayEchos",
			expectRepo: func(repo *echomock.MockRepository, showPrivate bool) {
				repo.EXPECT().GetTodayEchos(showPrivate, "UTC").Return([]echoModel.Echo{helpers.NewEcho()}).Once()
			},
			invoke: func(svc *echoService.EchoService, ctx context.Context) error {
				_, err := svc.GetTodayEchos(ctx, "UTC")
				return err
			},
		},
		{
			name: "GetOnThisDayEchos",
			expectRepo: func(repo *echomock.MockRepository, showPrivate bool) {
				repo.EXPECT().GetOnThisDayEchos(showPrivate, "UTC").Return([]echoModel.Echo{helpers.NewEcho()}).Once()
			},
			invoke: func(svc *echoService.EchoService, ctx context.Context) error {
				_, err := svc.GetOnThisDayEchos(ctx, "UTC")
				return err
			},
		},
		{
			name: "GetHotEchos",
			expectRepo: func(repo *echomock.MockRepository, showPrivate bool) {
				repo.EXPECT().GetHotEchos(5, showPrivate).Return([]echoModel.Echo{}, nil).Once()
			},
			invoke: func(svc *echoService.EchoService, ctx context.Context) error {
				_, err := svc.GetHotEchos(ctx, 5)
				return err
			},
		},
		{
			name: "GetRandomEcho",
			expectRepo: func(repo *echomock.MockRepository, showPrivate bool) {
				repo.EXPECT().GetRandomEcho(showPrivate).Return(nil, nil).Once()
			},
			invoke: func(svc *echoService.EchoService, ctx context.Context) error {
				_, err := svc.GetRandomEcho(ctx)
				return err
			},
		},
	}
}

// TestReadEchos_ShowPrivateResolution 锁定四个只读方法共享的可见性解析：
// 匿名 / 非管理员 → showPrivate=false（不查私密）；管理员 → showPrivate=true。
// 匿名分支不应触达 commonService。
func TestReadEchos_ShowPrivateResolution(t *testing.T) {
	viewerCases := []struct {
		name        string
		ctx         context.Context
		setupCommon func(common *commonmock.MockService)
		showPrivate bool
	}{
		{
			name:        "anonymous resolves to public-only",
			ctx:         helpers.CtxAnonymous(),
			setupCommon: func(*commonmock.MockService) {},
			showPrivate: false,
		},
		{
			name: "admin resolves to private-visible",
			ctx:  helpers.CtxAsUser(adminID),
			setupCommon: func(c *commonmock.MockService) {
				c.EXPECT().
					CommonGetUserByUserId(mock.Anything, adminID).
					Return(helpers.NewUser(helpers.AsAdmin), nil).
					Once()
			},
			showPrivate: true,
		},
		{
			name: "non-admin resolves to public-only",
			ctx:  helpers.CtxAsUser(userID),
			setupCommon: func(c *commonmock.MockService) {
				c.EXPECT().
					CommonGetUserByUserId(mock.Anything, userID).
					Return(helpers.NewUser(), nil).
					Once()
			},
			showPrivate: false,
		},
	}

	for _, m := range readMethods() {
		t.Run(m.name, func(t *testing.T) {
			for _, vc := range viewerCases {
				t.Run(vc.name, func(t *testing.T) {
					repo := echomock.NewMockRepository(t)
					common := commonmock.NewMockService(t)
					vc.setupCommon(common)
					m.expectRepo(repo, vc.showPrivate)

					svc := echoService.NewEchoService(nil, common, nil, repo, nilBus)
					require.NoError(t, m.invoke(svc, vc.ctx))
				})
			}
		})
	}
}

// TestReadEchos_UserLookupError 确认已认证调用方解析用户失败时原样上抛，且不触达仓储。
func TestReadEchos_UserLookupError(t *testing.T) {
	boom := errors.New("user lookup failed")
	for _, m := range readMethods() {
		t.Run(m.name, func(t *testing.T) {
			repo := echomock.NewMockRepository(t) // 无任何期望：仓储不应被调用
			common := commonmock.NewMockService(t)
			common.EXPECT().
				CommonGetUserByUserId(mock.Anything, adminID).
				Return(helpers.NewUser(helpers.AsAdmin), boom).
				Once()

			svc := echoService.NewEchoService(nil, common, nil, repo, nilBus)
			require.ErrorIs(t, m.invoke(svc, helpers.CtxAsUser(adminID)), boom)
		})
	}
}

// TestReadEchos_RepoErrorPropagates 确认带错误返回的只读方法把仓储错误上抛。
func TestReadEchos_RepoErrorPropagates(t *testing.T) {
	boom := errors.New("db down")

	t.Run("GetHotEchos", func(t *testing.T) {
		repo := echomock.NewMockRepository(t)
		common := commonmock.NewMockService(t)
		repo.EXPECT().GetHotEchos(3, false).Return(nil, boom).Once()

		svc := echoService.NewEchoService(nil, common, nil, repo, nilBus)
		_, err := svc.GetHotEchos(helpers.CtxAnonymous(), 3)
		require.ErrorIs(t, err, boom)
	})

	t.Run("GetRandomEcho", func(t *testing.T) {
		repo := echomock.NewMockRepository(t)
		common := commonmock.NewMockService(t)
		repo.EXPECT().GetRandomEcho(false).Return(nil, boom).Once()

		svc := echoService.NewEchoService(nil, common, nil, repo, nilBus)
		_, err := svc.GetRandomEcho(helpers.CtxAnonymous())
		require.ErrorIs(t, err, boom)
	})
}

// TestQueryEchos_ViewerResolution 覆盖 QueryEchos 的 showPrivate 解析：
// 管理员可见私密；已认证用户解析失败时上抛错误且不查询。
func TestQueryEchos_ViewerResolution(t *testing.T) {
	t.Run("admin queries with showPrivate true", func(t *testing.T) {
		repo := echomock.NewMockRepository(t)
		common := commonmock.NewMockService(t)
		common.EXPECT().
			CommonGetUserByUserId(mock.Anything, adminID).
			Return(helpers.NewUser(helpers.AsAdmin), nil).
			Once()
		repo.EXPECT().
			QueryEchos(mock.Anything, true).
			Return([]echoModel.Echo{helpers.NewEcho()}, int64(1), nil).
			Once()

		svc := echoService.NewEchoService(nil, common, nil, repo, nilBus)
		got, err := svc.QueryEchos(helpers.CtxAsUser(adminID), commonModel.EchoQueryDto{Page: 1, PageSize: 10})
		require.NoError(t, err)
		assert.Equal(t, int64(1), got.Total)
	})

	t.Run("user lookup error propagates", func(t *testing.T) {
		repo := echomock.NewMockRepository(t) // 不应触达 QueryEchos
		common := commonmock.NewMockService(t)
		boom := errors.New("lookup failed")
		common.EXPECT().
			CommonGetUserByUserId(mock.Anything, userID).
			Return(helpers.NewUser(), boom).
			Once()

		svc := echoService.NewEchoService(nil, common, nil, repo, nilBus)
		_, err := svc.QueryEchos(helpers.CtxAsUser(userID), commonModel.EchoQueryDto{Page: 1, PageSize: 10})
		require.ErrorIs(t, err, boom)
	})

	t.Run("repository error propagates", func(t *testing.T) {
		repo := echomock.NewMockRepository(t)
		common := commonmock.NewMockService(t)
		boom := errors.New("query failed")
		repo.EXPECT().QueryEchos(mock.Anything, false).Return(nil, int64(0), boom).Once()

		svc := echoService.NewEchoService(nil, common, nil, repo, nilBus)
		_, err := svc.QueryEchos(helpers.CtxAnonymous(), commonModel.EchoQueryDto{Page: 1, PageSize: 10})
		require.ErrorIs(t, err, boom)
	})
}

// TestGetEchosByPage_DelegatesToQuery 确认弃用的分页接口透传 Page/PageSize/Search 给 QueryEchos。
func TestGetEchosByPage_DelegatesToQuery(t *testing.T) {
	repo := echomock.NewMockRepository(t)
	common := commonmock.NewMockService(t)

	var captured commonModel.EchoQueryDto
	repo.EXPECT().
		QueryEchos(mock.Anything, false).
		Run(func(dto commonModel.EchoQueryDto, _ bool) { captured = dto }).
		Return([]echoModel.Echo{helpers.NewEcho()}, int64(1), nil).
		Once()

	svc := echoService.NewEchoService(nil, common, nil, repo, nilBus)
	got, err := svc.GetEchosByPage(helpers.CtxAnonymous(), commonModel.PageQueryDto{
		Page: 2, PageSize: 5, Search: "hi",
	})
	require.NoError(t, err)
	assert.Equal(t, int64(1), got.Total)
	assert.Len(t, got.Items, 1)
	assert.Equal(t, 2, captured.Page)
	assert.Equal(t, 5, captured.PageSize)
	assert.Equal(t, "hi", captured.Search)
	assert.Empty(t, captured.TagIDs)
}

// TestGetEchosByTagId_DelegatesToQuery 确认弃用的按标签接口把 tagId 放进 TagIDs 转发给 QueryEchos。
func TestGetEchosByTagId_DelegatesToQuery(t *testing.T) {
	repo := echomock.NewMockRepository(t)
	common := commonmock.NewMockService(t)

	var captured commonModel.EchoQueryDto
	repo.EXPECT().
		QueryEchos(mock.Anything, false).
		Run(func(dto commonModel.EchoQueryDto, _ bool) { captured = dto }).
		Return([]echoModel.Echo{}, int64(0), nil).
		Once()

	svc := echoService.NewEchoService(nil, common, nil, repo, nilBus)
	_, err := svc.GetEchosByTagId(helpers.CtxAnonymous(), "tag-7", commonModel.PageQueryDto{
		Page: 1, PageSize: 20,
	})
	require.NoError(t, err)
	assert.Equal(t, []string{"tag-7"}, captured.TagIDs)
}

// TestGetAllTags 确认标签列表直通仓储并传播错误。
func TestGetAllTags(t *testing.T) {
	t.Run("returns repository tags", func(t *testing.T) {
		repo := echomock.NewMockRepository(t)
		repo.EXPECT().GetAllTags().Return([]echoModel.Tag{{ID: "t1", Name: "go"}}, nil).Once()

		svc := echoService.NewEchoService(nil, nil, nil, repo, nilBus)
		got, err := svc.GetAllTags()
		require.NoError(t, err)
		require.Len(t, got, 1)
		assert.Equal(t, "go", got[0].Name)
	})

	t.Run("propagates repository error", func(t *testing.T) {
		repo := echomock.NewMockRepository(t)
		boom := errors.New("tags failed")
		repo.EXPECT().GetAllTags().Return(nil, boom).Once()

		svc := echoService.NewEchoService(nil, nil, nil, repo, nilBus)
		_, err := svc.GetAllTags()
		require.ErrorIs(t, err, boom)
	})
}
