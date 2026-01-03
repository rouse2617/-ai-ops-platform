package service

import (
	"context"
	"fmt"
	"time"

	"ai-ops/internal/model"
	"ai-ops/internal/repository"
	"ai-ops/internal/ssh"
	"ai-ops/pkg/logger"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// HealthService 健康检查服务
type HealthService struct {
	sshPool      *ssh.Pool
	hostRepo     repository.HostRepository
	healthRepo   *repository.HealthCheckRepository
}

// NewHealthService 创建健康检查服务
func NewHealthService(
	sshPool *ssh.Pool,
	hostRepo repository.HostRepository,
	healthRepo *repository.HealthCheckRepository,
) *HealthService {
	return &HealthService{
		sshPool:    sshPool,
		hostRepo:   hostRepo,
		healthRepo: healthRepo,
	}
}

// HealthCheckResult 健康检查结果
type HealthCheckResult struct {
	HostID      string                 `json:"host_id"`
	HostName    string                 `json:"host_name"`
	Status      string                 `json:"status"`
	Score       int                    `json:"score"`        // 健康分 0-100
	ScoreLevel  string                 `json:"score_level"`  // excellent, good, warning, critical
	Metrics     map[string]interface{} `json:"metrics"`
	Issues      []string               `json:"issues"`
	Suggestions []string               `json:"suggestions"`
	CheckedAt   time.Time              `json:"checked_at"`
}

// HostComparison 主机横向对比结果
type HostComparison struct {
	Hosts       []HostHealthSummary `json:"hosts"`
	Slowest     string              `json:"slowest"`       // 响应最慢的主机
	HighestCPU  string              `json:"highest_cpu"`   // CPU 最高的主机
	HighestMem  string              `json:"highest_mem"`   // 内存最高的主机
	HighestDisk string              `json:"highest_disk"`  // 磁盘最高的主机
	Anomalies   []string            `json:"anomalies"`     // 异常主机列表
	AvgScore    int                 `json:"avg_score"`     // 平均健康分
	CheckedAt   time.Time           `json:"checked_at"`
}

// HostHealthSummary 主机健康摘要
type HostHealthSummary struct {
	HostID     string  `json:"host_id"`
	HostName   string  `json:"host_name"`
	Score      int     `json:"score"`
	ScoreLevel string  `json:"score_level"`
	CPU        float64 `json:"cpu"`
	Memory     float64 `json:"memory"`
	Disk       float64 `json:"disk"`
	Status     string  `json:"status"`
	IssueCount int     `json:"issue_count"`
}

// CheckHost 检查单个主机健康状态
func (s *HealthService) CheckHost(ctx context.Context, hostID string) (*HealthCheckResult, error) {
	host, err := s.hostRepo.GetByID(hostID)
	if err != nil {
		return nil, fmt.Errorf("主机不存在: %w", err)
	}

	result := &HealthCheckResult{
		HostID:    hostID,
		HostName:  host.Name,
		Metrics:   make(map[string]interface{}),
		Issues:    make([]string, 0),
		Suggestions: make([]string, 0),
		CheckedAt: time.Now(),
	}

	// 检查 CPU
	cpuUsage, err := s.getCPUUsage(host.Name)
	if err == nil {
		result.Metrics["cpu_usage"] = cpuUsage
		if cpuUsage > 80 {
			result.Issues = append(result.Issues, fmt.Sprintf("CPU 使用率过高: %.1f%%", cpuUsage))
			result.Suggestions = append(result.Suggestions, "检查 CPU 占用最高的进程")
		}
	}

	// 检查内存
	memUsage, err := s.getMemoryUsage(host.Name)
	if err == nil {
		result.Metrics["memory_usage"] = memUsage
		if memUsage > 80 {
			result.Issues = append(result.Issues, fmt.Sprintf("内存使用率过高: %.1f%%", memUsage))
			result.Suggestions = append(result.Suggestions, "检查内存占用最高的进程")
		}
	}

	// 检查磁盘
	diskUsage, err := s.getDiskUsage(host.Name)
	if err == nil {
		result.Metrics["disk_usage"] = diskUsage
		if diskUsage > 85 {
			result.Issues = append(result.Issues, fmt.Sprintf("磁盘使用率过高: %.1f%%", diskUsage))
			result.Suggestions = append(result.Suggestions, "清理不必要的文件或扩容磁盘")
		}
	}

	// 检查负载
	loadAvg, err := s.getLoadAverage(host.Name)
	if err == nil {
		result.Metrics["load_average"] = loadAvg
	}

	// 确定整体状态
	result.Status = s.determineStatus(result)

	// 计算健康分
	result.Score, result.ScoreLevel = s.CalculateHealthScore(result)

	// 保存到数据库
	healthCheck := &model.HealthCheck{
		ID:          uuid.New().String(),
		HostID:      hostID,
		CheckType:   model.HealthCheckTypeSystem,
		Status:      result.Status,
		Metrics:     result.Metrics,
		Issues:      result.Issues,
		Suggestions: result.Suggestions,
		CheckedAt:   result.CheckedAt,
		CreatedAt:   time.Now(),
	}

	if err := s.healthRepo.Create(healthCheck); err != nil {
		logger.Error("保存健康检查记录失败", zap.Error(err))
	}

	return result, nil
}

// DailyHealthCheck 每日健康体检
func (s *HealthService) DailyHealthCheck(ctx context.Context) (map[string]*HealthCheckResult, error) {
	hosts, err := s.hostRepo.List(repository.HostFilter{})
	if err != nil {
		return nil, fmt.Errorf("获取主机列表失败: %w", err)
	}

	results := make(map[string]*HealthCheckResult)

	for _, host := range hosts {
		result, err := s.CheckHost(ctx, host.ID)
		if err != nil {
			logger.Error("健康检查失败", zap.String("host", host.Name), zap.Error(err))
			continue
		}
		results[host.ID] = result
	}

	return results, nil
}

// GetHealthHistory 获取健康检查历史
func (s *HealthService) GetHealthHistory(hostID string, limit int) ([]*model.HealthCheck, error) {
	return s.healthRepo.FindByHost(hostID, limit)
}

// getCPUUsage 获取 CPU 使用率
func (s *HealthService) getCPUUsage(hostName string) (float64, error) {
	cmd := "top -bn1 | grep 'Cpu(s)' | awk '{print $2}' | cut -d'%' -f1"
	output, err := s.sshPool.Exec(hostName, cmd)
	if err != nil {
		return 0, err
	}

	var usage float64
	_, err = fmt.Sscanf(output, "%f", &usage)
	return usage, err
}

// getMemoryUsage 获取内存使用率
func (s *HealthService) getMemoryUsage(hostName string) (float64, error) {
	cmd := "free | grep Mem | awk '{printf \"%.2f\", $3/$2 * 100.0}'"
	output, err := s.sshPool.Exec(hostName, cmd)
	if err != nil {
		return 0, err
	}

	var usage float64
	_, err = fmt.Sscanf(output, "%f", &usage)
	return usage, err
}

// getDiskUsage 获取磁盘使用率
func (s *HealthService) getDiskUsage(hostName string) (float64, error) {
	cmd := "df -h / | tail -1 | awk '{print $5}' | cut -d'%' -f1"
	output, err := s.sshPool.Exec(hostName, cmd)
	if err != nil {
		return 0, err
	}

	var usage float64
	_, err = fmt.Sscanf(output, "%f", &usage)
	return usage, err
}

// getLoadAverage 获取负载平均值
func (s *HealthService) getLoadAverage(hostName string) (map[string]float64, error) {
	cmd := "uptime | awk -F'load average:' '{print $2}' | awk '{print $1, $2, $3}' | tr -d ','"
	output, err := s.sshPool.Exec(hostName, cmd)
	if err != nil {
		return nil, err
	}

	var load1, load5, load15 float64
	_, err = fmt.Sscanf(output, "%f %f %f", &load1, &load5, &load15)
	if err != nil {
		return nil, err
	}

	return map[string]float64{
		"1min":  load1,
		"5min":  load5,
		"15min": load15,
	}, nil
}

// determineStatus 确定健康状态
func (s *HealthService) determineStatus(result *HealthCheckResult) string {
	if len(result.Issues) == 0 {
		return model.HealthCheckStatusHealthy
	}

	// 检查是否有严重问题
	if cpu, ok := result.Metrics["cpu_usage"].(float64); ok && cpu > 90 {
		return model.HealthCheckStatusCritical
	}
	if mem, ok := result.Metrics["memory_usage"].(float64); ok && mem > 90 {
		return model.HealthCheckStatusCritical
	}
	if disk, ok := result.Metrics["disk_usage"].(float64); ok && disk > 95 {
		return model.HealthCheckStatusCritical
	}

	return model.HealthCheckStatusWarning
}

// CalculateHealthScore 计算健康分 (0-100)
func (s *HealthService) CalculateHealthScore(result *HealthCheckResult) (int, string) {
	score := 100

	// CPU 扣分: 50% 以上开始扣分
	if cpu, ok := result.Metrics["cpu_usage"].(float64); ok && cpu > 50 {
		deduction := int((cpu - 50) * 0.6)
		score -= deduction
	}

	// 内存扣分: 60% 以上开始扣分
	if mem, ok := result.Metrics["memory_usage"].(float64); ok && mem > 60 {
		deduction := int((mem - 60) * 0.5)
		score -= deduction
	}

	// 磁盘扣分: 70% 以上开始扣分
	if disk, ok := result.Metrics["disk_usage"].(float64); ok && disk > 70 {
		deduction := int((disk - 70) * 0.8)
		score -= deduction
	}

	// 问题数量扣分
	score -= len(result.Issues) * 5

	// 确保分数在 0-100 范围内
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	// 确定等级
	var level string
	switch {
	case score >= 90:
		level = "excellent"
	case score >= 75:
		level = "good"
	case score >= 60:
		level = "warning"
	default:
		level = "critical"
	}

	return score, level
}

// CompareHosts 横向对比多个主机
func (s *HealthService) CompareHosts(ctx context.Context, hostIDs []string) (*HostComparison, error) {
	if len(hostIDs) == 0 {
		// 如果没有指定主机，获取所有主机
		hosts, err := s.hostRepo.List(repository.HostFilter{})
		if err != nil {
			return nil, fmt.Errorf("获取主机列表失败: %w", err)
		}
		for _, h := range hosts {
			hostIDs = append(hostIDs, h.ID)
		}
	}

	comparison := &HostComparison{
		Hosts:     make([]HostHealthSummary, 0, len(hostIDs)),
		Anomalies: make([]string, 0),
		CheckedAt: time.Now(),
	}

	var (
		maxCPU, maxMem, maxDisk float64
		totalScore              int
	)

	for _, hostID := range hostIDs {
		result, err := s.CheckHost(ctx, hostID)
		if err != nil {
			logger.Error("检查主机失败", zap.String("host_id", hostID), zap.Error(err))
			continue
		}

		cpu, _ := result.Metrics["cpu_usage"].(float64)
		mem, _ := result.Metrics["memory_usage"].(float64)
		disk, _ := result.Metrics["disk_usage"].(float64)

		summary := HostHealthSummary{
			HostID:     result.HostID,
			HostName:   result.HostName,
			Score:      result.Score,
			ScoreLevel: result.ScoreLevel,
			CPU:        cpu,
			Memory:     mem,
			Disk:       disk,
			Status:     result.Status,
			IssueCount: len(result.Issues),
		}
		comparison.Hosts = append(comparison.Hosts, summary)
		totalScore += result.Score

		// 记录最高值
		if cpu > maxCPU {
			maxCPU = cpu
			comparison.HighestCPU = result.HostName
		}
		if mem > maxMem {
			maxMem = mem
			comparison.HighestMem = result.HostName
		}
		if disk > maxDisk {
			maxDisk = disk
			comparison.HighestDisk = result.HostName
		}

		// 检测异常主机
		if result.Status == model.HealthCheckStatusCritical || result.Score < 60 {
			comparison.Anomalies = append(comparison.Anomalies, result.HostName)
		}
	}

	if len(comparison.Hosts) > 0 {
		comparison.AvgScore = totalScore / len(comparison.Hosts)
	}

	return comparison, nil
}

// GetOverallHealthScore 获取整体健康分
func (s *HealthService) GetOverallHealthScore(ctx context.Context) (int, string, error) {
	comparison, err := s.CompareHosts(ctx, nil)
	if err != nil {
		return 0, "", err
	}

	score := comparison.AvgScore
	var level string
	switch {
	case score >= 90:
		level = "excellent"
	case score >= 75:
		level = "good"
	case score >= 60:
		level = "warning"
	default:
		level = "critical"
	}

	return score, level, nil
}
