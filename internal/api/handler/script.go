package handler

import (
	"ai-ops/internal/model"
	"ai-ops/internal/repository"
	"ai-ops/internal/ssh"
	"ai-ops/internal/tool"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ScriptHandler 脚本处理器
type ScriptHandler struct {
	scriptRepo repository.ScriptRepository
	hostRepo   repository.HostRepository
	registry   *tool.Registry
	sshPool    *ssh.Pool
}

// NewScriptHandler 创建脚本处理器
func NewScriptHandler(scriptRepo repository.ScriptRepository, hostRepo repository.HostRepository, registry *tool.Registry, sshPool *ssh.Pool) *ScriptHandler {
	return &ScriptHandler{
		scriptRepo: scriptRepo,
		hostRepo:   hostRepo,
		registry:   registry,
		sshPool:    sshPool,
	}
}

// ListScripts 获取脚本列表
// GET /api/scripts
func (h *ScriptHandler) ListScripts(c *gin.Context) {
	keyword := c.Query("keyword")
	language := c.Query("language")

	var enabled *bool
	if enabledStr := c.Query("enabled"); enabledStr != "" {
		val := enabledStr == "true"
		enabled = &val
	}

	scripts, err := h.scriptRepo.List(repository.ScriptFilter{
		Keyword:  keyword,
		Language: language,
		Enabled:  enabled,
	})
	if err != nil {
		InternalError(c, "获取脚本列表失败: "+err.Error())
		return
	}

	Success(c, gin.H{
		"list":  scripts,
		"total": len(scripts),
	})
}

// GetScript 获取单个脚本
// GET /api/scripts/:id
func (h *ScriptHandler) GetScript(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		ParamError(c, "id 不能为空")
		return
	}

	script, err := h.scriptRepo.GetByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			NotFound(c, "脚本不存在")
			return
		}
		InternalError(c, "获取脚本失败: "+err.Error())
		return
	}

	Success(c, script)
}

// CreateScript 创建脚本
// POST /api/scripts
func (h *ScriptHandler) CreateScript(c *gin.Context) {
	var req struct {
		Name        string                  `json:"name" binding:"required"`
		Description string                  `json:"description"`
		Language    string                  `json:"language" binding:"required"`
		Content     string                  `json:"content" binding:"required"`
		Parameters  []model.ScriptParameter `json:"parameters"`
		Enabled     bool                    `json:"enabled"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		ParamError(c, "参数错误: "+err.Error())
		return
	}

	// 验证语言类型
	if req.Language != model.ScriptLanguageBash &&
		req.Language != model.ScriptLanguagePython &&
		req.Language != model.ScriptLanguagePowerShell {
		ParamError(c, "不支持的脚本语言: "+req.Language)
		return
	}

	// 检查名称是否已存在
	existing, _ := h.scriptRepo.GetByName(req.Name)
	if existing != nil {
		ParamError(c, "脚本名称已存在")
		return
	}

	script := &model.Script{
		ID:          uuid.New().String(),
		Name:        req.Name,
		Description: req.Description,
		Language:    req.Language,
		Content:     req.Content,
		Parameters:  req.Parameters,
		Enabled:     req.Enabled,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := h.scriptRepo.Create(script); err != nil {
		InternalError(c, "创建脚本失败: "+err.Error())
		return
	}

	// 注册到工具系统
	if err := h.registerScriptTool(script); err != nil {
		InternalError(c, "注册脚本工具失败: "+err.Error())
		return
	}

	Success(c, script)
}

// UpdateScript 更新脚本
// PUT /api/scripts/:id
func (h *ScriptHandler) UpdateScript(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		ParamError(c, "id 不能为空")
		return
	}

	var req struct {
		Name        string                  `json:"name"`
		Description string                  `json:"description"`
		Language    string                  `json:"language"`
		Content     string                  `json:"content"`
		Parameters  []model.ScriptParameter `json:"parameters"`
		Enabled     bool                    `json:"enabled"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		ParamError(c, "参数错误: "+err.Error())
		return
	}

	script, err := h.scriptRepo.GetByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			NotFound(c, "脚本不存在")
			return
		}
		InternalError(c, "获取脚本失败: "+err.Error())
		return
	}

	// 更新字段
	if req.Name != "" {
		script.Name = req.Name
	}
	if req.Description != "" {
		script.Description = req.Description
	}
	if req.Language != "" {
		script.Language = req.Language
	}
	if req.Content != "" {
		script.Content = req.Content
	}
	if req.Parameters != nil {
		script.Parameters = req.Parameters
	}
	script.Enabled = req.Enabled
	script.UpdatedAt = time.Now()

	if err := h.scriptRepo.Update(script); err != nil {
		InternalError(c, "更新脚本失败: "+err.Error())
		return
	}

	// 重新注册到工具系统
	h.registry.Unregister(script.Name)
	if err := h.registerScriptTool(script); err != nil {
		InternalError(c, "重新注册脚本工具失败: "+err.Error())
		return
	}

	Success(c, script)
}

