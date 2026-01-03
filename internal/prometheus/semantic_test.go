package prometheus

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNLQueryTranslator_PatternMatch(t *testing.T) {
	translator := &NLQueryTranslator{}

	tests := []struct {
		name          string
		query         string
		expectMatched bool
		expectPromQL  string
	}{
		{
			name:          "CPU 使用率查询",
			query:         "查看 CPU 使用率",
			expectMatched: true,
			expectPromQL:  `100 - (avg by(instance) (rate(node_cpu_seconds_total{mode="idle"}[5m])) * 100)`,
		},
		{
			name:          "内存使用率查询",
			query:         "查看内存占用情况",
			expectMatched: true,
			expectPromQL:  `100 - (node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes * 100)`,
		},
		{
			name:          "磁盘使用率查询",
			query:         "磁盘空间使用率",
			expectMatched: true,
			expectPromQL:  `100 - (node_filesystem_avail_bytes{mountpoint="/"} / node_filesystem_size_bytes{mountpoint="/"} * 100)`,
		},
		{
			name:          "网络流量查询",
			query:         "网络流量带宽",
			expectMatched: true,
			expectPromQL:  `rate(node_network_receive_bytes_total[5m]) + rate(node_network_transmit_bytes_total[5m])`,
		},
		{
			name:          "无法匹配的查询",
			query:         "今天天气怎么样",
			expectMatched: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matched, promQL, _, _, _ := translator.patternMatch(tt.query)
			assert.Equal(t, tt.expectMatched, matched)
			if tt.expectMatched {
				assert.Equal(t, tt.expectPromQL, promQL)
			}
		})
	}
}

func TestNLQueryTranslator_TimeRangeExtraction(t *testing.T) {
	translator := &NLQueryTranslator{}

	tests := []struct {
		query           string
		expectedRange   string
	}{
		{"查看过去 3 小时的 CPU", "3h"},
		{"查看 3h 内的内存", "3h"},
		{"过去 24 小时的磁盘", "24h"},
		{"一天内的网络流量", "24h"},
		{"过去 7 天的负载", "7d"},
		{"一周内的趋势", "7d"},
		{"查看 CPU 使用率", "1h"}, // 默认值
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			_, _, _, _, timeRange := translator.patternMatch(tt.query)
			if timeRange == "" {
				timeRange = "1h"
			}
			assert.Equal(t, tt.expectedRange, timeRange)
		})
	}
}

func TestDynamicAnomalyDetector_Statistics(t *testing.T) {
	tests := []struct {
		name     string
		values   []float64
		expected float64
		fn       func([]float64) float64
	}{
		{
			name:     "mean - 正常值",
			values:   []float64{1, 2, 3, 4, 5},
			expected: 3.0,
			fn:       mean,
		},
		{
			name:     "mean - 空数组",
			values:   []float64{},
			expected: 0,
			fn:       mean,
		},
		{
			name:     "min - 正常值",
			values:   []float64{5, 2, 8, 1, 9},
			expected: 1,
			fn:       min,
		},
		{
			name:     "max - 正常值",
			values:   []float64{5, 2, 8, 1, 9},
			expected: 9,
			fn:       max,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.fn(tt.values)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDynamicAnomalyDetector_Percentiles(t *testing.T) {
	values := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	percentiles := calculatePercentiles(values)

	assert.NotNil(t, percentiles)
	assert.Contains(t, percentiles, 50)
	assert.Contains(t, percentiles, 90)
	assert.Contains(t, percentiles, 95)
	assert.Contains(t, percentiles, 99)

	// P50 应该接近中位数
	assert.True(t, percentiles[50] >= 5 && percentiles[50] <= 6)
}

func TestAIPlaybook_MatchPlaybook(t *testing.T) {
	playbook := NewAIPlaybook(nil)

	tests := []struct {
		alertName string
		expectID  string
		expectNil bool
	}{
		{"DiskSpaceLow", "disk_space_low", false},
		{"HighCPU", "high_cpu", false},
		{"MemoryLow", "memory_low", false},
		{"ServiceDown", "service_down", false},
		{"UnknownAlert", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.alertName, func(t *testing.T) {
			result := playbook.MatchPlaybook(tt.alertName, nil)
			if tt.expectNil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectID, result.ID)
			}
		})
	}
}

func TestAIPlaybook_ListPlaybooks(t *testing.T) {
	playbook := NewAIPlaybook(nil)
	list := playbook.ListPlaybooks()

	assert.True(t, len(list) >= 4) // 至少有 4 个默认剧本
}

func TestDynamicDashboard_DetectTopic(t *testing.T) {
	dashboard := NewDynamicDashboard(nil)

	tests := []struct {
		text          string
		expectedTopic string
	}{
		{"数据库连接数太多了", "database"},
		{"MySQL 慢查询", "database"},
		{"网络丢包严重", "network"},
		{"TCP 连接数", "network"},
		{"CPU 负载很高", "cpu"},
		{"内存不足", "memory"},
		{"磁盘空间满了", "disk"},
		{"Docker 容器重启", "container"},
		{"Kubernetes pod", "container"},
		{"HTTP 请求延迟", "http"},
		{"API 错误率", "http"},
		{"JVM 堆内存", "application"},
		{"随便聊聊", "overview"},
	}

	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			topic := dashboard.detectTopic(tt.text)
			assert.Equal(t, tt.expectedTopic, topic)
		})
	}
}

func TestDynamicDashboard_GenerateDashboard(t *testing.T) {
	dashboard := NewDynamicDashboard(nil)

	topics := []string{"database", "network", "cpu", "memory", "disk", "container", "http", "application", "overview"}

	for _, topic := range topics {
		t.Run(topic, func(t *testing.T) {
			config := dashboard.GenerateDashboard(nil, topic)
			assert.NotNil(t, config)
			assert.Equal(t, topic, config.Topic)
			assert.True(t, len(config.Panels) > 0)
			assert.True(t, config.RefreshRate > 0)
		})
	}
}

func TestDynamicDashboard_MatchKeywordsInText(t *testing.T) {
	dashboard := NewDynamicDashboard(nil)

	text := "MySQL 数据库连接数太多，导致 CPU 负载升高"
	matches := dashboard.MatchKeywordsInText(text)

	assert.Contains(t, matches, "database")
	assert.Contains(t, matches, "cpu")
	assert.Contains(t, matches["database"], "mysql")
	assert.Contains(t, matches["database"], "数据库")
}
