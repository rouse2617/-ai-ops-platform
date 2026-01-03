# Slot Filling (参数补全) 功能设计文档

## 概述

Slot Filling 是一个智能参数补全功能，当用户输入不完整的命令时（如"重启服务"但未指定服务名），系统会自动检测缺失的参数，并在输入框下方弹出选择列表，用户可通过点击或键盘快捷键快速补全参数。

## 功能特性

### 1. 智能检测
- 实时分析用户输入的命令
- 识别常见运维操作模式（服务、文件、进程、端口等）
- 检测缺失的必需参数
- 防抖处理（800ms），避免频繁请求

### 2. 交互设计
- **视觉层次**：渐变色头部 + 清晰的选项列表
- **键盘导航**：↑↓ 选择、Enter 确认、Esc 取消
- **鼠标交互**：悬停高亮、点击选择
- **动画效果**：平滑的滑入动画（slideUp）

### 3. 参数类型支持

#### 服务类型 (service)
触发关键词：
- 重启服务、启动服务、停止服务、查看服务、服务状态
- restart service, start service, stop service

建议选项：
- Nginx (Web 服务器)
- Apache (HTTP 服务器)
- MySQL (数据库服务)
- Redis (缓存服务)
- Docker (容器服务)
- SSH (远程连接服务)
- PostgreSQL (数据库服务)
- MongoDB (文档数据库)

#### 文件类型 (file)
触发关键词：
- 查看文件、编辑文件、删除文件、查看日志
- tail, cat, vim

建议选项：
- /var/log/syslog (系统日志)
- /var/log/nginx/access.log (Nginx 访问日志)
- /var/log/nginx/error.log (Nginx 错误日志)
- /var/log/mysql/error.log (MySQL 错误日志)
- /etc/nginx/nginx.conf (Nginx 配置)
- /etc/mysql/my.cnf (MySQL 配置)

#### 进程类型 (process)
触发关键词：
- 杀死进程、停止进程、查看进程
- kill

建议选项：
- nginx (Web 服务器进程)
- java (Java 应用进程)
- python (Python 应用进程)
- node (Node.js 应用进程)
- mysqld (MySQL 数据库进程)

#### 端口类型 (port)
触发关键词：
- 查看端口、检查端口、端口占用
- netstat, lsof

建议选项：
- 80 (HTTP 端口)
- 443 (HTTPS 端口)
- 3306 (MySQL 端口)
- 6379 (Redis 端口)
- 8080 (应用端口)
- 22 (SSH 端口)

## 技术实现

### 前端架构

#### 1. 组件结构
```
SlotFillingDropdown.vue (新增)
├── 下拉框容器
├── 头部区域（渐变背景 + 提示文本）
├── 选项列表（可滚动）
│   ├── 图标
│   ├── 标签 + 描述
│   └── 标签（可选）
└── 底部提示（键盘快捷键）

InputBox.vue (修改)
├── 集成 SlotFillingDropdown
├── 添加防抖分析逻辑
├── 键盘导航支持
└── 状态管理
```

#### 2. 核心文件

**C:\Users\hrp\Downloads\ai-pro\web\src\components\chat\SlotFillingDropdown.vue**
- ���数补全下拉框组件
- 支持键盘导航（↑↓ Enter Esc）
- 动态头部文本根据参数类型变化
- 平滑动画效果

**C:\Users\hrp\Downloads\ai-pro\web\src\components\chat\InputBox.vue**
- 集成 SlotFillingDropdown 组件
- 添加 `analyzeMessageDebounced` 防抖函数
- 扩展键盘事件处理逻辑
- 管理 slot filling 状态

**C:\Users\hrp\Downloads\ai-pro\web\src\api\slotFilling.ts**
- API 接口定义
- TypeScript 类型声明
- 请求/响应数据结构

#### 3. 状态管理
```typescript
// Slot filling state
const showSlotFilling = ref(false)           // 是否显示下拉框
const slotSuggestions = ref<SlotSuggestion[]>([])  // 建议选项列表
const slotType = ref('')                     // 参数类型
const slotFillingRef = ref<any>()            // 组件引用
```

#### 4. 交互流程
```
用户输入 → 防抖分析 (800ms) → 后端检测 → 返回建议 → 显示下拉框
                                                    ↓
用户选择 ← 键盘/鼠标 ← 更新输入框 ← 补全参数 ← 关闭下拉框
```

### 后端架构

#### 1. 核心文件

**C:\Users\hrp\Downloads\ai-pro\internal\api\handler\slot_filling.go**
- SlotFillingHandler 处理器
- AnalyzeMessage 分析接口
- 参数检测逻辑
- 建议生成函数

#### 2. 检测算法
```go
// 检测流程
1. 标准化输入（去空格、转小写）
2. 关键词匹配（服务/文件/进程/端口）
3. 参数完整性检查
4. 返回建议列表
```

#### 3. API 路由
```
POST /api/slot-filling/analyze
```

请求体：
```json
{
  "message": "重启服务",
  "hosts": ["host1", "host2"]
}
```

响应体：
```json
{
  "needsSlotFilling": true,
  "slotType": "service",
  "suggestions": [
    {
      "value": "nginx",
      "label": "Nginx",
      "description": "Web 服务器",
      "icon": "Service",
      "iconColor": "#67C23A",
      "tag": "常用",
      "tagType": "success"
    }
  ],
  "originalMessage": "重启服务"
}
```

## 设计规范

### 视觉设计

