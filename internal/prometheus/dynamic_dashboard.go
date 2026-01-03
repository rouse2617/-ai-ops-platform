package prometheus

import (
	"context"
	"regexp"
	"strings"
)

// DynamicDashboard 动态监控看板
type DynamicDashboard struct {
	client *Client
}

// DashboardPanel 看板面板
type DashboardPanel struct {
	ID        string      `json:"id"`
	Title     string      `json:"title"`
	Type      string      `json:"type"` // line, gauge, stat, table, bar
	Query     string      `json:"query"`
	Unit      string      `json:"unit,omitempty"`
	Threshold []Threshold `json:"thresholds,omitempty"`
	TimeRange string      `json:"time_range"`
}

// Threshold 阈值配置
type Threshold struct {
	Value float64 `json:"value"`
	Color string  `json:"color"`
}

// DashboardConfig 看板配置
type DashboardConfig struct {
	Topic       string           `json:"topic"`
	Title       string           `json:"title"`
	Description string           `json:"description"`
	Panels      []DashboardPanel `json:"panels"`
	RefreshRate int              `json:"refresh_rate"` // 秒
}

// TopicKeywords 话题关键词映射
var TopicKeywords = map[string][]string{
	"database":    {"数据库", "mysql", "postgresql", "postgres", "mongodb", "redis", "sql", "查询", "慢查询", "连接池"},
	"network":     {"网络", "丢包", "延迟", "tcp", "连接", "带宽", "流量", "重传"},
	"cpu":         {"cpu", "处理器", "负载", "load", "计算"},
	"memory":      {"内存", "memory", "oom", "swap", "缓存"},
	"disk":        {"磁盘", "disk", "存储", "io", "iops", "空间"},
	"container":   {"容器", "docker", "kubernetes", "k8s", "pod", "deployment"},
	"http":        {"http", "api", "请求", "响应", "延迟", "qps", "错误率"},
	"application": {"应用", "服务", "进程", "jvm", "gc", "线程"},
}

// NewDynamicDashboard 创建动态看板
func NewDynamicDashboard(client *Client) *DynamicDashboard {
	return &DynamicDashboard{client: client}
}

// GenerateDashboard 根据话题生成看板
func (d *DynamicDashboard) GenerateDashboard(ctx context.Context, topic string) *DashboardConfig {
	// 检测话题类型
	detectedTopic := d.detectTopic(topic)

	// 根据话题生成看板配置
	switch detectedTopic {
	case "database":
		return d.generateDatabaseDashboard()
	case "network":
		return d.generateNetworkDashboard()
	case "cpu":
		return d.generateCPUDashboard()
	case "memory":
		return d.generateMemoryDashboard()
	case "disk":
		return d.generateDiskDashboard()
	case "container":
		return d.generateContainerDashboard()
	case "http":
		return d.generateHTTPDashboard()
	case "application":
		return d.generateApplicationDashboard()
	default:
		return d.generateOverviewDashboard()
	}
}

// detectTopic 检测话题类型
func (d *DynamicDashboard) detectTopic(text string) string {
	text = strings.ToLower(text)

	maxScore := 0
	detectedTopic := "overview"

	for topic, keywords := range TopicKeywords {
		score := 0
		for _, kw := range keywords {
			if strings.Contains(text, strings.ToLower(kw)) {
				score++
			}
		}
		if score > maxScore {
			maxScore = score
			detectedTopic = topic
		}
	}

	return detectedTopic
}

