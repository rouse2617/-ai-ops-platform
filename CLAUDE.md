# AI-Ops 智能运维平台

## 项目概述
AI 驱动的智能运维管理平台，支持多主机管理、自然语言运维、MCP 工具集成。

## 技术栈

### 后端 (Go 1.25)
- Web 框架: Gin
- ORM: GORM + SQLite
- LLM: Anthropic Claude API
- 认证: JWT
- 日志: Zap
- 监控: Prometheus

### 前端 (Vue 3)
- 构建: Vite
- UI: Element Plus
- 状态: Pinia
- 图表: ECharts
- 测试: Cypress + Vitest

## 项目结构
```
cmd/server/          # 主程序入口
internal/
  agent/             # AI Agent 核心逻辑
  api/handler/       # HTTP 处理器
  api/router/        # 路由配置
  llm/               # LLM 提供商适配
  mcp/               # MCP 协议实现
  model/             # 数据模型
  repository/        # 数据访问层
  tool/builtin/      # 内置工具
  ssh/               # SSH 执行器
  rag/               # RAG 检索增强
  monitor/           # 监控告警
web/src/
  components/        # Vue 组件
  views/             # 页面视图
  stores/            # Pinia 状态
  api/               # API 调用
```

---

## 开发规范

### 重要原则
- **不要随意改代码**：修改前先理解现有逻辑
- **大特性必须写单测**：修改重要功能后必须编写单元测试
- **及时 commit**：完成功能后及时提交，保持小步提交

---

## Go 后端规范

### 错误处理
- **必须包裹错误**：使用 `fmt.Errorf("描述: %w", err)` 保留错误堆栈，禁止直接返回 `err`
- **禁止 Panic**：业务逻辑中严禁 `panic`，仅限启动时致命错误
- **Sentinel Errors**：在 Domain 层定义如 `ErrUserNotFound` 用于逻辑判断

```go
// 正确
return fmt.Errorf("failed to get user: %w", err)

// 错误
return err
```

### 日志规范
- **结构化日志**：必须使用 `zap`，禁止 `fmt.Println` 或 `log.Printf`
- **日志级别**：
  - `Error`: 流程失败、需人工介入
  - `Info`: 关键业务事件
  - `Debug`: 开发调试详细数据
- **上下文**：日志必须携带 `ctx`，包含 `trace_id` 或业务 ID

### 分层架构（禁止跨层调用）
1. **Handler 层**：参数解析、校验、调用 Service、格式化响应。禁止业务逻辑
2. **Service 层**：业务逻辑、事务控制、领域规则。调用 Repository
3. **Repository 层**：仅数据库 CRUD。禁止业务逻辑

---

## Vue 3 + TypeScript 前端规范

### Composition API
- **必须使用** `<script setup lang="ts">`
- **禁止 Options API**：禁止 `data`, `methods`, `computed` 对象式写法
- **响应式变量**：
  - 基础类型用 `ref`
  - 对象用 `reactive`，解构时注意响应式丢失

### 命名规范
- **组件文件**：大驼峰 `PascalCase.vue` (如 `UserProfile.vue`)
- **Composables**：`useCamelCase.ts` (如 `useAuth.ts`)
- **事件处理**：以 `handle` 开头 (如 `handleSubmit`)
- **Props/Emits**：接口命名为 `Props` 和 `Emits`

### Pinia 状态管理
- **必须使用 Setup Store**（函数式），禁止 Option Store（对象式）
- **组件中使用**：状态用 `storeToRefs` 保持响应式，Action 直接解构

```typescript
// Store 定义
export const useUserStore = defineStore('user', () => {
  const user = ref(null)
  const isLoggedIn = computed(() => !!user.value)
  function login(data) { /* ... */ }
  return { user, isLoggedIn, login }
})

// 组件中使用
const store = useUserStore()
const { user, isLoggedIn } = storeToRefs(store)
const { login } = store
```

---

## 安全要求
- 不提交密钥、API key 或敏感配置
- 用户输入必须验证
- SQL 使用 GORM 参数化查询
- SSH 密码使用 AES 加密存储

## 常用命令

```bash
# 后端
go run cmd/server/main.go          # 启动服务
go test ./...                       # 运行测试
go build -o ai-ops.exe cmd/server/main.go

# 前端
cd web && npm run dev              # 开发服务器
cd web && npm run build            # 构建
cd web && npm run test:e2e         # E2E 测试
```

## 配置文件
- `config.yaml` - 主配置文件
- `.env` - 环境变量 (不提交)

## MCP 集成
项目同时作为 MCP Server 和 Client:
- Server: 暴露运维工具给外部 AI
- Client: 连接外部 MCP 服务扩展能力
