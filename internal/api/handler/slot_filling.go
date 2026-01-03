package handler

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// SlotFillingHandler 参数补全处理器
type SlotFillingHandler struct{}

// NewSlotFillingHandler 创建参数补全处理器
func NewSlotFillingHandler() *SlotFillingHandler {
	return &SlotFillingHandler{}
}

// SlotSuggestion 参数建议
type SlotSuggestion struct {
	Value       string `json:"value"`
	Label       string `json:"label"`
	Description string `json:"description,omitempty"`
	Icon        string `json:"icon,omitempty"`
	IconColor   string `json:"iconColor,omitempty"`
	Tag         string `json:"tag,omitempty"`
	TagType     string `json:"tagType,omitempty"`
}

// SlotFillingRequest 参数补全请求
type SlotFillingRequest struct {
	Message string   `json:"message" binding:"required"`
	Hosts   []string `json:"hosts"`
}

// SlotFillingResponse 参数补全响应
type SlotFillingResponse struct {
	NeedsSlotFilling bool             `json:"needsSlotFilling"`
	SlotType         string           `json:"slotType"`
	Suggestions      []SlotSuggestion `json:"suggestions"`
	OriginalMessage  string           `json:"originalMessage"`
}

// AnalyzeMessage 分析消息是否需要参数补全
// POST /api/slot-filling/analyze
func (h *SlotFillingHandler) AnalyzeMessage(c *gin.Context) {
	var req SlotFillingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ParamError(c, "参数错误: "+err.Error())
		return
	}

	message := strings.TrimSpace(strings.ToLower(req.Message))

	// 检测服务相关命令
	if h.isServiceCommand(message) {
		Success(c, SlotFillingResponse{
			NeedsSlotFilling: true,
			SlotType:         "service",
			Suggestions:      h.getServiceSuggestions(),
			OriginalMessage:  req.Message,
		})
		return
	}

	// 检测文件相关命令
	if h.isFileCommand(message) {
		Success(c, SlotFillingResponse{
			NeedsSlotFilling: true,
			SlotType:         "file",
			Suggestions:      h.getFileSuggestions(),
			OriginalMessage:  req.Message,
		})
		return
	}

	// 检测进程相关命令
	if h.isProcessCommand(message) {
		Success(c, SlotFillingResponse{
			NeedsSlotFilling: true,
			SlotType:         "process",
			Suggestions:      h.getProcessSuggestions(),
			OriginalMessage:  req.Message,
		})
		return
	}

	// 检测端口相关命令
	if h.isPortCommand(message) {
		Success(c, SlotFillingResponse{
			NeedsSlotFilling: true,
			SlotType:         "port",
			Suggestions:      h.getPortSuggestions(),
			OriginalMessage:  req.Message,
		})
		return
	}

	// 不需要参数补全
	Success(c, SlotFillingResponse{
		NeedsSlotFilling: false,
		OriginalMessage:  req.Message,
	})
}

// isServiceCommand 检测是否为服务命令
func (h *SlotFillingHandler) isServiceCommand(message string) bool {
	keywords := []string{"重启服务", "启动服务", "停止服务", "查看服务", "服务状态", "restart service", "start service", "stop service"}
	for _, kw := range keywords {
		if strings.Contains(message, kw) && !h.hasServiceName(message) {
			return true
		}
	}
	return false
}

// isFileCommand 检测是否为文件命令
func (h *SlotFillingHandler) isFileCommand(message string) bool {
	keywords := []string{"查看文件", "编辑文件", "删除文件", "查看日志", "tail", "cat", "vim"}
	for _, kw := range keywords {
		if strings.Contains(message, kw) && !h.hasFilePath(message) {
			return true
		}
	}
	return false
}

// isProcessCommand 检测是否为进程命令
func (h *SlotFillingHandler) isProcessCommand(message string) bool {
	keywords := []string{"杀死进程", "kill", "停止进程", "查看进程"}
	for _, kw := range keywords {
		if strings.Contains(message, kw) && !h.hasProcessName(message) {
			return true
		}
	}
	return false
}

// isPortCommand 检测是否为端口命令
func (h *SlotFillingHandler) isPortCommand(message string) bool {
	keywords := []string{"查看端口", "检查端口", "端口占用", "netstat", "lsof"}
	for _, kw := range keywords {
		if strings.Contains(message, kw) && !h.hasPortNumber(message) {
			return true
		}
	}
	return false
}

// hasServiceName 检测是否包含服务名
func (h *SlotFillingHandler) hasServiceName(message string) bool {
	services := []string{"nginx", "apache", "mysql", "redis", "docker", "ssh", "postgresql", "mongodb"}
	for _, svc := range services {
		if strings.Contains(message, svc) {
			return true
		}
	}
	return false
}

// hasFilePath 检测是否包含文件路径
func (h *SlotFillingHandler) hasFilePath(message string) bool {
	return strings.Contains(message, "/") || strings.Contains(message, ".log") || strings.Contains(message, ".conf")
}