// generateDatabaseDashboard 生成数据库看板
func (d *DynamicDashboard) generateDatabaseDashboard() *DashboardConfig {
	return &DashboardConfig{
		Topic:       "database",
		Title:       "数据库监控",
		Description: "MySQL/PostgreSQL 性能指标",
		RefreshRate: 10,
		Panels: []DashboardPanel{
			{ID: "db-connections", Title: "数据库连接数", Type: "line", Query: `mysql_global_status_threads_connected`, Unit: "connections", TimeRange: "1h"},
			{ID: "db-qps", Title: "查询 QPS", Type: "line", Query: `rate(mysql_global_status_queries[5m])`, Unit: "qps", TimeRange: "1h"},
			{ID: "db-slow-queries", Title: "慢查询数", Type: "stat", Query: `increase(mysql_global_status_slow_queries[1h])`, Unit: "queries", TimeRange: "1h"},
			{ID: "db-buffer-usage", Title: "Buffer Pool 使用率", Type: "gauge", Query: `mysql_global_status_innodb_buffer_pool_pages_data / mysql_global_status_innodb_buffer_pool_pages_total * 100`, Unit: "%", TimeRange: "1h", Threshold: []Threshold{{Value: 80, Color: "yellow"}, {Value: 95, Color: "red"}}},
			{ID: "db-replication-lag", Title: "复制延迟", Type: "stat", Query: `mysql_slave_status_seconds_behind_master`, Unit: "s", TimeRange: "1h"},
			{ID: "db-deadlocks", Title: "死锁数", Type: "stat", Query: `increase(mysql_global_status_innodb_deadlocks[1h])`, Unit: "deadlocks", TimeRange: "1h"},
		},
	}
}

// generateNetworkDashboard 生成网络看板
func (d *DynamicDashboard) generateNetworkDashboard() *DashboardConfig {
	return &DashboardConfig{
		Topic:       "network",
		Title:       "网络监控",
		Description: "网络流量和连接状态",
		RefreshRate: 5,
		Panels: []DashboardPanel{
			{ID: "net-traffic", Title: "网络流量", Type: "line", Query: `rate(node_network_receive_bytes_total[5m]) + rate(node_network_transmit_bytes_total[5m])`, Unit: "bytes/s", TimeRange: "1h"},
			{ID: "net-tcp-conn", Title: "TCP 连接数", Type: "line", Query: `node_netstat_Tcp_CurrEstab`, Unit: "connections", TimeRange: "1h"},
			{ID: "net-retrans", Title: "TCP 重传率", Type: "line", Query: `rate(node_netstat_Tcp_RetransSegs[5m]) / rate(node_netstat_Tcp_OutSegs[5m]) * 100`, Unit: "%", TimeRange: "1h", Threshold: []Threshold{{Value: 1, Color: "yellow"}, {Value: 5, Color: "red"}}},
			{ID: "net-errors", Title: "网络错误", Type: "line", Query: `rate(node_network_receive_errs_total[5m]) + rate(node_network_transmit_errs_total[5m])`, Unit: "errors/s", TimeRange: "1h"},
			{ID: "net-dropped", Title: "丢包数", Type: "stat", Query: `increase(node_network_receive_drop_total[1h]) + increase(node_network_transmit_drop_total[1h])`, Unit: "packets", TimeRange: "1h"},
			{ID: "net-bandwidth", Title: "带宽使用", Type: "gauge", Query: `(rate(node_network_receive_bytes_total[5m]) + rate(node_network_transmit_bytes_total[5m])) / 125000000 * 100`, Unit: "%", TimeRange: "1h"},
		},
	}
}

// generateCPUDashboard 生成 CPU 看板
func (d *DynamicDashboard) generateCPUDashboard() *DashboardConfig {
	return &DashboardConfig{
		Topic:       "cpu",
		Title:       "CPU 监控",
		Description: "CPU 使用率和负载",
		RefreshRate: 5,
		Panels: []DashboardPanel{
			{ID: "cpu-usage", Title: "CPU 使用率", Type: "line", Query: `100 - (avg by(instance) (rate(node_cpu_seconds_total{mode="idle"}[5m])) * 100)`, Unit: "%", TimeRange: "1h", Threshold: []Threshold{{Value: 70, Color: "yellow"}, {Value: 90, Color: "red"}}},
			{ID: "cpu-load1", Title: "1分钟负载", Type: "line", Query: `node_load1`, Unit: "load", TimeRange: "1h"},
			{ID: "cpu-load5", Title: "5分钟负载", Type: "line", Query: `node_load5`, Unit: "load", TimeRange: "1h"},
			{ID: "cpu-iowait", Title: "IO 等待", Type: "line", Query: `avg by(instance) (rate(node_cpu_seconds_total{mode="iowait"}[5m])) * 100`, Unit: "%", TimeRange: "1h"},
			{ID: "cpu-system", Title: "系统态 CPU", Type: "line", Query: `avg by(instance) (rate(node_cpu_seconds_total{mode="system"}[5m])) * 100`, Unit: "%", TimeRange: "1h"},
			{ID: "cpu-user", Title: "用户态 CPU", Type: "line", Query: `avg by(instance) (rate(node_cpu_seconds_total{mode="user"}[5m])) * 100`, Unit: "%", TimeRange: "1h"},
		},
	}
}

