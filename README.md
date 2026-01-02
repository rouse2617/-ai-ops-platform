# AI-Ops Platform

AI 驱动的智能运维平台，通过自然语言与服务器交互，实现智能化运维管理。

## 架构概览

```
┌─────────────────────────────────────────────────────────────────┐
│                         Web Frontend                             │
│                      (Vue 3 + TypeScript)                        │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                         Go Backend                               │
│                      (Gin + SQLite)                              │
│                       Port: 1280                                 │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │              LLM Client (多 Provider 支持)               │    │
│  │         OpenAI / Anthropic Claude API                    │    │
│  └─────────────────────────────────────────────────────────┘    │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │                    Tool Registry                         │    │
│  │    check_cpu / check_memory / check_disk / query_log     │    │
│  │    check_process / run_command / list_hosts              │    │
│  └─────────────────────────────────────────────────────────┘    │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │                     SSH Pool                             │    │
│  │              golang.org/x/crypto/ssh                     │    │
│  └─────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Target Hosts (SSH)                            │
└─────────────────────────────────────────────────────────────────┘
```

## 功能特性

- **自然语言交互**: 通过对话方式执行运维操作
- **多 LLM 支持**: 支持 OpenAI 和 Anthropic Claude API
- **多主机管理**: 支持管理多台服务器
- **实时监控**: CPU、内存、磁盘使用率监控
- **日志分析**: 智能日志查询和分析
- **进程管理**: 查看和管理系统进程
- **命令执行**: 安全的远程命令执行

## 快速开始

### 环境要求

- Go 1.21+
- Node.js 18+ (仅前端开发需要)
- pnpm (推荐) 或 npm

### 1. 克隆项目

```bash
git clone https://github.com/rouse2617/ai-ops-platform.git
cd ai-ops-platform
```

### 2. 配置文件

```bash
# 复制配置模板
cp config.yaml.example config.yaml

# 编辑配置文件，填入实际值
```

**config.yaml** 主要配置项：
```yaml
server:
  addr: ":1280"

llm:
  provider: "anthropic"  # openai / anthropic
  api_key: "your-api-key"
  model: "claude-sonnet-4-5-20250929"

hosts:
  - name: my-server
    host: 192.168.1.100
    user: root
    auth_type: password
    password: "your-password"
```

### 3. 启动服务

```bash
# 终端 1: 启动 Go 后端
go run cmd/server/main.go

# 终端 2: 启动前端 (开发模式)
cd web
pnpm install
pnpm dev
```

### 4. 访问应用

打开浏览器访问: http://localhost:5173

## 项目结构

```
ai-ops-platform/
├── cmd/                    # Go 入口
│   └── server/
├── internal/               # Go 内部包
│   ├── api/               # HTTP API
│   ├── config/            # 配置管理
│   ├── llm/               # LLM 客户端 (OpenAI/Anthropic)
│   ├── model/             # 数据模型
│   ├── repository/        # 数据访问
│   ├── service/           # 业务逻辑
│   ├── ssh/               # SSH 连接池
│   └── tool/              # 内置工具
├── web/                    # Vue 前端
│   └── src/
│       ├── components/    # Vue 组件
│       ├── stores/        # Pinia 状态
│       └── api/           # API 调用
├── config.yaml.example     # 配置模板
└── docker-compose.yml      # Docker 编排
```

## 内置工具列表

| 工具名 | 描述 |
|--------|------|
| `check_cpu` | 检查 CPU 使用率 |
| `check_memory` | 检查内存使用情况 |
| `check_disk` | 检查磁盘使用情况 |
| `query_log` | 查询系统日志 |
| `check_process` | 查看进程列表 |
| `run_command` | 执行远程命令 |
| `list_hosts` | 列出所有主机 |

## 使用示例

在聊天界面输入自然语言命令：

```
# 查看系统状态
"检查服务器的 CPU 和内存使用情况"

# 查看进程
"显示占用内存最多的前 5 个进程"

# 查看日志
"查看最近的系统错误日志"

# 执行命令
"执行 whoami && uname -a"
```

## LLM Provider 配置

### Anthropic Claude (推荐)

```yaml
llm:
  provider: "anthropic"
  endpoint: ""  # 留空使用官方 API
  model: "claude-sonnet-4-5-20250929"
  api_key: "your-anthropic-api-key"
```

### OpenAI

```yaml
llm:
  provider: "openai"
  endpoint: "https://api.openai.com/v1"
  model: "gpt-4"
  api_key: "your-openai-api-key"
```

### 使用代理

```yaml
llm:
  provider: "anthropic"
  endpoint: "https://your-proxy.com/api"  # 代理地址
  model: "claude-sonnet-4-5-20250929"
  api_key: "your-api-key"
```

## 技术栈

**后端:**
- Go 1.21 + Gin
- SQLite
- SSH (golang.org/x/crypto/ssh)
- Anthropic SDK Go / OpenAI API

**前端:**
- Vue 3 + TypeScript
- Pinia
- Element Plus
- Vite

## License

MIT
