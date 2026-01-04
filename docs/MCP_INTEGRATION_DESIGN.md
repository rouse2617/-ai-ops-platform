# MCP 集成技术方案

## 1. 背景与目标

### 1.1 背景
AI-Ops 平台目前有 17 个内置运维工具（check_disk、check_cpu、run_command 等），这些工具只能在平台内部使用。为了扩展平台能力并与外部生态互通，需要集成 MCP（Model Context Protocol）支持。

### 1.2 目标
- **保持兼容**：现有 17 个内置工具继续正常工作
- **扩展能力**：支持连接外部 MCP 服务器，获取更多工具
- **统一管理**：内置工具和 MCP 工具使用统一接口管理

### 1.3 范围
本方案实现 **MCP 客户端**功能，让 AI-Ops 平台可以��接外部 MCP 服务器。

---

## 2. 技术架构

### 2.1 整体架构

```
┌─────────────────────────────────────────────────────────────────┐
│                      AI-Ops 平台                                 │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │                   统一工具注册表 (Registry)                │   │
│  │  ┌─────────────────┐    ┌─────────────────────────────┐  │   │
│  │  │   内置工具适配器   │    │      MCP 工具适配器          │  │   │
│  │  │  (BuiltinAdapter) │    │     (MCPAdapter)           │  │   │
│  │  │                   │    │                            │  │   │
│  │  │  - check_disk     │    │  ┌─────────────────────┐   │  │   │
│  │  │  - check_cpu      │    │  │  MCP Client Pool    │   │  │   │
│  │  │  - run_command    │    │  │  - Server A         │   │  │   │
│  │  │  - ...            │    │  │  - Server B         │   │  │   │
│  │  │                   │    │  │  - ...              │   │  │   │
│  │  └─────────────────┘    │  └───────────────────��─┘   │  │   │
│  │                          └─────────────────────────────┘  │   │
│  └──────────────────────────────────────────────────────────┘   │
│                              │                                   │
│                              ▼                                   │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │              统一工具接口 (Tool Interface)                 │   │
│  │  - Name() string                                          │   │
│  │  - Description() string                                   │   │
│  │  - Parameters() []Parameter                               │   │
│  │  - Execute(ctx, params) (*Result, error)                  │   │
│  │  - Source() string  // "builtin" | "mcp:server-name"      │   │
│  └──────────────────────────────────────────────────────────┘   │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
                              │
                              │ JSON-RPC 2.0 over stdio/HTTP
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                    外部 MCP 服务器                               │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐             │
│  │ Prometheus  │  │  Database   │  │   Custom    │             │
│  │   Server    │  │   Server    │  │   Server    │             │
│  └─────────────┘  └─────────────┘  └─────────────┘             │
└─────────────────────────────────────────────────────────────────┘
```

### 2.2 MCP 协议简介

MCP 基于 JSON-RPC 2.0 协议，主要接口：

| 方法 | 描述 |
|------|------|
| `initialize` | 初始化连接，交换能力信息 |
| `tools/list` | 获取服务器提供的工具列表 |
| `tools/call` | 调用指定工具 |
| `resources/list` | 获取资源列表（可选） |
| `resources/read` | 读取资源内容（可选） |

### 2.3 传输方式

支持两种传输方式：
1. **stdio**：通过子进程的标准输入/输出通信（本地 MCP 服务器）
2. **HTTP/SSE**：通过 HTTP 请求和 Server-Sent Events 通信（远程 MCP 服务器）

---

## 3. 详细设计

### 3.1 目录结构

```
internal/
├── mcp/
│   ├── client.go          # MCP 客户端核心实现
│   ├── transport.go       # 传输层（stdio/HTTP）
│   ├── protocol.go        # MCP 协议消息定义
│   ├── tool_adapter.go    # MCP 工具适配器（实现 Tool 接口）
│   └── manager.go         # MCP 连接管理器
├── tool/
│   ├── interface.go       # 工具接口定义（扩展 Source 方法）
│   ├── registry.go        # 工具注册表（支持 MCP 工具）
│   └── builtin/           # 内置工具（保持不变）
```

### 3.2 核心接口定义

#### 3.2.1 扩展 Tool 接口

```go
// internal/tool/interface.go

type Tool interface {
    Name() string
    Description() string
    Parameters() []Parameter
    Execute(ctx *Context, params map[string]interface{}) (*Result, error)
    Source() string  // 新增：返回工具来源
}

// ToolSource 工具来源常量
const (
    SourceBuiltin = "builtin"
    SourceMCP     = "mcp"
)
```

#### 3.2.2 MCP 协议消息

```go
// internal/mcp/protocol.go

// JSONRPCRequest JSON-RPC 2.0 请求
type JSONRPCRequest struct {
    JSONRPC string      `json:"jsonrpc"`
    ID      int64       `json:"id"`
    Method  string      `json:"method"`
    Params  interface{} `json:"params,omitempty"`
}

// JSONRPCResponse JSON-RPC 2.0 响应
type JSONRPCResponse struct {
    JSONRPC string          `json:"jsonrpc"`
    ID      int64           `json:"id"`
    Result  json.RawMessage `json:"result,omitempty"`
    Error   *JSONRPCError   `json:"error,omitempty"`
}

// MCPToolInfo MCP 工具信息
type MCPToolInfo struct {
    Name        string                 `json:"name"`
    Description string                 `json:"description"`
    InputSchema map[string]interface{} `json:"inputSchema"`
}

// MCPToolCallResult MCP 工具调用结果
type MCPToolCallResult struct {
    Content []MCPContent `json:"content"`
    IsError bool         `json:"isError,omitempty"`
}

type MCPContent struct {
    Type string `json:"type"`
    Text string `json:"text,omitempty"`
}
```

