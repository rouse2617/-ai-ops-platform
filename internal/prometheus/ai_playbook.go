package prometheus

import (
	"context"
	"fmt"
	"strings"
	"time"

	"ai-pro/internal/model"
)

// AIPlaybook AI 排障剧本引擎
type AIPlaybook struct {
	client    *Client
	playbooks map[string]*Playbook
}

// Playbook 排障剧本
type Playbook struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Trigger     PlaybookTrigger `json:"trigger"`
	Steps       []PlaybookStep `json:"steps"`
	AutoExecute bool           `json:"auto_execute"` // 是否自动执行
	Cooldown    time.Duration  `json:"cooldown"`     // 冷却时间
	LastRun     time.Time      `json:"last_run"`
}

// PlaybookTrigger 触发条件
type PlaybookTrigger struct {
	AlertName   string            `json:"alert_name"`
	Metric      string            `json:"metric"`
	Condition   string            `json:"condition"` // gt, lt, eq
	Threshold   float64           `json:"threshold"`
	Labels      map[string]string `json:"labels"`
}

// PlaybookStep 剧本步骤
type PlaybookStep struct {
	Name             string   `json:"name"`
	Type             string   `json:"type"` // diagnose, action, confirm
	Command          string   `json:"command,omitempty"`
	Query            string   `json:"query,omitempty"` // PromQL 查询
	Description      string   `json:"description"`
	RequiresApproval bool     `json:"requires_approval"`
	Dangerous        bool     `json:"dangerous"`
	Alternatives     []string `json:"alternatives,omitempty"`
}

// PlaybookExecution 剧本执行记录
type PlaybookExecution struct {
	ID          string                   `json:"id"`
	PlaybookID  string                   `json:"playbook_id"`
	AlertName   string                   `json:"alert_name"`
	Instance    string                   `json:"instance"`
	Status      string                   `json:"status"` // pending, running, waiting_approval, completed, failed
	Steps       []StepExecution          `json:"steps"`
	StartTime   time.Time                `json:"start_time"`
	EndTime     time.Time                `json:"end_time,omitempty"`
	Actions     []model.QuickAction      `json:"actions"` // 可执行的快捷操作
	Summary     string                   `json:"summary"`
}

// StepExecution 步骤执行记录
type StepExecution struct {
	StepName  string      `json:"step_name"`
	Status    string      `json:"status"`
	Output    interface{} `json:"output,omitempty"`
	Error     string      `json:"error,omitempty"`
	StartTime time.Time   `json:"start_time"`
	EndTime   time.Time   `json:"end_time,omitempty"`
}

// NewAIPlaybook 创建 AI 排障剧本引擎
func NewAIPlaybook(client *Client) *AIPlaybook {
	p := &AIPlaybook{
		client:    client,
		playbooks: make(map[string]*Playbook),
	}
	p.registerDefaultPlaybooks()
	return p
}

