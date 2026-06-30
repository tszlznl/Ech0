// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package util

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
)

// MD5Encrypt 对内容进行 MD5 编码
func MD5Encrypt(text string) string {
	hash := md5.New()
	hash.Write([]byte(text))
	hashInBytes := hash.Sum(nil)
	return hex.EncodeToString(hashInBytes)
}

// randomCharset 是 GenerateRandomString 使用的字符集（62 个字符）。
const randomCharset = "abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// GenerateRandomString 使用密码学安全随机源（crypto/rand）生成指定长度的随机字符串。
//
// 该函数被用于生成 OAuth state、一次性 OAuth 兑换码、token JTI 等**安全敏感**的不可预测值，
// 因此必须使用 crypto/rand 而非 math/rand —— 后者可被预测，会让 CSRF state / 一次性码被伪造。
//
// 为避免取模偏置（256 不是 62 的整数倍），对落在不完整区间的随机字节做拒绝采样。
// length <= 0 时返回空串。crypto/rand 读取失败属系统级灾难，直接 panic，绝不退化为可预测值。
func GenerateRandomString(length int) string {
	if length <= 0 {
		return ""
	}

	// 拒绝采样上界：丢弃会引入取模偏置的高位字节（256 % 62 = 8 → 丢弃 248..255）。
	const limit = 256 - (256 % len(randomCharset))

	out := make([]byte, length)
	buf := make([]byte, length)
	for filled := 0; filled < length; {
		if _, err := rand.Read(buf); err != nil {
			panic("crypto: secure random source unavailable: " + err.Error())
		}
		for _, v := range buf {
			if int(v) >= limit {
				continue
			}
			out[filled] = randomCharset[int(v)%len(randomCharset)]
			filled++
			if filled == length {
				break
			}
		}
	}
	return string(out)
}
