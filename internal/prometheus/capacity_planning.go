package prometheus

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"
)

// CapacityPlanner 容量规划器
type CapacityPlanner struct {
	client *PrometheusClient
}

// NewCapacityPlanner 创建容量规划器
func NewCapacityPlanner(client *PrometheusClient) *CapacityPlanner {
	return &CapacityPlanner{client: client}
}

// LoadTrend 负载趋势
type LoadTrend struct {
	MetricName    string
	Host          string
	DataPoints    []TrendPoint
	Trend         TrendType
	GrowthRate    float64 // 每天增长率
	ProjectedPeak float64 // 预测峰值
	DaysToCapacity int     // 距离容量上限的天数
}

// TrendType 趋势类型
type TrendType string

const (
	TrendStable    TrendType = "stable"
	TrendIncreasing TrendType = "increasing"
	TrendDecreasing TrendType = "decreasing"
	TrendCyclic    TrendType = "cyclic"
)

// TrendPoint 趋势数据点
type TrendPoint struct {
	Timestamp time.Time
	Value     float64
	Predicted float64 // 预测值
}

// GetHistoricalData 获取历史负载数据
func (cp *CapacityPlanner) GetHistoricalData(ctx context.Context, query string, days int) ([]TrendPoint, error) {
	end := time.Now()
	start := end.AddDate(0, 0, -days)
	step := time.Duration(days*24) * time.Hour / 100 // 100 个数据点

	matrix, err := cp.client.QueryRange(ctx, query, start, end, step)
	if err != nil {
		return nil, err
	}

	points := make([]TrendPoint, 0)
	for _, series := range matrix {
		for _, sample := range series.Values {
			points = append(points, TrendPoint{
				Timestamp: time.Unix(0, int64(sample.Timestamp)*1e6),
				Value:     float64(sample.Value),
			})
		}
	}

	sort.Slice(points, func(i, j int) bool {
		return points[i].Timestamp.Before(points[j].Timestamp)
	})

	return points, nil
}

// AnalyzeTrend 分析趋势
func (cp *CapacityPlanner) AnalyzeTrend(points []TrendPoint) *LoadTrend {
	if len(points) < 2 {
		return &LoadTrend{Trend: TrendStable}
	}

	trend := &LoadTrend{
		DataPoints: points,
	}

	// 线性回归分析
	slope, intercept := cp.linearRegression(points)
	trend.GrowthRate = slope

	// 判断趋势类型
	if math.Abs(slope) < 0.01 {
		trend.Trend = TrendStable
	} else if slope > 0 {
		trend.Trend = TrendIncreasing
	} else {
		trend.Trend = TrendDecreasing
	}

	// 计算预测值
	for i := range points {
		x := float64(i)
		predicted := slope*x + intercept
		points[i].Predicted = predicted
	}

	// 预测峰值（30天后）
	lastIdx := float64(len(points) - 1)
	daysToPredict := 30.0
	pointsPerDay := lastIdx / float64(points[len(points)-1].Timestamp.Sub(points[0].Timestamp).Hours() / 24)
	futureIdx := lastIdx + (daysToPredict * pointsPerDay)
	trend.ProjectedPeak = slope*futureIdx + intercept

	return trend
}

// PredictCapacityExhaustion 预测容量耗尽时间
func (cp *CapacityPlanner) PredictCapacityExhaustion(trend *LoadTrend, capacityThreshold float64) int {
	if trend.GrowthRate <= 0 {
		return -1 // 不会耗尽
	}

	lastValue := trend.DataPoints[len(trend.DataPoints)-1].Value
	if lastValue >= capacityThreshold {
		return 0 // 已经超过容量
	}

	// 计算需要多少天达到容量
	daysNeeded := (capacityThreshold - lastValue) / trend.GrowthRate
	return int(math.Ceil(daysNeeded))
}

// linearRegression 线性回归
func (cp *CapacityPlanner) linearRegression(points []TrendPoint) (slope, intercept float64) {
	n := float64(len(points))
	if n < 2 {
		return 0, 0
	}

	var sumX, sumY, sumXY, sumX2 float64
	for i, point := range points {
		x := float64(i)
		y := point.Value
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}

	slope = (n*sumXY - sumX*sumY) / (n*sumX2 - sumX*sumX)
	intercept = (sumY - slope*sumX) / n

	return slope, intercept
}

