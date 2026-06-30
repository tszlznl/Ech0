// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package async

import (
	"errors"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWorkerPool_SubmitWaitRunsAllJobs 校验：提交的全部任务都会被执行，Wait 在全部完成后返回。
func TestWorkerPool_SubmitWaitRunsAllJobs(t *testing.T) {
	cases := []struct {
		name        string
		workerCount int
		queueSize   int
		jobs        int
	}{
		{"single worker", 1, 1, 50},
		{"multi worker", 4, 8, 200},
		{"more workers than jobs", 8, 16, 3},
		{"unbuffered queue", 2, 0, 30},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			pool := NewWorkerPool(tc.workerCount, tc.queueSize)
			defer pool.Stop()

			var done int32
			for range tc.jobs {
				pool.Submit(func() error {
					atomic.AddInt32(&done, 1)
					return nil
				})
			}
			pool.Wait()

			assert.Equal(t, int32(tc.jobs), atomic.LoadInt32(&done), "全部任务都应被执行")
		})
	}
}

// TestWorkerPool_JobErrorsDoNotBlockWait 校验：任务返回 error（被记录）不影响其它任务执行与 Wait 完成。
func TestWorkerPool_JobErrorsDoNotBlockWait(t *testing.T) {
	pool := NewWorkerPool(3, 4)
	defer pool.Stop()

	const total = 60
	var executed int32
	for i := range total {
		failing := i%2 == 0
		pool.Submit(func() error {
			atomic.AddInt32(&executed, 1)
			if failing {
				return errors.New("boom")
			}
			return nil
		})
	}
	pool.Wait()

	assert.Equal(t, int32(total), atomic.LoadInt32(&executed), "返回错误的任务也应计入执行并不阻塞 Wait")
}

// TestWorkerPool_SubmitAfterStopIsNoop 校验：Stop 之后 Submit 直接丢弃（no-op），不执行、不 panic、Wait 立即返回。
func TestWorkerPool_SubmitAfterStopIsNoop(t *testing.T) {
	pool := NewWorkerPool(2, 4)

	var before int32
	for range 10 {
		pool.Submit(func() error {
			atomic.AddInt32(&before, 1)
			return nil
		})
	}
	pool.Stop() // 内部会 Wait 直到前述任务全部完成
	require.Equal(t, int32(10), atomic.LoadInt32(&before), "Stop 前提交的任务应全部完成")

	var after int32
	require.NotPanics(t, func() {
		for range 10 {
			pool.Submit(func() error {
				atomic.AddInt32(&after, 1)
				return nil
			})
		}
	}, "Stop 之后 Submit 不应 panic")

	pool.Wait() // 不应阻塞
	assert.Zero(t, atomic.LoadInt32(&after), "Stop 之后提交的任务应被丢弃，不被执行")
}

// TestWorkerPool_StopIsIdempotent 校验：多次 Stop 不会因重复 close channel 而 panic（stopOnce 保护）。
func TestWorkerPool_StopIsIdempotent(t *testing.T) {
	pool := NewWorkerPool(2, 2)

	var done int32
	for range 5 {
		pool.Submit(func() error {
			atomic.AddInt32(&done, 1)
			return nil
		})
	}

	require.NotPanics(t, func() {
		pool.Stop()
		pool.Stop()
		pool.Stop()
	}, "重复 Stop 不应 panic")
	assert.Equal(t, int32(5), atomic.LoadInt32(&done))
}

// TestWorkerPool_StopWaitsForInFlightJobs 校验：Stop 会阻塞直到已入队任务全部执行完毕（内部 Wait）。
// 用任务自身的同步信号（channel）协调，不依赖计时。
func TestWorkerPool_StopWaitsForInFlightJobs(t *testing.T) {
	pool := NewWorkerPool(1, 8)

	release := make(chan struct{})
	started := make(chan struct{})
	var finished int32

	// 第一个任务阻塞直到收到 release，占住唯一 worker。
	pool.Submit(func() error {
		close(started)
		<-release
		atomic.AddInt32(&finished, 1)
		return nil
	})
	<-started // 确认任务已被 worker 取走并开始执行

	// 再排入若干任务，它们会在 release 之后由同一个 worker 依次执行。
	const queued = 5
	for range queued {
		pool.Submit(func() error {
			atomic.AddInt32(&finished, 1)
			return nil
		})
	}

	// 在后台调用 Stop，并在其返回后发信号；释放阻塞任务前 Stop 不应返回。
	stopReturned := make(chan struct{})
	go func() {
		pool.Stop()
		close(stopReturned)
	}()

	// 此刻 Stop 仍应被阻塞（still-running 任务未完成）。
	select {
	case <-stopReturned:
		t.Fatal("Stop 不应在仍有在途任务时返回")
	default:
	}

	close(release) // 放行，全部任务得以完成
	<-stopReturned // Stop 现在应当返回

	assert.Equal(t, int32(queued+1), atomic.LoadInt32(&finished), "Stop 返回时所有入队任务都应已执行")
}