// DeleteScript 删除脚本
// DELETE /api/scripts/:id
func (h *ScriptHandler) DeleteScript(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		ParamError(c, "id 不能为空")
		return
	}

	script, err := h.scriptRepo.GetByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			NotFound(c, "脚本不存在")
			return
		}
		InternalError(c, "获取脚本失败: "+err.Error())
		return
	}

	// 从工具系统注销
	h.registry.Unregister(script.Name)

	if err := h.scriptRepo.Delete(id); err != nil {
		InternalError(c, "删除脚本失败: "+err.Error())
		return
	}

	SuccessWithMessage(c, "删除成功", nil)
}

// ToggleScript 启用/禁用脚本
// PUT /api/scripts/:id/toggle
func (h *ScriptHandler) ToggleScript(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		ParamError(c, "id 不能为空")
		return
	}

	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		ParamError(c, "参数错误: "+err.Error())
		return
	}

	if err := h.scriptRepo.UpdateEnabled(id, req.Enabled); err != nil {
		InternalError(c, "更新脚本状态失败: "+err.Error())
		return
	}

	SuccessWithMessage(c, "操作成功", gin.H{
		"id":      id,
		"enabled": req.Enabled,
	})
}

// UploadScript 上传脚本文件
// POST /api/scripts/upload
func (h *ScriptHandler) UploadScript(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		ParamError(c, "文件上传失败: "+err.Error())
		return
	}

	// 检查文件扩展名
	ext := strings.ToLower(filepath.Ext(file.Filename))
	var language string
	switch ext {
	case ".sh":
		language = model.ScriptLanguageBash
	case ".py":
		language = model.ScriptLanguagePython
	case ".ps1":
		language = model.ScriptLanguagePowerShell
	default:
		ParamError(c, "不支持的文件类型: "+ext)
		return
	}

	// 读取文件内容
	f, err := file.Open()
	if err != nil {
		InternalError(c, "打开文件失败: "+err.Error())
		return
	}
	defer f.Close()

	content, err := io.ReadAll(f)
	if err != nil {
		InternalError(c, "读取文件失败: "+err.Error())
		return
	}

	// 使用文件名（去除扩展名）作为脚本名称
	name := strings.TrimSuffix(file.Filename, ext)

	script := &model.Script{
		ID:          uuid.New().String(),
		Name:        name,
		Description: "从文件上传: " + file.Filename,
		Language:    language,
		Content:     string(content),
		Parameters:  []model.ScriptParameter{},
		Enabled:     true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := h.scriptRepo.Create(script); err != nil {
		InternalError(c, "创建脚本失败: "+err.Error())
		return
	}

	// 注册到工具系统
	if err := h.registerScriptTool(script); err != nil {
		InternalError(c, "注册脚本工具失败: "+err.Error())
		return
	}

	Success(c, script)
}

