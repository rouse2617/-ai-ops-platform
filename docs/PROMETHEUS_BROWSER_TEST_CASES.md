# Prometheus 集成浏览器交互测试用例

## 测试环境准备

### 前置条件
1. AI-Ops 服务运行在 http://localhost:8080
2. Prometheus 服务运行在 http://localhost:9090
3. 浏览器: Chrome/Firefox/Safari 最新版本
4. 测试数据已导入 Prometheus

### 测试账户
- 用户名: test_user
- 密码: test_password

---

## 测试用例

### TC-001: 语义翻译 - CPU 指标翻译

**目标**: 验证 CPU 指标能正确翻译为自然语言

**步骤**:
1. 打开浏览器，访问 http://localhost:8080
2. 登录系统
3. 导航到 "Prometheus" → "指标翻译"
4. 输入以下信息:
   - 指标名称: `node_cpu_seconds_total`
   - 当前值: `85.5`
   - 标签: `{"instance": "localhost:9100", "job": "node"}`
5. 点击 "翻译" 按钮

**预期结果**:
- 显示标题: "CPU 使用率异常: 85.5%"
- 显示描述: 包含 "CPU 占用率过高，系统性能可能受到影响"
- 显示根因: "可能原因: 应用程序计算密集、死循环、或系统进程异常"
- 显示影响: "影响: 系统响应缓慢，用户体验下降，可能导致服务超时"
- 显示建议列表

**验证点**:
- [ ] 翻译结果准确
- [ ] 建议合理可行
- [ ] 响应时间 < 1s

---

### TC-002: 对话式 PromQL - 自然语言查询

**目标**: 验证自然语言能正确转换为 PromQL

**步骤**:
1. 导航到 "Prometheus" → "PromQL 生成器"
2. 在查询框输入: "最近一小时内存使用率超过80%的主机有哪些?"
3. 点击 "生成 PromQL" 按钮

**预期结果**:
- 生成的 PromQL: `(1 - (node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes)) * 100 > 80`
- 显示解释: "查询内存使用率超过 80% 的主机"
- 时间范围: 1h
- Step: 1m

**验证点**:
- [ ] PromQL 语法正确
- [ ] 查询能在 Prometheus 中执行
- [ ] 返回结果正确

**测试场景**:
1. 查询 CPU: "CPU 使用率超过 75% 的主机"
2. 查询磁盘: "磁盘使用率超过 90% 的主机"
3. 查询负载: "系统负载是多少"
4. 查询网络: "网络接收速率"

---

### TC-003: 动态阈值检测 - 异常检测

**目标**: 验证动态阈值检测功能

**步骤**:
1. 导航到 "Prometheus" → "异常检测"
2. 输入以下信息:
   - 指标名称: `node_cpu_seconds_total`
   - 当前值: `92.5`
3. 点击 "检测异常" 按钮

**预期结果**:
- 显示异常检测结果
- 严重程度: 0.8 以上 (高)
- 偏差: > 50%
- 建议: "立即采取行动，异常严重程度很高"

**验证点**:
- [ ] 异常检测准确
- [ ] 严重程度计算正确
- [ ] 建议合理

**测试场景**:
1. 正常值: 45% - 应显示 "指标正常"
2. 中等异常: 75% - 应显示 "需要关注"
3. 严重异常: 95% - 应显示 "立即采取行动"

---

### TC-004: 自愈剧本 - 创建和执行

**目标**: 验证自愈剧本的创建和执行

**步骤**:
1. 导航到 "Prometheus" → "自愈剧本"
2. 点击 "创建剧本" 按钮
3. 填写以下信息:
   - 名称: "测试高 CPU 处理"
   - 触发条件: "cpu_usage > 80%"
   - 自动执行: 关闭
4. 添加步骤:
   - 步骤 1: 诊断
     - 操作: `ps aux --sort=-%cpu | head -10`
     - 操作: `top -bn1 | head -20`
   - 步骤 2: 分析 (启用 AI 分析)
   - 步骤 3: 修复
     - 操作: `kill -9 <pid>`
     - 需要批准: 是
5. 点击 "保存" 按钮

**预期结果**:
- 剧本创建成功
- 显示剧本 ID
- 可以在列表中看到新创建的剧本

