package prometheus

import (
	"fmt"
	"sync"
	"time"
)

// DashboardGenerator 动态看板生成器
type DashboardGenerator struct {
	dashboards map[string]*DashboardConfig
	templates  map[string][]PanelConfig
	mu         sync.RWMutex
}

// NewDashboardGenerator 创建看板生成器
func NewDashboardGenerator() *DashboardGenerator {
	return &DashboardGenerator{
		dashboards: make(map[string]*DashboardConfig),
		templates:  initDashboardTemplates(),
	}
}

// GenerateDashboard 生成看板
func (dg *DashboardGenerator) GenerateDashboard(userID string, config DashboardConfig) (Dashboard, error) {
	dg.mu.Lock()
	defer dg.mu.Unlock()

	if config.ID == "" {
		config.ID = fmt.Sprintf("dash-%d", time.Now().UnixNano())
	}

	config.UserID = userID
	config.CreatedAt = time.Now()
	config.UpdatedAt = time.Now()

	dg.dashboards[config.ID] = &config

	return Dashboard{
		Config: config,
		Data:   make(map[string]interface{}),
	}, nil
}

// UpdateDashboard 更新看板
func (dg *DashboardGenerator) UpdateDashboard(dashboardID string, config DashboardConfig) error {
	dg.mu.Lock()
	defer dg.mu.Unlock()

	existing, ok := dg.dashboards[dashboardID]
	if !ok {
		return fmt.Errorf("dashboard not found: %s", dashboardID)
	}

	config.ID = dashboardID
	config.UserID = existing.UserID
	config.CreatedAt = existing.CreatedAt
	config.UpdatedAt = time.Now()

	dg.dashboards[dashboardID] = &config
	return nil
}

// GetDashboard 获取看板
func (dg *DashboardGenerator) GetDashboard(dashboardID string) (Dashboard, error) {
	dg.mu.RLock()
	defer dg.mu.RUnlock()

	config, ok := dg.dashboards[dashboardID]
	if !ok {
		return Dashboard{}, fmt.Errorf("dashboard not found: %s", dashboardID)
	}

	return Dashboard{
		Config: *config,
		Data:   make(map[string]interface{}),
	}, nil
}

// RecommendPanels 推荐面板
func (dg *DashboardGenerator) RecommendPanels(userID string) ([]PanelConfig, error) {
	dg.mu.RLock()
	defer dg.mu.RUnlock()

	// 根据用户历史推荐面板
	recommended := []PanelConfig{}

	// 添加系统概览面板
	recommended = append(recommended, PanelConfig{
		ID:    "panel-system-overview",
		Title: "系统概览",
		Type:  "stat",
		Query: "up",
		Options: map[string]interface{}{
			"thresholds": map[string]interface{}{
				"mode": "absolute",
				"steps": []map[string]interface{}{
					{"color": "green", "value": 1},
					{"color": "red", "value": 0},
				},
			},
		},
	})

	// 添加 CPU 使用率面板
	recommended = append(recommended, PanelConfig{
		ID:    "panel-cpu-usage",
		Title: "CPU 使用率",
		Type:  "graph",
		Query: "(1 - avg(rate(node_cpu_seconds_total{mode=\"idle\"}[5m])) by (instance)) * 100",
		Options: map[string]interface{}{
			"yaxes": []map[string]interface{}{
				{"format": "percent", "min": 0, "max": 100},
			},
		},
	})

	// 添加内存使用率面板
	recommended = append(recommended, PanelConfig{
		ID:    "panel-memory-usage",
		Title: "内存使用率",
		Type:  "graph",
		Query: "(1 - (node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes)) * 100",
		Options: map[string]interface{}{
			"yaxes": []map[string]interface{}{
				{"format": "percent", "min": 0, "max": 100},
			},
		},
	})

	// 添加磁盘使用率面板
	recommended = append(recommended, PanelConfig{
		ID:    "panel-disk-usage",
		Title: "磁盘使用率",
		Type:  "graph",
		Query: "(1 - (node_filesystem_avail_bytes / node_filesystem_size_bytes)) * 100",
		Options: map[string]interface{}{
			"yaxes": []map[string]interface{}{
				{"format": "percent", "min": 0, "max": 100},
			},
		},
	})

	return recommended, nil
}

