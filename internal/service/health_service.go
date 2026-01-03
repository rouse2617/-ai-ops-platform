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
	Metrics     map[string]interface{} `json:"metrics"`
	Issues      []string               `json:"issues"`
	Suggestions []string               `json:"suggestions"`
	CheckedAt   time.Time              `json:"checked_at"`
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
