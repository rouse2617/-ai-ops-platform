package prometheus

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"
)

// DynamicAnomalyDetector 动态异常检测器 - 基于历史基线
type DynamicAnomalyDetector struct {
	client    *Client
	baselines map[string]*MetricBaseline
}

// MetricBaseline 指标基线
type MetricBaseline struct {
	MetricName string             `json:"metric_name"`
	Instance   string             `json:"instance"`
	Mean       float64            `json:"mean"`
	StdDev     float64            `json:"stddev"`
	Min        float64            `json:"min"`
	Max        float64            `json:"max"`
	Percentiles map[int]float64   `json:"percentiles"` // p50, p90, p95, p99
	HourlyPattern []float64       `json:"hourly_pattern"` // 24小时模式
	DayOfWeekPattern []float64    `json:"day_of_week_pattern"` // 7天模式
	UpdatedAt  time.Time          `json:"updated_at"`
}

// AnomalyResult 异常检测结果
type AnomalyResult struct {
	IsAnomaly       bool    `json:"is_anomaly"`
	CurrentValue    float64 `json:"current_value"`
	ExpectedValue   float64 `json:"expected_value"`
	ExpectedRange   [2]float64 `json:"expected_range"` // [min, max]
	DeviationScore  float64 `json:"deviation_score"` // 偏离分数 (0-100)
	DeviationFactor float64 `json:"deviation_factor"` // 相对于历史的倍数
	Severity        string  `json:"severity"` // low, medium, high, critical
	Description     string  `json:"description"`
	Suggestion      string  `json:"suggestion"`
}

// NewDynamicAnomalyDetector 创建动态异常检测器
func NewDynamicAnomalyDetector(client *Client) *DynamicAnomalyDetector {
	return &DynamicAnomalyDetector{
		client:    client,
		baselines: make(map[string]*MetricBaseline),
	}
}

// LearnBaseline 学习指标基线（过去7天）
func (d *DynamicAnomalyDetector) LearnBaseline(ctx context.Context, metricName, instance string) (*MetricBaseline, error) {
	key := fmt.Sprintf("%s:%s", metricName, instance)

	// 查询过去7天的数据
	query := fmt.Sprintf(`%s{instance="%s"}`, metricName, instance)
	end := time.Now()
	start := end.Add(-7 * 24 * time.Hour)

	result, err := d.client.QueryRangeWithTime(ctx, query, start, end, "5m")
	if err != nil {
		return nil, fmt.Errorf("查询历史数据失败: %w", err)
	}

	if len(result.Data.Result) == 0 {
		return nil, fmt.Errorf("没有找到历史数据")
	}

	// 提取所有值
	var values []float64
	hourlyValues := make([][]float64, 24)
	dayOfWeekValues := make([][]float64, 7)

	for _, r := range result.Data.Result {
		for _, v := range r.Values {
			if len(v) >= 2 {
				ts, _ := v[0].(float64)
				valStr, _ := v[1].(string)
				var val float64
				fmt.Sscanf(valStr, "%f", &val)

				values = append(values, val)

				// 按小时分组
				t := time.Unix(int64(ts), 0)
				hour := t.Hour()
				hourlyValues[hour] = append(hourlyValues[hour], val)

				// 按星期分组
				dayOfWeek := int(t.Weekday())
				dayOfWeekValues[dayOfWeek] = append(dayOfWeekValues[dayOfWeek], val)
			}
		}
	}

	if len(values) == 0 {
		return nil, fmt.Errorf("没有有效的数据点")
	}

	// 计算统计量
	baseline := &MetricBaseline{
		MetricName:       metricName,
		Instance:         instance,
		Mean:             mean(values),
		StdDev:           stdDev(values),
		Min:              min(values),
		Max:              max(values),
		Percentiles:      calculatePercentiles(values),
		HourlyPattern:    make([]float64, 24),
		DayOfWeekPattern: make([]float64, 7),
		UpdatedAt:        time.Now(),
	}

	// 计算小时模式
	for i, hv := range hourlyValues {
		if len(hv) > 0 {
			baseline.HourlyPattern[i] = mean(hv)
		}
	}

	// 计算星期模式
	for i, dv := range dayOfWeekValues {
		if len(dv) > 0 {
			baseline.DayOfWeekPattern[i] = mean(dv)
		}
	}

	d.baselines[key] = baseline
	return baseline, nil
}