// ExponentialSmoothing 指数平滑预测
func (cp *CapacityPlanner) ExponentialSmoothing(points []TrendPoint, alpha float64, periods int) []TrendPoint {
	if len(points) == 0 {
		return nil
	}

	forecast := make([]TrendPoint, 0)
	s := points[0].Value // 初始平滑值

	for i := 0; i < len(points); i++ {
		s = alpha*points[i].Value + (1-alpha)*s
		forecast = append(forecast, TrendPoint{
			Timestamp: points[i].Timestamp,
			Value:     points[i].Value,
			Predicted: s,
		})
	}

	// 预测未来周期
	lastTime := points[len(points)-1].Timestamp
	interval := lastTime.Sub(points[0].Timestamp) / time.Duration(len(points)-1)

	for i := 0; i < periods; i++ {
		futureTime := lastTime.Add(interval * time.Duration(i+1))
		forecast = append(forecast, TrendPoint{
			Timestamp: futureTime,
			Predicted: s,
		})
	}

	return forecast
}

// WhatIfScenario What-if 场景分析
type WhatIfScenario struct {
	Name              string
	GrowthMultiplier  float64 // 增长倍数
	PeakMultiplier    float64 // 峰值倍数
	ProjectedMetrics  map[string]float64
	RecommendedAction string
	TimeToAction      int // 建议采取行动的天数
}

// AnalyzeWhatIfScenarios 分析 What-if 场景
func (cp *CapacityPlanner) AnalyzeWhatIfScenarios(trend *LoadTrend, capacityThreshold float64) []WhatIfScenario {
	scenarios := make([]WhatIfScenario, 0)

	// 场景1：正常增长
	scenario1 := WhatIfScenario{
		Name:             "正常增长",
		GrowthMultiplier: 1.0,
		PeakMultiplier:   1.0,
	}
	scenario1.ProjectedMetrics = map[string]float64{
		"30day_peak":       trend.ProjectedPeak,
		"days_to_capacity": float64(cp.PredictCapacityExhaustion(trend, capacityThreshold)),
	}
	if daysToCapacity := scenario1.ProjectedMetrics["days_to_capacity"]; daysToCapacity > 0 && daysToCapacity < 30 {
		scenario1.RecommendedAction = "立即规划扩容"
		scenario1.TimeToAction = 7
	}
	scenarios = append(scenarios, scenario1)

	// 场景2：加速增长（1.5倍）
	scenario2 := WhatIfScenario{
		Name:             "加速增长(1.5x)",
		GrowthMultiplier: 1.5,
		PeakMultiplier:   1.2,
	}
	acceleratedTrend := &LoadTrend{
		GrowthRate:    trend.GrowthRate * 1.5,
		ProjectedPeak: trend.ProjectedPeak * 1.2,
	}
	scenario2.ProjectedMetrics = map[string]float64{
		"30day_peak": acceleratedTrend.ProjectedPeak,
		"days_to_capacity": float64(cp.PredictCapacityExhaustion(acceleratedTrend, capacityThreshold)),
	}
	scenario2.RecommendedAction = "紧急扩容计划"
	scenario2.TimeToAction = 3
	scenarios = append(scenarios, scenario2)

	// 场景3：业务高峰（2倍）
	scenario3 := WhatIfScenario{
		Name:             "业务高峰(2x)",
		GrowthMultiplier: 2.0,
		PeakMultiplier:   2.0,
	}
	peakTrend := &LoadTrend{
		GrowthRate:    trend.GrowthRate * 2.0,
		ProjectedPeak: trend.ProjectedPeak * 2.0,
	}
	scenario3.ProjectedMetrics = map[string]float64{
		"30day_peak": peakTrend.ProjectedPeak,
		"days_to_capacity": float64(cp.PredictCapacityExhaustion(peakTrend, capacityThreshold)),
	}
	scenario3.RecommendedAction = "立即启动应急扩容"
	scenario3.TimeToAction = 1
	scenarios = append(scenarios, scenario3)

	return scenarios
}

