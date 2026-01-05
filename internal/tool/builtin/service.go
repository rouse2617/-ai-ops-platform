package builtin

import (
	"fmt"
	"strings"

	"ai-ops/internal/tool"
)

// ManageServiceTool 服务管理工具
type ManageServiceTool struct{}

func (t *ManageServiceTool) Name() string { return "manage_service" }

func (t *ManageServiceTool) Description() string {
	return `# 服务管理工具

## 功能说明
通过 systemctl 管理 Linux 系统服务，支持启动、停止、重启和查询状态。

## 适用场景
- 启动/停止/重启系统服务（nginx、mysql、redis 等）
- 查询服务运行状态
- 服务故障恢复
- 部署后服务重启

## 支持的操作
- start: 启动服务
- stop: 停止服务
- restart: 重启服务
- status: 查询服务状态

## 使用示例
- 启动 nginx: manage_service(host="web1", service="nginx", action="start")
- 重启 mysql: manage_service(host="db1", service="mysql", action="restart")
- 查看状态: manage_service(host="app1", service="redis", action="status")

## 注意事项
- 需要目标主机有 sudo 权限
- 服务名称必须是系统中存在的服务
- 操作会立即生效，请谨慎使用 stop 操作`
}

func (t *ManageServiceTool) Parameters() []tool.Parameter {
	return []tool.Parameter{
		{
			Name:        "host",
			Type:        "string",
			Description: "目标主机名称（必填）",
			Required:    true,
		},
		{
			Name:        "service",
			Type:        "string",
			Description: "服务名称，如 nginx、mysql、redis（必填）",
			Required:    true,
		},
		{
			Name:        "action",
			Type:        "string",
			Description: "操作类型: start（启动）、stop（停止）、restart（重启）、status（状态）",
			Required:    true,
			Enum:        []interface{}{"start", "stop", "restart", "status"},
		},
	}
}

func (t *ManageServiceTool) Execute(ctx *tool.Context, params map[string]interface{}) (*tool.Result, error) {
	host := tool.GetStringParam(params, "host", "")
	service := tool.GetStringParam(params, "service", "")
	action := tool.GetStringParam(params, "action", "")

	if host == "" {
		return tool.NewErrorResult("参数 host 不能为空"), nil
	}
	if service == "" {
		return tool.NewErrorResult("参数 service 不能为空"), nil
	}
	if action == "" {
		return tool.NewErrorResult("参数 action 不能为空"), nil
	}

	if ctx == nil || ctx.SSH == nil {
		return tool.NewErrorResult("SSH 上下文未初始化"), nil
	}

	cmd := fmt.Sprintf("sudo systemctl %s %s", action, service)
	output, err := ctx.SSH.Exec(host, cmd)
	if err != nil {
		return tool.NewErrorResult(fmt.Sprintf("执行失败: %v", err)), nil
	}

	return tool.NewResult(map[string]interface{}{
		"host":    host,
		"service": service,
		"action":  action,
		"output":  strings.TrimSpace(output),
	}, fmt.Sprintf("成功在 %s 上执行 %s 服务的 %s 操作", host, service, action)), nil
}

func NewManageServiceTool() *ManageServiceTool { return &ManageServiceTool{} }

// CheckServiceHealthTool 服务健康检查工具
type CheckServiceHealthTool struct{}

func (t *CheckServiceHealthTool) Name() string { return "check_service_health" }

func (t *CheckServiceHealthTool) Description() string {
	return `# 服务健康检查工具

## 功能说明
通过 HTTP 或 TCP 协议检查服务的健康状态，验证服务是否正常响应。

## 适用场景
- 检查 Web 服务是否可访问
- 验证 API 接口是否正常
- 检查数据库端口是否开放
- 服务部署后的健康验证
- 负载均衡健康检查

## 支持的协议
- http: HTTP 健康检查，使用 curl 命令
- tcp: TCP 端口检查，使用 nc 命令

## 使用示例
- HTTP 检查: check_service_health(host="web1", target="localhost", port=80, protocol="http", path="/health")
- TCP 检查: check_service_health(host="db1", target="localhost", port=3306, protocol="tcp")
- API 检查: check_service_health(host="api1", target="api.example.com", port=443, protocol="http", path="/api/v1/status")

## 输出说明
- HTTP: 返回 HTTP 状态码和响应内容
- TCP: 返回端口连接状态

## 注意事项
- HTTP 检查需要目标主机安装 curl
- TCP 检查需要目标主机安装 nc（netcat）
- path 参数仅在 HTTP 协议时有效`
}

func (t *CheckServiceHealthTool) Parameters() []tool.Parameter {
	return []tool.Parameter{
		{
			Name:        "host",
			Type:        "string",
			Description: "执行检查的主机名称（必填）",
			Required:    true,
		},
		{
			Name:        "target",
			Type:        "string",
			Description: "目标地址，如 localhost、IP 或域名（必填）",
			Required:    true,
		},
		{
			Name:        "port",
			Type:        "int",
			Description: "目标端口号（必填）",
			Required:    true,
		},
		{
			Name:        "protocol",
			Type:        "string",
			Description: "检查协议: http 或 tcp（必填）",
			Required:    true,
			Enum:        []interface{}{"http", "tcp"},
		},
		{
			Name:        "path",
			Type:        "string",
			Description: "HTTP 路径，如 /health、/api/status（仅 HTTP 协议时使用）",
			Required:    false,
			Default:     "/",
		},
	}
}