// generateMemoryDashboard 生成内存看板
func (d *DynamicDashboard) generateMemoryDashboard() *DashboardConfig {
	return &DashboardConfig{
		Topic:       "memory",
		Title:       "内存监控",
		Description: "内存使用和 Swap 状态",
		RefreshRate: 10,
		Panels: []DashboardPanel{
			{ID: "mem-usage", Title: "内存使用率", Type: "gauge", Query: `100 - (node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes * 100)`, Unit: "%", TimeRange: "1h", Threshold: []Threshold{{Value: 70, Color: "yellow"}, {Value: 90, Color: "red"}}},
			{ID: "mem-used", Title: "已用内存", Type: "line", Query: `node_memory_MemTotal_bytes - node_memory_MemAvailable_bytes`, Unit: "bytes", TimeRange: "1h"},
			{ID: "mem-cached", Title: "缓存内存", Type: "line", Query: `node_memory_Cached_bytes`, Unit: "bytes", TimeRange: "1h"},
			{ID: "mem-buffers", Title: "Buffer 内存", Type: "line", Query: `node_memory_Buffers_bytes`, Unit: "bytes", TimeRange: "1h"},
			{ID: "mem-swap", Title: "Swap 使用率", Type: "gauge", Query: `(node_memory_SwapTotal_bytes - node_memory_SwapFree_bytes) / node_memory_SwapTotal_bytes * 100`, Unit: "%", TimeRange: "1h"},
			{ID: "mem-oom", Title: "OOM 事件", Type: "stat", Query: `increase(node_vmstat_oom_kill[24h])`, Unit: "events", TimeRange: "24h"},
		},
	}
}

// generateDiskDashboard 生成磁盘看板
func (d *DynamicDashboard) generateDiskDashboard() *DashboardConfig {
	return &DashboardConfig{
		Topic:       "disk",
		Title:       "磁盘监控",
		Description: "磁盘空间和 IO 性能",
		RefreshRate: 30,
		Panels: []DashboardPanel{
			{ID: "disk-usage", Title: "磁盘使用率", Type: "gauge", Query: `100 - (node_filesystem_avail_bytes{mountpoint="/"} / node_filesystem_size_bytes{mountpoint="/"} * 100)`, Unit: "%", TimeRange: "1h", Threshold: []Threshold{{Value: 70, Color: "yellow"}, {Value: 90, Color: "red"}}},
			{ID: "disk-iops", Title: "磁盘 IOPS", Type: "line", Query: `rate(node_disk_reads_completed_total[5m]) + rate(node_disk_writes_completed_total[5m])`, Unit: "iops", TimeRange: "1h"},
			{ID: "disk-throughput", Title: "磁盘吞吐", Type: "line", Query: `rate(node_disk_read_bytes_total[5m]) + rate(node_disk_written_bytes_total[5m])`, Unit: "bytes/s", TimeRange: "1h"},
			{ID: "disk-util", Title: "磁盘利用率", Type: "line", Query: `rate(node_disk_io_time_seconds_total[5m]) * 100`, Unit: "%", TimeRange: "1h"},
			{ID: "disk-latency", Title: "IO 延迟", Type: "line", Query: `rate(node_disk_read_time_seconds_total[5m]) / rate(node_disk_reads_completed_total[5m])`, Unit: "s", TimeRange: "1h"},
			{ID: "disk-inode", Title: "Inode 使用率", Type: "gauge", Query: `100 - (node_filesystem_files_free{mountpoint="/"} / node_filesystem_files{mountpoint="/"} * 100)`, Unit: "%", TimeRange: "1h"},
		},
	}
}

