package monitor

import (
	"context"
	"fmt"
	"sync"
	"time"

	"ai-ops/internal/llm"
	"ai-ops/internal/ssh"
)

// AnomalyDetector 异常检测器
type AnomalyDetector struct {
	sshPool   *ssh.Pool
	llmClient *llm.OpenAIClient
	rules     map[string]*AlertRule
	history   *MetricsHistory
	mu        sync.RWMutex
}

// NewAnomalyDetector 创建异常检测器
func NewAnomalyDetector(sshPool *ssh.Pool, llmClient *llm.OpenAIClient) *AnomalyDetector {
	return &AnomalyDetector{
		sshPool:   sshPool,
		llmClient: llmClient,
		rules:     make(map[string]*AlertRule),
		history:   NewMetricsHistory(1000),
	}
}

// AddRule 添加告警规则
func (d *AnomalyDetector) AddRule(rule *AlertRule) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.rules[rule.ID] = rule
}

// DetectAnomalies 检测异常
func (d *AnomalyDetector) DetectAnomalies(ctx context.Context, hosts []string) ([]*Alert, error) {
	var alerts []*Alert
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, hostName := range hosts {
		wg.Add(1)
		go func(host string) {
			defer wg.Done()

			metrics, err := d.collectMetrics(host)
			if err != nil {
				return
			}

			d.history.Add(metrics)

			// 规则检测
			ruleAlerts := d.checkRules(metrics)

			// AI 预测性检测
			predictiveAlerts := d.predictiveAnalysis(ctx, host, metrics)

			mu.Lock()
			alerts = append(alerts, ruleAlerts...)
			alerts = append(alerts, predictiveAlerts...)
			mu.Unlock()
		}(hostName)
	}

	wg.Wait()
	return alerts, nil
}

// collectMetrics 采集指标
func (d *AnomalyDetector) collectMetrics(hostName string) (*HostMetrics, error) {
	metrics := &HostMetrics{
		HostName:  hostName,
		Timestamp: time.Now(),
	}

	// CPU
	cpuCmd := "top -bn1 | grep 'Cpu(s)' | awk '{print $2}' | cut -d'%' -f1"
	if result, err := d.sshPool.Exec(hostName, cpuCmd); err == nil {
		fmt.Sscanf(result, "%f", &metrics.CPU)
	}

	// Memory
	memCmd := "free | grep Mem | awk '{printf \"%.2f\", $3/$2 * 100.0}'"
	if result, err := d.sshPool.Exec(hostName, memCmd); err == nil {
		fmt.Sscanf(result, "%f", &metrics.Memory)
	}

	// Disk
	diskCmd := "df -h / | tail -1 | awk '{print $5}' | cut -d'%' -f1"
	if result, err := d.sshPool.Exec(hostName, diskCmd); err == nil {
		fmt.Sscanf(result, "%f", &metrics.Disk)
	}

	return metrics, nil
}

// checkRules 检查规则
func (d *AnomalyDetector) checkRules(metrics *HostMetrics) []*Alert {
	d.mu.RLock()
	defer d.mu.RUnlock()

	var alerts []*Alert

	for _, rule := range d.rules {
		if !rule.Enabled {
			continue
		}

		if d.matchRule(rule, metrics) {
			alert := d.createAlert(rule, metrics)
			alerts = append(alerts, alert)
		}
	}

	return alerts
}

// matchRule 匹配规则
func (d *AnomalyDetector) matchRule(rule *AlertRule, metrics *HostMetrics) bool {
	threshold, ok := rule.Condition["threshold"].(float64)
	if !ok {
		return false
	}

	switch rule.Type {
	case AlertTypeCPU:
		return metrics.CPU > threshold
	case AlertTypeMemory:
		return metrics.Memory > threshold
	case AlertTypeDisk:
		return metrics.Disk > threshold
	default:
		return false
	}
}

// createAlert 创建告警
func (d *AnomalyDetector) createAlert(rule *AlertRule, metrics *HostMetrics) *Alert {
	alert := &Alert{
		ID:       fmt.Sprintf("alert-%d", time.Now().UnixNano()),
		HostName: metrics.HostName,
		Type:     rule.Type,
		Level:    rule.Level,
		Title:    rule.Name,
		Metrics: map[string]interface{}{
			"cpu":    metrics.CPU,
			"memory": metrics.Memory,
			"disk":   metrics.Disk,
		},
		CreatedAt: time.Now(),
		Status:    "active",
	}

	// 生成消息和建议
	alert.Message = d.generateMessage(rule, metrics)
	alert.Suggestions = d.generateSuggestions(rule.Type, metrics)
	alert.Actions = d.generateQuickActions(rule.Type, metrics)

	return alert
}