// registerDefaultPlaybooks 注册默认剧本
func (p *AIPlaybook) registerDefaultPlaybooks() {
	// 磁盘空间不足剧本
	p.playbooks["disk_space_low"] = &Playbook{
		ID:          "disk_space_low",
		Name:        "磁盘空间不足处理",
		Description: "当磁盘使用率超过 90% 时自动诊断并提供清理选项",
		Trigger: PlaybookTrigger{
			AlertName: "DiskSpaceLow",
			Metric:    "node_filesystem_avail_bytes",
			Condition: "lt",
			Threshold: 10, // 剩余空间小于 10%
		},
		Steps: []PlaybookStep{
			{
				Name:        "扫描大文件",
				Type:        "diagnose",
				Command:     "find /var -type f -size +100M -exec ls -lh {} \\; 2>/dev/null | head -20",
				Description: "查找 /var 目录下大于 100MB 的文件",
			},
			{
				Name:        "检查日志目录",
				Type:        "diagnose",
				Command:     "du -sh /var/log/* 2>/dev/null | sort -rh | head -10",
				Description: "检查日志目录占用情况",
			},
			{
				Name:             "清理旧日志",
				Type:             "action",
				Command:          "find /var/log -name '*.gz' -mtime +7 -delete",
				Description:      "删除 7 天前的压缩日志",
				RequiresApproval: true,
				Alternatives:     []string{"清理 3 天前的日志", "清理 30 天前的日志", "稍后提醒我"},
			},
			{
				Name:             "清理临时文件",
				Type:             "action",
				Command:          "find /tmp -type f -mtime +3 -delete",
				Description:      "删除 3 天前的临时文件",
				RequiresApproval: true,
			},
		},
		Cooldown: 30 * time.Minute,
	}

	// CPU 过高剧本
	p.playbooks["high_cpu"] = &Playbook{
		ID:          "high_cpu",
		Name:        "CPU 使用率过高处理",
		Description: "当 CPU 使用率超过 90% 时自动诊断",
		Trigger: PlaybookTrigger{
			AlertName: "HighCPU",
			Metric:    "node_cpu_seconds_total",
			Condition: "gt",
			Threshold: 90,
		},
		Steps: []PlaybookStep{
			{
				Name:        "查看 CPU 占用进程",
				Type:        "diagnose",
				Command:     "ps aux --sort=-%cpu | head -15",
				Description: "列出 CPU 占用最高的进程",
			},
			{
				Name:        "检查系统负载",
				Type:        "diagnose",
				Query:       "node_load1",
				Description: "查看系统 1 分钟负载",
			},
			{
				Name:        "检查 IO 等待",
				Type:        "diagnose",
				Query:       `rate(node_cpu_seconds_total{mode="iowait"}[5m]) * 100`,
				Description: "检查 IO 等待时间",
			},
			{
				Name:             "重启高 CPU 服务",
				Type:             "action",
				Command:          "systemctl restart <service>",
				Description:      "重启占用 CPU 最高的服务",
				RequiresApproval: true,
				Dangerous:        true,
			},
		},
		Cooldown: 15 * time.Minute,
	}

	// 内存不足剧本
	p.playbooks["memory_low"] = &Playbook{
		ID:          "memory_low",
		Name:        "内存不足处理",
		Description: "当内存使用率超过 90% 时自动诊断",
		Trigger: PlaybookTrigger{
			AlertName: "MemoryLow",
			Metric:    "node_memory_MemAvailable_bytes",
			Condition: "lt",
			Threshold: 10,
		},
		Steps: []PlaybookStep{
			{
				Name:        "查看内存占用进程",
				Type:        "diagnose",
				Command:     "ps aux --sort=-%mem | head -15",
				Description: "列出内存占用最高的进程",
			},
			{
				Name:        "检查 Swap 使用",
				Type:        "diagnose",
				Query:       `(node_memory_SwapTotal_bytes - node_memory_SwapFree_bytes) / node_memory_SwapTotal_bytes * 100`,
				Description: "检查 Swap 使用率",
			},
			{
				Name:        "检查 OOM 事件",
				Type:        "diagnose",
				Command:     "dmesg | grep -i 'out of memory' | tail -5",
				Description: "检查最近的 OOM 事件",
			},
			{
				Name:             "清理缓存",
				Type:             "action",
				Command:          "sync && echo 3 > /proc/sys/vm/drop_caches",
				Description:      "清理系统缓存",
				RequiresApproval: true,
			},
		},
		Cooldown: 15 * time.Minute,
	}

	// 服务不可用剧本
	p.playbooks["service_down"] = &Playbook{
		ID:          "service_down",
		Name:        "服务不可用处理",
		Description: "当服务探测失败时自动诊断",
		Trigger: PlaybookTrigger{
			AlertName: "ServiceDown",
			Metric:    "up",
			Condition: "eq",
			Threshold: 0,
		},
		Steps: []PlaybookStep{
			{
				Name:        "检查服务状态",
				Type:        "diagnose",
				Command:     "systemctl status <service> --no-pager",
				Description: "查看服务当前状态",
			},
			{
				Name:        "查看服务日志",
				Type:        "diagnose",
				Command:     "journalctl -u <service> -n 50 --no-pager",
				Description: "查看最近 50 条服务日志",
			},
			{
				Name:        "检查端口监听",
				Type:        "diagnose",
				Command:     "ss -tunlp | grep <port>",
				Description: "检查端口是否在监听",
			},
			{
				Name:             "重启服务",
				Type:             "action",
				Command:          "systemctl restart <service>",
				Description:      "重启服务",
				RequiresApproval: true,
				Dangerous:        true,
				Alternatives:     []string{"查看更多日志", "稍后提醒我"},
			},
		},
		Cooldown: 5 * time.Minute,
	}
}

