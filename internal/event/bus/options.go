// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package bus

import (
	"github.com/lin-snow/ech0/internal/config"
	"github.com/lin-snow/ech0/pkg/busen"
)

// AsyncParallel 是“异步、多 worker 并行消费”的订阅策略，供按事件并行处理且无需保序的订阅者使用
// （agent 缓存失效、embedding 增量索引）。buffer/parallelism 沿用历史的 ECH0_EVENT_AGENT_* 配置项。
func AsyncParallel() []busen.SubscribeOption {
	ec := config.Config().Event
	return []busen.SubscribeOption{
		busen.Async(),
		busen.WithParallelism(ec.AgentParallelism),
		busen.WithBuffer(ec.AgentBuffer),
		busen.WithOverflow(MapOverflow(ec.DefaultOverflow)),
	}
}

// AsyncSequential 是“异步、单 worker FIFO”的订阅策略，供需要保序处理的系统事件订阅者使用
// （快照计划重配）。buffer 取 ECH0_EVENT_SYSTEM_*。
func AsyncSequential() []busen.SubscribeOption {
	ec := config.Config().Event
	return []busen.SubscribeOption{
		busen.Async(),
		busen.Sequential(),
		busen.WithBuffer(ec.SystemBuffer),
		busen.WithOverflow(MapOverflow(ec.DefaultOverflow)),
	}
}

// DeadLetter 是死信重试订阅者的策略：异步、单 worker FIFO、独立 buffer（ECH0_EVENT_DEADLETTER_*）。
func DeadLetter() []busen.SubscribeOption {
	ec := config.Config().Event
	return []busen.SubscribeOption{
		busen.Async(),
		busen.Sequential(),
		busen.WithBuffer(ec.DeadLetterBuffer),
		busen.WithOverflow(MapOverflow(ec.DefaultOverflow)),
	}
}