// TestScript 测试脚本执行
// POST /api/scripts/:id/test
func (h *ScriptHandler) TestScript(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		ParamError(c, "id 不能为空")
		return
	}

	var req struct {
		Params map[string]interface{} `json:"params"`
		HostID string                 `json:"hostId"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		ParamError(c, "参数错误: "+err.Error())
		return
	}

	script, err := h.scriptRepo.GetByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			NotFound(c, "脚本不存在")
			return
		}
		InternalError(c, "获取脚本失败: "+err.Error())
		return
	}

	// 获取主机信息
	if req.HostID == "" {
		ParamError(c, "hostId 不能为空")
		return
	}

	host, err := h.hostRepo.GetByID(req.HostID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			NotFound(c, "主机不存在")
			return
		}
		InternalError(c, "获取主机失败: "+err.Error())
		return
	}

	// 添加主机到 SSH Pool
	h.sshPool.AddHost(ssh.HostInfo{
		Name:       host.Name,
		Host:       host.IP,
		Port:       host.Port,
		User:       host.User,
		Password:   host.Password,
		KeyPath:    host.KeyPath,
		KeyContent: host.KeyContent,
	})

	// 根据语言选择执行方式
	var cmd string
	switch script.Language {
	case model.ScriptLanguageBash:
		cmd = script.Content
	case model.ScriptLanguagePython:
		cmd = "python3 << 'EOF'\n" + script.Content + "\nEOF"
	case model.ScriptLanguagePowerShell:
		cmd = "powershell -Command \"" + strings.ReplaceAll(script.Content, "\"", "\\\"") + "\""
	default:
		ParamError(c, "不支持的脚本语言: "+script.Language)
		return
	}

	// 执行脚本
	output, err := h.sshPool.Exec(host.Name, cmd)
	if err != nil {
		ExecError(c, "脚本执行失败: "+err.Error())
		return
	}

	Success(c, gin.H{
		"output":   output,
		"exitCode": 0,
	})
}

// registerScriptTool 注册脚本为工具
func (h *ScriptHandler) registerScriptTool(script *model.Script) error {
	// 转换参数格式
	params := make([]tool.Parameter, len(script.Parameters))
	for i, p := range script.Parameters {
		params[i] = tool.Parameter{
			Name:        p.Name,
			Type:        p.Type,
			Description: p.Description,
			Required:    p.Required,
			Default:     p.Default,
		}
	}

	// 添加 host 参数（如果不存在）
	hasHostParam := false
	for _, p := range params {
		if p.Name == "host" {
			hasHostParam = true
			break
		}
	}
	if !hasHostParam {
		params = append(params, tool.Parameter{
			Name:        "host",
			Type:        "string",
			Description: "目标主机ID",
			Required:    true,
		})
	}

	// 创建脚本工具
	scriptTool := &ScriptTool{
		name:        script.Name,
		description: script.Description,
		parameters:  params,
		script:      script,
		sshPool:     h.sshPool,
		hostRepo:    h.hostRepo,
	}

	return h.registry.RegisterScript(scriptTool)
}

// RegisterAllScripts 注册所有已启用的脚本到工具系统（启动时调用）
func (h *ScriptHandler) RegisterAllScripts() {
	scripts, err := h.scriptRepo.List(repository.ScriptFilter{})
	if err != nil {
		return
	}

	for _, script := range scripts {
		if script.Enabled {
			h.registerScriptTool(script)
		}
	}
}

// ScriptTool 脚本工具实现
type ScriptTool struct {
	name        string
	description string
	parameters  []tool.Parameter
	script      *model.Script
	sshPool     *ssh.Pool
	hostRepo    repository.HostRepository
}

func (t *ScriptTool) Name() string {
	return t.name
}

func (t *ScriptTool) Description() string {
	return t.description
}

func (t *ScriptTool) Parameters() []tool.Parameter {
	return t.parameters
}

func (t *ScriptTool) Execute(ctx *tool.Context, params map[string]interface{}) (*tool.Result, error) {
	// 获取主机参数
	hostID := tool.GetStringParam(params, "host", "")
	if hostID == "" {
		return tool.NewErrorResult("缺少 host 参数"), nil
	}

	// 获取主机信息
	host, err := t.hostRepo.GetByID(hostID)
	if err != nil {
		return tool.NewErrorResult(fmt.Sprintf("获取主机失败: %v", err)), nil
	}

	// 添加主机到 SSH Pool
	t.sshPool.AddHost(ssh.HostInfo{
		Name:       host.Name,
		Host:       host.IP,
		Port:       host.Port,
		User:       host.User,
		Password:   host.Password,
		KeyPath:    host.KeyPath,
		KeyContent: host.KeyContent,
	})

	// 构建执行命令
	var cmd string
	switch t.script.Language {
	case model.ScriptLanguageBash:
		cmd = t.script.Content
	case model.ScriptLanguagePython:
		cmd = "python3 << 'EOF'\n" + t.script.Content + "\nEOF"
	case model.ScriptLanguagePowerShell:
		cmd = "powershell -Command \"" + strings.ReplaceAll(t.script.Content, "\"", "\\\"") + "\""
	default:
		return tool.NewErrorResult("不支持的脚本语言: " + t.script.Language), nil
	}

	// 执行脚本
	output, err := t.sshPool.Exec(host.Name, cmd)
	if err != nil {
		return tool.NewErrorResult(fmt.Sprintf("脚本执行失败: %v", err)), nil
	}

	return tool.NewResult(gin.H{
		"output": output,
		"host":   host.Name,
	}, "脚本执行成功"), nil
}
