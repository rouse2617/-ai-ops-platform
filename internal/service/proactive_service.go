package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"ai-ops/internal/model"
	"ai-ops/internal/monitor"
	"ai-ops/pkg/logger"

	"go.uber.org/zap"
)

// ProactiveService 主动服务 - 负责生成主动消息
type ProactiveService struct {
	healthService *HealthService
	notifier      *monitor.AlertNotifier
}

// NewProactiveService 创建主动服务
func NewProactiveService(healthService *HealthService, notifier *monitor.AlertNotifier) *ProactiveService {
	return &ProactiveService{
		healthService: healthService,
		notifier:      notifier,
	}
}

// GenerateMorningReport 生成早安报告
func (s *ProactiveService) GenerateMorningReport(ctx context.Context) error {
	comparison, err := s.healthService.CompareHosts(ctx, nil)
	if err != nil {
		return fmt.Errorf("获取健康数据失败: %w", err)
	}

	// 构建报告内容
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("早安！过去 8 小时内系统整体健康分 **%d** 分。\n\n", comparison.AvgScore))

	if len(comparison.Anomalies) == 0 {
		sb.WriteString("所有主机运行平稳，无异常情况。")
	} else {
		sb.WriteString(fmt.Sprintf("发现 **%d** 台主机需要关注：\n", len(comparison.Anomalies)))
		for _, host := range comparison.Anomalies {
			sb.WriteString(fmt.Sprintf("- %s\n", host))
		}
	}

	// 添加资源占用最高的主机信息
	if comparison.HighestCPU != "" || comparison.HighestMem != "" || comparison.HighestDisk != "" {
		sb.WriteString("\n**资源占用情况：**\n")
		if comparison.HighestCPU != "" {
			sb.WriteString(fmt.Sprintf("- CPU 最高: %s\n", comparison.HighestCPU))
		}
		if comparison.HighestMem != "" {
			sb.WriteString(fmt.Sprintf("- 内存最高: %s\n", comparison.HighestMem))
		}
		if comparison.HighestDisk != "" {
			sb.WriteString(fmt.Sprintf("- 磁盘最高: %s\n", comparison.HighestDisk))
		}
	}

	// 构建指标数据
	metrics := map[string]interface{}{
		"avg_score":    comparison.AvgScore,
		"total_hosts":  len(comparison.Hosts),
		"anomaly_count": len(comparison.Anomalies),
	}

	// 获取主机名列表
	var hostNames []string
	for _, h := range comparison.Hosts {
		hostNames = append(hostNames, h.HostName)
	}

	// 发送早安报告
	return s.notifier.SendMorningReport(sb.String(), hostNames, metrics)
}

// GenerateHealthSummary 生成健康摘要
func (s *ProactiveService) GenerateHealthSummary(ctx context.Context) error {
	score, level, err := s.healthService.GetOverallHealthScore(ctx)
	if err != nil {
		return fmt.Errorf("获取健康分失败: %w", err)
	}

	var levelText string
	switch level {
	case "excellent":
		levelText = "优秀"
	case "good":
		levelText = "良好"
	case "warning":
		levelText = "需关注"
	default:
		levelText = "异常"
	}

	summary := fmt.Sprintf("当前系统健康度：**%d** 分（%s）", score, levelText)

	metrics := map[string]interface{}{
		"score": score,
		"level": level,
	}

	return s.notifier.SendHealthSummary(summary, metrics)
}

// CheckAndNotifyAnomalies 检查并通知异常
func (s *ProactiveService) CheckAndNotifyAnomalies(ctx context.Context) error {
	comparison, err := s.healthService.CompareHosts(ctx, nil)
	if err != nil {
		return fmt.Errorf("获取健康数据失败: %w", err)
	}

	if len(comparison.Anomalies) == 0 {
		return nil
	}

	// 为每个异常主机发送告警
	for _, hostName := range comparison.Anomalies {
		// 找到对应的主机详情
		var hostDetail *HostHealthSummary
		for _, h := range comparison.Hosts {
			if h.HostName == hostName {
				hostDetail = &h
				break
			}
		}

		if hostDetail == nil {
			continue
		}

		title := fmt.Sprintf("主机 %s 健康异常", hostName)
		content := fmt.Sprintf("健康分: %d，CPU: %.1f%%，内存: %.1f%%，磁盘: %.1f%%",
			hostDetail.Score, hostDetail.CPU, hostDetail.Memory, hostDetail.Disk)

		actions := []model.QuickAction{
			{
				ID:          "check_detail",
				Label:       "查看详情",
				Description: "查看主机详细健康状态",
			},
			{
				ID:          "run_diagnosis",
				Label:       "运行诊断",
				Description: "运行自动诊断程序",
			},
		}

		if err := s.notifier.SendAnomalyAlert(title, content, []string{hostName}, actions); err != nil {
			logger.Error("发送异常告警失败", zap.String("host", hostName), zap.Error(err))
		}
	}

	return nil
}

// StartScheduler 启动定时任务
func (s *ProactiveService) StartScheduler(ctx context.Context) {
	// 早安报告定时器 - 每天早上 8 点
	go s.scheduleMorningReport(ctx)

	// 健康检查定时器 - 每 30 分钟
	go s.scheduleHealthCheck(ctx)

	logger.Info("主动服务定时任务已启动")
}

func (s *ProactiveService) scheduleMorningReport(ctx context.Context) {
	for {
		now := time.Now()
		// 计算下一个 8:00
		next := time.Date(now.Year(), now.Month(), now.Day(), 8, 0, 0, 0, now.Location())
		if now.After(next) {
			next = next.Add(24 * time.Hour)
		}

		timer := time.NewTimer(time.Until(next))
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
			if err := s.GenerateMorningReport(ctx); err != nil {
				logger.Error("生成早安报告失败", zap.Error(err))
			}
		}
	}
}

func (s *ProactiveService) scheduleHealthCheck(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := s.CheckAndNotifyAnomalies(ctx); err != nil {
				logger.Error("健康检查失败", zap.Error(err))
			}
		}
	}
}
