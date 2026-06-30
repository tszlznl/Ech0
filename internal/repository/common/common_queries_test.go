// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package repository

import (
	"context"
	"errors"
	"testing"

	echoModel "github.com/lin-snow/ech0/internal/model/echo"
	fileModel "github.com/lin-snow/ech0/internal/model/file"
	userModel "github.com/lin-snow/ech0/internal/model/user"
	"github.com/lin-snow/ech0/internal/test/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// newCommonRepo 构造一个绑定到测试内存库的 CommonRepository。
func newCommonRepo(t *testing.T) (*CommonRepository, *gorm.DB) {
	t.Helper()
	db := helpers.NewTestDB(t)
	return NewCommonRepository(func() *gorm.DB { return db }), db
}

// seedEcho 插入一条 echo 行（显式 ID / created_at，绕开 autoCreateTime 以保证排序确定）。
func seedEcho(t *testing.T, db *gorm.DB, id, userID string, private bool, createdAt int64) {
	t.Helper()
	require.NoError(t, db.Create(&echoModel.Echo{
		ID:        id,
		Content:   "content-" + id,
		UserID:    userID,
		Private:   private,
		CreatedAt: createdAt,
	}).Error)
}

// seedFile 插入一条 external 类型 File 行（external 下 AfterFind 直接保留快照 URL，断言稳定）。
func seedFile(t *testing.T, db *gorm.DB, id, userID string) {
	t.Helper()
	require.NoError(t, db.Create(&fileModel.File{
		ID:          id,
		Key:         "key-" + id,
		StorageType: "external",
		URL:         "https://example.com/" + id,
		UserID:      userID,
		Category:    "image",
	}).Error)
}

// linkEchoFile 直接写 echo_files 关系行（携带 sort_order，便于断言 Preload 排序）。
func linkEchoFile(t *testing.T, db *gorm.DB, id, echoID, fileID string, sortOrder int) {
	t.Helper()
	require.NoError(t, db.Create(&fileModel.EchoFile{
		ID:        id,
		EchoID:    echoID,
		FileID:    fileID,
		SortOrder: sortOrder,
	}).Error)
}

func echoIDs(echos []echoModel.Echo) []string {
	ids := make([]string, len(echos))
	for i, e := range echos {
		ids[i] = e.ID
	}
	return ids
}

func TestCommonRepository_GetAllEchos_VisibilityAndOrder(t *testing.T) {
	repo, db := newCommonRepo(t)
	// created_at 升序：pub-old(100) < prv(200) < pub-new(300)；默认排序 created_at DESC。
	seedEcho(t, db, "pub-old", "u1", false, 100)
	seedEcho(t, db, "prv", "u1", true, 200)
	seedEcho(t, db, "pub-new", "u1", false, 300)

	t.Run("showPrivate=false excludes private, DESC order", func(t *testing.T) {
		echos, err := repo.GetAllEchos(context.Background(), false)
		require.NoError(t, err)
		assert.Equal(t, []string{"pub-new", "pub-old"}, echoIDs(echos))
	})

	t.Run("showPrivate=true includes private, DESC order", func(t *testing.T) {
		echos, err := repo.GetAllEchos(context.Background(), true)
		require.NoError(t, err)
		assert.Equal(t, []string{"pub-new", "prv", "pub-old"}, echoIDs(echos))
	})
}

func TestCommonRepository_GetAllEchos_PreloadFilesOrderedAndTags(t *testing.T) {
	repo, db := newCommonRepo(t)
	seedEcho(t, db, "e1", "u1", false, 100)

	// 两个文件，sort_order 故意倒序插入；Preload 须按 sort_order ASC 回放。
	seedFile(t, db, "f-a", "u1")
	seedFile(t, db, "f-b", "u1")
	linkEchoFile(t, db, "ef-b", "e1", "f-b", 2)
	linkEchoFile(t, db, "ef-a", "e1", "f-a", 1)

	// 一个标签，many2many 预载。
	require.NoError(t, db.Create(&echoModel.Tag{ID: "t1", Name: "alpha"}).Error)
	require.NoError(t, db.Create(&echoModel.EchoTag{EchoID: "e1", TagID: "t1"}).Error)

	echos, err := repo.GetAllEchos(context.Background(), false)
	require.NoError(t, err)
	require.Len(t, echos, 1)

	got := echos[0]
	// EchoFiles 按 sort_order ASC：f-a(1) 在 f-b(2) 之前。
	require.Len(t, got.EchoFiles, 2)
	assert.Equal(t, "f-a", got.EchoFiles[0].FileID)
	assert.Equal(t, "f-b", got.EchoFiles[1].FileID)
	// EchoFiles.File 被预载（external 保留快照 URL）。
	assert.Equal(t, "f-a", got.EchoFiles[0].File.ID)
	assert.Equal(t, "https://example.com/f-a", got.EchoFiles[0].File.URL)
	// Tags 被预载。
	require.Len(t, got.Tags, 1)
	assert.Equal(t, "alpha", got.Tags[0].Name)
}

func TestCommonRepository_GetAllEchos_Empty(t *testing.T) {
	repo, _ := newCommonRepo(t)
	echos, err := repo.GetAllEchos(context.Background(), true)
	require.NoError(t, err)
	assert.Empty(t, echos)
}

func TestCommonRepository_GetHeatMap_HalfOpenWindow(t *testing.T) {
	repo, db := newCommonRepo(t)
	seedEcho(t, db, "e100", "u1", false, 100)
	seedEcho(t, db, "e200", "u1", false, 200)
	seedEcho(t, db, "e300", "u1", false, 300)

	t.Run("start inclusive, end exclusive", func(t *testing.T) {
		// [100, 300) 应包含 100、200，排除 300。
		got, err := repo.GetHeatMap(context.Background(), 100, 300)
		require.NoError(t, err)
		assert.Equal(t, []int64{100, 200}, got)
	})

	t.Run("end past max includes all, ASC order", func(t *testing.T) {
		got, err := repo.GetHeatMap(context.Background(), 100, 301)
		require.NoError(t, err)
		assert.Equal(t, []int64{100, 200, 300}, got)
	})

	t.Run("lower bound is inclusive at exact start", func(t *testing.T) {
		// [200, 300) 排除 100（< start）与 300（>= end），仅留 200。
		got, err := repo.GetHeatMap(context.Background(), 200, 300)
		require.NoError(t, err)
		assert.Equal(t, []int64{200}, got)
	})

	t.Run("empty window returns no rows", func(t *testing.T) {
		got, err := repo.GetHeatMap(context.Background(), 1000, 2000)
		require.NoError(t, err)
		assert.Empty(t, got)
	})
}

func TestCommonRepository_GetOwner(t *testing.T) {
	repo, db := newCommonRepo(t)

	t.Run("no owner returns record-not-found", func(t *testing.T) {
		_, err := repo.GetOwner(context.Background())
		require.Error(t, err)
		assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))
	})

	t.Run("returns the owner among many users", func(t *testing.T) {
		require.NoError(t, db.Create(&userModel.User{ID: "u-normal", Username: "normal"}).Error)
		require.NoError(t, db.Create(&userModel.User{ID: "u-owner", Username: "owner", IsOwner: true, IsAdmin: true}).Error)

		owner, err := repo.GetOwner(context.Background())
		require.NoError(t, err)
		assert.Equal(t, "u-owner", owner.ID)
		assert.True(t, owner.IsOwner)
	})
}