// MatchPlaybook 匹配告警对应的剧本
func (p *AIPlaybook) MatchPlaybook(alertName string, labels map[string]string) *Playbook {
	for _, playbook := range p.playbooks {
		if strings.Contains(strings.ToLower(alertName), strings.ToLower(playbook.Trigger.AlertName)) {
			return playbook
		}
	}
	return nil
}

// ExecutePlaybook 执行剧本
func (p *AIPlaybook) ExecutePlaybook(ctx context.Context, playbook *Playbook, instance string) (*PlaybookExecution, error) {
	// 检查冷却时间
	if time.Since(playbook.LastRun) < playbook.Cooldown {
		return nil, fmt.Errorf("剧本在冷却中，请稍后再试")
	}

	execution := &PlaybookExecution{
		ID:         fmt.Sprintf("exec-%d", time.Now().UnixNano()),
		PlaybookID: playbook.ID,
		Instance:   instance,
		Status:     "running",
		Steps:      make([]StepExecution, 0),
		StartTime:  time.Now(),
		Actions:    make([]model.QuickAction, 0),
	}

	var summaryParts []string

	// 执行诊断步骤
	for _, step := range playbook.Steps {
		stepExec := StepExecution{
			StepName:  step.Name,
			Status:    "running",
			StartTime: time.Now(),
		}

		if step.Type == "diagnose" {
			// 执行诊断（这里只是模拟，实际需要 SSH 执行）
			if step.Query != "" {
				// 执行 PromQL 查询
				result, err := p.client.Query(ctx, step.Query)
				if err != nil {
					stepExec.Error = err.Error()
					stepExec.Status = "failed"
				} else {
					stepExec.Output = result.Data.Result
					stepExec.Status = "completed"
				}
			} else {
				// 命令执行需要 SSH，这里返回命令供前端展示
				stepExec.Output = map[string]string{
					"command":     step.Command,
					"description": step.Description,
				}
				stepExec.Status = "pending_execution"
			}
			summaryParts = append(summaryParts, fmt.Sprintf("- %s: %s", step.Name, step.Description))
		} else if step.Type == "action" {
			// 需要用户确认的操作
			action := model.QuickAction{
				ID:          fmt.Sprintf("action-%s-%d", playbook.ID, len(execution.Actions)),
				Label:       step.Name,
				Command:     step.Command,
				Description: step.Description,
				Dangerous:   step.Dangerous,
			}
			execution.Actions = append(execution.Actions, action)

			// 添加替代选项
			for i, alt := range step.Alternatives {
				execution.Actions = append(execution.Actions, model.QuickAction{
					ID:          fmt.Sprintf("alt-%s-%d-%d", playbook.ID, len(execution.Actions), i),
					Label:       alt,
					Description: alt,
				})
			}

			stepExec.Status = "waiting_approval"
		}

		stepExec.EndTime = time.Now()
		execution.Steps = append(execution.Steps, stepExec)
	}

	// 生成摘要
	execution.Summary = p.generateSummary(playbook, execution, summaryParts)
	execution.Status = "waiting_approval"
	playbook.LastRun = time.Now()

	return execution, nil
}

// generateSummary 生成执行摘要
func (p *AIPlaybook) generateSummary(playbook *Playbook, exec *PlaybookExecution, diagResults []string) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("**%s**\n\n", playbook.Name))
	sb.WriteString(fmt.Sprintf("实例: %s\n\n", exec.Instance))

	if len(diagResults) > 0 {
		sb.WriteString("**诊断结果：**\n")
		for _, r := range diagResults {
			sb.WriteString(r + "\n")
		}
		sb.WriteString("\n")
	}

	if len(exec.Actions) > 0 {
		sb.WriteString("**可执行操作：**\n")
		sb.WriteString("请选择要执行的操作，或稍后手动处理。\n")
	}

	return sb.String()
}

// GetPlaybook 获取剧本
func (p *AIPlaybook) GetPlaybook(id string) *Playbook {
	return p.playbooks[id]
}

// ListPlaybooks 列出所有剧本
func (p *AIPlaybook) ListPlaybooks() []*Playbook {
	result := make([]*Playbook, 0, len(p.playbooks))
	for _, pb := range p.playbooks {
		result = append(result, pb)
	}
	return result
}

// RegisterPlaybook 注册自定义剧本
func (p *AIPlaybook) RegisterPlaybook(playbook *Playbook) {
	p.playbooks[playbook.ID] = playbook
}
