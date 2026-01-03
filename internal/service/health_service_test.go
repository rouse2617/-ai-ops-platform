package service

import (
	"context"
	"testing"

	"ai-ops/internal/model"

	"github.com/stretchr/testify/assert"
)

func TestHealthService_DetermineStatus(t *testing.T) {
	service := &HealthService{}

	tests := []struct {
		name   string
		result *HealthCheckResult
		want   string
	}{
		{
			name: "健康状态",
			result: &HealthCheckResult{
				Metrics: map[string]interface{}{
					"cpu_usage":    50.0,
					"memory_usage": 60.0,
					"disk_usage":   70.0,
				},
				Issues: []string{},
			},
			want: model.HealthCheckStatusHealthy,
		},
		{
			name: "警告状态",
			result: &HealthCheckResult{
				Metrics: map[string]interface{}{
					"cpu_usage":    85.0,
					"memory_usage": 60.0,
					"disk_usage":   70.0,
				},
				Issues: []string{"CPU 使用率过高: 85.0%"},
			},
			want: model.HealthCheckStatusWarning,
		},
		{
			name: "严重状态 - CPU",
			result: &HealthCheckResult{
				Metrics: map[string]interface{}{
					"cpu_usage":    95.0,
					"memory_usage": 60.0,
					"disk_usage":   70.0,
				},
				Issues: []string{"CPU 使用率过高: 95.0%"},
			},
			want: model.HealthCheckStatusCritical,
		},
		{
			name: "严重状态 - 内存",
			result: &HealthCheckResult{
				Metrics: map[string]interface{}{
					"cpu_usage":    50.0,
					"memory_usage": 95.0,
					"disk_usage":   70.0,
				},
				Issues: []string{"内存使用率过高: 95.0%"},
			},
			want: model.HealthCheckStatusCritical,
		},
		{
			name: "严重状态 - 磁盘",
			result: &HealthCheckResult{
				Metrics: map[string]interface{}{
					"cpu_usage":    50.0,
					"memory_usage": 60.0,
					"disk_usage":   96.0,
				},
				Issues: []string{"磁盘使用率过高: 96.0%"},
			},
			want: model.HealthCheckStatusCritical,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := service.determineStatus(tt.result)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestHealthService_CheckHost_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	// 这个测试需要真实的 SSH 连接和数据库
	// 在实际环境中运行
}

func BenchmarkHealthService_DetermineStatus(b *testing.B) {
	service := &HealthService{}
	result := &HealthCheckResult{
		Metrics: map[string]interface{}{
			"cpu_usage":    85.0,
			"memory_usage": 75.0,
			"disk_usage":   80.0,
		},
		Issues: []string{"CPU 使用率过高: 85.0%"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.determineStatus(result)
	}
}
