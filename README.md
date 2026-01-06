<p align="center">
  <img src="docs/assets/logo.svg" alt="AI-Ops Logo" width="120" height="120">
</p>

<h1 align="center">AI-Ops Platform</h1>

<p align="center">
  <strong>企业级 AI 驱动智能运维平台</strong>
</p>

<p align="center">
  通过自然语言与基础设施交互，实现智能化运维管理
</p>

<p align="center">
  <a href="#核心特性">核心特性</a> •
  <a href="#快速开始">快速开始</a> •
  <a href="#架构设计">架构设计</a> •
  <a href="#部署指南">部署指南</a> •
  <a href="#api-文档">API 文档</a>
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat-square&logo=go" alt="Go Version">
  <img src="https://img.shields.io/badge/Vue-3.x-4FC08D?style=flat-square&logo=vue.js" alt="Vue Version">
  <img src="https://img.shields.io/badge/TypeScript-5.x-3178C6?style=flat-square&logo=typescript" alt="TypeScript">
  <img src="https://img.shields.io/badge/License-MIT-yellow?style=flat-square" alt="License">
</p>

---

## 产品概述

**AI-Ops Platform** 是新一代企业级智能运维平台，将大语言模型（LLM）能力与传统运维工具深度融合，让运维工程师通过自然语言即可完成复杂的运维操作。

### 为什么选择 AI-Ops？

| 传��运维 | AI-Ops 智能运维 |
|---------|----------------|
| 记忆大量命令和参数 | 自然语言描述意图 |
| 手动分析日志和指标 | AI 自动关联分析 |
| 逐台主机执行操作 | 批量智能编排 |
| 被动响应告警 | 主动异常检测 |
| 依赖专家经验 | 知识沉淀复用 |

---

## 核心特性

### 自然语言运维

```
用户: "检查所有生产服务器的磁盘使用情况，找出使用率超过 80% 的"

AI-Ops: 正在检查 12 台生产服务器...
        发现 3 台服务器磁盘使用率超过阈值：
        • prod-web-03: /data 分区 87%
        • prod-db-01: /var/lib/mysql 92%
        • prod-log-02: /var/log 85%
        建议: 清理日志或扩容磁盘
```

### 多 LLM 支持

- **Anthropic Claude** - 推荐，强大的推理和工具调用能力
- **OpenAI GPT-4** - 广泛兼容，生态丰富
- **私有化部署** - 支持自建 LLM 服务

### MCP 协议集成

