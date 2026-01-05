package builtin

import (
	"fmt"
	"strings"

	"ai-ops/internal/tool"
)

// ListContainersTool Docker 容器列表工具
type ListContainersTool struct{}

func (t *ListContainersTool) Name() string { return "list_containers" }

func (t *ListContainersTool) Description() string {
	return `列出 Docker 容器信息。
可以查看所有运行中的容器或包含停止的容器，支持按条件过滤。
当用户询问"容器列表"、"有哪些容器"、"Docker 容器状态"时使用。`
}

func (t *ListContainersTool) Parameters() []tool.Parameter {
	return []tool.Parameter{
		{
			Name:        "host",
			Type:        "string",
			Description: "目标节点名称",
			Required:    true,
		},
		{
			Name:        "all",
			Type:        "bool",
			Description: "是否包含停止的容器",
			Required:    false,
			Default:     false,
		},
		{
			Name:        "filter",
			Type:        "string",
			Description: "过滤条件（如：name=nginx, status=running）",
			Required:    false,
		},
	}
}

func (t *ListContainersTool) Execute(ctx *tool.Context, params map[string]interface{}) (*tool.Result, error) {
	host := tool.GetStringParam(params, "host", "")
	all := tool.GetBoolParam(params, "all", false)
	filter := tool.GetStringParam(params, "filter", "")

	if host == "" {
		return tool.NewErrorResult("参数 host 不能为空"), nil
	}

	if ctx == nil || ctx.SSH == nil {
		return tool.NewErrorResult("SSH 上下文未初始化"), nil
	}

	cmd := "docker ps"
	if all {
		cmd += " -a"
	}
	if filter != "" {
		cmd += fmt.Sprintf(" --filter '%s'", filter)
	}

	output, err := ctx.SSH.Exec(host, cmd)
	if err != nil {
		return tool.NewErrorResult(fmt.Sprintf("执行失败: %v", err)), nil
	}

	if strings.TrimSpace(output) == "" {
		return tool.NewResult("", "未找到容器"), nil
	}

	return tool.NewResult(map[string]interface{}{
		"host":   host,
		"output": strings.TrimSpace(output),
	}, fmt.Sprintf("成功获取 %s 的容器列表", host)), nil
}

func NewListContainersTool() *ListContainersTool { return &ListContainersTool{} }

// CheckContainerLogsTool 容器日志查询工具
type CheckContainerLogsTool struct{}

func (t *CheckContainerLogsTool) Name() string { return "check_container_logs" }

func (t *CheckContainerLogsTool) Description() string {
	return `查询 Docker 容器日志。
可以指定查询的行数和时间范围，用于排查容器问题。
当用户询问"容器日志"、"查看日志"、"容器报错"时使用。`
}

func (t *CheckContainerLogsTool) Parameters() []tool.Parameter {
	return []tool.Parameter{
		{
			Name:        "host",
			Type:        "string",
			Description: "目标节点名称",
			Required:    true,
		},
		{
			Name:        "container",
			Type:        "string",
			Description: "容器名称或 ID",
			Required:    true,
		},
		{
			Name:        "lines",
			Type:        "int",
			Description: "显示最后 N 行日志",
			Required:    false,
			Default:     100,
		},
		{
			Name:        "since",
			Type:        "string",
			Description: "时间范围（如：1h, 30m, 2024-01-01）",
			Required:    false,
		},
	}
}

func (t *CheckContainerLogsTool) Execute(ctx *tool.Context, params map[string]interface{}) (*tool.Result, error) {
	host := tool.GetStringParam(params, "host", "")
	container := tool.GetStringParam(params, "container", "")
	lines := tool.GetIntParam(params, "lines", 100)
	since := tool.GetStringParam(params, "since", "")

	if host == "" {
		return tool.NewErrorResult("参数 host 不能为空"), nil
	}
	if container == "" {
		return tool.NewErrorResult("参数 container 不能为空"), nil
	}

	if ctx == nil || ctx.SSH == nil {
		return tool.NewErrorResult("SSH 上下文未初始化"), nil
	}

	cmd := fmt.Sprintf("docker logs --tail %d", lines)
	if since != "" {
		cmd += fmt.Sprintf(" --since '%s'", since)
	}
	cmd += fmt.Sprintf(" %s", container)

	output, err := ctx.SSH.Exec(host, cmd)
	if err != nil {
		return tool.NewErrorResult(fmt.Sprintf("执行失败: %v", err)), nil
	}

	if strings.TrimSpace(output) == "" {
		return tool.NewResult("", fmt.Sprintf("容器 %s 暂无日志", container)), nil
	}

	return tool.NewResult(map[string]interface{}{
		"host":      host,
		"container": container,
		"output":    strings.TrimSpace(output),
	}, fmt.Sprintf("成功获取容器 %s 的日志", container)), nil
}

func NewCheckContainerLogsTool() *CheckContainerLogsTool { return &CheckContainerLogsTool{} }