// predictiveAnalysis AI 预测性分析
func (d *AnomalyDetector) predictiveAnalysis(ctx context.Context, hostName string, current *HostMetrics) []*Alert {
	historical := d.history.GetRecent(hostName, 10)
	if len(historical) < 5 {
		return nil
	}

	// 检测趋势
	if d.detectTrend(historical, "cpu", 10.0) {
		return []*Alert{{
			ID:       fmt.Sprintf("pred-%d", time.Now().UnixNano()),
			HostName: hostName,
			Type:     AlertTypePredictive,
			Level:    AlertLevelMedium,
			Title:    "CPU 使用率持续上升",
			Message:  fmt.Sprintf("检测到 %s CPU 使用率持续上升趋势，当前 %.1f%%", hostName, current.CPU),
			Suggestions: []string{
				"检查是否有异常进程",
				"考虑扩容或优化应用",
			},
			Actions: []QuickAction{
				{ID: "top", Label: "查看进程", Command: "top -bn1 | head -20"},
				{ID: "ps", Label: "CPU Top 10", Command: "ps aux --sort=-%cpu | head -11"},
			},
			CreatedAt: time.Now(),
			Status:    "active",
		}}
	}

	return nil
}

// detectTrend 检测趋势
func (d *AnomalyDetector) detectTrend(metrics []*HostMetrics, field string, threshold float64) bool {
	if len(metrics) < 3 {
		return false
	}

	var values []float64
	for _, m := range metrics {
		switch field {
		case "cpu":
			values = append(values, m.CPU)
		case "memory":
			values = append(values, m.Memory)
		}
	}

	// 简单线性趋势检测
	increase := values[len(values)-1] - values[0]
	return increase > threshold
}

// generateMessage 生成告警消息
func (d *AnomalyDetector) generateMessage(rule *AlertRule, metrics *HostMetrics) string {
	threshold := rule.Condition["threshold"].(float64)

	switch rule.Type {
	case AlertTypeCPU:
		return fmt.Sprintf("主机 %s CPU 使用率 %.1f%% 超过阈值 %.0f%%",
			metrics.HostName, metrics.CPU, threshold)
	case AlertTypeMemory:
		return fmt.Sprintf("主机 %s 内存使用率 %.1f%% 超过阈值 %.0f%%",
			metrics.HostName, metrics.Memory, threshold)
	case AlertTypeDisk:
		return fmt.Sprintf("主机 %s 磁盘使用率 %.1f%% 超过阈值 %.0f%%",
			metrics.HostName, metrics.Disk, threshold)
	default:
		return fmt.Sprintf("主机 %s 触发告警规则: %s", metrics.HostName, rule.Name)
	}
}

// generateSuggestions 生成建议
func (d *AnomalyDetector) generateSuggestions(alertType AlertType, metrics *HostMetrics) []string {
	switch alertType {
	case AlertTypeCPU:
		return []string{
			"检查 CPU 占用最高的进程",
			"分析是否有死循环或异常计算",
			"考虑增加 CPU 资源或优化代码",
		}
	case AlertTypeMemory:
		return []string{
			"检查内存占用最高的进程",
			"排查是否存在内存泄漏",
			"考虑增加内存或优化内存使用",
		}
	case AlertTypeDisk:
		return []string{
			"清理不必要的日志文件",
			"检查大文件占用情况",
			"考虑扩容磁盘或归档旧数据",
		}
	default:
		return []string{"请及时处理该告警"}
	}
}

// generateQuickActions 生成快捷操作
func (d *AnomalyDetector) generateQuickActions(alertType AlertType, metrics *HostMetrics) []QuickAction {
	switch alertType {
	case AlertTypeCPU:
		return []QuickAction{
			{ID: "top", Label: "查看进程", Command: "top -bn1 | head -20"},
			{ID: "cpu-top", Label: "CPU Top 10", Command: "ps aux --sort=-%cpu | head -11"},
		}
	case AlertTypeMemory:
		return []QuickAction{
			{ID: "free", Label: "内存状态", Command: "free -h"},
			{ID: "mem-top", Label: "内存 Top 10", Command: "ps aux --sort=-%mem | head -11"},
		}
	case AlertTypeDisk:
		return []QuickAction{
			{ID: "df", Label: "磁盘使用", Command: "df -h"},
			{ID: "du", Label: "目录占用", Command: "du -sh /* 2>/dev/null | sort -hr | head -10"},
		}
	default:
		return []QuickAction{}
	}
}

// MetricsHistory 指标历史
type MetricsHistory struct {
	data map[string][]*HostMetrics
	size int
	mu   sync.RWMutex
}

func NewMetricsHistory(size int) *MetricsHistory {
	return &MetricsHistory{
		data: make(map[string][]*HostMetrics),
		size: size,
	}
}

func (h *MetricsHistory) Add(metrics *HostMetrics) {
	h.mu.Lock()
	defer h.mu.Unlock()

	history := h.data[metrics.HostName]
	history = append(history, metrics)

	if len(history) > h.size {
		history = history[len(history)-h.size:]
	}

	h.data[metrics.HostName] = history
}

func (h *MetricsHistory) GetRecent(hostName string, count int) []*HostMetrics {
	h.mu.RLock()
	defer h.mu.RUnlock()

	history := h.data[hostName]
	if len(history) <= count {
		return history
	}
	return history[len(history)-count:]
}