基于 [Model Context Protocol](https://modelcontextprotocol.io/) 实现工具扩展：

- **作为 MCP 客户端**: 连接外部 MCP 服务，扩展工具能力
- **作为 MCP 服务端**: 暴露内置工具给 Claude Desktop 等客户端

### 企业级安全

- SSH 密钥/密码加密存储 (AES-256)
- 操作审计日志
- 危险命令拦截与确认
- 基于角色的访问控制 (RBAC)

---

## 功能矩阵

| 模块 | 功能 | 状态 |
|-----|------|-----|
| **主机管理** | 多主机注册、分组、标签 | ✅ |
| **智能对话** | 自然语言交互、上下文理解 | ✅ |
| **命令执行** | 远程命令、批量执行、结果聚合 | ✅ |
| **系统监控** | CPU/内存/磁盘/网络实时监控 | ✅ |
| **日志分析** | 智能日志查询、异常检测 | ✅ |
| **进程管理** | 进程列表、资源占用分析 | ✅ |
| **脚本管理** | 脚本库、参数化执行 | ✅ |
| **MCP 集成** | 工具扩展、协议适配 | ✅ |
| **Prometheus** | 指标查询、语义化告警 | ✅ |
| **根因分析** | 故障诊断、关联分析 | ✅ |
| **趋势预测** | 容量规划、异常预警 | ✅ |

---

## 架构设计

```
┌─────────────────────────────────────────────────────────────────────────┐
│                           Web Frontend                                   │
│                    Vue 3 + TypeScript + Bento UI                        │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                           API Gateway                                    │
│                         Go + Gin Framework                               │
├─────────────────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐    │
│  │  Chat API   │  │  Host API   │  │  Tool API   │  │  MCP API    │    │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘    │
├─────────────────────────────────────────────────────────────────────────┤
│                          Service Layer                                   │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │                      AI Agent Engine                             │    │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐        │    │
│  │  │ Planning │  │ Routing  │  │ Execution│  │ Learning │        │    │
│  │  └──────────┘  └──────────┘  └──────────┘  └──────────┘        │    │
│  └─────────────────────────────────────────────────────────────────┘    │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │                      Tool Registry                               │    │
│  │  ┌──────────┐  ┌──────────┐  ┌──���───────┐  ┌──────────┐        │    │
│  │  │ Built-in │  │   MCP    │  │  Script  │  │ Prometheus│        │    │
│  │  │  Tools   │  │  Tools   │  │  Tools   │  │  Tools   │        │    │
│  │  └──────────┘  └──────────┘  └──────────┘  └──────────┘        │    │
│  └─��───────────────────────────────────────────────────────────────┘    │
├─────────────────────────────────────────────────────────────────────────┤
│                        Infrastructure Layer                              │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌���─────────┐               │
│  │ SSH Pool │  │ LLM Client│  │ MCP Client│  │ Database │               │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘               │
└─────────────────────────────────────────────────────────────────────────┘
         │                │                │
         ▼                ▼                ▼
��──────────────┐  ┌──────────────┐  ┌──────────────┐
│ Target Hosts │  │  LLM APIs    │  │ MCP Servers  │
│    (SSH)     │  │ Claude/GPT   │  │  (External)  │
└──────────────┘  └──────────────┘  └──────────────┘
```

---

## 快速开始

### 环境要求

- Go 1.21+
- Node.js 18+
- SQLite 3.x (内置)

### 1. 克隆项目

```bash
git clone https://github.com/your-org/ai-ops-platform.git
cd ai-ops-platform
```

### 2. 配置文件

```bash
cp config.yaml.example config.yaml
```

编辑 `config.yaml`:

```yaml
server:
  addr: ":1280"
  mode: "release"

llm:
  provider: "anthropic"          # anthropic / openai
  api_key: "${ANTHROPIC_API_KEY}"
  model: "claude-sonnet-4-5-20250929"

database:
  driver: "sqlite"
  dsn: "./data/ai-ops.db"

security:
  encryption_key: "${ENCRYPTION_KEY}"  # 32 字节密钥
  jwt_secret: "${JWT_SECRET}"
```

### 3. 启动服务

```bash
# 后端服务
go run cmd/server/main.go

# 前端开发服务器
cd web && npm install && npm run dev
```

### 4. 访问应用

- **Web 界面**: http://localhost:1281
- **API 服务**: http://localhost:1280
- **API 文档**: http://localhost:1280/swagger

---

## 部署指南

### Docker 部署

```bash
# 构建镜像
docker build -t ai-ops-platform:latest .

# 运行容器
docker run -d \
  --name ai-ops \
  -p 1280:1280 \
  -v $(pwd)/config.yaml:/app/config.yaml \
  -v $(pwd)/data:/app/data \
  -e ANTHROPIC_API_KEY=your-key \
  ai-ops-platform:latest
```

### Docker Compose

```yaml
version: '3.8'
services:
  ai-ops:
    image: ai-ops-platform:latest
    ports:
      - "1280:1280"
    volumes:
      - ./config.yaml:/app/config.yaml
      - ./data:/app/data
    environment:
      - ANTHROPIC_API_KEY=${ANTHROPIC_API_KEY}
    restart: unless-stopped
```

### Kubernetes

```bash
kubectl apply -f deploy/kubernetes/
```

---

## 内置工具

| 工具 | 描述 | 风险等级 |
|-----|------|---------|
| `list_hosts` | 列出所有主机 | 低 |
| `check_cpu` | 检查 CPU 使用率 | 低 |
| `check_memory` | 检查内存使用情况 | 低 |
| `check_disk` | 检查磁盘使用情况 | 低 |
| `check_process` | 查看进程列表 | 低 |
| `query_log` | 查询系统日志 | 低 |
| `exec_command` | 执行远程命令 | 高 |
| `anomaly_detection` | 异常检测 | 低 |
| `trend_analysis` | 趋势分析 | 低 |
| `root_cause_diagnosis` | 根因诊断 | 低 |

---

## API 文档

### 核心接口

```http
# 发送聊天消息
POST /api/chat/send
Content-Type: application/json

{
  "message": "检查服务器状态",
  "host_ids": ["host-1", "host-2"],
  "session_id": "session-xxx"
}
```

```http
# 获取主机列表
GET /api/hosts

# 获取工具列表
GET /api/tools

# MCP RPC 端点
POST /api/mcp/rpc
```

### 流式响应

聊天接口支持 Server-Sent Events (SSE) 流式响应：

```javascript
const eventSource = new EventSource('/api/chat/stream?session_id=xxx');

eventSource.onmessage = (event) => {
  const data = JSON.parse(event.data);
  // data.type: 'thinking' | 'tool_call' | 'tool_result' | 'content' | 'done'
};
```

---

## 技术栈

### 后端

| 技术 | 用途 |
|-----|------|
| Go 1.21 | 主语言 |
| Gin | Web 框架 |
| GORM | ORM |
| SQLite | 数据存储 |
| golang.org/x/crypto/ssh | SSH 连接 |
| Anthropic SDK | LLM 集成 |
| Zap | 结构化日志 |

### 前端

| 技术 | 用途 |
|-----|------|
| Vue 3 | UI 框架 |
| TypeScript | 类型安全 |
| Pinia | 状态管理 |
| Element Plus | UI 组件库 |
| Vite | 构建工具 |
| ECharts | 数据可视化 |

---

## 项目结构

```
ai-ops-platform/
├── cmd/server/              # 应用入口
├── internal/
│   ├── api/                 # HTTP API 层
│   │   ├── handler/         # 请求处理器
│   │   └── router/          # 路由配置
│   ├── agent/               # AI Agent 引擎
│   ├── llm/                 # LLM 提供商适配
│   ├── mcp/                 # MCP 协议实现
│   ├── model/               # 数据模型
│   ├── repository/          # 数据访问层
│   ├── service/             # 业务逻辑层
│   ├── ssh/                 # SSH 连接池
│   └── tool/                # 内置工具
├── web/                     # Vue 前端
│   └─�� src/
│       ├── components/      # Vue 组件
│       ├── views/           # 页面视图
│       ├── stores/          # Pinia 状态
│       ├── api/             # API 调用
│       └── styles/          # 样式文件
├── deploy/                  # 部署配置
│   ├── docker/
│   └── kubernetes/
├── docs/                    # 文档
└── config.yaml.example      # 配置模板
```

---

## 路线图

- [x] 核心对话引擎
- [x] 多主机管理
- [x] MCP 协议支持
- [x] Prometheus 集成
- [x] Bento Grid UI
- [ ] RBAC 权限管理
- [ ] 工作流编排
- [ ] 知识库 RAG
- [ ] 多租户支持
- [ ] Kubernetes 原生集成

---

## 贡献指南

我们欢迎社区贡献！请阅读 [CONTRIBUTING.md](CONTRIBUTING.md) 了解如何参与。

```bash
# 开发环境
go mod download
cd web && npm install

# 运行测试
go test ./...
cd web && npm run test

# 代码检查
golangci-lint run
cd web && npm run lint
```

---

## 许可证

本项目采用 [MIT License](LICENSE) 开源许可证。

---

<p align="center">
  <sub>Built with ❤��� by AI-Ops Team</sub>
</p>
