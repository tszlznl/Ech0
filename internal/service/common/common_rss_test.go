// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service_test

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lin-snow/ech0/internal/cache"
	echoModel "github.com/lin-snow/ech0/internal/model/echo"
	fileModel "github.com/lin-snow/ech0/internal/model/file"
	commonService "github.com/lin-snow/ech0/internal/service/common"
	commonmock "github.com/lin-snow/ech0/internal/test/mocks/commonmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// fakeCache 是 cache.ICache 的极简内存实现，避免 ristretto 的异步最终一致性带来测试抖动。
type fakeCache struct {
	mu   sync.Mutex
	data map[string]any
}

var _ cache.ICache[string, any] = (*fakeCache)(nil)

func newFakeCache() *fakeCache {
	return &fakeCache{data: make(map[string]any)}
}

func (c *fakeCache) Set(key string, value any, _ int64) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[key] = value
	return true
}

func (c *fakeCache) SetWithTTL(key string, value any, cost int64, _ time.Duration) bool {
	return c.Set(key, value, cost)
}

func (c *fakeCache) Get(key string) (any, bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	v, ok := c.data[key]
	return v, ok, nil
}

func (c *fakeCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.data, key)
}

func (c *fakeCache) Close() error { return nil }

// newRSSContext 构造一个最小可用的 *gin.Context（仅设置 Request.Host / 明文 http）。
func newRSSContext(t *testing.T, host string) *gin.Context {
	t.Helper()
	gin.SetMode(gin.TestMode)
	req := httptest.NewRequest(http.MethodGet, "http://"+host+"/rss", nil)
	req.Host = host
	req.TLS = nil // 明文 → schema=http
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request = req
	return ctx
}

// TestGenerateRSS_NormalFeed 正常路径：渲染出带 XSLT 样式声明的 Atom，含条目内容，
// 并把缓存键登记到 TrackRSSCacheKey。
func TestGenerateRSS_NormalFeed(t *testing.T) {
	repo := commonmock.NewMockCommonRepository(t)
	svc := commonService.NewCommonService(repo, newFakeCache())

	echos := []echoModel.Echo{
		{
			ID:        "echo-1",
			Username:  "alice",
			Content:   "hello rss world",
			CreatedAt: time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC).Unix(),
		},
	}

	repo.EXPECT().GetAllEchos(mock.Anything, false).Return(echos, nil).Once()
	repo.EXPECT().TrackRSSCacheKey("rss:http:example.com").Return().Once()

	ctx := newRSSContext(t, "example.com")
	atom, err := svc.GenerateRSS(ctx)
	require.NoError(t, err)

	// XSLT 样式表声明被注入到 XML 声明之后。
	assert.Contains(t, atom, `<?xml-stylesheet type="text/xsl" href="/rss.xsl"?>`)
	// Feed 元信息与条目内容。
	assert.Contains(t, atom, "<title>Ech0</title>")
	assert.Contains(t, atom, "hello rss world")
	assert.Contains(t, atom, "alice")
	// 条目链接含 host 与 echo id。
	assert.Contains(t, atom, "http://example.com/echo/echo-1")
}

