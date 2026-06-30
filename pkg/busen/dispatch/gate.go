// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

// Package dispatch provides small primitives for coordinating in-process delivery.
package dispatch

import (
	"context"
	"sync"
)

// Gate coordinates "accept new work" and "wait for in-flight work" semantics.
type Gate struct {
	mu     sync.Mutex
	closed bool
	active int
	idle   chan struct{}
}

// NewGate creates a gate in the open state.
func NewGate() *Gate {
	g := &Gate{
		idle: make(chan struct{}),
	}
	close(g.idle)
	return g
}

// Enter registers one in-flight operation if the gate is still open.
func (g *Gate) Enter() bool {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.closed {
		return false
	}

	if g.active == 0 {
		g.idle = make(chan struct{})
	}
	g.active++
	return true
}

// Leave marks one in-flight operation as completed.
func (g *Gate) Leave() {
	g.mu.Lock()
	defer g.mu.Unlock()

	// 未配对/多余的 Leave（active 已为 0）：此时 idle 已处于 closed 状态，必须直接
	// no-op 返回。否则会再次 close 一个已关闭的 channel → panic: close of closed channel。
	if g.active == 0 {
		return
	}
	g.active--
	// 仅在 active 由 >0 降到 0 的「沿」上关闭 idle，唤醒 Wait。
	if g.active == 0 {
		close(g.idle)
	}
}

// Closed reports whether the gate has been closed for new work.
func (g *Gate) Closed() bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.closed
}

// Close prevents future Enter calls from succeeding.
func (g *Gate) Close() {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.closed = true
}

// Wait blocks until all in-flight work has completed or the context is canceled.
func (g *Gate) Wait(ctx context.Context) error {
	g.mu.Lock()
	idle := g.idle
	g.mu.Unlock()

	select {
	case <-idle:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
