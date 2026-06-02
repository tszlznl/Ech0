// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package util

import (
	"strings"
)

// TrimURL 去除 URL 前后的空格和斜杠
func TrimURL(url string) string {
	if url == "" {
		return ""
	}

	// 去除连接地址前后的空格和斜杠
	url = strings.TrimSpace(url)
	url = strings.TrimPrefix(url, "/")
	url = strings.TrimSuffix(url, "/")
	return url
}

// ExtractDomain 从 URL 中提取域名
func ExtractDomain(url string) string {
	// 去除协议部分
	if strings.HasPrefix(url, "http://") {
		url = strings.TrimPrefix(url, "http://")
	} else if strings.HasPrefix(url, "https://") {
		url = strings.TrimPrefix(url, "https://")
	}

	// 提取域名部分 (到第一个斜杠为止)
	slashIndex := strings.Index(url, "/")
	if slashIndex != -1 {
		url = url[:slashIndex]
	}

	return url
}

// GetMIMETypeFromFilenameOrURL 根据文件名或 URL 获取 MIME 类型
func GetMIMETypeFromFilenameOrURL(filenameOrURL string) string {
	lowerFilename := strings.ToLower(filenameOrURL)
	switch {
	case strings.HasSuffix(lowerFilename, ".jpg"), strings.HasSuffix(lowerFilename, ".jpeg"):
		return "image/jpeg"
	case strings.HasSuffix(lowerFilename, ".png"):
		return "image/png"
	case strings.HasSuffix(lowerFilename, ".gif"):
		return "image/gif"
	case strings.HasSuffix(lowerFilename, ".bmp"):
		return "image/bmp"
	case strings.HasSuffix(lowerFilename, ".webp"):
		return "image/webp"
	case strings.HasSuffix(lowerFilename, ".mp4"):
		return "video/mp4"
	case strings.HasSuffix(lowerFilename, ".mov"):
		return "video/quicktime"
	case strings.HasSuffix(lowerFilename, ".mp3"):
		return "audio/mpeg"
	case strings.HasSuffix(lowerFilename, ".wav"):
		return "audio/wav"
	case strings.HasSuffix(lowerFilename, ".ogg"):
		return "audio/ogg"
	case strings.HasSuffix(lowerFilename, ".pdf"):
		return "application/pdf"
	case strings.HasSuffix(lowerFilename, ".doc"):
		return "application/msword"
	case strings.HasSuffix(lowerFilename, ".docx"):
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case strings.HasSuffix(lowerFilename, ".xls"):
		return "application/vnd.ms-excel"
	case strings.HasSuffix(lowerFilename, ".xlsx"):
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	case strings.HasSuffix(lowerFilename, ".ppt"):
		return "application/vnd.ms-powerpoint"
	case strings.HasSuffix(lowerFilename, ".pptx"):
		return "application/vnd.openxmlformats-officedocument.presentationml.presentation"
	case strings.HasSuffix(lowerFilename, ".txt"):
		return "text/plain"
	case strings.HasSuffix(lowerFilename, ".html"), strings.HasSuffix(lowerFilename, ".htm"):
		return "text/html"
	case strings.HasSuffix(lowerFilename, ".csv"):
		return "text/csv"
	default:
		return "application/octet-stream" // 默认二进制流
	}
}
