# 声明式权限系统

## 概述

类似 `.gitignore` 的声明式权限控制系统，用于管理 AI Agent 工具的执行权限。

## 文件结构

```
internal/permission/
├── rule.go          # 规则定义和配置加载
├── matcher.go       # 通配符模式匹配
├── checker.go       # 权限检查器
├── *_test.go        # 单元测试
└── example_test.go  # 使用示例

configs/
└── permissions.yaml # 默认权限配置
```

## 快速开始

### 1. 初始化权限检查器

```go
import "ai-ops/internal/permission"

checker, err := permission.NewChecker("configs/permissions.yaml")
if err != nil {
    log.Fatalf("初始化权限检查器失败: %v", err)
}
```

### 2. 检查工具权限

```go
result := checker.Check("exec_command")

switch result.Action {
case permission.ActionAllow:
    // 自动允许，直接执行
    executeTool()
case permission.ActionDeny:
    // 拒绝执行
    return fmt.Errorf("工具被拒绝执行")
case permission.ActionAsk:
    // 需要用户确认
    if result.Message != "" {
        fmt.Println(result.Message)
    }
    if userConfirm() {
        executeTool()
    }
}
```

## 配置文件格式

```yaml
# configs/permissions.yaml

default: ask  # 默认策略: allow/deny/ask

rules:
  # 只读操作 - 自动允许
  - pattern: "check_*"
    action: allow

  # 写操作 - 需要确认
  - pattern: "exec_command"
    action: ask
    message: "执行远程命令需要确认"

  # 危险操作 - 拒绝
  - pattern: "rm_*"
    action: deny
```

## 通配符规则

- `*` - 匹配任意字符（不包括路径分隔符）
- `?` - 匹配单个字符
- `**` - 匹配所有

### 示例

| 模式 | 匹配 | 不匹配 |
|------|------|--------|
| `check_*` | `check_disk`, `check_memory` | `list_files` |
| `*_command` | `exec_command`, `run_command` | `command_exec` |
| `deploy_*_prod` | `deploy_app_prod` | `deploy_prod` |
| `**` | 所有工具 | - |

## 集成到 Agent

在 `internal/agent/agent.go` 中集成：

```go
type Agent struct {
    permChecker *permission.Checker
    // ... 其他字段
}

func NewAgent(configPath string) (*Agent, error) {
    checker, err := permission.NewChecker(configPath)
    if err != nil {
        return nil, fmt.Errorf("初始化权限检查器: %w", err)
    }

    return &Agent{
        permChecker: checker,
    }, nil
}

func (a *Agent) ExecuteTool(toolName string, params map[string]any) error {
    // 检查权限
    result := a.permChecker.Check(toolName)

    switch result.Action {
    case permission.ActionDeny:
        return fmt.Errorf("工具 %s 被拒绝执行", toolName)
    case permission.ActionAsk:
        // 请求用户确认
        if !a.requestUserConfirmation(toolName, result.Message) {
            return fmt.Errorf("用户拒绝执行工具 %s", toolName)
        }
    }

    // 执行工具
    return a.doExecuteTool(toolName, params)
}
```

## 测试

运行所有测试：

```bash
go test ./internal/permission/... -v
```

运行示例测试：

```bash
go test ./internal/permission/... -v -run Example
```

## 最佳实践

1. **分环境配置**：开发环境可以更宽松，生产环境更严格
2. **定期审查规则**：根据实际使用情况调整权限规则
3. **记录审计日志**：记录所有权限检查和用户确认操作
4. **最小权限原则**：默认拒绝，显式允许需要的操作

## API 参考

### Types

```go
type Action string
const (
    ActionAllow Action = "allow"
    ActionDeny  Action = "deny"
    ActionAsk   Action = "ask"
)

type CheckResult struct {
    Action  Action
    Message string
}
```

### Functions

```go
// 创建权限检查器
func NewChecker(configPath string) (*Checker, error)

// 检查工具权限
func (c *Checker) Check(toolName string) CheckResult

// 加载配置文件
func LoadConfig(path string) (*Config, error)

// 模式匹配
func Match(pattern, name string) bool
```
