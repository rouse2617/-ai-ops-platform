package prometheus

import (
	"fmt"
	"strings"
)

// SemanticEngine 语义翻译引擎
type SemanticEngine struct {
	metricDescriptions map[string]string
	alertPatterns      map[string]string
}

// NewSemanticEngine 创建语义引��
func NewSemanticEngine() *SemanticEngine {
	return &SemanticEngine{
		metricDescriptions: initMetricDescriptions(),
		alertPatterns:      initAlertPatterns(),
	}
}

// TranslateMetric 翻译指标为自然语言
func (se *SemanticEngine) TranslateMetric(metric MetricSample) SemanticTranslation {
	metricName := string(metric.Metric["__name__"])
	value := metric.Value

	desc := se.getMetricDescription(metricName)

	return SemanticTranslation{
		Title:       se.generateTitle(metricName, value),
		Description: se.generateDescription(metricName, value, desc),
		RootCause:   se.analyzeRootCause(metricName, value),
		Impact:      se.analyzeImpact(metricName, value),
		Suggestions: se.generateSuggestions(metricName, value),
	}
}

// GenerateAlertDescription 生成告警描述
func (se *SemanticEngine) GenerateAlertDescription(alert AlertMetadata, context AlertContext) string {
	parts := []string{
		fmt.Sprintf("告警: %s", alert.AlertName),
		fmt.Sprintf("严重级别: %s", alert.Severity),
		fmt.Sprintf("当前值: %.2f", context.Metric.Value),
		fmt.Sprintf("基线值: %.2f", context.Baseline.Mean),
		fmt.Sprintf("偏差: %.2f%%", context.Trend.Rate*100),
	}

	if context.Trend.Direction != "" {
		parts = append(parts, fmt.Sprintf("趋势: %s", context.Trend.Direction))
	}

	return strings.Join(parts, "\n")
}

// GenerateRecommendations 生成建议
func (se *SemanticEngine) GenerateRecommendations(anomaly AnomalyResult) []string {
	recommendations := []string{}

	if anomaly.Severity > 0.8 {
		recommendations = append(recommendations, "立即采取行动，异常严重程度很高")
	} else if anomaly.Severity > 0.5 {
		recommendations = append(recommendations, "需要关注，异常严重程度中等")
	}

	if anomaly.Deviation > 50 {
		recommendations = append(recommendations, "偏差超过50%，建议进行深入分析")
	}

	recommendations = append(recommendations, anomaly.Recommendation)

	return recommendations
}

// 私有方法

func (se *SemanticEngine) getMetricDescription(metricName string) string {
	if desc, ok := se.metricDescriptions[metricName]; ok {
		return desc
	}
	return "未知指标"
}

func (se *SemanticEngine) calculateSeverity(metricName string, value float64) AlertLevel {
	switch {
	case value > 90:
		return AlertLevelCritical
	case value > 75:
		return AlertLevelHigh
	case value > 50:
		return AlertLevelMedium
	default:
		return AlertLevelLow
	}
}

func (se *SemanticEngine) generateTitle(metricName string, value float64) string {
	switch {
	case strings.Contains(metricName, "cpu"):
		return fmt.Sprintf("CPU 使用率异常: %.1f%%", value)
	case strings.Contains(metricName, "memory"):
		return fmt.Sprintf("内存使用率异常: %.1f%%", value)
	case strings.Contains(metricName, "disk"):
		return fmt.Sprintf("磁盘使用率异常: %.1f%%", value)
	case strings.Contains(metricName, "network"):
		return fmt.Sprintf("网络流量异常: %.2f", value)
	default:
		return fmt.Sprintf("%s 异常: %.2f", metricName, value)
	}
}

func (se *SemanticEngine) generateDescription(metricName string, value float64, desc string) string {
	return fmt.Sprintf("%s 当前值为 %.2f，%s", desc, value, se.getStatusDescription(metricName, value))
}

