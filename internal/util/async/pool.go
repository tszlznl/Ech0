// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package async

import (
	"sync"

	logUtil "github.com/lin-snow/ech0/internal/util/log"
	"go.uber.org/zap"
)

// WorkerPool 是一个可复用的通用异步任务池
type WorkerPool struct {
	workerCount int               // 并发数
	jobs        chan func() error // 任务通道
	wg          sync.WaitGroup    // 用于等待所有任务完成
	mu          sync.RWMutex
	stopped     bool
	stopOnce    sync.Once
}

// NewWorkerPool 创建一个新的 WorkerPool
func NewWorkerPool(workerCount, jobQueueSize int) *WorkerPool {
	workerPool := &WorkerPool{
		workerCount: workerCount,
		jobs:        make(chan func() error, jobQueueSize),
	}
	workerPool.start()
	return workerPool
}

// Start 启动工作池
func (p *WorkerPool) start() {
	for i := 0; i < p.workerCount; i++ {
		go func() {
			for job := range p.jobs {
				func() {
					defer p.wg.Done()
					if err := job(); err != nil {
						logUtil.GetLogger().
							Error("worker job failed", zap.Error(err))
					}
				}()
			}
		}()
	}
}

// Submit 提交一个任务到工作池。
//
// 关停安全：全程持读锁直到 send 完成，而 Stop 的 close(p.jobs) 只在写锁内发生，
// 二者互斥 —— 因此 send 永远不会落到已关闭的 channel 上（无需再用 recover 兜 panic）。
// 缓冲满时 send 阻塞，但 worker 不持锁、持续 drain，排空后 send 即返回、随后 Stop 才能拿到写锁，
// 故不会死锁（前提 workerCount > 0，与原行为一致）。
func (p *WorkerPool) Submit(job func() error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.stopped {
		return
	}
	p.wg.Add(1)
	p.jobs <- job
}

// Wait 等待所有任务完成
func (p *WorkerPool) Wait() {
	p.wg.Wait()
}

// Stop 停止工作池
func (p *WorkerPool) Stop() {
	p.stopOnce.Do(func() {
		p.mu.Lock()
		p.stopped = true
		close(p.jobs)
		p.mu.Unlock()
	})
	p.Wait()
}
