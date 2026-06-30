// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package util

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMD5Encrypt(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "empty", in: "", want: "d41d8cd98f00b204e9800998ecf8427e"},
		{name: "abc", in: "abc", want: "900150983cd24fb0d6963f7d28e17f72"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, MD5Encrypt(tc.in))
		})
	}
}

func TestGenerateRandomString_Length(t *testing.T) {
	for _, n := range []int{1, 8, 16, 32, 64, 256} {
		got := GenerateRandomString(n)
		assert.Len(t, got, n, "length %d", n)
	}
}

func TestGenerateRandomString_NonPositiveReturnsEmpty(t *testing.T) {
	assert.Equal(t, "", GenerateRandomString(0))
	assert.Equal(t, "", GenerateRandomString(-5))
}

func TestGenerateRandomString_CharsetOnly(t *testing.T) {
	// 生成的每个字符都必须落在白名单字符集内（防止拒绝采样引入非法字节）。
	s := GenerateRandomString(10000)
	for i, r := range s {
		require.True(t, strings.ContainsRune(randomCharset, r),
			"index %d produced out-of-charset rune %q", i, r)
	}
}

// TestGenerateRandomString_NoCollisions 验证安全令牌的不可预测/唯一性：
// 大量生成长度 32 的串不应出现碰撞。若退回 math/rand 的可预测序列，
// 在并发/同源情况下更易产生可被利用的重复，本测试作为回归护栏。
func TestGenerateRandomString_NoCollisions(t *testing.T) {
	const (
		count  = 50000
		length = 32
	)
	seen := make(map[string]struct{}, count)
	for i := 0; i < count; i++ {
		s := GenerateRandomString(length)
		_, dup := seen[s]
		require.Falsef(t, dup, "collision after %d generations: %q", i, s)
		seen[s] = struct{}{}
	}
}

// TestGenerateRandomString_UsesAllCharsetSymbols 粗略验证分布：足量采样下，
// 字符集中每个符号都至少出现一次（拒绝采样不应系统性遗漏任何字符）。
func TestGenerateRandomString_UsesAllCharsetSymbols(t *testing.T) {
	s := GenerateRandomString(100000)
	for _, c := range randomCharset {
		assert.Truef(t, strings.ContainsRune(s, c), "charset symbol %q never appeared", c)
	}
}
