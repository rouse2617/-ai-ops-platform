package prometheus

import (
	"fmt"
	"math"
	"sort"
	"sync"
	"time"
)

// ThresholdDetector 动态阈值检测器
type ThresholdDetector struct {
	baselines map[string]*BaselineMetrics
	history   map[string][]MetricSample
	mu        sync.RWMutex
	maxSize   int
}

// NewThresholdDetector 创建动态阈值检测器
func NewThresholdDetector(maxHistorySize int) *ThresholdDetector {
	return &ThresholdDetector{
		baselines: make(map[string]*BaselineMetrics),
		history:   make(map[string][]MetricSample),
		maxSize:   maxHistorySize,
	}
}

// LearnBaseline 学习基线
func (td *ThresholdDetector) LearnBaseline(metrics []MetricSample) (BaselineMetrics, error) {
	if len(metrics) == 0 {
		return BaselineMetrics{}, fmt.Errorf("no metrics provided")
	}

	td.mu.Lock()
	defer td.mu.Unlock()

	metricName := metrics[0].Metric["__name__"]
	values := make([]float64, len(metrics))

	for i, m := range metrics {
		values[i] = m.Value
		td.addToHistory(metricName, m)
	}

	baseline := td.calculateBaseline(metricName, values)
	td.baselines[metricName] = &baseline

	return baseline, nil
}

// DetectAnomaly 检测异常
func (td *ThresholdDetector) DetectAnomaly(current MetricSample, baseline BaselineMetrics) AnomalyResult {
	td.mu.RLock()
	defer td.mu.RUnlock()

	deviation := math.Abs(current.Value - baseline.Mean)
	stdDevs := deviation / baseline.StdDev

	// 使用 3-sigma 规则
	isAnomaly := stdDevs > 3.0

	severity := td.calculateSeverity(current.Value, baseline)

	recommendation := td.generateRecommendation(current.Value, baseline, isAnomaly)

	return AnomalyResult{
		IsAnomaly:      isAnomaly,
		Severity:       severity,
		Deviation:      (deviation / baseline.Mean) * 100,
		Recommendation: recommendation,
	}
}

// UpdateThreshold 更新动态阈值
func (td *ThresholdDetector) UpdateThreshold(metric string, baseline BaselineMetrics) error {
	td.mu.Lock()
	defer td.mu.Unlock()

	td.baselines[metric] = &baseline
	return nil
}

// GetBaseline 获取基线
func (td *ThresholdDetector) GetBaseline(metric string) (*BaselineMetrics, error) {
	td.mu.RLock()
	defer td.mu.RUnlock()

	baseline, ok := td.baselines[metric]
	if !ok {
		return nil, fmt.Errorf("baseline not found for metric: %s", metric)
	}

	return baseline, nil
}

// 私有方法

func (td *ThresholdDetector) addToHistory(metricName string, sample MetricSample) {
	history := td.history[metricName]
	history = append(history, sample)

	if len(history) > td.maxSize {
		history = history[len(history)-td.maxSize:]
	}

	td.history[metricName] = history
}

func (td *ThresholdDetector) calculateBaseline(metricName string, values []float64) BaselineMetrics {
	sort.Float64s(values)

	mean := td.calculateMean(values)
	stdDev := td.calculateStdDev(values, mean)
	p95 := td.calculatePercentile(values, 0.95)
	p99 := td.calculatePercentile(values, 0.99)

	return BaselineMetrics{
		MetricName: metricName,
		Mean:       mean,
		StdDev:     stdDev,
		P95:        p95,
		P99:        p99,
		UpdatedAt:  time.Now(),
	}
}

func (td *ThresholdDetector) calculateMean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	sum := 0.0
	for _, v := range values {
		sum += v
	}

	return sum / float64(len(values))
}

func (td *ThresholdDetector) calculateStdDev(values []float64, mean float64) float64 {
	if len(values) <= 1 {
		return 0
	}

	variance := 0.0
	for _, v := range values {
		diff := v - mean
		variance += diff * diff
	}

	variance /= float64(len(values) - 1)
	return math.Sqrt(variance)
}

func (td *ThresholdDetector) calculatePercentile(values []float64, percentile float64) float64 {
	if len(values) == 0 {
		return 0
	}

	index := int(float64(len(values)) * percentile)
	if index >= len(values) {
		index = len(values) - 1
	}

	return values[index]
}

func (td *ThresholdDetector) calculateSeverity(value float64, baseline BaselineMetrics) float64 {
	deviation := math.Abs(value - baseline.Mean)
	stdDevs := deviation / baseline.StdDev

	// 将标准差转换为 0-1 的严重程度
	severity := math.Min(1.0, stdDevs/5.0)
	return severity
}

func (td *ThresholdDetector) generateRecommendation(value float64, baseline BaselineMetrics, isAnomaly bool) string {
	if !isAnomaly {
		return "指标正常，无需采取行动"
	}

	if value > baseline.P99 {
		return "指标值超过 99 分位数，建议立即调查"
	}

	if value > baseline.P95 {
		return "指标值超过 95 分位数，建议关注"
	}

	deviation := (math.Abs(value - baseline.Mean) / baseline.Mean) * 100
	if deviation > 50 {
		return fmt.Sprintf("指标值偏离基线 %.1f%%，建议进行分析", deviation)
	}

	return "检测到异常，建议进一步调查"
}