#### 1. 颜色系统
```css
/* 头部渐变 */
background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);

/* 激活状态 */
background: #ecf5ff;
border-left: 3px solid #409eff;

/* 悬停状态 */
background: #f5f7fa;

/* 图标颜色 */
--success: #67C23A
--info: #409EFF
--warning: #E6A23C
--danger: #F56C6C
--secondary: #909399
```

#### 2. 间距系统
```css
/* 内边距 */
--padding-header: 12px 16px
--padding-item: 10px 16px
--padding-footer: 8px 16px

/* 外边距 */
--gap-content: 12px
--gap-text: 2px
```

#### 3. 字体规范
```css
/* 头部文本 */
font-size: 14px
font-weight: 500

/* 选项标签 */
font-size: 14px
font-weight: 500
color: #303133

/* 选项描述 */
font-size: 12px
color: #909399

/* 底部提示 */
font-size: 11px
color: #909399
```

#### 4. 动��效果
```css
/* 滑入动画 */
@keyframes slideUp {
  from {
    opacity: 0;
    transform: translateY(8px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

/* 过渡效果 */
transition: all 0.2s ease;
```

### 交互规范

#### 1. 键盘快捷键
- `↑` - 向上选择
- `↓` - 向下选择
- `Enter` - 确认选择
- `Esc` - 取消/关闭

#### 2. 鼠标交互
- 悬停 - 高亮选项
- 点击 - 选择并补全
- 外部点击 - 关闭下拉框

#### 3. 触摸交互
- 滚动 - 浏览选项
- 点击 - 选择并补全

## 可访问性 (Accessibility)

### ARIA 属性
```html
<!-- 下拉框 -->
role="listbox"
aria-label="Parameter suggestions"

<!-- 选项 -->
role="option"
:aria-selected="index === activeIndex"
tabindex="-1"
```

### 键盘导航
- 完整的键盘支持
- 焦点管理
- 视觉反馈

### 屏幕阅读器
- 语义化 HTML
- ARIA 标签
- 状态通知

## 性能优化

### 1. 防抖处理
```typescript
const analyzeMessageDebounced = debounce(async (message: string) => {
  // 分析逻辑
}, 800)
```

### 2. 条件渲染
```vue
<SlotFillingDropdown
  v-if="showSlotFilling"
  :visible="showSlotFilling"
/>
```

### 3. 虚拟滚动
- 当选项超过 50 个时考虑实现
- 使用 `vue-virtual-scroller`

## 扩展性

### 1. 新增参数类型
```go
// 1. 添加检测函数
func (h *SlotFillingHandler) isNewTypeCommand(message string) bool {
  // 检测逻辑
}

// 2. 添加建议函数
func (h *SlotFillingHandler) getNewTypeSuggestions() []SlotSuggestion {
  // 返回建议
}

// 3. 在 AnalyzeMessage 中调用
if h.isNewTypeCommand(message) {
  Success(c, AnalyzeResponse{
    NeedsSlotFilling: true,
    SlotType:         "newType",
    Suggestions:      h.getNewTypeSuggestions(),
  })
  return
}
```

### 2. 动态建议
- 从数据库加载常用服务
- 根据历史记录排序
- 个性化推荐

### 3. 智能学习
- 记录用户选择
- 优化建议排序
- 自适应关键词

## 测试用例

### 功能测试
```
1. 输入"重启服务" → 显示服务列表
2. 输入"查看日志" → 显示日志文件列表
3. 输入"杀死进程" → 显示进程列表
4. 输入"查看端口" → 显示端口列表
5. 输入完整命令 → 不显示下拉框
```

### 交互测试
```
1. 键盘导航 ↑↓ → 高亮切换
2. Enter 键 → 选择并补全
3. Esc 键 → 关闭下拉框
4. 鼠标悬停 → 高亮选项
5. 点击选项 → 选择并补全
6. 外部点击 → 关闭下拉框
```

### 性能测试
```
1. 防抖延迟 800ms → 避免频繁请求
2. 大量选项 → 滚动流畅
3. 快速输入 → 响应及时
```

## 使用示例

### 示例 1：重启服务
```
用户输入：重启服务
系统响应：显示服务列表（Nginx, MySQL, Redis...）
用户选择：Nginx
最终命令：重启服务 nginx
```

### 示例 2：查看日志
```
用户输入：查看日志
系统响应：显示日志文件列表
用户选择：/var/log/nginx/error.log
最终命令：查看日志 /var/log/nginx/error.log
```

### 示例 3：检查端口
```
用户输入：检查端口
系统响应：显示常用端口列表
用户选择：80
最终命令：检查端口 80
```

## 未来优化

### 短期（1-2 周）
- [ ] 添加更多参数类型（用户、目录、配置项）
- [ ] 支持多参数补全
- [ ] 优化检测算法准确率

### 中期（1-2 月）
- [ ] 集成 LLM 进行智能分析
- [ ] 从实际主机动态获取服务列表
- [ ] 添加参数验证和提示

### 长期（3-6 月）
- [ ] 机器学习优化建议排序
- [ ] 上下文感知的智能补全
- [ ] 多语言支持（英文、中文）

## 总结

Slot Filling 功能通过智能检测和友好的交互设计，显著提升了用户输入命令的效率和体验。该功能具有良好的扩展性和可维护性，为后续的智能化运维操作奠定了基础。

---

**文档版本**: 1.0
**创建日期**: 2026-01-03
**作者**: AI-Ops Team
