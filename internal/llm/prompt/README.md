# Prompt Loader 使用示例

## 基本用法

```go
package main

import (
    "fmt"
    "ai-ops/internal/llm/prompt"
)

func main() {
    // 创建 Prompt 加载器
    loader := prompt.NewPromptLoader("./prompts")

    // 加载 OpenAI 的 prompt 模板
    openaiPrompt, err := loader.Load("openai")
    if err != nil {
        panic(err)
    }
    fmt.Println(openaiPrompt)

    // 渲染带变量的模板
    data := map[string]any{
        "Tools": "execute_command, read_file, write_file",
        "Context": "Current host: prod-server-01",
        "UserMessage": "Check disk usage",
    }

    rendered, err := loader.Render("anthropic", data)
    if err != nil {
        panic(err)
    }
    fmt.Println(rendered)
}
```

## 集成到 Agent

```go
package agent

import (
    "ai-ops/internal/llm"
    "ai-ops/internal/llm/prompt"
)

type Agent struct {
    provider llm.Provider
    loader   *prompt.PromptLoader
}

func NewAgent(provider llm.Provider) *Agent {
    return &Agent{
        provider: provider,
        loader:   prompt.NewPromptLoader("./prompts"),
    }
}

func (a *Agent) Chat(userMsg string, tools []llm.ToolDef) (*llm.ChatResponse, error) {
    // 渲染系统 prompt
    systemPrompt, err := a.loader.Render(a.provider.ID(), map[string]any{
        "Tools":       formatTools(tools),
        "Context":     getSystemContext(),
        "UserMessage": userMsg,
    })
    if err != nil {
        return nil, err
    }

    messages := []llm.Message{
        llm.NewSystemMessage(systemPrompt),
        llm.NewUserMessage(userMsg),
    }

    return a.provider.ChatWithTools(ctx, messages, tools)
}
```

## 模板变量

所有模板支持以下变量：

- `{{.Tools}}` - 可用工具列表
- `{{.Context}}` - 上下文信息（主机、环境等）
- `{{.UserMessage}}` - 用户消息

## 模板文件位置

```
prompts/
  providers/
    anthropic.txt   # Claude 专用 prompt
    openai.txt      # OpenAI 专用 prompt
    ollama.txt      # Ollama 专用 prompt
  base.txt          # 默认 prompt（回退）
```

## 自定义模板

创建新的 provider 模板：

```bash
# 创建新模板文件
cat > prompts/providers/custom.txt << 'EOF'
You are a custom AI assistant.

{{if .Tools}}Tools: {{.Tools}}{{end}}
{{if .Context}}Context: {{.Context}}{{end}}
{{if .UserMessage}}Query: {{.UserMessage}}{{end}}
EOF
```

然后在代码中使用：

```go
loader := prompt.NewPromptLoader("./prompts")
customPrompt, err := loader.Render("custom", data)
```
