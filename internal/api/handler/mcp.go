package handler

import (
	"io"
	"net/http"
	"time"

	"ai-ops/internal/mcp"
	"ai-ops/internal/tool"

	"github.com/gin-gonic/gin"
)

// MCPHandler MCP 处理器
type MCPHandler struct {
	manager  *mcp.Manager
	server   *mcp.Server
	registry *tool.Registry
}

// NewMCPHandler 创建 MCP 处理器
func NewMCPHandler(manager *mcp.Manager, registry *tool.Registry) *MCPHandler {
	return &MCPHandler{
		manager:  manager,
		server:   mcp.NewServer(registry),
		registry: registry,
	}
}

// HandleJSONRPC 处理 MCP JSON-RPC 请求（服务端功能）
// POST /api/mcp/rpc
func (h *MCPHandler) HandleJSONRPC(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "读取请求失败"})
		return
	}

	resp, err := h.server.HandleJSONRPC(c.Request.Context(), body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 通知消息无响应
	if resp == nil {
		c.Status(http.StatusNoContent)
		return
	}

	c.Data(http.StatusOK, "application/json", resp)
}

// AddServerRequest 添加 MCP 服务器请求
type AddServerRequest struct {
	Name    string `json:"name" binding:"required"`
	URL     string `json:"url" binding:"required"`
	Timeout int    `json:"timeout"` // 秒
}

// AddServer 动态添加 MCP 服务器
// POST /api/mcp/servers
func (h *MCPHandler) AddServer(c *gin.Context) {
	var req AddServerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ParamError(c, "参数错误: "+err.Error())
		return
	}

	timeout := time.Duration(req.Timeout) * time.Second
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	cfg := mcp.Config{
		Name:    req.Name,
		URL:     req.URL,
		Timeout: timeout,
		Enabled: true,
	}

	if err := h.manager.RegisterClient(cfg); err != nil {
		InternalError(c, "注册 MCP 服务器失败: "+err.Error())
		return
	}

	Success(c, gin.H{
		"message": "MCP 服务器添加成功",
		"name":    req.Name,
	})
}

// RemoveServer 移除 MCP 服务器
// DELETE /api/mcp/servers/:name
func (h *MCPHandler) RemoveServer(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		ParamError(c, "服务器名称不能为空")
		return
	}

	h.manager.UnregisterClient(name)

	Success(c, gin.H{
		"message": "MCP 服务器已移除",
		"name":    name,
	})
}

// GetServerTools 获取指定服务器的工具列表
// GET /api/mcp/servers/:name/tools
func (h *MCPHandler) GetServerTools(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		ParamError(c, "服务器名称不能为空")
		return
	}

	tools := h.manager.GetServerTools(name)
	Success(c, gin.H{
		"server": name,
		"tools":  tools,
		"total":  len(tools),
	})
}

// GetAllTools 获取所有 MCP 工具
// GET /api/mcp/tools
func (h *MCPHandler) GetAllTools(c *gin.Context) {
	tools := h.manager.GetAllTools()

	// 计算总数
	total := 0
	for _, t := range tools {
		total += len(t)
	}

	Success(c, gin.H{
		"tools_by_server": tools,
		"total":           total,
	})
}

// ListServers 获取 MCP 服务器列表
// GET /api/mcp/servers
func (h *MCPHandler) ListServers(c *gin.Context) {
	if h.manager == nil {
		Success(c, gin.H{
			"servers": []interface{}{},
			"total":   0,
		})
		return
	}

	clients := h.manager.ListClients()
	servers := make([]gin.H, 0, len(clients))

	healthResults := h.manager.HealthCheck()

	for _, name := range clients {
		status := "healthy"
		if err := healthResults[name]; err != nil {
			status = "unhealthy"
		}
		servers = append(servers, gin.H{
			"name":   name,
			"status": status,
		})
	}

	Success(c, gin.H{
		"servers": servers,
		"total":   len(servers),
	})
}

// GetStats 获取 MCP 统计信息
// GET /api/mcp/stats
func (h *MCPHandler) GetStats(c *gin.Context) {
	if h.manager == nil {
		Success(c, gin.H{
			"clients":  0,
			"adapters": 0,
		})
		return
	}

	stats := h.manager.GetStats()
	Success(c, stats)
}

// HealthCheck 检查所有 MCP 服务器健康状态
// GET /api/mcp/health
func (h *MCPHandler) HealthCheck(c *gin.Context) {
	if h.manager == nil {
		Success(c, gin.H{})
		return
	}

	results := h.manager.HealthCheck()
	response := make(map[string]string)

	for name, err := range results {
		if err != nil {
			response[name] = err.Error()
		} else {
			response[name] = "ok"
		}
	}

	Success(c, response)
}