// hasProcessName 检测是否包含进程名
func (h *SlotFillingHandler) hasProcessName(message string) bool {
	processes := []string{"java", "python", "node", "nginx", "apache"}
	for _, proc := range processes {
		if strings.Contains(message, proc) {
			return true
		}
	}
	return false
}

// hasPortNumber 检测是否包含端口号
func (h *SlotFillingHandler) hasPortNumber(message string) bool {
	// 简单检测是否包含数字
	for _, char := range message {
		if char >= '0' && char <= '9' {
			return true
		}
	}
	return false
}

// getServiceSuggestions 获取服务建议
func (h *SlotFillingHandler) getServiceSuggestions() []SlotSuggestion {
	return []SlotSuggestion{
		{Value: "nginx", Label: "Nginx", Description: "Web 服务器", Icon: "Service", IconColor: "#67C23A", Tag: "常用", TagType: "success"},
		{Value: "apache2", Label: "Apache", Description: "HTTP 服务器", Icon: "Service", IconColor: "#E6A23C"},
		{Value: "mysql", Label: "MySQL", Description: "数据库服务", Icon: "Service", IconColor: "#409EFF", Tag: "常用", TagType: "success"},
		{Value: "redis", Label: "Redis", Description: "缓存服务", Icon: "Service", IconColor: "#F56C6C"},
		{Value: "docker", Label: "Docker", Description: "容器服务", Icon: "Service", IconColor: "#409EFF"},
		{Value: "ssh", Label: "SSH", Description: "远程连接服务", Icon: "Service", IconColor: "#909399"},
		{Value: "postgresql", Label: "PostgreSQL", Description: "数据库服务", Icon: "Service", IconColor: "#409EFF"},
		{Value: "mongodb", Label: "MongoDB", Description: "文档数据库", Icon: "Service", IconColor: "#67C23A"},
	}
}

// getFileSuggestions 获取文件建议
func (h *SlotFillingHandler) getFileSuggestions() []SlotSuggestion {
	return []SlotSuggestion{
		{Value: "/var/log/syslog", Label: "系统日志", Description: "/var/log/syslog", Icon: "Files", IconColor: "#409EFF", Tag: "常用", TagType: "success"},
		{Value: "/var/log/nginx/access.log", Label: "Nginx 访问日志", Description: "/var/log/nginx/access.log", Icon: "Files", IconColor: "#67C23A"},
		{Value: "/var/log/nginx/error.log", Label: "Nginx 错误日志", Description: "/var/log/nginx/error.log", Icon: "Files", IconColor: "#F56C6C"},
		{Value: "/var/log/mysql/error.log", Label: "MySQL 错误日志", Description: "/var/log/mysql/error.log", Icon: "Files", IconColor: "#E6A23C"},
		{Value: "/etc/nginx/nginx.conf", Label: "Nginx 配置", Description: "/etc/nginx/nginx.conf", Icon: "Files", IconColor: "#909399"},
		{Value: "/etc/mysql/my.cnf", Label: "MySQL 配置", Description: "/etc/mysql/my.cnf", Icon: "Files", IconColor: "#909399"},
	}
}

// getProcessSuggestions 获取进程建议
func (h *SlotFillingHandler) getProcessSuggestions() []SlotSuggestion {
	return []SlotSuggestion{
		{Value: "nginx", Label: "Nginx", Description: "Web 服务器进程", Icon: "Operation", IconColor: "#67C23A"},
		{Value: "java", Label: "Java", Description: "Java 应用进程", Icon: "Operation", IconColor: "#E6A23C"},
		{Value: "python", Label: "Python", Description: "Python 应用进程", Icon: "Operation", IconColor: "#409EFF"},
		{Value: "node", Label: "Node.js", Description: "Node.js 应用进程", Icon: "Operation", IconColor: "#67C23A"},
		{Value: "mysqld", Label: "MySQL", Description: "MySQL 数据库进程", Icon: "Operation", IconColor: "#409EFF"},
	}
}

// getPortSuggestions 获取端口建议
func (h *SlotFillingHandler) getPortSuggestions() []SlotSuggestion {
	return []SlotSuggestion{
		{Value: "80", Label: "80", Description: "HTTP 端口", Icon: "Setting", IconColor: "#409EFF", Tag: "常用", TagType: "success"},
		{Value: "443", Label: "443", Description: "HTTPS 端口", Icon: "Setting", IconColor: "#67C23A", Tag: "常用", TagType: "success"},
		{Value: "3306", Label: "3306", Description: "MySQL 端口", Icon: "Setting", IconColor: "#409EFF"},
		{Value: "6379", Label: "6379", Description: "Redis 端口", Icon: "Setting", IconColor: "#F56C6C"},
		{Value: "8080", Label: "8080", Description: "应用端口", Icon: "Setting", IconColor: "#E6A23C"},
		{Value: "22", Label: "22", Description: "SSH 端口", Icon: "Setting", IconColor: "#909399"},
	}
}
