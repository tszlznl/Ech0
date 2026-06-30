// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package humares

import (
	"bytes"
	"compress/gzip"
	_ "embed"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// 本文件用 Scalar API Reference 离线接管 docs UI（替换 Huma 内置的 Stoplight Elements）。
// 资源完全 self-host、不联网 CDN，符合「单二进制、离线可用」的部署风格。
//
// 资源来源（standalone IIFE，自包含、无运行时外链/动态 chunk/外部字体）：
//
//	https://cdn.jsdelivr.net/npm/@scalar/api-reference@1.62.0/dist/browser/standalone.js
//
// 仓库里存的是 gzip -9 预压版本（raw ~3.7MB → ~1MB），按 Accept-Encoding 决定直接吐 gz
// 还是即时解压。升级方式：重新下载上方 standalone.js → `gzip -9 -c` 覆盖
// assets/scalar.standalone.js.gz，并同步更新此处版本号注释。
const scalarVersion = "1.62.0"

//go:embed assets/scalar.standalone.js.gz
var scalarBundleGz []byte

// DocsRenderer 选择 /api/docs 的文档面板。默认 Stoplight（Huma 内置），Scalar 为离线自托管可选项。
type DocsRenderer string

const (
	// DocsRendererStoplight 用 Huma 内置面板（Stoplight Elements）——默认。
	DocsRendererStoplight DocsRenderer = "stoplight"
	// DocsRendererScalar 用本文件自托管的离线 Scalar。
	DocsRendererScalar DocsRenderer = "scalar"
)

// ParseDocsRenderer 把配置里的原始字符串归一化为受支持的渲染器；未知值一律回退到默认（Stoplight）。
func ParseDocsRenderer(s string) DocsRenderer {
	if strings.EqualFold(strings.TrimSpace(s), string(DocsRendererScalar)) {
		return DocsRendererScalar
	}
	return DocsRendererStoplight
}

// 以下三条均为「Huma group 前缀（basePath）之后」的相对路径；页面里要用的绝对地址由
// buildScalarHTML 拼 basePath 得到（如 basePath=/api → /api/docs、/api/docs/scalar.standalone.js）。
const (
	// scalarDocsRoute 是 docs 页路由。
	scalarDocsRoute = "/docs"
	// scalarScriptRoute 是离线 bundle 路由。
	scalarScriptRoute = "/docs/scalar.standalone.js"
	// scalarSpecRoute 是已有的 OpenAPI spec 路由（humagin 默认 OpenAPIPath=/openapi）。
	scalarSpecRoute = "/openapi.json"
)

// registerScalarDocs 在 Huma 的 group 上挂 docs UI 与离线 bundle 两条路由。
// basePath 即该 group 的前缀（/api），用于把页面内的脚本/spec 拼成根绝对路径。
// 调用前须把 huma.Config.DocsPath 置空，避免与 Huma 内置 docs 路由冲突。
func registerScalarDocs(group *gin.RouterGroup, basePath string) {
	html := buildScalarHTML(basePath)
	group.GET(scalarDocsRoute, func(ctx *gin.Context) {
		ctx.Header("Cache-Control", "no-cache")
		ctx.Data(http.StatusOK, "text/html; charset=utf-8", html)
	})
	group.GET(scalarScriptRoute, serveScalarBundle)
}

// buildScalarHTML 生成 docs 页：内联调用 createApiReference，脚本与 spec 均走 basePath 下的绝对路径。
func buildScalarHTML(basePath string) []byte {
	scriptSrc := basePath + scalarScriptRoute // 如 /api/docs/scalar.standalone.js
	specURL := basePath + scalarSpecRoute     // 如 /api/openapi.json
	return []byte(`<!doctype html>
<!-- Scalar API Reference (self-hosted) v` + scalarVersion + ` -->
<html lang="zh-CN">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Ech0 API 文档</title>
</head>
<body>
  <div id="app"></div>
  <script src="` + scriptSrc + `"></script>
  <script>
    Scalar.createApiReference('#app', { url: '` + specURL + `' })
  </script>
</body>
</html>`)
}

// serveScalarBundle 提供离线 bundle：客户端支持 gzip 时直接吐预压字节，否则即时解压。
func serveScalarBundle(ctx *gin.Context) {
	const contentType = "application/javascript; charset=utf-8"
	ctx.Header("Cache-Control", "public, max-age=86400")
	ctx.Header("Vary", "Accept-Encoding")

	if strings.Contains(ctx.GetHeader("Accept-Encoding"), "gzip") {
		ctx.Header("Content-Encoding", "gzip")
		ctx.Data(http.StatusOK, contentType, scalarBundleGz)
		return
	}

	reader, err := gzip.NewReader(bytes.NewReader(scalarBundleGz))
	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		return
	}
	defer func() { _ = reader.Close() }()
	ctx.Header("Content-Type", contentType)
	ctx.Status(http.StatusOK)
	_, _ = io.Copy(ctx.Writer, reader)
}
