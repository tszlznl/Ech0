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
	txmock "github.com/lin-snow/ech0/internal/test/mocks/txmock"
	"github.com/lin-snow/ech0/pkg/busen"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// nilBus 满足 NewEchoService 的 busProvider 入参；被测方法（GetEchoById/LikeEcho/QueryEchos）
// 都不向 bus 发事件，因此返回 nil 即可。
func nilBus() *busen.Bus { return nil }

const (
	adminID = "admin-0001"
	userID  = "user-0002"
	echoID  = "echo-0001"
)

// TestGetEchoById_Visibility 覆盖私密 echo 的可见性规则：
// 匿名/非作者非管理员看不到 private；管理员可见；公开 echo 任何人可见。
func TestGetEchoById_Visibility(t *testing.T) {
	t.Run("anonymous cannot read private echo", func(t *testing.T) {
		repo := echomock.NewMockRepository(t)
		common := commonmock.NewMockService(t)
		private := helpers.NewEcho(helpers.AsPrivate)
		repo.EXPECT().GetEchosById(mock.Anything, echoID).Return(&private, nil).Once()

		svc := echoService.NewEchoService(nil, common, nil, repo, nilBus)
		got, err := svc.GetEchoById(helpers.CtxAnonymous(), echoID)

		require.Error(t, err)
		require.EqualError(t, err, commonModel.NO_PERMISSION_DENIED)
		assert.Nil(t, got)
	})

	t.Run("anonymous can read public echo", func(t *testing.T) {
		repo := echomock.NewMockRepository(t)
		common := commonmock.NewMockService(t)
		public := helpers.NewEcho()
		repo.EXPECT().GetEchosById(mock.Anything, echoID).Return(&public, nil).Once()

		svc := echoService.NewEchoService(nil, common, nil, repo, nilBus)
		got, err := svc.GetEchoById(helpers.CtxAnonymous(), echoID)

		require.NoError(t, err)
		require.NotNil(t, got)
		assert.False(t, got.Private)
	})

	t.Run("non-admin user cannot read private echo", func(t *testing.T) {
		repo := echomock.NewMockRepository(t)
		common := commonmock.NewMockService(t)
		private := helpers.NewEcho(helpers.AsPrivate)
		repo.EXPECT().GetEchosById(mock.Anything, echoID).Return(&private, nil).Once()
		common.EXPECT().
			CommonGetUserByUserId(mock.Anything, userID).
			Return(helpers.NewUser(), nil).
			Once()

		svc := echoService.NewEchoService(nil, common, nil, repo, nilBus)
		got, err := svc.GetEchoById(helpers.CtxAsUser(userID), echoID)

		require.EqualError(t, err, commonModel.NO_PERMISSION_DENIED)
		assert.Nil(t, got)
	})

	t.Run("admin can read private echo", func(t *testing.T) {
		repo := echomock.NewMockRepository(t)
		common := commonmock.NewMockService(t)
		private := helpers.NewEcho(helpers.AsPrivate)
		repo.EXPECT().GetEchosById(mock.Anything, echoID).Return(&private, nil).Once()
		common.EXPECT().
			CommonGetUserByUserId(mock.Anything, adminID).
			Return(helpers.NewUser(helpers.AsAdmin), nil).
			Once()

		svc := echoService.NewEchoService(nil, common, nil, repo, nilBus)
		got, err := svc.GetEchoById(helpers.CtxAsUser(adminID), echoID)

		require.NoError(t, err)
		require.NotNil(t, got)
		assert.True(t, got.Private)
	})

	t.Run("not found returns ECHO_NOT_FOUND", func(t *testing.T) {
		repo := echomock.NewMockRepository(t)
		common := commonmock.NewMockService(t)
		repo.EXPECT().GetEchosById(mock.Anything, echoID).Return(nil, nil).Once()

		svc := echoService.NewEchoService(nil, common, nil, repo, nilBus)
		got, err := svc.GetEchoById(helpers.CtxAnonymous(), echoID)

		require.EqualError(t, err, commonModel.ECHO_NOT_FOUND)
		assert.Nil(t, got)
	})

	t.Run("repository error is propagated", func(t *testing.T) {
		repo := echomock.NewMockRepository(t)
		common := commonmock.NewMockService(t)
		repoErr := errors.New("db down")
		repo.EXPECT().GetEchosById(mock.Anything, echoID).Return(nil, repoErr).Once()

		svc := echoService.NewEchoService(nil, common, nil, repo, nilBus)
		got, err := svc.GetEchoById(helpers.CtxAsUser(adminID), echoID)

		require.ErrorIs(t, err, repoErr)
		assert.Nil(t, got)
	})
}

// runTx 让 MockTransactor.Run 真正执行内部回调，从而触发 repo.LikeEcho。
func runTx(_ context.Context, fn func(ctx context.Context) error) error {
	return fn(context.Background())
}