// CapacityRecommendation 容量建议
type CapacityRecommendation struct {
	CurrentUsage      float64
	ProjectedUsage    float64
	CapacityThreshold float64
	RecommendedCapacity float64
	UtilizationRate   float64
	SafetyMargin      float64
	ActionItems       []string
	Priority          string // critical, high, medium, low
}

// GenerateCapacityRecommendation 生成容量建议
func (cp *CapacityPlanner) GenerateCapacityRecommendation(trend *LoadTrend, currentCapacity, capacityThreshold float64) *CapacityRecommendation {
	rec := &CapacityRecommendation{
		CurrentUsage:        trend.DataPoints[len(trend.DataPoints)-1].Value,
		ProjectedUsage:      trend.ProjectedPeak,
		CapacityThreshold:   capacityThreshold,
		UtilizationRate:     trend.DataPoints[len(trend.DataPoints)-1].Value / currentCapacity,
		SafetyMargin:        (currentCapacity - trend.ProjectedPeak) / currentCapacity,
		ActionItems:         make([]string, 0),
	}

	// 计算推荐容量（预测峰值 + 20% 安全余量）
	rec.RecommendedCapacity = trend.ProjectedPeak * 1.2

	// 判断优先级
	if rec.SafetyMargin < 0 {
		rec.Priority = "critical"
		rec.ActionItems = append(rec.ActionItems, "立即扩容，已超过安全容量")
	} else if rec.SafetyMargin < 0.1 {
		rec.Priority = "high"
		rec.ActionItems = append(rec.ActionItems, "紧急扩容计划，安全余量不足10%")
	} else if rec.SafetyMargin < 0.2 {
		rec.Priority = "medium"
		rec.ActionItems = append(rec.ActionItems, "计划扩容，安全余量不足20%")
	} else {
		rec.Priority = "low"
		rec.ActionItems = append(rec.ActionItems, "监控使用趋势，准备扩容计划")
	}

	// 根据增长率添加建议
	if trend.GrowthRate > 0.1 {
		rec.ActionItems = append(rec.ActionItems, fmt.Sprintf("增长率为 %.2f/天，建议加快扩容进度", trend.GrowthRate))
	}

	if rec.UtilizationRate > 0.8 {
		rec.ActionItems = append(rec.ActionItems, "当前利用率超过80%，建议优化资源使用")
	}

	return rec
}

// ResourceAllocationStrategy 资源分配策略
type ResourceAllocationStrategy struct {
	MetricName        string
	CurrentAllocation float64
	RecommendedAllocation float64
	AllocationRatio   float64
	Timeline          string // 立即, 1周内, 2周内, 1月内
	CostEstimate      float64
	ROI               float64
}

// CalculateResourceAllocation 计算资源分配
func (cp *CapacityPlanner) CalculateResourceAllocation(trend *LoadTrend, currentAllocation, costPerUnit float64) *ResourceAllocationStrategy {
	strategy := &ResourceAllocationStrategy{
		MetricName:            trend.MetricName,
		CurrentAllocation:     currentAllocation,
		RecommendedAllocation: trend.ProjectedPeak * 1.2,
	}

	strategy.AllocationRatio = strategy.RecommendedAllocation / currentAllocation

	// 确定时间线
	daysToCapacity := cp.PredictCapacityExhaustion(trend, currentAllocation)
	if daysToCapacity <= 0 {
		strategy.Timeline = "立即"
	} else if daysToCapacity <= 7 {
		strategy.Timeline = "1周内"
	} else if daysToCapacity <= 14 {
		strategy.Timeline = "2周内"
	} else {
		strategy.Timeline = "1月内"
	}

	// 计算成本
	additionalAllocation := strategy.RecommendedAllocation - currentAllocation
	strategy.CostEstimate = additionalAllocation * costPerUnit

	// 计算 ROI（假设避免故障的收益）
	strategy.ROI = (additionalAllocation * costPerUnit * 10) / strategy.CostEstimate // 假设收益是成本的10倍

	return strategy
}
