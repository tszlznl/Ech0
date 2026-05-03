// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type rateLimiter struct {
	mu      sync.Mutex
	buckets map[string]*tokenBucket
	rps     int
	burst   int
}

type tokenBucket struct {
	tokens    float64
	lastTime  time.Time
	ratePerNs float64
	burst     float64
}

func newRateLimiter(rps, burst int) *rateLimiter {
	return &rateLimiter{
		buckets: make(map[string]*tokenBucket),
		rps:     rps,
		burst:   burst,
	}
}

func (rl *rateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	b, ok := rl.buckets[key]
	if !ok {
		b = &tokenBucket{
			tokens:    float64(rl.burst),
			lastTime:  time.Now(),
			ratePerNs: float64(rl.rps) / float64(time.Second),
			burst:     float64(rl.burst),
		}
		rl.buckets[key] = b
	}

	now := time.Now()
	elapsed := now.Sub(b.lastTime)
	b.tokens += float64(elapsed) * b.ratePerNs
	if b.tokens > b.burst {
		b.tokens = b.burst
	}
	b.lastTime = now

	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

func RateLimit(rps, burst int) gin.HandlerFunc {
	limiter := newRateLimiter(rps, burst)
	startBucketGC(limiter, 5*time.Minute, 10*time.Minute)

	return func(c *gin.Context) {
		key := c.ClientIP()
		if !limiter.allow(key) {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// RateLimitWithIdempotency 在 RateLimit 基础上叠加幂等窗口，适用于"写多读少且天然幂等"的接口
// （例如点赞）。
//
//   - 单 IP 令牌桶限速：超过阈值返回 429。
//   - (IP, 资源 ID) 维度的去重窗口：同一 IP 在 dedupTTL 内对同一资源的请求被视作已处理，
//     调用 onIdempotent 写出响应并中止后续处理，避免重复请求触发数据库写入与缓存失效。
//
// resourceParam 是 gin 路径参数名（如 "id"）。当 IP 或资源 ID 缺失时跳过去重检查，
// 由 handler 进行业务校验并返回业务错误，避免幂等逻辑掩盖参数问题。
//
// onIdempotent 必须自行写出响应并 Abort；推荐返回与正常成功路径形状一致的响应，
// 使客户端无感知。
func RateLimitWithIdempotency(
	rps, burst int,
	dedupTTL time.Duration,
	resourceParam string,
	onIdempotent gin.HandlerFunc,
) gin.HandlerFunc {
	limiter := newRateLimiter(rps, burst)
	dedup := newIdempotencyStore(dedupTTL)

	gcInterval := max(dedupTTL, time.Minute)
	startBucketGC(limiter, gcInterval, 10*time.Minute)
	startIdempotencyGC(dedup, gcInterval)

	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !limiter.allow(ip) {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			c.Abort()
			return
		}

		resourceID := c.Param(resourceParam)
		if ip != "" && resourceID != "" && !dedup.acquire(ip+"|"+resourceID, time.Now()) {
			onIdempotent(c)
			c.Abort()
			return
		}

		c.Next()
	}
}

func startBucketGC(rl *rateLimiter, interval, idle time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C {
			rl.mu.Lock()
			cutoff := time.Now().Add(-idle)
			for k, b := range rl.buckets {
				if b.lastTime.Before(cutoff) {
					delete(rl.buckets, k)
				}
			}
			rl.mu.Unlock()
		}
	}()
}

// idempotencyStore 维护 (IP, 资源 ID) → 最近一次命中时间的映射，由后台 goroutine
// 周期性回收过期条目，使内存占用与活跃来源数成正比而非历史累积。
type idempotencyStore struct {
	mu   sync.Mutex
	seen map[string]time.Time
	ttl  time.Duration
}

func newIdempotencyStore(ttl time.Duration) *idempotencyStore {
	return &idempotencyStore{
		seen: make(map[string]time.Time),
		ttl:  ttl,
	}
}

// acquire 在窗口内未命中时记录并返回 true（允许放行）；命中时返回 false（应作幂等处理）。
func (s *idempotencyStore) acquire(key string, now time.Time) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if t, ok := s.seen[key]; ok && now.Sub(t) < s.ttl {
		return false
	}
	s.seen[key] = now
	return true
}

func (s *idempotencyStore) gc(now time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for k, t := range s.seen {
		if now.Sub(t) >= s.ttl {
			delete(s.seen, k)
		}
	}
}

func startIdempotencyGC(s *idempotencyStore, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C {
			s.gc(time.Now())
		}
	}()
}