// TestGenerateRSS_TagHTMLEntityEscaping 绑定 GHSA-3v85-fqvh-7rxf：
// 标签名进入 <summary type="html"> 前必须先做 HTML 实体转义，阻断订阅器二次解码触发的 stored XSS。
func TestGenerateRSS_TagHTMLEntityEscaping(t *testing.T) {
	repo := commonmock.NewMockCommonRepository(t)
	svc := commonService.NewCommonService(repo, newFakeCache())

	echos := []echoModel.Echo{
		{
			ID:        "echo-xss",
			Username:  "mallory",
			Content:   "benign body",
			CreatedAt: time.Now().UTC().Unix(),
			Tags: []echoModel.Tag{
				{Name: `<script>alert(1)</script>`},
			},
		},
	}

	repo.EXPECT().GetAllEchos(mock.Anything, false).Return(echos, nil).Once()
	repo.EXPECT().TrackRSSCacheKey(mock.Anything).Return().Once()

	ctx := newRSSContext(t, "example.com")
	atom, err := svc.GenerateRSS(ctx)
	require.NoError(t, err)

	// 不得出现原始 <script>。
	assert.NotContains(t, atom, "<script>", "RSS 不应含原始 script 标签")
	// 关键回归断言：若漏掉 HTML 实体转义，标签的 '<' 仅经一次 XML 转义会变成 &lt;script&gt;，
	// 订阅器解码一次即得到可执行的 <script>。转义到位后此单层形态不应出现。
	assert.NotContains(t, atom, "&lt;script&gt;", "标签名必须先做 HTML 实体转义，杜绝单层转义形态")
	// 正向证据：标签先 HTML 转义(&lt;)再被 Atom 序列化 XML 转义(&amp;)，呈现为双层转义形态。
	assert.Contains(t, atom, "&amp;lt;script&amp;gt;", "应为双层转义，证明 HTML 实体转义已生效")
}

// TestGenerateRSS_RendersEchoImages EchoFiles 会被渲染为内联 <img>，src 取文件直链快照。
func TestGenerateRSS_RendersEchoImages(t *testing.T) {
	repo := commonmock.NewMockCommonRepository(t)
	svc := commonService.NewCommonService(repo, newFakeCache())

	echos := []echoModel.Echo{
		{
			ID:        "echo-img",
			Username:  "bob",
			Content:   "look at this",
			CreatedAt: time.Now().UTC().Unix(),
			EchoFiles: []echoModel.EchoFile{
				{File: fileModel.File{URL: "http://example.com/files/pic.png"}},
			},
		},
	}

	repo.EXPECT().GetAllEchos(mock.Anything, false).Return(echos, nil).Once()
	repo.EXPECT().TrackRSSCacheKey(mock.Anything).Return().Once()

	ctx := newRSSContext(t, "example.com")
	atom, err := svc.GenerateRSS(ctx)
	require.NoError(t, err)

	assert.Contains(t, atom, "http://example.com/files/pic.png", "图片直链应出现在条目描述里")
	assert.Contains(t, atom, "look at this")
}

// TestGenerateRSS_ReadThrough 读穿透：相同 host 第二次调用命中缓存，不再回源仓库。
func TestGenerateRSS_ReadThrough(t *testing.T) {
	repo := commonmock.NewMockCommonRepository(t)
	svc := commonService.NewCommonService(repo, newFakeCache())

	echos := []echoModel.Echo{{ID: "e1", Username: "u", Content: "c", CreatedAt: time.Now().UTC().Unix()}}

	// GetAllEchos 与 TrackRSSCacheKey 都只允许发生一次。
	repo.EXPECT().GetAllEchos(mock.Anything, false).Return(echos, nil).Once()
	repo.EXPECT().TrackRSSCacheKey("rss:http:example.com").Return().Once()

	ctx := newRSSContext(t, "example.com")

	first, err := svc.GenerateRSS(ctx)
	require.NoError(t, err)
	second, err := svc.GenerateRSS(ctx)
	require.NoError(t, err)

	assert.Equal(t, first, second, "缓存命中应返回与首回相同的内容")
}

// TestGenerateRSS_RepositoryError 仓库取数据失败时透传错误，且不登记缓存键。
func TestGenerateRSS_RepositoryError(t *testing.T) {
	repo := commonmock.NewMockCommonRepository(t)
	svc := commonService.NewCommonService(repo, newFakeCache())

	repo.EXPECT().GetAllEchos(mock.Anything, false).Return(nil, assert.AnError).Once()
	// 不设置 TrackRSSCacheKey 期望：mock 会校验它确实未被调用。

	ctx := newRSSContext(t, "example.com")
	atom, err := svc.GenerateRSS(ctx)
	require.Error(t, err)
	assert.ErrorIs(t, err, assert.AnError)
	assert.Empty(t, atom)
}