func (se *SemanticEngine) getStatusDescription(metricName string, value float64) string {
	switch {
	case strings.Contains(metricName, "cpu"):
		if value > 80 {
			return "CPU 占用率过高，系统性能可能受到影响"
		}
		return "CPU 使用率正常"
	case strings.Contains(metricName, "memory"):
		if value > 85 {
			return "内存占用率过高，可能导致系统缓慢"
		}
		return "内存使用率正常"
	case strings.Contains(metricName, "disk"):
		if value > 90 {
			return "磁盘空间即将耗尽，需要立即清理"
		}
		return "磁盘空间充足"
	default:
		return "指标异常"
	}
}

func (se *SemanticEngine) analyzeRootCause(metricName string, value float64) string {
	switch {
	case strings.Contains(metricName, "cpu"):
		return "可能原因: 应用程序计算密集、死循环、或系统进程异常"
	case strings.Contains(metricName, "memory"):
		return "可能原因: 内存泄漏、缓存过大、或应用程序内存占用过多"
	case strings.Contains(metricName, "disk"):
		return "可能原因: 日志文件过大、临时文件堆积、或数据库文件增长"
	case strings.Contains(metricName, "network"):
		return "可能原因: 网络流量突增、DDoS 攻击、或应用程序网络请求过多"
	default:
		return "需要进一步分析"
	}
}

func (se *SemanticEngine) analyzeImpact(metricName string, value float64) string {
	switch {
	case strings.Contains(metricName, "cpu"):
		if value > 90 {
			return "影响: 系统响应缓慢，用户体验下降，可能导致服务超时"
		}
		return "影响: 轻微，系统仍可正常运行"
	case strings.Contains(metricName, "memory"):
		if value > 90 {
			return "影响: 系统可能出现 OOM，导致进程被杀死"
		}
		return "影响: 轻微，系统仍可正常运行"
	case strings.Contains(metricName, "disk"):
		if value > 95 {
			return "影响: 磁盘满，无法写入数据，服务可能中断"
		}
		return "影响: 轻微，仍有可用空间"
	default:
		return "影响: 需要进一步评估"
	}
}

func (se *SemanticEngine) generateSuggestions(metricName string, value float64) []string {
	suggestions := []string{}

	switch {
	case strings.Contains(metricName, "cpu"):
		suggestions = append(suggestions, "检查 CPU 占用最高的进程")
		suggestions = append(suggestions, "分析应用程序是否存在性能问题")
		suggestions = append(suggestions, "考虑增加 CPU 资源或优化代码")
	case strings.Contains(metricName, "memory"):
		suggestions = append(suggestions, "检查内存占用最高的进程")
		suggestions = append(suggestions, "排查是否存在内存泄漏")
		suggestions = append(suggestions, "考虑增加内存或优化内存使用")
	case strings.Contains(metricName, "disk"):
		suggestions = append(suggestions, "清理不必要的日志文件")
		suggestions = append(suggestions, "检查大文件占用情况")
		suggestions = append(suggestions, "考虑扩容磁盘或归档旧数据")
	}

	return suggestions
}

func initMetricDescriptions() map[string]string {
	return map[string]string{
		"node_cpu_seconds_total":           "CPU 使用时间总计",
		"node_memory_MemAvailable_bytes":   "可用内存",
		"node_memory_MemTotal_bytes":       "总内存",
		"node_filesystem_avail_bytes":      "文件系统可用空间",
		"node_filesystem_size_bytes":       "文件系统总大小",
		"node_network_receive_bytes_total": "网络接收字节总计",
		"node_network_transmit_bytes_total": "网络发送字节总计",
		"node_load1":                       "1分钟负载",
		"node_load5":                       "5分钟负载",
		"node_load15":                      "15分钟负载",
	}
}

func initAlertPatterns() map[string]string {
	return map[string]string{
		"HighCPU":    "CPU 使用率超过 80%",
		"HighMemory": "内存使用率超过 85%",
		"HighDisk":   "磁盘使用率超过 90%",
		"HighLoad":   "系统负载过高",
	}
}
