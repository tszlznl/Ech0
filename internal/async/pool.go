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

// Submit 提交一个任务到工作池
func (p *WorkerPool) Submit(job func() error) {
	p.mu.RLock()
	if p.stopped {
		p.mu.RUnlock()
		return
	}
	p.wg.Add(1)
	jobs := p.jobs
	p.mu.RUnlock()

	defer func() {
		if recover() != nil {
			// channel 已关闭时回收计数，避免 Wait 永久阻塞。
			p.wg.Done()
		}
	}()
	jobs <- job
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