// generateContainerDashboard 生成容器看板
func (d *DynamicDashboard) generateContainerDashboard() *DashboardConfig {
	return &DashboardConfig{
		Topic:       "container",
		Title:       "容器监控",
		Description: "Docker/Kubernetes 容器状态",
		RefreshRate: 10,
		Panels: []DashboardPanel{
			{ID: "container-cpu", Title: "容器 CPU 使用", Type: "line", Query: `sum(rate(container_cpu_usage_seconds_total[5m])) by (container_name) * 100`, Unit: "%", TimeRange: "1h"},
			{ID: "container-mem", Title: "容器内存使用", Type: "line", Query: `sum(container_memory_usage_bytes) by (container_name)`, Unit: "bytes", TimeRange: "1h"},
			{ID: "container-restart", Title: "容器重启次数", Type: "stat", Query: `sum(increase(kube_pod_container_status_restarts_total[24h])) by (container)`, Unit: "restarts", TimeRange: "24h"},
			{ID: "container-running", Title: "运行中容器", Type: "stat", Query: `count(container_last_seen) - 1`, Unit: "containers", TimeRange: "1h"},
			{ID: "pod-pending", Title: "Pending Pods", Type: "stat", Query: `sum(kube_pod_status_phase{phase="Pending"})`, Unit: "pods", TimeRange: "1h"},
			{ID: "container-network", Title: "容器网络流量", Type: "line", Query: `sum(rate(container_network_receive_bytes_total[5m])) by (container_name)`, Unit: "bytes/s", TimeRange: "1h"},
		},
	}
}

// generateHTTPDashboard 生成 HTTP 看板
func (d *DynamicDashboard) generateHTTPDashboard() *DashboardConfig {
	return &DashboardConfig{
		Topic:       "http",
		Title:       "HTTP 监控",
		Description: "API 请求和响应指标",
		RefreshRate: 5,
		Panels: []DashboardPanel{
			{ID: "http-qps", Title: "请求 QPS", Type: "line", Query: `sum(rate(http_requests_total[5m]))`, Unit: "req/s", TimeRange: "1h"},
			{ID: "http-latency", Title: "响应延迟 P99", Type: "line", Query: `histogram_quantile(0.99, sum(rate(http_request_duration_seconds_bucket[5m])) by (le))`, Unit: "s", TimeRange: "1h"},
			{ID: "http-errors", Title: "错误率", Type: "line", Query: `sum(rate(http_requests_total{status=~"5.."}[5m])) / sum(rate(http_requests_total[5m])) * 100`, Unit: "%", TimeRange: "1h", Threshold: []Threshold{{Value: 1, Color: "yellow"}, {Value: 5, Color: "red"}}},
			{ID: "http-status", Title: "状态码分布", Type: "bar", Query: `sum(increase(http_requests_total[1h])) by (status)`, Unit: "requests", TimeRange: "1h"},
			{ID: "http-latency-p50", Title: "响应延迟 P50", Type: "stat", Query: `histogram_quantile(0.50, sum(rate(http_request_duration_seconds_bucket[5m])) by (le))`, Unit: "s", TimeRange: "1h"},
			{ID: "http-active", Title: "活跃连接", Type: "stat", Query: `sum(http_connections_active)`, Unit: "connections", TimeRange: "1h"},
		},
	}
}