func (t *CheckServiceHealthTool) Execute(ctx *tool.Context, params map[string]interface{}) (*tool.Result, error) {
	host := tool.GetStringParam(params, "host", "")
	target := tool.GetStringParam(params, "target", "")
	port := tool.GetIntParam(params, "port", 0)
	protocol := tool.GetStringParam(params, "protocol", "")
	path := tool.GetStringParam(params, "path", "/")

	if host == "" {
		return tool.NewErrorResult("参数 host 不能为空"), nil
	}
	if target == "" {
		return tool.NewErrorResult("参数 target 不能为空"), nil
	}
	if port <= 0 {
		return tool.NewErrorResult("参数 port 必须大于 0"), nil
	}
	if protocol == "" {
		return tool.NewErrorResult("参数 protocol 不能为空"), nil
	}

	if ctx == nil || ctx.SSH == nil {
		return tool.NewErrorResult("SSH 上下文未初始化"), nil
	}

	var cmd string
	if protocol == "http" {
		url := fmt.Sprintf("http://%s:%d%s", target, port, path)
		cmd = fmt.Sprintf("curl -s -o /dev/null -w '%%{http_code}' -m 5 %s", url)
	} else if protocol == "tcp" {
		cmd = fmt.Sprintf("nc -zv -w 5 %s %d 2>&1", target, port)
	} else {
		return tool.NewErrorResult(fmt.Sprintf("不支持的协议: %s", protocol)), nil
	}

	output, err := ctx.SSH.Exec(host, cmd)
	if err != nil {
		return tool.NewErrorResult(fmt.Sprintf("执行失败: %v", err)), nil
	}

	result := map[string]interface{}{
		"host":     host,
		"target":   target,
		"port":     port,
		"protocol": protocol,
		"output":   strings.TrimSpace(output),
	}

	if protocol == "http" {
		result["path"] = path
	}

	return tool.NewResult(result, fmt.Sprintf("成功检查 %s 上 %s:%d 的健康状态", host, target, port)), nil
}

func NewCheckServiceHealthTool() *CheckServiceHealthTool { return &CheckServiceHealthTool{} }

// ReloadNginxTool Nginx 配置重载工具
type ReloadNginxTool struct{}

func (t *ReloadNginxTool) Name() string { return "reload_nginx" }

func (t *ReloadNginxTool) Description() string {
	return `# Nginx 配置重载工具

## 功能说明
验证 Nginx 配置文件的正确性，并在验证通过后重载配置，实现零停机更新。

## 适用场景
- 修改 Nginx 配置后重载
- 添加新的虚拟主机配置
- 更新 SSL 证书后重载
- 修改反向代理规则后生效
- 配置文件语法验证

## 工作流程
1. 执行 nginx -t 验证配置文件语法
2. 如果验证通过且 validate_only=false，执行 systemctl reload nginx
3. ��回验证和重载结果

## 使用示例
- 仅验证配置: reload_nginx(host="web1", validate_only=true)
- 验证并重载: reload_nginx(host="web1", validate_only=false)
- 快速重载: reload_nginx(host="web1")

## 优势
- 零停机重载配置
- 自动验证配置正确性
- 避免错误配置导致服务中断
- 安全可靠的配置更新方式

## 注意事项
- 需要目标主机有 sudo 权限
- 配置验证失败时不会执行重载
- 建议先使用 validate_only=true 验证配置`
}

func (t *ReloadNginxTool) Parameters() []tool.Parameter {
	return []tool.Parameter{
		{
			Name:        "host",
			Type:        "string",
			Description: "目标主机名称（必填）",
			Required:    true,
		},
		{
			Name:        "validate_only",
			Type:        "bool",
			Description: "是否仅验证配置不重载，默认 false",
			Required:    false,
			Default:     false,
		},
	}
}

func (t *ReloadNginxTool) Execute(ctx *tool.Context, params map[string]interface{}) (*tool.Result, error) {
	host := tool.GetStringParam(params, "host", "")
	validateOnly := tool.GetBoolParam(params, "validate_only", false)

	if host == "" {
		return tool.NewErrorResult("参数 host 不能为空"), nil
	}

	if ctx == nil || ctx.SSH == nil {
		return tool.NewErrorResult("SSH 上下文未初始化"), nil
	}

	// 验证配置
	validateCmd := "sudo nginx -t"
	validateOutput, err := ctx.SSH.Exec(host, validateCmd)
	if err != nil {
		return tool.NewErrorResult(fmt.Sprintf("配置验证失败: %v\n%s", err, validateOutput)), nil
	}

	result := map[string]interface{}{
		"host":            host,
		"validate_output": strings.TrimSpace(validateOutput),
		"validate_only":   validateOnly,
	}

	// 如果仅验证，直接返回
	if validateOnly {
		return tool.NewResult(result, fmt.Sprintf("Nginx 配置验证通过: %s", host)), nil
	}

	// 重载配置
	reloadCmd := "sudo systemctl reload nginx"
	reloadOutput, err := ctx.SSH.Exec(host, reloadCmd)
	if err != nil {
		return tool.NewErrorResult(fmt.Sprintf("重载失败: %v\n%s", err, reloadOutput)), nil
	}

	result["reload_output"] = strings.TrimSpace(reloadOutput)

	return tool.NewResult(result, fmt.Sprintf("成功验证并重载 %s 的 Nginx 配置", host)), nil
}

func NewReloadNginxTool() *ReloadNginxTool { return &ReloadNginxTool{} }