**验证点**:
- [ ] 剧本创建成功
- [ ] 所有字段保存正确
- [ ] 可以编辑和删除

**执行测试**:
1. 点击 "执行" 按钮
2. 输入上下文信息 (可选)
3. 点击 "开始执行"

**预期结果**:
- 显示执行进度
- 每个步骤显示执行状态
- 显示执行输出
- 最终显示执行结果

**验证点**:
- [ ] 剧本执行成功
- [ ] 步骤按顺序执行
- [ ] 输出正确显示
- [ ] 需要批准的步骤暂停

---

### TC-005: 动态看板 - 生成和查看

**目标**: 验证动态看板生成功能

**步骤**:
1. 导航到 "Prometheus" → "动态看板"
2. 点击 "生成看板" 按钮
3. 选择模板: "系统概览"
4. 输入看板名称: "我的系统看板"
5. 点击 "生成" 按钮

**预期结果**:
- 看板生成成功
- 显示看板 ID
- 自动跳转到看板详情页面
- 看板包含以下面板:
  - 主机在线状态
  - CPU 使用率
  - 内存使用率
  - 磁盘使用率

**验证点**:
- [ ] 看板生成成功
- [ ] 所有面板正确显示
- [ ] 数据实时更新
- [ ] 响应时间 < 2s

**看板交互测试**:
1. 点击面板标题，查看详细信息
2. 调整时间范围 (1h, 6h, 24h, 7d)
3. 刷新看板数据
4. 导出看板配置

**预期结果**:
- [ ] 面板详情正确显示
- [ ] 时间范围切换正常
- [ ] 数据随时间范围更新
- [ ] 导��配置成功

---

### TC-006: 推荐面板 - 个性化推荐

**目标**: 验证推荐面板功能

**步骤**:
1. 导航到 "Prometheus" → "推荐面板"
2. 系统自动推荐面板

**预期结果**:
- 显示推荐的面板列表
- 包含系统概览、应用性能、业务指标等
- 每个面板显示标题、类型、查询

**验证点**:
- [ ] 推荐面板合理
- [ ] 面板类型多样
- [ ] 可以添加到看板

**添加推荐面板**:
1. 点击推荐面板的 "添加到看板" 按钮
2. 选择目标看板
3. 点击 "确认"

**预期结果**:
- 面板成功添加到看板
- 看板自动刷新
- 新面板显示在看板中

**验证点**:
- [ ] 面板添加成功
- [ ] 面板数据正确显示

---

### TC-007: 告警规则管理

**目标**: 验证告警规则的创建和管理

**步骤**:
1. 导航�� "Prometheus" → "告警规则"
2. 点击 "创建规则" 按钮
3. 填写以下信息:
   - 规则名称: "高 CPU 告警"
   - 规则类型: CPU
   - 告警级别: High
   - 条件: `{"threshold": 80}`
   - 启用: 是
4. 点击 "保存" 按钮

**预期结果**:
- 规则创建成功
- 显示规则 ID
- 可以在列表中看到新规则

**验证点**:
- [ ] 规则创建成功
- [ ] 规则参数保存正确
- [ ] 规则立即生效

**规则测试**:
1. 模拟 CPU 使用率超过 80%
2. 观察是否触发告警

**预期结果**:
- [ ] 告警被触发
- [ ] 告警信息正确
- [ ] 告警通知发送

---

### TC-008: 性能测试

**目标**: 验证系统性能指标

**测试场景**:

#### 查询响应时间
1. 执行 100 次 PromQL 查询
2. 记录平均响应时间

**预期结果**:
- 平均响应时间 < 500ms
- 99 分位数 < 1s

#### 语义翻译延迟
1. 执行 50 次语义翻译
2. 记录平均延迟

**预期结果**:
- 平均延迟 < 1s
- 99 分位数 < 2s

#### 看板加载时间
1. 加载包含 10 个面板的看板
2. 记录加载时间

**预期结果**:
- 加载时间 < 2s
- 所有面板数据完整

#### 并发用户测试
1. 模拟 10 个并发用户
2. 每个用户执行 10 次操作
3. 记录成功率和响应时间

**预期结果**:
- 成功率 > 99%
- 平均响应时间 < 1s

---

### TC-009: 错误处理

**目标**: 验证系统错误处理

**测试场景**:

#### Prometheus 连接失败
1. 停止 Prometheus 服务
2. 尝试执行查询

