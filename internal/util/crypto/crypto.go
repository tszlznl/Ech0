// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package util

import (
	"crypto/md5"
	"encoding/hex"
	"math/rand"
	"time"
)

var seededRand = rand.New(rand.NewSource(time.Now().UTC().UnixNano()))

// MD5Encrypt 对内容进行 MD5 编码
func MD5Encrypt(text string) string {
	hash := md5.New()
	hash.Write([]byte(text))
	hashInBytes := hash.Sum(nil)
	return hex.EncodeToString(hashInBytes)
}

// GenerateRandomString 生成指定长度的随机字符串
func GenerateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz" +
		"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}