// DetectAnomaly 检测异常
func (d *DynamicAnomalyDetector) DetectAnomaly(ctx context.Context, metricName, instance string, currentValue float64) (*AnomalyResult, error) {
	key := fmt.Sprintf("%s:%s", metricName, instance)

	baseline, exists := d.baselines[key]
	if !exists {
		// 尝试学习基线
		var err error
		baseline, err = d.LearnBaseline(ctx, metricName, instance)
		if err != nil {
			return nil, err
		}
	}

	// 获取当前时间的预期值
	now := time.Now()
	hour := now.Hour()
	dayOfWeek := int(now.Weekday())

	// 综合预期值（基于小时和星期模式）
	expectedValue := baseline.Mean
	if baseline.HourlyPattern[hour] > 0 {
		expectedValue = (expectedValue + baseline.HourlyPattern[hour]) / 2
	}
	if baseline.DayOfWeekPattern[dayOfWeek] > 0 {
		expectedValue = (expectedValue + baseline.DayOfWeekPattern[dayOfWeek]) / 2
	}

	// 计算预期范围（3-sigma）
	expectedMin := expectedValue - 3*baseline.StdDev
	expectedMax := expectedValue + 3*baseline.StdDev
	if expectedMin < 0 {
		expectedMin = 0
	}

	// 计算偏离分数
	deviation := math.Abs(currentValue - expectedValue)
	deviationScore := 0.0
	if baseline.StdDev > 0 {
		deviationScore = (deviation / baseline.StdDev) * 20 // 标准化到 0-100
		if deviationScore > 100 {
			deviationScore = 100
		}
	}

	// 计算偏离倍数
	deviationFactor := 1.0
	if expectedValue > 0 {
		deviationFactor = currentValue / expectedValue
	}

	// 判断是否异常
	isAnomaly := currentValue < expectedMin || currentValue > expectedMax

	// 确定严重程度
	severity := "low"
	if deviationScore > 80 {
		severity = "critical"
	} else if deviationScore > 60 {
		severity = "high"
	} else if deviationScore > 40 {
		severity = "medium"
	}

	// 生成描述
	description := d.generateDescription(metricName, currentValue, expectedValue, deviationFactor, isAnomaly)
	suggestion := d.generateSuggestion(metricName, severity, deviationFactor)

	return &AnomalyResult{
		IsAnomaly:       isAnomaly,
		CurrentValue:    currentValue,
		ExpectedValue:   expectedValue,
		ExpectedRange:   [2]float64{expectedMin, expectedMax},
		DeviationScore:  deviationScore,
		DeviationFactor: deviationFactor,
		Severity:        severity,
		Description:     description,
		Suggestion:      suggestion,
	}, nil
}

// generateDescription 生成描述
func (d *DynamicAnomalyDetector) generateDescription(metricName string, current, expected, factor float64, isAnomaly bool) string {
	if !isAnomaly {
		return fmt.Sprintf("当前值 %.2f 在正常范围内（预期 %.2f）", current, expected)
	}

	if factor > 1 {
		return fmt.Sprintf("虽然目前系统指标在安全范围内，但当前 %s 是历史同期的 %.1f 倍（当前 %.2f，预期 %.2f），存在潜在风���",
			metricName, factor, current, expected)
	}
	return fmt.Sprintf("当前 %s 低于历史同期（当前 %.2f，预期 %.2f）",
		metricName, current, expected)
}

// generateSuggestion 生成建议
func (d *DynamicAnomalyDetector) generateSuggestion(metricName, severity string, factor float64) string {
	switch severity {
	case "critical":
		return "建议立即检查相关服务，可能存在严重问题"
	case "high":
		return "建议密切关注，并准备好应急预案"
	case "medium":
		return "建议持续观察，如持续异常请进一步排查"
	default:
		return "当前状态正常，建议保持监控"
	}
}

// BatchDetect 批量检测多个指标
func (d *DynamicAnomalyDetector) BatchDetect(ctx context.Context, metrics []struct {
	Name     string
	Instance string
	Value    float64
}) ([]*AnomalyResult, error) {
	results := make([]*AnomalyResult, 0, len(metrics))

	for _, m := range metrics {
		result, err := d.DetectAnomaly(ctx, m.Name, m.Instance, m.Value)
		if err != nil {
			continue
		}
		results = append(results, result)
	}

	return results, nil
}

// GetBaseline 获取基线
func (d *DynamicAnomalyDetector) GetBaseline(metricName, instance string) *MetricBaseline {
	key := fmt.Sprintf("%s:%s", metricName, instance)
	return d.baselines[key]
}

// 统计函数
func mean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func stdDev(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	m := mean(values)
	sum := 0.0
	for _, v := range values {
		sum += (v - m) * (v - m)
	}
	return math.Sqrt(sum / float64(len(values)))
}

func min(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	m := values[0]
	for _, v := range values[1:] {
		if v < m {
			m = v
		}
	}
	return m
}

func max(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	m := values[0]
	for _, v := range values[1:] {
		if v > m {
			m = v
		}
	}
	return m
}

func calculatePercentiles(values []float64) map[int]float64 {
	if len(values) == 0 {
		return nil
	}

	sorted := make([]float64, len(values))
	copy(sorted, values)
	sort.Float64s(sorted)

	percentiles := map[int]float64{
		50: percentile(sorted, 50),
		90: percentile(sorted, 90),
		95: percentile(sorted, 95),
		99: percentile(sorted, 99),
	}

	return percentiles
}

func percentile(sorted []float64, p int) float64 {
	if len(sorted) == 0 {
		return 0
	}
	idx := int(float64(len(sorted)-1) * float64(p) / 100)
	return sorted[idx]
}
