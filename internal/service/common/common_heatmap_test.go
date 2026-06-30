// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service_test

import (
	"context"
	"testing"
	"time"
	_ "time/tzdata" // 内嵌 IANA 时区库，保证 LoadLocation("Asia/Shanghai") 在任意平台可用

	commonModel "github.com/lin-snow/ech0/internal/model/common"
	commonService "github.com/lin-snow/ech0/internal/service/common"
	commonmock "github.com/lin-snow/ech0/internal/test/mocks/commonmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const secondsPerDay = int64(24 * 60 * 60)

// countForDate 在热力图结果里按日期查计数；返回是否命中该日期格子。
func countForDate(t *testing.T, hm []commonModel.Heatmap, date string) (int, bool) {
	t.Helper()
	for _, h := range hm {
		if h.Date == date {
			return h.Count, true
		}
	}
	return 0, false
}

// sumCounts 汇总热力图所有格子的计数。
func sumCounts(hm []commonModel.Heatmap) int {
	total := 0
	for _, h := range hm {
		total += h.Count
	}
	return total
}

// TestGetHeatMap_BucketingStructure 校验 UTC 下的桶结构：固定 30 天窗口、
// 仓库时间戳按本地日正确归桶、且向仓库请求的查询区间恰为 [本地午夜, +30天)。
func TestGetHeatMap_BucketingStructure(t *testing.T) {
	repo := commonmock.NewMockCommonRepository(t)
	svc := commonService.NewCommonService(repo, nil) // GetHeatMap 不触碰 cache，传 nil 安全

	loc := time.UTC
	now := time.Now().In(loc)
	todayNoon := time.Date(now.Year(), now.Month(), now.Day(), 12, 0, 0, 0, loc)

	// 3 条今天、2 条昨天、1 条窗口最早一天(29 天前)，其余应全为 0。
	timestamps := []int64{
		todayNoon.Unix(),
		todayNoon.Unix(),
		todayNoon.Unix(),
		todayNoon.AddDate(0, 0, -1).Unix(),
		todayNoon.AddDate(0, 0, -1).Unix(),
		todayNoon.AddDate(0, 0, -29).Unix(),
	}

	var gotStart, gotEnd int64
	repo.EXPECT().
		GetHeatMap(mock.Anything, mock.Anything, mock.Anything).
		Run(func(_ context.Context, start int64, end int64) {
			gotStart, gotEnd = start, end
		}).
		Return(timestamps, nil).
		Once()

	hm, err := svc.GetHeatMap("UTC")
	require.NoError(t, err)
	require.Len(t, hm, 30, "热力图固定返回 30 天")

	// 查询区间：宽度恰为 30 天，起点对齐 UTC 本地午夜。
	assert.Equal(t, 30*secondsPerDay, gotEnd-gotStart, "查询窗口宽度应为 30 天")
	assert.Zero(t, gotStart%secondsPerDay, "UTC 起点应对齐午夜")

	// 桶位：最新一格(index 29)是今天，倒数第二是昨天，第一格是 29 天前。
	assert.Equal(t, 3, hm[29].Count, "今天应有 3 条")
	assert.Equal(t, 2, hm[28].Count, "昨天应有 2 条")
	assert.Equal(t, 1, hm[0].Count, "29 天前应有 1 条")
	assert.Equal(t, 6, sumCounts(hm), "落在窗口内的计数总和")

	// 日期连续且递增。
	for i := 0; i < len(hm)-1; i++ {
		prev, perr := time.Parse("2006-01-02", hm[i].Date)
		require.NoError(t, perr)
		assert.Equal(t, prev.AddDate(0, 0, 1).Format("2006-01-02"), hm[i+1].Date, "相邻格子应相差一天")
	}
}

// TestGetHeatMap_CrossTimezoneBucketing 校验同一时间戳在不同时区下归入不同本地日期：
// 5 天前 23:00 UTC 这一刻，在 UTC 算作当天，在 Asia/Shanghai(UTC+8) 已是次日。
func TestGetHeatMap_CrossTimezoneBucketing(t *testing.T) {
	shLoc, err := time.LoadLocation("Asia/Shanghai")
	require.NoError(t, err)

	nowUTC := time.Now().UTC()
	base := nowUTC.AddDate(0, 0, -5) // 远离 30 天窗口两端，避免边界抖动
	instant := time.Date(base.Year(), base.Month(), base.Day(), 23, 0, 0, 0, time.UTC)

	utcDate := instant.Format("2006-01-02")          // UTC 视角：5 天前
	shDate := instant.In(shLoc).Format("2006-01-02") // 上海视角：4 天前(23:00+8h 跨日)
	require.NotEqual(t, utcDate, shDate, "构造的时刻应跨日，否则用例无意义")

	cases := []struct {
		name      string
		timezone  string
		hitDate   string
		emptyDate string
	}{
		{name: "UTC 归当天", timezone: "UTC", hitDate: utcDate, emptyDate: shDate},
		{name: "Shanghai 归次日", timezone: "Asia/Shanghai", hitDate: shDate, emptyDate: utcDate},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := commonmock.NewMockCommonRepository(t)
			svc := commonService.NewCommonService(repo, nil)

			repo.EXPECT().
				GetHeatMap(mock.Anything, mock.Anything, mock.Anything).
				Return([]int64{instant.Unix()}, nil).
				Once()

			hm, hmErr := svc.GetHeatMap(tc.timezone)
			require.NoError(t, hmErr)
			require.Len(t, hm, 30)

			hit, ok := countForDate(t, hm, tc.hitDate)
			require.True(t, ok, "命中日期 %s 应在窗口内", tc.hitDate)
			assert.Equal(t, 1, hit, "该时间戳应只落在 %s 这一格", tc.hitDate)

			empty, ok := countForDate(t, hm, tc.emptyDate)
			require.True(t, ok, "对照日期 %s 应在窗口内", tc.emptyDate)
			assert.Equal(t, 0, empty, "另一时区的日期 %s 不应计数", tc.emptyDate)

			assert.Equal(t, 1, sumCounts(hm), "整个窗口只有一条记录")
		})
	}
}

// TestGetHeatMap_RepositoryError 仓库出错时直接透传错误，不返回部分数据。
func TestGetHeatMap_RepositoryError(t *testing.T) {
	repo := commonmock.NewMockCommonRepository(t)
	svc := commonService.NewCommonService(repo, nil)

	wantErr := assert.AnError
	repo.EXPECT().
		GetHeatMap(mock.Anything, mock.Anything, mock.Anything).
		Return(nil, wantErr).
		Once()

	hm, err := svc.GetHeatMap("UTC")
	require.Error(t, err)
	assert.ErrorIs(t, err, wantErr)
	assert.Nil(t, hm)
}

// TestGetHeatMap_InvalidTimezoneFallsBackToUTC 非法时区名回退 UTC，仍返回完整 30 天。
func TestGetHeatMap_InvalidTimezoneFallsBackToUTC(t *testing.T) {
	repo := commonmock.NewMockCommonRepository(t)
	svc := commonService.NewCommonService(repo, nil)

	var gotStart int64
	repo.EXPECT().
		GetHeatMap(mock.Anything, mock.Anything, mock.Anything).
		Run(func(_ context.Context, start int64, _ int64) { gotStart = start }).
		Return([]int64{}, nil).
		Once()

	hm, err := svc.GetHeatMap("Not/A_Zone")
	require.NoError(t, err)
	require.Len(t, hm, 30)
	// 回退 UTC 后起点仍对齐午夜。
	assert.Zero(t, gotStart%secondsPerDay)
	assert.Zero(t, sumCounts(hm))
}
