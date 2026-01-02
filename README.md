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
│                    (Gin + SQLite + SSH)                          │
│                      Port: 1280                                  │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌───────────────────────────────────────��─────────────────────────┐
│                      Agent Service                               │
│                 (Node.js + Claude API)                           │
│                      Port: 3001                                  │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                       MCP Server                                 │
│              (Model Context Protocol - stdio)                    │
│                                                                  │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐        │
│  │ System   │  │   Log    │  │ Process  │  │   Host   │        │
│  │  Tools   │  │  Tools   │  │  Tools   │  │  Tools   │        │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘        │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Target Hosts (SSH)                            │
│              Direct SSH via ssh2 library                         │
└─────────────────────────────────────────────────────────────────┘
```

## 功能特性

- **自然语言交互**: 通过对话方式执行运维操作
- **多主机管理**: 支持管理多台服务器
- **实时监控**: CPU、内存、磁盘使用率监控
- **日志分析**: 智能日志查询和分析
- **进程管理**: 查看和管理系统进程
- **命令执行**: 安全的远程命令执行
- **MCP 协议**: 标准化的工具调用接口

## 快速开始

### 环境要求

- Go 1.21+
- Node.js 18+
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
cp agent-service/.env.example agent-service/.env

# 编辑配置文件，填入实际值
```

**config.yaml** 主要配置项：
```yaml
server:
  addr: ":1280"

llm:
  api_key: "your-api-key"  # LLM API Key

hosts:
  - name: my-server
    host: 192.168.1.100
    user: root
    auth_type: password
    password: "your-password"
```

**agent-service/.env** 配置：
```env
ANTHROPIC_API_KEY=your_api_key_here
GO_BACKEND_URL=http://localhost:1280
```

### 3. 启动服务

#### 方式一：分别启动

```bash
# 终端 1: 启动 Go 后端
go run cmd/server/main.go

# 终端 2: 构建并启动 MCP Server
cd mcp-server
npm install
npm run build

# 终端 3: 启动 Agent Service
cd agent-service
npm install
npm run dev

# 终端 4: 启动前端
cd web
pnpm install
pnpm dev
```

#### 方式二：Docker Compose

```bash
docker-compose up -d
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
│   ├── model/             # 数据模型
│   ├── repository/        # 数据访问
│   └── service/           # 业务逻辑
├── agent-service/          # Node.js Agent 服务
│   └── src/
│       ├── agent/         # Claude API 集成
│       └── types/         # TypeScript 类型
├── mcp-server/             # MCP 协议服务
│   └── src/
│       ├── ssh/           # SSH 客户端
│       └── tools/         # MCP 工具定义
├── web/                    # Vue 前端
│   └── src/
│       ├── components/    # Vue 组件
│       ├── stores/        # Pinia 状态
│       └── api/           # API 调用
├── config.yaml.example     # 配置模板
└── docker-compose.yml      # Docker 编排
```

## MCP 工具列表

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

## 开发指南

### 添加新的 MCP 工具

1. 在 `mcp-server/src/tools/` 创建工具类
2. 在 `mcp-server/src/tools/index.ts` 注册工具
3. 重新构建 MCP Server

### 前端开发

```bash
cd web
pnpm dev
```

### 后端开发

```bash
go run cmd/server/main.go
```

## 技术栈

**后端:**
- Go 1.21 + Gin
- SQLite
- SSH (golang.org/x/crypto/ssh)

**Agent Service:**
- Node.js + TypeScript
- Anthropic Claude API
- MCP SDK

**MCP Server:**
- Node.js + TypeScript
- ssh2 (直接 SSH 连接)
- Model Context Protocol

**前端:**
- Vue 3 + TypeScript
- Pinia
- Element Plus
- Vite

## License

MIT
