package service

import (
	"context"
	"fmt"
	"math"
	"time"

	"ai-ops/internal/model"
	"ai-ops/internal/repository"
	"ai-ops/internal/ssh"
	"ai-ops/pkg/logger"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// TrendService 趋势分析服务
type TrendService struct {
	sshPool      *ssh.Pool
	hostRepo     repository.HostRepository
	trendRepo    *repository.TrendPredictionRepository
	healthRepo   *repository.HealthCheckRepository
}

// NewTrendService 创建趋势分析服务
func NewTrendService(
	sshPool *ssh.Pool,
	hostRepo repository.HostRepository,
	trendRepo *repository.TrendPredictionRepository,
	healthRepo *repository.HealthCheckRepository,
) *TrendService {
	return &TrendService{
		sshPool:    sshPool,
		hostRepo:   hostRepo,
		trendRepo:  trendRepo,
		healthRepo: healthRepo,
	}
}

// TrendAnalysisResult 趋势分析结果
type TrendAnalysisResult struct {
	HostID         string                 `json:"host_id"`
	HostName       string                 `json:"host_name"`
	MetricType     string                 `json:"metric_type"`
	CurrentValue   float64                `json:"current_value"`
	PredictedValue float64                `json:"predicted_value"`
	Trend          string                 `json:"trend"` // increasing, decreasing, stable
	Confidence     float64                `json:"confidence"`
	AlertLevel     string                 `json:"alert_level,omitempty"`
	Recommendation string                 `json:"recommendation,omitempty"`
	PredictionTime time.Time              `json:"prediction_time"`
}

// AnalyzeTrends 分析趋势
func (s *TrendService) AnalyzeTrends(ctx context.Context, hostID string) ([]*TrendAnalysisResult, error) {
	host, err := s.hostRepo.GetByID(hostID)
	if err != nil {
		return nil, fmt.Errorf("主机不存在: %w", err)
	}

	// 获取历史数据
	history, err := s.healthRepo.FindByHost(hostID, 20)
	if err != nil {
		return nil, fmt.Errorf("获取历史数据失败: %w", err)
	}

	if len(history) < 5 {
		return nil, fmt.Errorf("历史数据不足，至少需要 5 条记录")
	}

	results := make([]*TrendAnalysisResult, 0)

	// 分析 CPU 趋势
	if cpuResult := s.analyzeCPUTrend(host, history); cpuResult != nil {
		results = append(results, cpuResult)
		s.savePrediction(cpuResult)
	}

	// 分析内存趋势
	if memResult := s.analyzeMemoryTrend(host, history); memResult != nil {
		results = append(results, memResult)
		s.savePrediction(memResult)
	}

	// 分析磁盘趋势
	if diskResult := s.analyzeDiskTrend(host, history); diskResult != nil {
		results = append(results, diskResult)
		s.savePrediction(diskResult)
	}

	return results, nil
}

// analyzeCPUTrend 分析 CPU 趋势
func (s *TrendService) analyzeCPUTrend(host *model.Host, history []*model.HealthCheck) *TrendAnalysisResult {
	values := make([]float64, 0)
	for _, h := range history {
		if cpu, ok := h.Metrics["cpu_usage"].(float64); ok {
			values = append(values, cpu)
		}
	}

	if len(values) < 5 {
		return nil
	}

	current := values[len(values)-1]
	predicted, confidence := s.linearRegression(values)
	trend := s.determineTrend(values)

	result := &TrendAnalysisResult{
		HostID:         host.ID,
		HostName:       host.Name,
		MetricType:     model.MetricTypeCPU,
		CurrentValue:   current,
		PredictedValue: predicted,
		Trend:          trend,
		Confidence:     confidence,
		PredictionTime: time.Now().Add(1 * time.Hour),
	}

	// 判断告警级别
	if predicted > 90 {
		result.AlertLevel = model.RiskLevelCritical
		result.Recommendation = "CPU 使用率预计将超过 90%，建议立即检查并优化"
	} else if predicted > 80 {
		result.AlertLevel = model.RiskLevelHigh
		result.Recommendation = "CPU 使用率持续上升，建议关注并准备优化措施"
	} else if predicted > 70 && trend == "increasing" {
		result.AlertLevel = model.RiskLevelMedium
		result.Recommendation = "CPU 使用率呈上升趋势，建议监控"
	}

	return result
}

// analyzeMemoryTrend 分析内存趋势
func (s *TrendService) analyzeMemoryTrend(host *model.Host, history []*model.HealthCheck) *TrendAnalysisResult {
	values := make([]float64, 0)
	for _, h := range history {
		if mem, ok := h.Metrics["memory_usage"].(float64); ok {
			values = append(values, mem)
		}
	}

	if len(values) < 5 {
		return nil
	}

	current := values[len(values)-1]
	predicted, confidence := s.linearRegression(values)
	trend := s.determineTrend(values)

	result := &TrendAnalysisResult{
		HostID:         host.ID,
		HostName:       host.Name,
		MetricType:     model.MetricTypeMemory,
		CurrentValue:   current,
		PredictedValue: predicted,
		Trend:          trend,
		Confidence:     confidence,
		PredictionTime: time.Now().Add(1 * time.Hour),
	}

	if predicted > 90 {
		result.AlertLevel = model.RiskLevelCritical
		result.Recommendation = "内存使用率预计将超过 90%，可能存在内存泄漏"
	} else if predicted > 85 {
		result.AlertLevel = model.RiskLevelHigh
		result.Recommendation = "内存使用率持续上升，建议检查内存占用"
	}

	return result
}

// analyzeDiskTrend 分析磁盘趋势
func (s *TrendService) analyzeDiskTrend(host *model.Host, history []*model.HealthCheck) *TrendAnalysisResult {
	values := make([]float64, 0)
	for _, h := range history {
		if disk, ok := h.Metrics["disk_usage"].(float64); ok {
			values = append(values, disk)
		}
	}

	if len(values) < 5 {
		return nil
	}

	current := values[len(values)-1]
	predicted, confidence := s.linearRegression(values)
	trend := s.determineTrend(values)

	result := &TrendAnalysisResult{
		HostID:         host.ID,
		HostName:       host.Name,
		MetricType:     model.MetricTypeDisk,
		CurrentValue:   current,
		PredictedValue: predicted,
		Trend:          trend,
		Confidence:     confidence,
		PredictionTime: time.Now().Add(24 * time.Hour),
	}

	if predicted > 95 {
		result.AlertLevel = model.RiskLevelCritical
		result.Recommendation = "磁盘空间即将耗尽，请立即清理或扩容"
	} else if predicted > 90 {
		result.AlertLevel = model.RiskLevelHigh
		result.Recommendation = "磁盘使用率持续上升，建议清理或扩容"
	}

	return result
}

// linearRegression 简单线性回归预测
func (s *TrendService) linearRegression(values []float64) (predicted float64, confidence float64) {
	n := float64(len(values))
	if n < 2 {
		return values[len(values)-1], 0
	}

	// 计算均值
	var sumX, sumY, sumXY, sumX2 float64
	for i, y := range values {
		x := float64(i)
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}

	// 计算斜率和截距
	slope := (n*sumXY - sumX*sumY) / (n*sumX2 - sumX*sumX)
	intercept := (sumY - slope*sumX) / n

	// 预测下一个值
	predicted = slope*n + intercept

	// 计算置信度（基于 R²）
	var ssRes, ssTot float64
	meanY := sumY / n
	for i, y := range values {
		x := float64(i)
		predicted := slope*x + intercept
		ssRes += math.Pow(y-predicted, 2)
		ssTot += math.Pow(y-meanY, 2)
	}

	if ssTot == 0 {
		confidence = 0
	} else {
		rSquared := 1 - (ssRes / ssTot)
		confidence = math.Max(0, math.Min(1, rSquared))
	}

	return predicted, confidence
}

// determineTrend 判断趋势
func (s *TrendService) determineTrend(values []float64) string {
	if len(values) < 2 {
		return "stable"
	}

	// 计算最近几个值的平均变化率
	changes := 0.0
	for i := 1; i < len(values); i++ {
		changes += values[i] - values[i-1]
	}
	avgChange := changes / float64(len(values)-1)

	if avgChange > 2 {
		return "increasing"
	} else if avgChange < -2 {
		return "decreasing"
	}
	return "stable"
}

// savePrediction 保存预测结果
func (s *TrendService) savePrediction(result *TrendAnalysisResult) {
	prediction := &model.TrendPrediction{
		ID:             uuid.New().String(),
		HostID:         result.HostID,
		MetricType:     result.MetricType,
		CurrentValue:   result.CurrentValue,
		PredictedValue: result.PredictedValue,
		PredictionTime: result.PredictionTime,
		Confidence:     result.Confidence,
		AlertLevel:     result.AlertLevel,
		CreatedAt:      time.Now(),
	}

	if err := s.trendRepo.Create(prediction); err != nil {
		logger.Error("保存趋势预测失败", zap.Error(err))
	}
}

// GetAlerts 获取告警
func (s *TrendService) GetAlerts(hostID string) ([]*model.TrendPrediction, error) {
	return s.trendRepo.FindAlerts(hostID)
}

// GetPredictions 获取预测历史
func (s *TrendService) GetPredictions(hostID string, limit int) ([]*model.TrendPrediction, error) {
	return s.trendRepo.FindByHost(hostID, limit)
}
