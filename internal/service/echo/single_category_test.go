// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service_test

import (
	"context"
	"testing"

	commonModel "github.com/lin-snow/ech0/internal/model/common"
	echoModel "github.com/lin-snow/ech0/internal/model/echo"
	fileModel "github.com/lin-snow/ech0/internal/model/file"
	echoService "github.com/lin-snow/ech0/internal/service/echo"
	"github.com/lin-snow/ech0/internal/test/helpers"
	commonmock "github.com/lin-snow/ech0/internal/test/mocks/commonmock"
	echomock "github.com/lin-snow/ech0/internal/test/mocks/echomock"
	filemock "github.com/lin-snow/ech0/internal/test/mocks/filemock"
	txmock "github.com/lin-snow/ech0/internal/test/mocks/txmock"
	"github.com/lin-snow/ech0/pkg/busen"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// echoWithFileIDs 构造一条带若干 EchoFile（仅填 FileID）的 echo，用于覆盖单类别校验。
func echoWithFileIDs(ids ...string) *echoModel.Echo {
	files := make([]fileModel.EchoFile, 0, len(ids))
	for i, id := range ids {
		files = append(files, fileModel.EchoFile{FileID: id, SortOrder: i})
	}
	e := helpers.NewEcho(func(e *echoModel.Echo) {
		e.ID = echoID
		e.Content = "hi"
		e.EchoFiles = files
	})
	return &e
}

func fileDto(id, category string) commonModel.FileDto {
	return commonModel.FileDto{ID: id, Category: category}
}

// TestValidateSingleFileCategory_Reject 覆盖 PostEcho 在进入事务前因文件类别不合法而快速失败：
// 混合类别、多个音频、多个视频均被拒，且不触达 transactor / repo 写入。
func TestValidateSingleFileCategory_Reject(t *testing.T) {
	cases := []struct {
		name  string
		files []commonModel.FileDto
	}{
		{
			name:  "image mixed with audio rejected",
			files: []commonModel.FileDto{fileDto("f1", "image"), fileDto("f2", "audio")},
		},
		{
			name:  "image mixed with video rejected",
			files: []commonModel.FileDto{fileDto("f1", "image"), fileDto("f2", "video")},
		},
		{
			name:  "two audios rejected",
			files: []commonModel.FileDto{fileDto("f1", "audio"), fileDto("f2", "audio")},
		},
		{
			name:  "two videos rejected",
			files: []commonModel.FileDto{fileDto("f1", "video"), fileDto("f2", "video")},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := echomock.NewMockRepository(t)
			common := commonmock.NewMockService(t)
			file := filemock.NewMockService(t)
			common.EXPECT().
				CommonGetUserByUserId(mock.Anything, adminID).
				Return(helpers.NewUser(helpers.AsAdmin), nil).
				Once()
			file.EXPECT().
				GetFilesByIDs(mock.Anything, mock.Anything).
				Return(tc.files, nil).
				Once()

			svc := echoService.NewEchoService(nil, common, file, repo, nilBus)
			ids := make([]string, len(tc.files))
			for i, f := range tc.files {
				ids[i] = f.ID
			}
			err := svc.PostEcho(helpers.CtxAsUser(adminID), echoWithFileIDs(ids...))

			require.EqualError(t, err, commonModel.ECHO_MIXED_FILE_CATEGORIES)
		})
	}
}

// TestValidateSingleFileCategory_MultipleImagesAllowed 确认同类别多图不被拒，
// 校验通过后走完整成功路径（事务 / 建 echo / 缓存失效 / 回查 / 确认临时文件）。
func TestValidateSingleFileCategory_MultipleImagesAllowed(t *testing.T) {
	repo := echomock.NewMockRepository(t)
	common := commonmock.NewMockService(t)
	file := filemock.NewMockService(t)
	tx := txmock.NewMockTransactor(t)
	bus := helpers.NewTestBus(t)

	common.EXPECT().
		CommonGetUserByUserId(mock.Anything, adminID).
		Return(helpers.NewUser(helpers.AsAdmin), nil).
		Once()
	file.EXPECT().
		GetFilesByIDs(mock.Anything, mock.Anything).
		Return([]commonModel.FileDto{fileDto("f1", "image"), fileDto("f2", "image")}, nil).
		Once()

	tx.EXPECT().Run(mock.Anything, mock.Anything).RunAndReturn(runTx).Once()
	repo.EXPECT().GetTagsByNames(mock.Anything, mock.Anything).Return([]*echoModel.Tag{}, nil).Once()

	var created echoModel.Echo
	repo.EXPECT().
		CreateEcho(mock.Anything, mock.Anything).
		Run(func(_ context.Context, e *echoModel.Echo) { created = *e }).
		Return(nil).
		Once()
	repo.EXPECT().InvalidateEchoCaches().Once()
	saved := helpers.NewEcho(func(e *echoModel.Echo) { e.ID = "saved-1" })
	repo.EXPECT().GetEchosById(mock.Anything, mock.Anything).Return(&saved, nil).Once()
	file.EXPECT().ConfirmTempFiles(mock.Anything, mock.Anything).Return(nil).Once()

	svc := echoService.NewEchoService(tx, common, file, repo, func() *busen.Bus { return bus })
	require.NoError(t, svc.PostEcho(helpers.CtxAsUser(adminID), echoWithFileIDs("f1", "f2")))

	assert.Len(t, created.EchoFiles, 2)
}

// TestUpdateEcho_MixedFileCategoriesRejected 确认更新路径同样在进入事务前拒绝混合类别。
func TestUpdateEcho_MixedFileCategoriesRejected(t *testing.T) {
	repo := echomock.NewMockRepository(t)
	common := commonmock.NewMockService(t)
	file := filemock.NewMockService(t)

	common.EXPECT().
		CommonGetUserByUserId(mock.Anything, adminID).
		Return(helpers.NewUser(helpers.AsAdmin), nil).
		Once()
	file.EXPECT().
		GetFilesByIDs(mock.Anything, mock.Anything).
		Return([]commonModel.FileDto{fileDto("f1", "image"), fileDto("f2", "audio")}, nil).
		Once()

	svc := echoService.NewEchoService(nil, common, file, repo, nilBus)
	err := svc.UpdateEcho(helpers.CtxAsUser(adminID), echoWithFileIDs("f1", "f2"))

	require.EqualError(t, err, commonModel.ECHO_MIXED_FILE_CATEGORIES)
}