// TestLikeEcho_Visibility 覆盖点赞的私密可见性规则，应与 GetEchoById 一致。
func TestLikeEcho_Visibility(t *testing.T) {
	t.Run("anonymous cannot like private echo", func(t *testing.T) {
		repo := echomock.NewMockRepository(t)
		common := commonmock.NewMockService(t)
		private := helpers.NewEcho(helpers.AsPrivate)
		repo.EXPECT().GetEchosById(mock.Anything, echoID).Return(&private, nil).Once()

		svc := echoService.NewEchoService(nil, common, nil, repo, nilBus)
		err := svc.LikeEcho(helpers.CtxAnonymous(), echoID)

		require.EqualError(t, err, commonModel.NO_PERMISSION_DENIED)
	})

	t.Run("non-admin user cannot like private echo", func(t *testing.T) {
		repo := echomock.NewMockRepository(t)
		common := commonmock.NewMockService(t)
		private := helpers.NewEcho(helpers.AsPrivate)
		repo.EXPECT().GetEchosById(mock.Anything, echoID).Return(&private, nil).Once()
		common.EXPECT().
			CommonGetUserByUserId(mock.Anything, userID).
			Return(helpers.NewUser(), nil).
			Once()

		svc := echoService.NewEchoService(nil, common, nil, repo, nilBus)
		err := svc.LikeEcho(helpers.CtxAsUser(userID), echoID)

		require.EqualError(t, err, commonModel.NO_PERMISSION_DENIED)
	})

	t.Run("admin can like private echo", func(t *testing.T) {
		repo := echomock.NewMockRepository(t)
		common := commonmock.NewMockService(t)
		tx := txmock.NewMockTransactor(t)
		private := helpers.NewEcho(helpers.AsPrivate)
		repo.EXPECT().GetEchosById(mock.Anything, echoID).Return(&private, nil).Once()
		common.EXPECT().
			CommonGetUserByUserId(mock.Anything, adminID).
			Return(helpers.NewUser(helpers.AsAdmin), nil).
			Once()
		tx.EXPECT().Run(mock.Anything, mock.Anything).RunAndReturn(runTx).Once()
		repo.EXPECT().LikeEcho(mock.Anything, echoID).Return(nil).Once()
		repo.EXPECT().InvalidateEchoCaches(echoID).Once()

		svc := echoService.NewEchoService(tx, common, nil, repo, nilBus)
		err := svc.LikeEcho(helpers.CtxAsUser(adminID), echoID)

		require.NoError(t, err)
	})

	t.Run("anonymous can like public echo", func(t *testing.T) {
		repo := echomock.NewMockRepository(t)
		common := commonmock.NewMockService(t)
		tx := txmock.NewMockTransactor(t)
		public := helpers.NewEcho()
		repo.EXPECT().GetEchosById(mock.Anything, echoID).Return(&public, nil).Once()
		tx.EXPECT().Run(mock.Anything, mock.Anything).RunAndReturn(runTx).Once()
		repo.EXPECT().LikeEcho(mock.Anything, echoID).Return(nil).Once()
		repo.EXPECT().InvalidateEchoCaches(echoID).Once()

		svc := echoService.NewEchoService(tx, common, nil, repo, nilBus)
		err := svc.LikeEcho(helpers.CtxAnonymous(), echoID)

		require.NoError(t, err)
	})

	t.Run("not found returns ECHO_NOT_FOUND", func(t *testing.T) {
		repo := echomock.NewMockRepository(t)
		common := commonmock.NewMockService(t)
		repo.EXPECT().GetEchosById(mock.Anything, echoID).Return(nil, nil).Once()

		svc := echoService.NewEchoService(nil, common, nil, repo, nilBus)
		err := svc.LikeEcho(helpers.CtxAnonymous(), echoID)

		require.EqualError(t, err, commonModel.ECHO_NOT_FOUND)
	})
}

// TestQueryEchos_PageSizeClamp 守护公开 /echo/query 端点的 DoS 护栏：
// pageSize<1 回落 10，>100 钳到 100，区间内原样保留；page<1 钳到 1。
func TestQueryEchos_PageSizeClamp(t *testing.T) {
	cases := []struct {
		name         string
		inPage       int
		inPageSize   int
		wantPage     int
		wantPageSize int
	}{
		{"zero pagesize falls back to 10", 1, 0, 1, 10},
		{"negative pagesize falls back to 10", 1, -5, 1, 10},
		{"oversized pagesize clamps to 100", 1, 500, 1, 100},
		{"exactly 100 is kept", 1, 100, 1, 100},
		{"in-range pagesize is kept", 2, 50, 2, 50},
		{"zero page clamps to 1", 0, 20, 1, 20},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := echomock.NewMockRepository(t)
			common := commonmock.NewMockService(t)

			var captured commonModel.EchoQueryDto
			repo.EXPECT().
				QueryEchos(mock.Anything, false).
				Run(func(dto commonModel.EchoQueryDto, _ bool) { captured = dto }).
				Return([]echoModel.Echo{}, int64(0), nil).
				Once()

			svc := echoService.NewEchoService(nil, common, nil, repo, nilBus)
			_, err := svc.QueryEchos(helpers.CtxAnonymous(), commonModel.EchoQueryDto{
				Page:     tc.inPage,
				PageSize: tc.inPageSize,
			})

			require.NoError(t, err)
			assert.Equal(t, tc.wantPageSize, captured.PageSize, "pageSize clamp")
			assert.Equal(t, tc.wantPage, captured.Page, "page clamp")
			// 默认排序兜底也应生效
			assert.Equal(t, "created_at", captured.SortBy)
			assert.Equal(t, "desc", captured.SortOrder)
		})
	}
}
