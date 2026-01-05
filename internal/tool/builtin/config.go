package builtin

import (
	"fmt"
	"strings"
	"time"

	"ai-ops/internal/tool"
)

// GetConfigTool 获取配置文件工具
type GetConfigTool struct{}

func (t *GetConfigTool) Name() string { return "get_config" }

func (t *GetConfigTool) Description() string {
	return `读取远程主机的配置文件内容。
适用于查看配置文件、检查配置项、对比配置差异等场景。
当用户询问"配置文件内容"、"查看配置"时使用。`
}

func (t *GetConfigTool) Parameters() []tool.Parameter {
	return []tool.Parameter{
		{
			Name:        "host",
			Type:        "string",
			Description: "目标节点名称",
			Required:    true,
		},
		{
			Name:        "path",
			Type:        "string",
			Description: "配置文件路径",
			Required:    true,
		},
	}
}

func (t *GetConfigTool) Execute(ctx *tool.Context, params map[string]interface{}) (*tool.Result, error) {
	host := tool.GetStringParam(params, "host", "")
	path := tool.GetStringParam(params, "path", "")

	if host == "" {
		return tool.NewErrorResult("参数 host 不能为空"), nil
	}
	if path == "" {
		return tool.NewErrorResult("参数 path 不能为空"), nil
	}

	if ctx == nil || ctx.SSH == nil {
		return tool.NewErrorResult("SSH 上下文未初始化"), nil
	}

	cmd := fmt.Sprintf("cat '%s'", path)
	output, err := ctx.SSH.Exec(host, cmd)
	if err != nil {
		return tool.NewErrorResult(fmt.Sprintf("读取配置文件失败: %v", err)), nil
	}

	return tool.NewResult(map[string]interface{}{
		"host":    host,
		"path":    path,
		"content": strings.TrimSpace(output),
	}, fmt.Sprintf("成功读取 %s 的配置文件 %s", host, path)), nil
}

func NewGetConfigTool() *GetConfigTool { return &GetConfigTool{} }

// UpdateConfigTool 更新配置文件工具
type UpdateConfigTool struct{}

func (t *UpdateConfigTool) Name() string { return "update_config" }

func (t *UpdateConfigTool) Description() string {
	return `更新远程主机的配置文件内容，支持自动备份。
默认会在修改前创建备份文件（.bak 后缀）。
适用于修改配置、批量更新配置等场景。
当用户询问"修改配置"、"更新配置文件"时使用。`
}

func (t *UpdateConfigTool) Parameters() []tool.Parameter {
	return []tool.Parameter{
		{
			Name:        "host",
			Type:        "string",
			Description: "目标节点名称",
			Required:    true,
		},
		{
			Name:        "path",
			Type:        "string",
			Description: "配置文件路径",
			Required:    true,
		},
		{
			Name:        "content",
			Type:        "string",
			Description: "新的配置文件内容",
			Required:    true,
		},
		{
			Name:        "backup",
			Type:        "bool",
			Description: "是否备份原文件",
			Required:    false,
			Default:     true,
		},
	}
}

func (t *UpdateConfigTool) Execute(ctx *tool.Context, params map[string]interface{}) (*tool.Result, error) {
	host := tool.GetStringParam(params, "host", "")
	path := tool.GetStringParam(params, "path", "")
	content := tool.GetStringParam(params, "content", "")
	backup := tool.GetBoolParam(params, "backup", true)

	if host == "" {
		return tool.NewErrorResult("参数 host 不能为空"), nil
	}
	if path == "" {
		return tool.NewErrorResult("参数 path 不能为空"), nil
	}
	if content == "" {
		return tool.NewErrorResult("参数 content 不能为空"), nil
	}

	if ctx == nil || ctx.SSH == nil {
		return tool.NewErrorResult("SSH 上下文未初始化"), nil
	}

	var cmd string
	timestamp := time.Now().Format("20060102150405")

	if backup {
		backupPath := fmt.Sprintf("%s.bak.%s", path, timestamp)
		cmd = fmt.Sprintf("cp '%s' '%s' && cat > '%s' << 'EOF'\n%s\nEOF", path, backupPath, path, content)
	} else {
		cmd = fmt.Sprintf("cat > '%s' << 'EOF'\n%s\nEOF", path, content)
	}

	output, err := ctx.SSH.Exec(host, cmd)
	if err != nil {
		return tool.NewErrorResult(fmt.Sprintf("更新配置文件失败: %v", err)), nil
	}

	result := map[string]interface{}{
		"host":   host,
		"path":   path,
		"backup": backup,
	}
	if backup {
		result["backup_path"] = fmt.Sprintf("%s.bak.%s", path, timestamp)
	}

	message := fmt.Sprintf("成功更新 %s 的配置文件 %s", host, path)
	if backup {
		message += " (已备份)"
	}

	if strings.TrimSpace(output) != "" {
		result["output"] = strings.TrimSpace(output)
	}

	return tool.NewResult(result, message), nil
}

func NewUpdateConfigTool() *UpdateConfigTool { return &UpdateConfigTool{} }

// CheckSSLCertTool SSL 证书检查工具
type CheckSSLCertTool struct{}

func (t *CheckSSLCertTool) Name() string { return "check_ssl_cert" }

func (t *CheckSSLCertTool) Description() string {
	return `检查 SSL 证书的有效期和详细信息。
可以检查证书过期时间、颁发者、主题等信息。
适用于证书监控、过期提醒等场景。
当用户询问"证书过期时间"、"SSL 证书"、"HTTPS 证书"时使用。`
}

func (t *CheckSSLCertTool) Parameters() []tool.Parameter {
	return []tool.Parameter{
		{
			Name:        "host",
			Type:        "string",
			Description: "目标节点名称",
			Required:    true,
		},
		{
			Name:        "domain",
			Type:        "string",
			Description: "要检查的域名",
			Required:    true,
		},
		{
			Name:        "port",
			Type:        "int",
			Description: "端口号",
			Required:    false,
			Default:     443,
		},
	}
}

func (t *CheckSSLCertTool) Execute(ctx *tool.Context, params map[string]interface{}) (*tool.Result, error) {
	host := tool.GetStringParam(params, "host", "")
	domain := tool.GetStringParam(params, "domain", "")
	port := tool.GetIntParam(params, "port", 443)

	if host == "" {
		return tool.NewErrorResult("参数 host 不能为空"), nil
	}
	if domain == "" {
		return tool.NewErrorResult("参数 domain 不能为空"), nil
	}

	if ctx == nil || ctx.SSH == nil {
		return tool.NewErrorResult("SSH 上下文未初始化"), nil
	}

	cmd := fmt.Sprintf(
		"echo | openssl s_client -servername %s -connect %s:%d 2>/dev/null | openssl x509 -noout -dates -subject -issuer",
		domain, domain, port,
	)

	output, err := ctx.SSH.Exec(host, cmd)
	if err != nil {
		return tool.NewErrorResult(fmt.Sprintf("检查 SSL 证书失败: %v", err)), nil
	}

	if strings.TrimSpace(output) == "" {
		return tool.NewErrorResult("无法获取证书信息，请检查域名和端口是否正确"), nil
	}

	return tool.NewResult(map[string]interface{}{
		"host":   host,
		"domain": domain,
		"port":   port,
		"output": strings.TrimSpace(output),
	}, fmt.Sprintf("成功检查 %s 的 SSL 证书 (%s:%d)", host, domain, port)), nil
}

func NewCheckSSLCertTool() *CheckSSLCertTool { return &CheckSSLCertTool{} }
