package agent

import (
	"context"
)

// ExpertType 专家类型
type ExpertType string

const (
	ExpertTroubleshoot ExpertType = "troubleshoot" // 问题定位
	ExpertMonitor      ExpertType = "monitor"      // 监控查询
	ExpertDatabase     ExpertType = "database"     // 数据库
	ExpertSecurity     ExpertType = "security"     // 安全审计
	ExpertStorage      ExpertType = "storage"      // 存储（已有）
	ExpertGeneral      ExpertType = "general"      // 通用
)

// ExpertAgent 专家 Agent 接口
type ExpertAgent interface {
	// Type 返回专家类型
	Type() ExpertType

	// Description 返回专家描述（用于 LLM 路由决策）
	Description() string

	// GetTools 返回该专家使用的工具列表
	GetTools() []string

	// Execute 执行专家分析（独立上下文）
	Execute(ctx context.Context, req ChatRequest) (*ChatResponse, error)
}

// ExpertInfo 专家信息（用于路由决策）
type ExpertInfo struct {
	Type        ExpertType `json:"type"`
	Description string     `json:"description"`
	Tools       []string   `json:"tools"`
}

// GetAllExpertInfo 获取所有专家信息
func GetAllExpertInfo() []ExpertInfo {
	return []ExpertInfo{
		{
			Type:        ExpertTroubleshoot,
			Description: "问题定位专家：擅长日志分析、错误追踪、根因定位、故障排查。当用户询问'为什么'、'报错'、'失败'、'挂了'等问题时使用。",
			Tools:       []string{"query_log", "check_process", "root_cause_diagnosis", "anomaly_detection"},
		},
		{
			Type:        ExpertMonitor,
			Description: "监控查询专家：擅长 CPU/内存/磁盘监控、指标查询、趋势分析、告警管理。当用户询问系统指标、性能数据、告警信息时使用。",
			Tools:       []string{"check_cpu", "check_memory", "check_disk", "trend_analysis", "alert_manager"},
		},
		{
			Type:        ExpertDatabase,
			Description: "数据库专家：擅长 MySQL/Redis/PostgreSQL 诊断、连接池分析、慢查询优化、数据库备份。当用户询问数据库相关问题时使用。",
			Tools:       []string{"check_mysql_status", "check_redis_status", "backup_database"},
		},
		{
			Type:        ExpertSecurity,
			Description: "安全审计专家：擅长入侵检测、异常登录分析、安全日志审计、漏洞扫描。当用户询问安全相关问题时使用。",
			Tools:       []string{"intrusion_detection", "query_log"},
		},
		{
			Type:        ExpertStorage,
			Description: "存储专家：擅长磁盘空间分析、文件系统检查、RAID 状态、存储性能。当用户询问存储、磁盘、空间相关问题时使用。",
			Tools:       []string{"check_disk", "check_inode", "storage_check", "raid_check"},
		},
		{
			Type:        ExpertGeneral,
			Description: "通用运维专家：处理一般性运维操作，如服务管理、命令执行、配置查看等。当其他专家都不适合时使用。",
			Tools:       []string{"run_command", "manage_service", "get_config", "list_hosts"},
		},
	}
}