func TestCommonRepository_GetAllUsers(t *testing.T) {
	repo, db := newCommonRepo(t)

	t.Run("empty table returns empty slice", func(t *testing.T) {
		users, err := repo.GetAllUsers(context.Background())
		require.NoError(t, err)
		assert.Empty(t, users)
	})

	t.Run("returns every user", func(t *testing.T) {
		require.NoError(t, db.Create(&userModel.User{ID: "u1", Username: "alice"}).Error)
		require.NoError(t, db.Create(&userModel.User{ID: "u2", Username: "bob"}).Error)

		users, err := repo.GetAllUsers(context.Background())
		require.NoError(t, err)
		assert.Len(t, users, 2)
	})
}

func TestCommonRepository_GetUserByUserId(t *testing.T) {
	repo, db := newCommonRepo(t)
	require.NoError(t, db.Create(&userModel.User{ID: "u-hit", Username: "hit"}).Error)

	t.Run("hit returns the user", func(t *testing.T) {
		user, err := repo.GetUserByUserId(context.Background(), "u-hit")
		require.NoError(t, err)
		assert.Equal(t, "hit", user.Username)
	})

	t.Run("miss returns record-not-found", func(t *testing.T) {
		_, err := repo.GetUserByUserId(context.Background(), "nope")
		require.Error(t, err)
		assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))
	})
}
