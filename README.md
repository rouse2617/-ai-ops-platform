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
│  ┌─────────────────────────────────────────────────────────────┐ │
│  │              LLM Client (多 Provider 支持)                  │ │
│  │         OpenAI / Anthropic Claude API                       │ │
│  └─────────────────────────────────────────────────────────────┘ │
│  ┌─────────────────────────────────────────────────────────────┐ │
│  │                    Tool Registry                            │ │
│  │    内置工具 + MCP 工具 + 脚本工具 (统一管理)                 │ │
│  └─────────────────────────────────────────────────────────────┘ │
│  ┌─────────────────────────────────────────────────────────────┐ │
│  │              MCP (Model Context Protocol)                   │ │
│  │         客户端 (连接外部服务) + 服务端 (暴露工具)            │ │
│  └─────────────────────────────────────────────────────────────┘ │
│  ┌─────────────────────────────────────────────────────────────┐ │
│  │                     SSH Pool                                │ │
│  │              golang.org/x/crypto/ssh                        │ │
│  └─────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
          │                                    │
          ▼                                    ▼
┌──────────────────────┐          ┌──────────────────────┐
│   Target Hosts (SSH) │          │  External MCP Servers │
└──────────────────────┘          └──────────────────────┘
```

## 功能特性

- **自然语言交互**: 通过对话方式执行运维操作
- **多 LLM 支持**: 支持 OpenAI 和 Anthropic Claude API
- **多主机管理**: 支持管理多台服务器
- **实时监控**: CPU、内存、磁盘使用率监控
- **日志分析**: 智能日志查询和分析
- **进程管理**: 查看和管理系统进程
- **命令执行**: 安全的远程命令执行
- **MCP 集成**: 支持 Model Context Protocol，可扩展外部工具

## MCP (Model Context Protocol) 支持

本平台支持 MCP 协议，实现双角色架构：

### 作为 MCP 客户端
连接外部 MCP 服务，扩展工具能力：

```yaml
# config.yaml
mcp:
  servers:
    - name: "demo-tools"
      url: "http://localhost:3001"
      enabled: true
```

### 作为 MCP 服务端
通过 `/api/mcp/rpc` 端点暴露所有内置工具给外部 MCP 客户端：

```bash
# 初始化
curl -X POST http://localhost:1280/api/mcp/rpc \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05"}}'

# 获取工具列表
curl -X POST http://localhost:1280/api/mcp/rpc \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":2,"method":"tools/list"}'

# 调用工具
curl -X POST http://localhost:1280/api/mcp/rpc \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"list_hosts","arguments":{}}}'
```

### MCP 工具流程

```
┌─────────────────────────────────────────────────────────────────┐
│                       Tool Registry                              │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐           │
│  │ 内置工具(17) │  │ MCP工具(N)   │  │ 脚本工具     │           │
│  │ list_hosts   │  │ get_weather  │  │ custom.sh    │           │
│  │ exec_command │  │ calculate    │  │ ...          │           │
│  └──────────────┘  └──────────────┘  └──────────────┘           │
└─────────────────────────────────────────────────────────────────┘
         ↑                   ↑
         │                   │
    直接注册            mcp.Adapter 适配
                             ↑
                      外部 MCP 服务 (HTTP/stdio)
```

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

# MCP 服务配置 (可选)
mcp:
  servers:
    - name: "demo-tools"
      url: "http://localhost:3001"
      enabled: true
```

### 3. 启动服务

```bash
# 终端 1: 启动 Go 后端
go run cmd/server/main.go

# 终端 2: 启动前端 (开发模式)
cd web
pnpm install
pnpm dev

# 可选: 启动示例 MCP 服务
cd examples/mcp-server
node server.js
```

### 4. 访问应用

- 前端界面: http://localhost:5173
- 后端 API: http://localhost:1280
- MCP 管理: http://localhost:5173/mcp

## 项目结构

```
ai-ops-platform/
├── cmd/                    # Go 入口
│   └── server/
├── internal/               # Go 内部包
│   ├── api/               # HTTP API
│   ├── config/            # 配置管理
│   ├── llm/               # LLM 客户端 (OpenAI/Anthropic)
│   ├── mcp/               # MCP 客户端/服务端
│   │   ├── client.go      # HTTP 协议客户端
│   │   ├── stdio_client.go # stdio 协议客户端
│   │   ├── server.go      # MCP 服务端
│   │   ├── manager.go     # MCP 管理器
│   │   └── adapter.go     # 工具适配器
│   ├── model/             # 数据模型
│   ├── repository/        # 数据访问
│   ├── service/           # 业务逻辑
│   ├── ssh/               # SSH 连接池
│   └── tool/              # 内置工具
├── web/                    # Vue 前端
│   └── src/
│       ├── components/    # Vue 组件
│       ├── stores/        # Pinia 状态
│       ├── views/         # 页面视图
│       │   └── MCPSettings.vue  # MCP 管理页面
│       └── api/           # API 调用
├── examples/               # 示例代码
│   └── mcp-server/        # 示例 MCP 服务
├── docs/                   # 文档
│   └── MCP_INTEGRATION_DESIGN.md
├── config.yaml.example     # 配置模板
└── docker-compose.yml      # Docker 编排
```

## 内置工具列表

| 工具名 | 描述 |
|--------|------|
| `list_hosts` | 列出所有主机 |
| `check_cpu` | 检查 CPU 使用率 |
| `check_memory` | 检查内存使用情况 |
| `check_disk` | 检查磁盘使用情况 |
| `query_log` | 查询系统日志 |
| `check_process` | 查看进程列表 |
| `exec_command` | 执行远程命令 |
| `anomaly_detection` | 异常检测 |
| `trend_analysis` | 趋势分析 |
| `root_cause_diagnosis` | 根因诊断 |
| ... | 更多工具 |

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

# 使用 MCP 工具 (如果已配置)
"北京天气怎么样？"
"帮我生成 3 个 UUID"
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

## API 端点

### 核心 API
- `POST /api/chat/send` - 发送聊天消息
- `GET /api/hosts` - 获取主机列表
- `GET /api/tools` - 获取工具列表

### MCP API
- `GET /api/mcp/servers` - 获取 MCP 服务器列表
- `POST /api/mcp/servers` - 添加 MCP 服务器
- `DELETE /api/mcp/servers/:name` - 移除 MCP 服务器
- `GET /api/mcp/tools` - 获取所有 MCP 工具
- `POST /api/mcp/rpc` - MCP JSON-RPC 端点 (服务端)

## 技术栈

**后端:**
- Go 1.21 + Gin
- SQLite
- SSH (golang.org/x/crypto/ssh)
- Anthropic SDK Go / OpenAI API
- MCP (Model Context Protocol)

**前端:**
- Vue 3 + TypeScript
- Pinia
- Element Plus
- Vite

## License

MIT