// ListDashboards 列表看板
func (dg *DashboardGenerator) ListDashboards(userID string) ([]DashboardConfig, error) {
	dg.mu.RLock()
	defer dg.mu.RUnlock()

	var result []DashboardConfig

	for _, config := range dg.dashboards {
		if config.UserID == userID {
			result = append(result, *config)
		}
	}

	return result, nil
}

// DeleteDashboard 删除看板
func (dg *DashboardGenerator) DeleteDashboard(dashboardID string) error {
	dg.mu.Lock()
	defer dg.mu.Unlock()

	delete(dg.dashboards, dashboardID)
	return nil
}

// 私有方法

func initDashboardTemplates() map[string][]PanelConfig {
	return map[string][]PanelConfig{
		"system_overview": {
			{
				ID:    "panel-1",
				Title: "主机在线状态",
				Type:  "stat",
				Query: "up",
			},
			{
				ID:    "panel-2",
				Title: "CPU 使用率",
				Type:  "graph",
				Query: "(1 - avg(rate(node_cpu_seconds_total{mode=\"idle\"}[5m])) by (instance)) * 100",
			},
			{
				ID:    "panel-3",
				Title: "内存使用率",
				Type:  "graph",
				Query: "(1 - (node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes)) * 100",
			},
			{
				ID:    "panel-4",
				Title: "磁盘使用率",
				Type:  "graph",
				Query: "(1 - (node_filesystem_avail_bytes / node_filesystem_size_bytes)) * 100",
			},
		},
		"application_performance": {
			{
				ID:    "panel-1",
				Title: "请求速率",
				Type:  "graph",
				Query: "rate(http_requests_total[5m])",
			},
			{
				ID:    "panel-2",
				Title: "错误率",
				Type:  "graph",
				Query: "rate(http_requests_total{status=~\"5..\"}[5m])",
			},
			{
				ID:    "panel-3",
				Title: "响应时间",
				Type:  "graph",
				Query: "histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))",
			},
		},
		"business_metrics": {
			{
				ID:    "panel-1",
				Title: "交易量",
				Type:  "stat",
				Query: "increase(transactions_total[1h])",
			},
			{
				ID:    "panel-2",
				Title: "成功率",
				Type:  "stat",
				Query: "increase(transactions_success_total[1h]) / increase(transactions_total[1h])",
			},
			{
				ID:    "panel-3",
				Title: "平均交易金额",
				Type:  "stat",
				Query: "avg(transaction_amount)",
			},
		},
		"alerts": {
			{
				ID:    "panel-1",
				Title: "活跃告警",
				Type:  "table",
				Query: "ALERTS{alertstate=\"firing\"}",
			},
			{
				ID:    "panel-2",
				Title: "告警趋势",
				Type:  "graph",
				Query: "increase(alerts_total[1h])",
			},
		},
	}
}

// GetDashboardTemplate 获取看板模板
func (dg *DashboardGenerator) GetDashboardTemplate(templateName string) ([]PanelConfig, error) {
	dg.mu.RLock()
	defer dg.mu.RUnlock()

	panels, ok := dg.templates[templateName]
	if !ok {
		return nil, fmt.Errorf("template not found: %s", templateName)
	}

	return panels, nil
}

// ListTemplates 列表模板
func (dg *DashboardGenerator) ListTemplates() []string {
	dg.mu.RLock()
	defer dg.mu.RUnlock()

	templates := make([]string, 0, len(dg.templates))
	for name := range dg.templates {
		templates = append(templates, name)
	}

	return templates
}
