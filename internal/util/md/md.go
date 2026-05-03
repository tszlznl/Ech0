// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package util

import (
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

// MdToHTML 渲染 Markdown 为 HTML
func MdToHTML(md []byte) []byte {
	// 创建 Markdown 解析器
	extensions := parser.NoIntraEmphasis |
		parser.Tables |
		parser.FencedCode |
		parser.Autolink |
		parser.Strikethrough |
		parser.SpaceHeadings |
		parser.BackslashLineBreak |
		parser.DefinitionLists |
		parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse(md)

	// 创建 HTML 渲染器
	// SkipHTML 丢弃 markdown 中的原始 HTML 块/内联，阻止 <script> 等标签被原样输出。
	// 该函数仅服务于 RSS Atom <summary type="html"> 渲染，前端正文走客户端 markdown-it，
	// 因此关闭原始 HTML 透传不会影响 Web UI。
	htmlFlags := html.CommonFlags |
		html.Safelink |
		html.HrefTargetBlank |
		html.NoopenerLinks |
		html.NoreferrerLinks |
		html.SkipHTML
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	// 渲染并返回 HTML
	return markdown.Render(doc, renderer)
}
