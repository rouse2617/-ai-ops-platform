package service

import (
	"testing"
	"time"

	"ai-ops/internal/model"

	"github.com/stretchr/testify/assert"
)

func TestTrendService_LinearRegression(t *testing.T) {
	service := &TrendService{}

	tests := []struct {
		name       string
		values     []float64
		wantTrend  string
		minConfid  float64
	}{
		{
			name:      "上升趋势",
			values:    []float64{10, 15, 20, 25, 30},
			wantTrend: "increasing",
			minConfid: 0.9,
		},
		{
			name:      "下降趋势",
			values:    []float64{30, 25, 20, 15, 10},
			wantTrend: "decreasing",
			minConfid: 0.9,
		},
		{
			name:      "稳定趋势",
			values:    []float64{50, 51, 50, 49, 50},
			wantTrend: "stable",
			minConfid: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			predicted, confidence := service.linearRegression(tt.values)
			trend := service.determineTrend(tt.values)

			assert.Equal(t, tt.wantTrend, trend)
			assert.GreaterOrEqual(t, confidence, tt.minConfid)
			assert.Greater(t, predicted, 0.0)
		})
	}
}

func TestTrendService_DetermineTrend(t *testing.T) {
	service := &TrendService{}

	tests := []struct {
		name   string
		values []float64
		want   string
	}{
		{
			name:   "明显上升",
			values: []float64{10, 15, 20, 25, 30},
			want:   "increasing",
		},
		{
			name:   "明显下降",
			values: []float64{30, 25, 20, 15, 10},
			want:   "decreasing",
		},
		{
			name:   "基本稳定",
			values: []float64{50, 51, 50, 49, 50},
			want:   "stable",
		},
		{
			name:   "轻微波动",
			values: []float64{50, 51, 52, 51, 52},
			want:   "stable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := service.determineTrend(tt.values)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestTrendService_AnalyzeCPUTrend(t *testing.T) {
	service := &TrendService{}

	host := &model.Host{
		ID:   "test-host",
		Name: "test-server",
	}

	history := []*model.HealthCheck{
		{Metrics: map[string]interface{}{"cpu_usage": 50.0}, CheckedAt: time.Now().Add(-5 * time.Minute)},
		{Metrics: map[string]interface{}{"cpu_usage": 60.0}, CheckedAt: time.Now().Add(-4 * time.Minute)},
		{Metrics: map[string]interface{}{"cpu_usage": 70.0}, CheckedAt: time.Now().Add(-3 * time.Minute)},
		{Metrics: map[string]interface{}{"cpu_usage": 80.0}, CheckedAt: time.Now().Add(-2 * time.Minute)},
		{Metrics: map[string]interface{}{"cpu_usage": 85.0}, CheckedAt: time.Now().Add(-1 * time.Minute)},
	}

	result := service.analyzeCPUTrend(host, history)

	assert.NotNil(t, result)
	assert.Equal(t, "test-host", result.HostID)
	assert.Equal(t, model.MetricTypeCPU, result.MetricType)
	assert.Equal(t, 85.0, result.CurrentValue)
	assert.Greater(t, result.PredictedValue, 85.0)
	assert.Equal(t, "increasing", result.Trend)
	assert.NotEmpty(t, result.AlertLevel)
}

func BenchmarkLinearRegression(b *testing.B) {
	service := &TrendService{}
	values := []float64{10, 15, 20, 25, 30, 35, 40, 45, 50}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.linearRegression(values)
	}
}

func BenchmarkDetermineTrend(b *testing.B) {
	service := &TrendService{}
	values := []float64{10, 15, 20, 25, 30, 35, 40, 45, 50}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.determineTrend(values)
	}
}
