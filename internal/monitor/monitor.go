package monitor

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/lin-snow/ech0/internal/metric"
	model "github.com/lin-snow/ech0/internal/model/metric"
	logUtil "github.com/lin-snow/ech0/internal/util/log"
	"go.uber.org/zap"
)

var (
	instance *Monitor
	once     sync.Once
)

type Monitor struct {
	collector metric.MetricCollector
	metrics   model.Metrics
	mu        sync.RWMutex
	interval  time.Duration
	stopChan  chan struct{}
	running   atomic.Bool
}

// NewMonitor 创建一个新的监控器（单例）。
func NewMonitor(collector metric.MetricCollector) *Monitor {
	once.Do(func() {
		instance = &Monitor{
			collector: collector,
			interval:  30 * time.Second,
			stopChan:  make(chan struct{}),
		}
		instance.Start()
	})
	return instance
}

// Start 开始定时采集系统指标。
func (m *Monitor) Start() {
	if m.running.Load() {
		return
	}
	m.running.Store(true)

	// 首次采样
	if err := m.collect(); err != nil {
		logUtil.Warn("monitor initial collect failed", zap.String("module", "monitor"), zap.Error(err))
	}

	go func() {
		ticker := time.NewTicker(m.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := m.collect(); err != nil {
					logUtil.Warn("monitor collect failed", zap.String("module", "monitor"), zap.Error(err))
				}
			case <-m.stopChan:
				logUtil.Info("monitor stopped", zap.String("module", "monitor"))
				m.running.Store(false)
				return
			}
		}
	}()
}

// Stop 停止监控（支持重启）。
func (m *Monitor) Stop() {
	if !m.running.Load() {
		return
	}
	close(m.stopChan)
	m.stopChan = make(chan struct{}) // 重建通道以支持重启
	m.running.Store(false)
}

// collect 内部采样逻辑。
func (m *Monitor) collect() error {
	data, err := m.collector.Collect()
	if err != nil {
		return err
	}
	m.mu.Lock()
	m.metrics = data
	m.mu.Unlock()
	return nil
}

// GetMetrics 获取当前缓存的最新指标。
func (m *Monitor) GetMetrics() model.Metrics {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.metrics
}

// ForceCollect 立即采样一次（同步）。
func (m *Monitor) ForceCollect() (model.Metrics, error) {
	data, err := m.collector.Collect()
	if err != nil {
		return data, err
	}
	m.mu.Lock()
	m.metrics = data
	m.mu.Unlock()
	return data, nil
}