// generateApplicationDashboard 生成应用看板
func (d *DynamicDashboard) generateApplicationDashboard() *DashboardConfig {
	return &DashboardConfig{
		Topic:       "application",
		Title:       "应用监控",
		Description: "应用进程和 JVM 指标",
		RefreshRate: 10,
		Panels: []DashboardPanel{
			{ID: "app-process", Title: "进程数", Type: "stat", Query: `node_procs_running`, Unit: "processes", TimeRange: "1h"},
			{ID: "app-threads", Title: "线程数", Type: "line", Query: `jvm_threads_current`, Unit: "threads", TimeRange: "1h"},
			{ID: "app-heap", Title: "JVM 堆内存", Type: "line", Query: `jvm_memory_bytes_used{area="heap"}`, Unit: "bytes", TimeRange: "1h"},
			{ID: "app-gc", Title: "GC 次数", Type: "line", Query: `rate(jvm_gc_collection_seconds_count[5m])`, Unit: "gc/s", TimeRange: "1h"},
			{ID: "app-gc-time", Title: "GC 耗时", Type: "line", Query: `rate(jvm_gc_collection_seconds_sum[5m])`, Unit: "s", TimeRange: "1h"},
			{ID: "app-uptime", Title: "运行时间", Type: "stat", Query: `process_uptime_seconds`, Unit: "s", TimeRange: "1h"},
		},
	}
}

// generateOverviewDashboard 生成概览看板
func (d *DynamicDashboard) generateOverviewDashboard() *DashboardConfig {
	return &DashboardConfig{
		Topic:       "overview",
		Title:       "系统概览",
		Description: "核心系统指标",
		RefreshRate: 10,
		Panels: []DashboardPanel{
			{ID: "overview-cpu", Title: "CPU 使用率", Type: "gauge", Query: `100 - (avg(rate(node_cpu_seconds_total{mode="idle"}[5m])) * 100)`, Unit: "%", TimeRange: "1h", Threshold: []Threshold{{Value: 70, Color: "yellow"}, {Value: 90, Color: "red"}}},
			{ID: "overview-mem", Title: "内存使用率", Type: "gauge", Query: `100 - (avg(node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes) * 100)`, Unit: "%", TimeRange: "1h", Threshold: []Threshold{{Value: 70, Color: "yellow"}, {Value: 90, Color: "red"}}},
			{ID: "overview-disk", Title: "磁盘使用率", Type: "gauge", Query: `100 - (avg(node_filesystem_avail_bytes{mountpoint="/"} / node_filesystem_size_bytes{mountpoint="/"}) * 100)`, Unit: "%", TimeRange: "1h", Threshold: []Threshold{{Value: 70, Color: "yellow"}, {Value: 90, Color: "red"}}},
			{ID: "overview-load", Title: "系统负载", Type: "line", Query: `node_load1`, Unit: "load", TimeRange: "1h"},
			{ID: "overview-network", Title: "网络流量", Type: "line", Query: `sum(rate(node_network_receive_bytes_total[5m])) + sum(rate(node_network_transmit_bytes_total[5m]))`, Unit: "bytes/s", TimeRange: "1h"},
			{ID: "overview-uptime", Title: "运行时间", Type: "stat", Query: `node_time_seconds - node_boot_time_seconds`, Unit: "s", TimeRange: "1h"},
		},
	}
}

// ExtractTopicFromMessage 从消息中提取话题
func (d *DynamicDashboard) ExtractTopicFromMessage(message string) string {
	return d.detectTopic(message)
}

// GetAvailableTopics 获取可用话题列表
func (d *DynamicDashboard) GetAvailableTopics() []string {
	topics := make([]string, 0, len(TopicKeywords)+1)
	for topic := range TopicKeywords {
		topics = append(topics, topic)
	}
	topics = append(topics, "overview")
	return topics
}

// MatchKeywordsInText 在文本中匹配关键词
func (d *DynamicDashboard) MatchKeywordsInText(text string) map[string][]string {
	text = strings.ToLower(text)
	matches := make(map[string][]string)

	for topic, keywords := range TopicKeywords {
		for _, kw := range keywords {
			pattern := regexp.MustCompile(`(?i)\b` + regexp.QuoteMeta(kw) + `\b`)
			if pattern.MatchString(text) {
				matches[topic] = append(matches[topic], kw)
			}
		}
	}

	return matches
}
