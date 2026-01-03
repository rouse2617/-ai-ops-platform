package prometheus

import (
	"context"
	"fmt"
	"time"
)

// IncidentReplayQuery 故障复盘查询
type IncidentReplayQuery struct {
	client *PrometheusClient
}

// NewIncidentReplayQuery 创建故障复盘查询
func NewIncidentReplayQuery(client *PrometheusClient) *IncidentReplayQuery {
	return &IncidentReplayQuery{client: client}
}

// IncidentTimeline 故障时间线
type IncidentTimeline struct {
	IncidentTime time.Time
	PreWindow    time.Time // 故障前15分钟
	PostWindow   time.Time // 恢复后15分钟
	Metrics      map[string]*MetricTimeSeries
}

// MetricTimeSeries 指标时间序列
type MetricTimeSeries struct {
	MetricName string
	Labels     map[string]string
	Values     []MetricPoint
	Baseline   float64 // 基线值
	Peak       float64 // 峰值
	Anomaly    bool    // 是否异常
}

// MetricPoint 指标数据点
type MetricPoint struct {
	Timestamp time.Time
	Value     float64
}

// QueryIncidentMetrics 查询故障时间段的关键指标
func (irq *IncidentReplayQuery) QueryIncidentMetrics(ctx context.Context, incidentTime time.Time, host string) (*IncidentTimeline, error) {
	preWindow := incidentTime.Add(-15 * time.Minute)
	postWindow := incidentTime.Add(15 * time.Minute)

	timeline := &IncidentTimeline{
		IncidentTime: incidentTime,
		PreWindow:    preWindow,
		PostWindow:   postWindow,
		Metrics:      make(map[string]*MetricTimeSeries),
	}

	// 关键指标列表
	keyMetrics := []string{
		fmt.Sprintf(`node_cpu_seconds_total{instance="%s"}`, host),
		fmt.Sprintf(`node_memory_MemAvailable_bytes{instance="%s"}`, host),
		fmt.Sprintf(`node_disk_io_time_seconds_total{instance="%s"}`, host),
		fmt.Sprintf(`node_network_receive_bytes_total{instance="%s"}`, host),
		fmt.Sprintf(`node_network_transmit_bytes_total{instance="%s"}`, host),
		fmt.Sprintf(`node_load1{instance="%s"}`, host),
	}

	step := 30 * time.Second

	for _, metric := range keyMetrics {
		matrix, err := irq.client.QueryRange(ctx, metric, preWindow, postWindow, step)
		if err != nil {
			continue
		}

		for _, series := range matrix {
			metricName := string(series.Metric["__name__"])
			ts := &MetricTimeSeries{
				MetricName: metricName,
				Labels:     make(map[string]string),
				Values:     make([]MetricPoint, 0),
			}

			// 提取标签
			for k, v := range series.Metric {
				if k != "__name__" {
					ts.Labels[string(k)] = string(v)
				}
			}

			// 提取数据点
			for _, sample := range series.Values {
				ts.Values = append(ts.Values, MetricPoint{
					Timestamp: time.Unix(0, int64(sample.Timestamp)*1e6),
					Value:     float64(sample.Value),
				})
			}

			// 计算统计信息
			if len(ts.Values) > 0 {
				ts.Baseline = irq.calculateBaseline(ts.Values, preWindow, incidentTime)
				ts.Peak = irq.calculatePeak(ts.Values)
				ts.Anomaly = irq.detectAnomaly(ts.Values, ts.Baseline)
			}

			timeline.Metrics[metricName] = ts
		}
	}

	return timeline, nil
}

// calculateBaseline 计算基线值（故障前的平均值）
func (irq *IncidentReplayQuery) calculateBaseline(values []MetricPoint, start, end time.Time) float64 {
	var sum float64
	var count int

	for _, point := range values {
		if point.Timestamp.After(start) && point.Timestamp.Before(end) {
			sum += point.Value
			count++
		}
	}

	if count == 0 {
		return 0
	}
	return sum / float64(count)
}

// calculatePeak 计算峰值
func (irq *IncidentReplayQuery) calculatePeak(values []MetricPoint) float64 {
	if len(values) == 0 {
		return 0
	}

	peak := values[0].Value
	for _, point := range values {
		if point.Value > peak {
			peak = point.Value
		}
	}
	return peak
}

