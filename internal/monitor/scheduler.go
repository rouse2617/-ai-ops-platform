package monitor

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"
)

// MonitorScheduler 监控调度器
type MonitorScheduler struct {
	detector  *AnomalyDetector
	notifier  *AlertNotifier
	interval  time.Duration
	hosts     []string
	running   bool
	mu        sync.RWMutex
	logger    *zap.Logger
	ctx       context.Context
	cancel    context.CancelFunc
}

// NewMonitorScheduler 创建监控调度器
func NewMonitorScheduler(detector *AnomalyDetector, notifier *AlertNotifier, interval time.Duration, logger *zap.Logger) *MonitorScheduler {
	ctx, cancel := context.WithCancel(context.Background())
	return &MonitorScheduler{
		detector: detector,
		notifier: notifier,
		interval: interval,
		logger:   logger,
		ctx:      ctx,
		cancel:   cancel,
	}
}

// SetHosts 设置监控主机列表
func (s *MonitorScheduler) SetHosts(hosts []string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.hosts = hosts
}

// Start 启动调度器
func (s *MonitorScheduler) Start() {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return
	}
	s.running = true
	s.mu.Unlock()

	s.logger.Info("监控调度器启动", zap.Duration("interval", s.interval))

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	// 立即执行一次
	s.runCheck()

	for {
		select {
		case <-ticker.C:
			s.runCheck()
		case <-s.ctx.Done():
			s.logger.Info("监控调度器停止")
			return
		}
	}
}

// Stop 停止调度器
func (s *MonitorScheduler) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return
	}

	s.running = false
	s.cancel()
}

// runCheck 执行检查
func (s *MonitorScheduler) runCheck() {
	s.mu.RLock()
	hosts := make([]string, len(s.hosts))
	copy(hosts, s.hosts)
	s.mu.RUnlock()

	if len(hosts) == 0 {
		return
	}

	s.logger.Debug("执行监控检查", zap.Int("host_count", len(hosts)))

	alerts, err := s.detector.DetectAnomalies(s.ctx, hosts)
	if err != nil {
		s.logger.Error("异常检测失败", zap.Error(err))
		return
	}

	if len(alerts) > 0 {
		s.logger.Info("检测到告警", zap.Int("count", len(alerts)))

		for _, alert := range alerts {
			if err := s.notifier.Notify(alert); err != nil {
				s.logger.Error("告警推送失败",
					zap.String("alert_id", alert.ID),
					zap.Error(err))
			}
		}
	}
}
