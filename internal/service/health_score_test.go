package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalculateHealthScore(t *testing.T) {
	s := &HealthService{}

	tests := []struct {
		name          string
		metrics       map[string]interface{}
		issues        []string
		expectedScore int
		expectedLevel string
	}{
		{
			name: "健康状态 - 所有指标正常",
			metrics: map[string]interface{}{
				"cpu_usage":    30.0,
				"memory_usage": 40.0,
				"disk_usage":   50.0,
			},
			issues:        []string{},
			expectedScore: 100,
			expectedLevel: "excellent",
		},
		{
			name: "良好状态 - CPU 稍高",
			metrics: map[string]interface{}{
				"cpu_usage":    70.0, // 扣分: (70-50)*0.6 = 12
				"memory_usage": 50.0,
				"disk_usage":   60.0,
			},
			issues:        []string{},
			expectedScore: 88,
			expectedLevel: "good",
		},
		{
			name: "警告状态 - 多项指标偏高",
			metrics: map[string]interface{}{
				"cpu_usage":    80.0, // 扣分: (80-50)*0.6 = 18
				"memory_usage": 75.0, // 扣分: (75-60)*0.5 = 7.5 -> 7
				"disk_usage":   80.0, // 扣分: (80-70)*0.8 = 8
			},
			issues:        []string{"CPU 使用率过高"},
			expectedScore: 62, // 100 - 18 - 7 - 8 - 5 = 62
			expectedLevel: "warning",
		},
		{
			name: "严重状态 - 所有指标过高",
			metrics: map[string]interface{}{
				"cpu_usage":    95.0, // 扣分: (95-50)*0.6 = 27
				"memory_usage": 90.0, // 扣分: (90-60)*0.5 = 15
				"disk_usage":   95.0, // 扣分: (95-70)*0.8 = 20
			},
			issues:        []string{"CPU 过高", "内存过高", "磁盘过高"},
			expectedScore: 23, // 100 - 27 - 15 - 20 - 15 = 23
			expectedLevel: "critical",
		},
		{
			name: "边界情况 - 分数不低于0",
			metrics: map[string]interface{}{
				"cpu_usage":    100.0,
				"memory_usage": 100.0,
				"disk_usage":   100.0,
			},
			issues:        []string{"问题1", "问题2", "问题3", "问题4", "问题5"},
			expectedScore: 0,
			expectedLevel: "critical",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &HealthCheckResult{
				Metrics: tt.metrics,
				Issues:  tt.issues,
			}

			score, level := s.CalculateHealthScore(result)

			assert.Equal(t, tt.expectedScore, score, "健康分不匹配")
			assert.Equal(t, tt.expectedLevel, level, "健康等级不匹配")
		})
	}
}

func TestDetermineStatus(t *testing.T) {
	s := &HealthService{}

	tests := []struct {
		name           string
		metrics        map[string]interface{}
		issues         []string
		expectedStatus string
	}{
		{
			name: "健康 - 无问题",
			metrics: map[string]interface{}{
				"cpu_usage":    50.0,
				"memory_usage": 50.0,
				"disk_usage":   50.0,
			},
			issues:         []string{},
			expectedStatus: "healthy",
		},
		{
			name: "警告 - 有问题但不严重",
			metrics: map[string]interface{}{
				"cpu_usage":    85.0,
				"memory_usage": 85.0,
				"disk_usage":   90.0,
			},
			issues:         []string{"CPU 使用率过高"},
			expectedStatus: "warning",
		},
		{
			name: "严重 - CPU 超过 90%",
			metrics: map[string]interface{}{
				"cpu_usage":    95.0,
				"memory_usage": 50.0,
				"disk_usage":   50.0,
			},
			issues:         []string{"CPU 使用率过高"},
			expectedStatus: "critical",
		},
		{
			name: "严重 - 内存超过 90%",
			metrics: map[string]interface{}{
				"cpu_usage":    50.0,
				"memory_usage": 95.0,
				"disk_usage":   50.0,
			},
			issues:         []string{"内存使用率过高"},
			expectedStatus: "critical",
		},
		{
			name: "严重 - 磁盘超过 95%",
			metrics: map[string]interface{}{
				"cpu_usage":    50.0,
				"memory_usage": 50.0,
				"disk_usage":   98.0,
			},
			issues:         []string{"磁盘使用率过高"},
			expectedStatus: "critical",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &HealthCheckResult{
				Metrics: tt.metrics,
				Issues:  tt.issues,
			}

			status := s.determineStatus(result)
			assert.Equal(t, tt.expectedStatus, status)
		})
	}
}
