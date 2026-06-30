// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package humares

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() { gin.SetMode(gin.TestMode) }

// 测试用的 group 前缀，与生产 humaAPIBase 保持一致。
const testBasePath = "/api"

// 页面里实际会请求的绝对路径（= basePath + 组内相对路由）。
var (
	scriptAbsPath = testBasePath + scalarScriptRoute // /api/docs/scalar.standalone.js
	specAbsURL    = testBasePath + scalarSpecRoute   // /api/openapi.json
)

// newDocsRouter 建一个仅注册 Scalar docs 路由的引擎（复用生产注册逻辑）。
func newDocsRouter() *gin.Engine {
	r := gin.New()
	registerScalarDocs(r.Group(testBasePath), testBasePath)
	return r
}

// newAPIRouter 走完整 NewAPI 装配（按 renderer 决定挂内置 docs 还是 Scalar），用于验证默认/可选切换。
func newAPIRouter(docs DocsRenderer) *gin.Engine {
	r := gin.New()
	NewAPI(r, r.Group(testBasePath), "test", "1.0", testBasePath, docs)
	return r
}

func TestParseDocsRenderer(t *testing.T) {
	cases := map[string]DocsRenderer{
		"scalar":     DocsRendererScalar,
		"Scalar":     DocsRendererScalar,
		"  scalar  ": DocsRendererScalar,
		"stoplight":  DocsRendererStoplight,
		"":           DocsRendererStoplight,
		"swagger-ui": DocsRendererStoplight, // 未知值回退默认
		"nonsense":   DocsRendererStoplight,
	}
	for in, want := range cases {
		if got := ParseDocsRenderer(in); got != want {
			t.Errorf("ParseDocsRenderer(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestNewAPI_DefaultUsesHumaBuiltinDocs(t *testing.T) {
	r := newAPIRouter(DocsRendererStoplight)

	// 默认应保留 Huma 内置 docs（Stoplight），且不注册自托管 Scalar bundle 路由。
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, testBasePath+"/docs", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("built-in docs: expected 200, got %d", rec.Code)
	}
	if !strings.Contains(strings.ToLower(rec.Body.String()), "stoplight") {
		t.Fatalf("built-in docs should render Stoplight; got: %s", rec.Body.String())
	}

	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, testBasePath+scalarScriptRoute, nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("scalar bundle must NOT be registered under default; got %d", rec.Code)
	}
}

func TestNewAPI_ScalarRendererServesScalar(t *testing.T) {
	r := newAPIRouter(DocsRendererScalar)

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, testBasePath+"/docs", nil))
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "Scalar.createApiReference") {
		t.Fatalf("scalar renderer should serve Scalar HTML; code=%d body=%s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, testBasePath+scalarScriptRoute, nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("scalar bundle should be served; got %d", rec.Code)
	}
}

func TestScalarDocs_HTMLReferencesLocalAssets(t *testing.T) {
	r := newDocsRouter()

	req := httptest.NewRequest(http.MethodGet, testBasePath+scalarDocsRoute, nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/html") {
		t.Fatalf("expected text/html, got %q", ct)
	}
	body := rec.Body.String()
	// 必须自托管：脚本指向本地离线 bundle 的根绝对路径，且不出现联网 CDN。
	if !strings.Contains(body, `src="`+scriptAbsPath+`"`) {
		t.Fatalf("docs HTML missing local bundle path %q; got: %s", scriptAbsPath, body)
	}
	if strings.Contains(body, "cdn.jsdelivr.net") || strings.Contains(body, "unpkg.com") {
		t.Fatalf("docs HTML must not reference any CDN; got: %s", body)
	}
	if !strings.Contains(body, "Scalar.createApiReference") || !strings.Contains(body, specAbsURL) {
		t.Fatalf("docs HTML missing Scalar init / spec url; got: %s", body)
	}
}

func TestScalarBundle_GzipPassthrough(t *testing.T) {
	r := newDocsRouter()

	req := httptest.NewRequest(http.MethodGet, scriptAbsPath, nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if enc := rec.Header().Get("Content-Encoding"); enc != "gzip" {
		t.Fatalf("expected Content-Encoding gzip, got %q", enc)
	}
	// 直吐的应当与 embed 的预压字节一致，且能解压回真实 bundle。
	if !bytes.Equal(rec.Body.Bytes(), scalarBundleGz) {
		t.Fatalf("gzip passthrough body differs from embedded asset")
	}
	gz, err := gzip.NewReader(bytes.NewReader(rec.Body.Bytes()))
	if err != nil {
		t.Fatalf("response is not valid gzip: %v", err)
	}
	out, _ := io.ReadAll(gz)
	if !bytes.Contains(out, []byte("createApiReference")) {
		t.Fatalf("decompressed bundle missing createApiReference")
	}
}

func TestScalarBundle_PlainFallbackDecompresses(t *testing.T) {
	r := newDocsRouter()

	req := httptest.NewRequest(http.MethodGet, scriptAbsPath, nil)
	req.Header.Set("Accept-Encoding", "identity")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if enc := rec.Header().Get("Content-Encoding"); enc != "" {
		t.Fatalf("expected no Content-Encoding for plain fallback, got %q", enc)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/javascript") {
		t.Fatalf("expected javascript content-type, got %q", ct)
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("createApiReference")) {
		t.Fatalf("plain response missing decompressed bundle content")
	}
}