**预期结果**:
- 显示友好的错误提示
- 建议用户检查 Prometheus 连接

#### 无效的 PromQL
1. 输入无效的 PromQL: `invalid query`
2. 点击 "执行"

**预期结果**:
- 显示错误信息
- 提示 PromQL 语法错误

#### 超时处理
1. 执行耗时的查询
2. 等待超时

**预期结果**:
- 显示超时错误
- 提供重试选项

---

### TC-010: 数据安全

**目标**: 验证数据安全性

**测试场景**:

#### 认证检查
1. 不登录直接访问 API
2. 验证是否返回 401 错误

**预期结果**:
- [ ] 返回 401 Unauthorized
- [ ] 重定向到登录页面

#### 授权检查
1. 使用普通用户账户
2. 尝试访问管理员功能

**预期结果**:
- [ ] 返回 403 Forbidden
- [ ] 显示权限不足提示

#### 敏感数据脱敏
1. 查看告警详情
2. 验证敏感信息是否脱敏

**预期结果**:
- [ ] 密码等敏感信息不显示
- [ ] 关键信息正确显示

---

## 测试报告模板

### 测试执行记录

| 测试用例 | 执行时间 | 结果 | 备注 |
|---------|---------|------|------|
| TC-001 | 2026-01-03 | PASS | - |
| TC-002 | 2026-01-03 | PASS | - |
| TC-003 | 2026-01-03 | PASS | - |
| TC-004 | 2026-01-03 | PASS | - |
| TC-005 | 2026-01-03 | PASS | - |
| TC-006 | 2026-01-03 | PASS | - |
| TC-007 | 2026-01-03 | PASS | - |
| TC-008 | 2026-01-03 | PASS | - |
| TC-009 | 2026-01-03 | PASS | - |
| TC-010 | 2026-01-03 | PASS | - |

### 缺陷记录

| ID | 标题 | 严重程度 | 状态 |
|----|------|---------|------|
| BUG-001 | 示例缺陷 | High | Open |

### 测试总结

- 总测试用例数: 10
- 通过: 10
- 失败: 0
- 通过率: 100%
- 测试周期: 1 天

---

## 自动化测试脚本

### Cypress 测试示例

```javascript
describe('Prometheus Integration Tests', () => {
  beforeEach(() => {
    cy.visit('http://localhost:8080');
    cy.login('test_user', 'test_password');
  });

  it('TC-001: Should translate CPU metric to natural language', () => {
    cy.visit('/prometheus/translate');
    cy.get('[data-testid="metric-name"]').type('node_cpu_seconds_total');
    cy.get('[data-testid="metric-value"]').type('85.5');
    cy.get('[data-testid="translate-btn"]').click();

    cy.get('[data-testid="translation-title"]')
      .should('contain', 'CPU 使用率异常');
    cy.get('[data-testid="translation-description"]')
      .should('contain', 'CPU 占用率过高');
  });

  it('TC-002: Should generate PromQL from natural language', () => {
    cy.visit('/prometheus/promql-generator');
    cy.get('[data-testid="query-input"]')
      .type('最近一小时内存使用率超过80%的主机有哪些?');
    cy.get('[data-testid="generate-btn"]').click();

    cy.get('[data-testid="promql-output"]')
      .should('contain', 'node_memory_MemAvailable_bytes');
  });

  it('TC-005: Should generate dynamic dashboard', () => {
    cy.visit('/prometheus/dashboards');
    cy.get('[data-testid="create-dashboard-btn"]').click();
    cy.get('[data-testid="dashboard-name"]').type('我的系统看板');
    cy.get('[data-testid="template-select"]').select('system_overview');
    cy.get('[data-testid="generate-btn"]').click();

    cy.get('[data-testid="dashboard-title"]')
      .should('contain', '我的系统看板');
    cy.get('[data-testid="panel"]').should('have.length', 4);
  });
});
```

---

## 测试环境配置

### Docker Compose 配置

```yaml
version: '3.8'
services:
  prometheus:
    image: prom/prometheus:latest
    ports:
      - "9090:9090"
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'

  ai-ops:
    build: .
    ports:
      - "8080:8080"
    environment:
      - PROMETHEUS_URL=http://prometheus:9090
    depends_on:
      - prometheus
```