#### 3.2.3 MCP 客户端

```go
// internal/mcp/client.go

type Client struct {
    name      string           // 服务器名称
    transport Transport        // 传输层
    tools     []MCPToolInfo    // 工具列表缓存
    mu        sync.RWMutex
}

// NewClient 创建 MCP 客户端
func NewClient(name string, transport Transport) *Client

// Initialize 初始化连接
func (c *Client) Initialize(ctx context.Context) error

// ListTools 获取工具列表
func (c *Client) ListTools(ctx context.Context) ([]MCPToolInfo, error)

// CallTool 调用工具
func (c *Client) CallTool(ctx context.Context, name string, args map[string]interface{}) (*MCPToolCallResult, error)

// Close 关闭连接
func (c *Client) Close() error
```

#### 3.2.4 MCP 工具适配器

```go
// internal/mcp/tool_adapter.go

// MCPTool 将 MCP 工具适配为 Tool 接口
type MCPTool struct {
    client     *Client
    serverName string
    info       MCPToolInfo
}

func (t *MCPTool) Name() string {
    return t.info.Name
}

func (t *MCPTool) Description() string {
    return t.info.Description
}

func (t *MCPTool) Parameters() []tool.Parameter {
    // 从 inputSchema 转换为 Parameter
}

func (t *MCPTool) Execute(ctx *tool.Context, params map[string]interface{}) (*tool.Result, error) {
    result, err := t.client.CallTool(context.Background(), t.info.Name, params)
    // 转换结果
}

func (t *MCPTool) Source() string {
    return "mcp:" + t.serverName
}
```

### 3.3 配置设计

```yaml
# config/config.yaml

mcp:
  enabled: true
  servers:
    - name: "prometheus"
      type: "stdio"                    # stdio | http
      command: "prometheus-mcp-server" # stdio 模式的命令
      args: ["--config", "/etc/prometheus/prometheus.yml"]

    - name: "database"
      type: "http"
      url: "http://localhost:8080/mcp"

    - name: "custom"
      type: "stdio"
      command: "node"
      args: ["/path/to/custom-mcp-server.js"]
```

### 3.4 工具注册表扩展

```go
// internal/tool/registry.go

type Registry struct {
    builtinTools map[string]Tool      // 内置工具
    mcpTools     map[string]Tool      // MCP 工具
    mcpManager   *mcp.Manager         // MCP 连接管理器
    mu           sync.RWMutex
}

// LoadMCPServers 加载 MCP 服务器配置
func (r *Registry) LoadMCPServers(configs []MCPServerConfig) error

// ListAll 列出所有工具（内置 + MCP）
func (r *Registry) ListAll() []Tool

// Get 获取工具（优先内置，其次 MCP）
func (r *Registry) Get(name string) (Tool, bool)

// GetBySource 按来源获取工具
func (r *Registry) GetBySource(source string) []Tool
```

---

## 4. 实现计划

### 阶段 1：核心 MCP 客户端
1. 实现 MCP 协议消息定义
2. 实现 stdio 传输层
3. 实现 MCP 客户端核心逻辑
4. 单元测试

### 阶段 2：工具适配与注册
1. 扩展 Tool 接口（添加 Source 方法）
2. 实现 MCP 工具适配器
3. 扩展工具注册表
4. 集成测试

### 阶段 3：配置与 API
1. 添加 MCP 配置支持
2. 添加 MCP 工具管理 API
3. 前端工具列表显示 MCP 工具
4. E2E 测试

---

## 5. API 设计

### 5.1 工具列表 API（扩展）

**GET /api/tools**

响应：
```json
{
  "total": 20,
  "tools": [
    {
      "name": "check_disk",
      "description": "检查磁盘使用情况",
      "type": "builtin",
      "source": "builtin",
      "enabled": true,
      "parameters": [...]
    },
    {
      "name": "query_prometheus",
      "description": "查询 Prometheus 指标",
      "type": "mcp",
      "source": "mcp:prometheus",
      "enabled": true,
      "parameters": [...]
    }
  ]
}
```

### 5.2 MCP 服务器管理 API

**GET /api/mcp/servers** - 获取 MCP 服务器列表

**POST /api/mcp/servers/:name/reconnect** - 重连 MCP 服务器

**GET /api/mcp/servers/:name/tools** - 获取指定服务器的工具列表

---

## 6. 测试计划

### 6.1 单元测试
- MCP 协议消息序列化/反序列化
- MCP 客户端连接/断开
- 工具适配器转换逻辑

### 6.2 集成测试
- 连接本地 MCP 服务器
- 工具发现和调用
- 错误处理和重连

### 6.3 E2E 测试
- 前端显示 MCP 工具
- 通过 AI 对话调用 MCP 工具
- MCP 工具启用/禁用

---

## 7. 风险与缓解

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| MCP 服务器不稳定 | 工具调用失败 | 实现重连机制和超时控制 |
| 工具名称冲突 | 调用错误工具 | MCP 工具名称加前缀 |
| 性能影响 | 响应变慢 | 工具列表缓存，异步加载 |

---

## 8. 参考资料

- [MCP 官方规范](https://modelcontextprotocol.io/specification)
- [MCP TypeScript SDK](https://github.com/modelcontextprotocol/typescript-sdk)
- [MCP Go SDK](https://github.com/mark3labs/mcp-go)