// detectAnomaly 检测异常
func (irq *IncidentReplayQuery) detectAnomaly(values []MetricPoint, baseline float64) bool {
	if baseline == 0 {
		return false
	}

	for _, point := range values {
		deviation := (point.Value - baseline) / baseline
		if deviation > 0.5 || deviation < -0.5 { // 50% 偏差
			return true
		}
	}
	return false
}

// PromQLQueries 常用 PromQL 查询语句
type PromQLQueries struct{}

// CPUUsageQuery CPU 使用率查询
func (pq *PromQLQueries) CPUUsageQuery(host string) string {
	return fmt.Sprintf(`100 - (avg by (instance) (rate(node_cpu_seconds_total{mode="idle",instance="%s"}[5m])) * 100)`, host)
}

// MemoryUsageQuery 内存使用率查询
func (pq *PromQLQueries) MemoryUsageQuery(host string) string {
	return fmt.Sprintf(`(1 - (node_memory_MemAvailable_bytes{instance="%s"} / node_memory_MemTotal_bytes{instance="%s"})) * 100`, host, host)
}

// DiskIOQuery 磁盘 I/O 查询
func (pq *PromQLQueries) DiskIOQuery(host string) string {
	return fmt.Sprintf(`rate(node_disk_io_time_seconds_total{instance="%s"}[5m])`, host)
}

// NetworkBytesQuery 网络流量查询
func (pq *PromQLQueries) NetworkBytesQuery(host string) string {
	return fmt.Sprintf(`rate(node_network_receive_bytes_total{instance="%s"}[5m]) + rate(node_network_transmit_bytes_total{instance="%s"}[5m])`, host, host)
}

// LoadAverageQuery 负载平均值查询
func (pq *PromQLQueries) LoadAverageQuery(host string) string {
	return fmt.Sprintf(`node_load1{instance="%s"}`, host)
}

// ContextSwitchQuery 上下文切换查询
func (pq *PromQLQueries) ContextSwitchQuery(host string) string {
	return fmt.Sprintf(`rate(node_context_switches_total{instance="%s"}[5m])`, host)
}

// ProcessCountQuery 进程数查询
func (pq *PromQLQueries) ProcessCountQuery(host string) string {
	return fmt.Sprintf(`node_procs_running{instance="%s"}`, host)
}

// FileDescriptorQuery 文件描述符查询
func (pq *PromQLQueries) FileDescriptorQuery(host string) string {
	return fmt.Sprintf(`node_filefd_allocated{instance="%s"}`, host)
}

// TCPConnectionQuery TCP 连接数查询
func (pq *PromQLQueries) TCPConnectionQuery(host string) string {
	return fmt.Sprintf(`node_sockstat_TCP_inuse{instance="%s"}`, host)
}

// IncidentReport 故障报告
type IncidentReport struct {
	IncidentTime    time.Time
	Duration        time.Duration
	AffectedMetrics []string
	RootCauses      []string
	Timeline        string
	Recommendations []string
}

// GenerateIncidentReport 生成故障报告
func (irq *IncidentReplayQuery) GenerateIncidentReport(timeline *IncidentTimeline) *IncidentReport {
	report := &IncidentReport{
		IncidentTime:    timeline.IncidentTime,
		AffectedMetrics: make([]string, 0),
		RootCauses:      make([]string, 0),
		Recommendations: make([]string, 0),
	}

	// 分析异常指标
	for metricName, ts := range timeline.Metrics {
		if ts.Anomaly {
			report.AffectedMetrics = append(report.AffectedMetrics, metricName)

			// 根据指标类型推断根因
			if ts.Peak > ts.Baseline*1.5 {
				switch metricName {
				case "node_cpu_seconds_total":
					report.RootCauses = append(report.RootCauses, "CPU 使用率异常升高")
					report.Recommendations = append(report.Recommendations, "检查高 CPU 进程，考虑优化代码或扩容")
				case "node_memory_MemAvailable_bytes":
					report.RootCauses = append(report.RootCauses, "内存使用率异常升高")
					report.Recommendations = append(report.Recommendations, "检查内存泄漏，考虑增加内存或优化应用")
				case "node_disk_io_time_seconds_total":
					report.RootCauses = append(report.RootCauses, "磁盘 I/O 异常")
					report.Recommendations = append(report.Recommendations, "检查磁盘写入进程，考虑使用 SSD 或优化 I/O")
				}
			}
		}
	}

	return report
}
