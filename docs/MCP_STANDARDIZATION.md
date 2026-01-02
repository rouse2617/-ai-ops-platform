# MCP 协议标准化方案

## 1. 背景

### 当前实现（自定义 HTTP REST）
```
Agent Service --HTTP--> MCP Server
  GET  /tools           获取工具列表
  POST /tools/:name     执行工具
```

### 目标（标准 MCP 协议）
```
Agent Service --stdio/SSE--> MCP Server
  JSON-RPC 2.0 格式
  标准方法: tools/list, tools/call
```

## 2. 改动范围

| 组件 | 文件 | 改动内容 |
|------|------|----------|
| MCP Server | `mcp-server/src/index.ts` | 移除 HTTP 路由，使用 MCP SDK Server |
| Agent Service | `agent-service/src/agent/index.ts` | 使用 MCP Client SDK 替代 HTTP fetch |
| 工具实现 | `mcp-server/src/tools/*.ts` | **不需要改动** |

## 3. 技术方案

### 3.1 MCP Server 改造

**改造前：**
```typescript
// HTTP REST API
app.get('/tools', () => res.json(getToolsList()));
app.post('/tools/:name', async (req, res) => {
  const result = await executeToolByName(name, input);
  res.json(result);
});
```

**改造后：**
```typescript
import { Server } from '@modelcontextprotocol/sdk/server/index.js';
import { StdioServerTransport } from '@modelcontextprotocol/sdk/server/stdio.js';

const server = new Server({
  name: 'ai-ops-mcp-server',
  version: '1.0.0'
}, {
  capabilities: { tools: {} }
});

// 注册工具列表
server.setRequestHandler(ListToolsRequestSchema, async () => ({
  tools: getToolsList()
}));

// 注册工具调用
server.setRequestHandler(CallToolRequestSchema, async (request) => {
  const { name, arguments: args } = request.params;
  const result = await executeToolByName(name, args);
  return {
    content: [{ type: 'text', text: JSON.stringify(result) }]
  };
});

// 启动 stdio 传输
const transport = new StdioServerTransport();
await server.connect(transport);
```

### 3.2 Agent Service 改造

**改造前：**
```typescript
// HTTP 调用
private async loadMCPTools(): Promise<void> {
  const response = await fetch(`${this.mcpServerUrl}/tools`);
  const mcpTools = await response.json();
  this.tools = mcpTools;
}

private async callTool(name: string, input: Record<string, unknown>): Promise<unknown> {
  const response = await fetch(`${this.mcpServerUrl}/tools/${name}`, {
    method: 'POST',
    body: JSON.stringify(input)
  });
  return response.json();
}
```

**改造后：**
```typescript
import { Client } from '@modelcontextprotocol/sdk/client/index.js';
import { StdioClientTransport } from '@modelcontextprotocol/sdk/client/stdio.js';

private mcpClient: Client;

async initialize(): Promise<void> {
  // 启动 MCP Server 子进程并连接
  const transport = new StdioClientTransport({
    command: 'node',
    args: ['path/to/mcp-server/dist/index.js']
  });

  this.mcpClient = new Client({
    name: 'agent-service',
    version: '1.0.0'
  });

  await this.mcpClient.connect(transport);
  await this.loadMCPTools();
}

private async loadMCPTools(): Promise<void> {
  const result = await this.mcpClient.listTools();
  this.tools = result.tools;
}

private async callTool(name: string, input: Record<string, unknown>): Promise<unknown> {
  const result = await this.mcpClient.callTool({ name, arguments: input });
  return JSON.parse(result.content[0].text);
}
```

## 4. 传输方式选择

| 方式 | 优点 | 缺点 | 适用场景 |
|------|------|------|----------|
| **stdio** | 标准、简单、可靠 | 需要子进程管理 | 本地部署 |
| **SSE** | 支持 HTTP、可跨网络 | 稍复杂 | 分布式部署 |

**建议**：先用 stdio，后续需要分布式再改 SSE。

## 5. 实施步骤

### 第一步：MCP Server 改造
1. 保留现有 HTTP 接口（兼容）
2. 新增 stdio 入口 `mcp-server/src/stdio.ts`
3. 测试 stdio 模式

### 第二步：Agent Service 改造
1. 添加 MCP Client SDK 依赖
2. 实现 stdio 连接逻辑
3. 替换 HTTP 调用为 MCP Client 调用

### 第三步：清理
1. 移除 MCP Server 的 HTTP 接口
2. 更新部署配置

## 6. 兼容性考虑

### 过渡期方案
可以同时支持两种模式：
```typescript
// MCP Server 同时支持 HTTP 和 stdio
if (process.env.MCP_MODE === 'stdio') {
  // stdio 模式
  const transport = new StdioServerTransport();
  await server.connect(transport);
} else {
  // HTTP 模式（兼容旧版）
  app.listen(PORT);
}
```

## 7. 好处

1. **标准兼容**：可在 Claude Desktop 等客户端直接使用
2. **生态复用**：可使用社区 MCP 工具
3. **协议规范**：遵循 JSON-RPC 2.0 标准
4. **未来扩展**：支持 Resources、Prompts 等 MCP 特性

## 8. 预估工作量

- MCP Server 改造：约 50 行代码
- Agent Service 改造：约 80 行代码
- 测试验证：需要完整测试工具调用流程

## 9. 风险点

1. **子进程管理**：需要处理 MCP Server 进程的启动/重启
2. **错误处理**：stdio 通信的错误处理与 HTTP 不同
3. **调试难度**：stdio 比 HTTP 更难调试

---

**确认后我可以开始实施改造。**
