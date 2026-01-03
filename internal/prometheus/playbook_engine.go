package prometheus

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// PlaybookEngine 自愈剧本引擎
type PlaybookEngine struct {
	playbooks  map[string]*Playbook
	executions map[string]*PlaybookExecution
	mu         sync.RWMutex
}

// NewPlaybookEngine 创建剧本引擎
func NewPlaybookEngine() *PlaybookEngine {
	return &PlaybookEngine{
		playbooks:  make(map[string]*Playbook),
		executions: make(map[string]*PlaybookExecution),
	}
}

// CreatePlaybook 创建剧本
func (pe *PlaybookEngine) CreatePlaybook(playbook Playbook) error {
	pe.mu.Lock()
	defer pe.mu.Unlock()

	if playbook.ID == "" {
		playbook.ID = fmt.Sprintf("pb-%d", time.Now().UnixNano())
	}

	playbook.CreatedAt = time.Now()
	playbook.UpdatedAt = time.Now()

	pe.playbooks[playbook.ID] = &playbook
	return nil
}

// ExecutePlaybook 执行剧本
func (pe *PlaybookEngine) ExecutePlaybook(ctx context.Context, playbookID string, context map[string]interface{}) (ExecutionResult, error) {
	pe.mu.RLock()
	playbook, ok := pe.playbooks[playbookID]
	pe.mu.RUnlock()

	if !ok {
		return ExecutionResult{}, fmt.Errorf("playbook not found: %s", playbookID)
	}

	execution := &PlaybookExecution{
		ID:         fmt.Sprintf("exec-%d", time.Now().UnixNano()),
		PlaybookID: playbookID,
		Status:     "running",
		StartTime:  time.Now(),
	}

	pe.mu.Lock()
	pe.executions[execution.ID] = execution
	pe.mu.Unlock()

	result := ExecutionResult{
		ExecutionID: execution.ID,
		PlaybookID:  playbookID,
		Status:      "running",
		Steps:       []StepResult{},
	}

	// 执行每个步骤
	for _, step := range playbook.Steps {
		stepResult := pe.executeStep(ctx, step, context)
		result.Steps = append(result.Steps, stepResult)

		if stepResult.Status == "failed" && !step.AIAnalysis {
			result.Status = "failed"
			execution.Status = "failed"
			execution.Error = stepResult.Error
			break
		}
	}

	if result.Status != "failed" {
		result.Status = "success"
		execution.Status = "success"
	}

	now := time.Now()
	execution.EndTime = &now
	execution.Output = fmt.Sprintf("Playbook %s executed with status: %s", playbookID, result.Status)

	pe.mu.Lock()
	pe.executions[execution.ID] = execution
	pe.mu.Unlock()

	return result, nil
}

// GetPlaybook 获取剧本
func (pe *PlaybookEngine) GetPlaybook(playbookID string) (Playbook, error) {
	pe.mu.RLock()
	defer pe.mu.RUnlock()

	playbook, ok := pe.playbooks[playbookID]
	if !ok {
		return Playbook{}, fmt.Errorf("playbook not found: %s", playbookID)
	}

	return *playbook, nil
}

// ListPlaybooks 列表剧本
func (pe *PlaybookEngine) ListPlaybooks(filter PlaybookFilter) ([]Playbook, error) {
	pe.mu.RLock()
	defer pe.mu.RUnlock()

	var result []Playbook

	for _, pb := range pe.playbooks {
		if filter.Name != "" && pb.Name != filter.Name {
			continue
		}

		if filter.Trigger != "" && pb.Trigger != filter.Trigger {
			continue
		}

		result = append(result, *pb)
	}

	return result, nil
}

// GetExecution 获取执行记录
func (pe *PlaybookEngine) GetExecution(executionID string) (PlaybookExecution, error) {
	pe.mu.RLock()
	defer pe.mu.RUnlock()

	exec, ok := pe.executions[executionID]
	if !ok {
		return PlaybookExecution{}, fmt.Errorf("execution not found: %s", executionID)
	}

	return *exec, nil
}

// 私有方法

func (pe *PlaybookEngine) executeStep(ctx context.Context, step PlaybookStep, context map[string]interface{}) StepResult {
	result := StepResult{
		Name:      step.Name,
		Status:    "success",
		StartTime: time.Now(),
	}

	// 执行操作
	for _, action := range step.Actions {
		output, err := pe.executeAction(ctx, action, context)
		result.Output += output + "\n"

		if err != nil {
			result.Status = "failed"
			result.Error = err.Error()
			break
		}
	}

	result.EndTime = time.Now()
	return result
}

func (pe *PlaybookEngine) executeAction(ctx context.Context, action string, context map[string]interface{}) (string, error) {
	// 这里应该调用实际的执行引擎（如 SSH 执行）
	// 现在只是模拟
	return fmt.Sprintf("Executed: %s\n", action), nil
}

// ExecutionResult 执行结果
type ExecutionResult struct {
	ExecutionID string
	PlaybookID  string
	Status      string
	Steps       []StepResult
	StartTime   time.Time
	EndTime     time.Time
}

// StepResult 步骤结果
type StepResult struct {
	Name      string
	Status    string
	Output    string
	Error     string
	StartTime time.Time
	EndTime   time.Time
}

// PlaybookFilter 剧本过滤器
type PlaybookFilter struct {
	Name    string
	Trigger string
}

// DefaultPlaybooks 返回默认剧本
func DefaultPlaybooks() []Playbook {
	return []Playbook{
		{
			ID:          "pb-high-cpu",
			Name:        "高 CPU 处理",
			Trigger:     "cpu_usage > 80%",
			AutoExecute: false,
			Steps: []PlaybookStep{
				{
					Name:       "诊断",
					Actions:    []string{"ps aux --sort=-%cpu | head -10", "top -bn1 | head -20"},
					AIAnalysis: false,
				},
				{
					Name:       "分析",
					Actions:    []string{},
					AIAnalysis: true,
				},
				{
					Name:             "修复",
					Actions:          []string{"kill -9 <pid>"},
					RequiresApproval: true,
				},
				{
					Name:    "验证",
					Actions: []string{"ps aux | grep <process>"},
				},
			},
		},
		{
			ID:          "pb-high-memory",
			Name:        "高内存处理",
			Trigger:     "memory_usage > 85%",
			AutoExecute: false,
			Steps: []PlaybookStep{
				{
					Name:       "诊断",
					Actions:    []string{"free -h", "ps aux --sort=-%mem | head -10"},
					AIAnalysis: false,
				},
				{
					Name:       "分析",
					Actions:    []string{},
					AIAnalysis: true,
				},
				{
					Name:             "修复",
					Actions:          []string{"sync && echo 3 > /proc/sys/vm/drop_caches"},
					RequiresApproval: true,
				},
			},
		},
		{
			ID:          "pb-high-disk",
			Name:        "高磁盘使用率处理",
			Trigger:     "disk_usage > 90%",
			AutoExecute: false,
			Steps: []PlaybookStep{
				{
					Name:       "诊断",
					Actions:    []string{"df -h", "du -sh /* 2>/dev/null | sort -hr | head -10"},
					AIAnalysis: false,
				},
				{
					Name:       "分析",
					Actions:    []string{},
					AIAnalysis: true,
				},
				{
					Name:             "清理",
					Actions:          []string{"find /var/log -name '*.log' -mtime +30 -delete"},
					RequiresApproval: true,
				},
			},
		},
	}
}
