<template>
  <div class="contextual-actions-panel">
    <div class="panel-header">
      <el-icon><Operation /></el-icon>
      <span class="header-title">智能操作</span>
      <el-badge v-if="aiRecommendedCount > 0" :value="aiRecommendedCount" type="primary" />
    </div>

    <div class="actions-content">
      <!-- AI 推荐操作 -->
      <div class="action-group" v-if="aiRecommendedActions.length > 0">
        <div class="group-label">
          <el-icon class="ai-icon"><MagicStick /></el-icon>
          <span>AI 推荐</span>
        </div>
        <div class="actions-list">
          <div
            v-for="action in aiRecommendedActions"
            :key="action.id"
            class="action-item ai-recommended"
            :class="{ highlight: action.highlight, [`risk-${action.risk_level}`]: true }"
            @click="handleAction(action)"
          >
            <div class="action-main">
              <el-icon class="action-icon">
                <component :is="getIcon(action.icon)" />
              </el-icon>
              <span class="action-label">{{ action.label }}</span>
            </div>
            <div class="action-meta">
              <el-tag
                v-if="action.confidence"
                size="small"
                type="primary"
                effect="plain"
                class="confidence-tag"
              >
                {{ Math.round(action.confidence * 100) }}% 置信度
              </el-tag>
              <el-tag
                size="small"
                :type="getRiskTagType(action.risk_level)"
                effect="dark"
              >
                {{ getRiskText(action.risk_level) }}
              </el-tag>
            </div>
            <div class="action-description" v-if="action.description">
              {{ action.description }}
            </div>
          </div>
        </div>
      </div>

      <!-- 常规操作 -->
      <div class="action-group">
        <div class="group-label">
          <el-icon><Tools /></el-icon>
          <span>快速操作</span>
        </div>
        <div class="actions-grid">
          <el-button
            v-for="action in regularActions"
            :key="action.id"
            :type="getButtonType(action)"
            :icon="getIcon(action.icon)"
            :loading="loadingActions.has(action.id)"
            :disabled="loadingActions.size > 0"
            @click="handleAction(action)"
            class="action-button"
          >
            {{ action.label }}
          </el-button>
        </div>
      </div>
    </div>

    <!-- 加载状态 -->
    <div class="loading-overlay" v-if="loading">
      <el-icon class="is-loading"><Loading /></el-icon>
      <span>分析中...</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Operation, MagicStick, Tools, Loading,
  Delete, Document, Monitor, Refresh, Plus,
  Setting, Search, Download, Upload, Warning
} from '@element-plus/icons-vue'
import { getSuggestedActions, type ContextualAction } from '@/api/analysis'

interface Props {
  hostId?: string
  toolResults?: Array<{
    tool_name: string
    result: unknown
    host_id?: string
  }>
}

const props = defineProps<Props>()

const emit = defineEmits<{
  execute: [command: string, action: ContextualAction]
}>()

const loading = ref(false)
const actions = ref<ContextualAction[]>([])
const loadingActions = ref<Set<string>>(new Set())

// 图标映射
const iconMap: Record<string, any> = {
  Delete, Document, Monitor, Refresh, Plus,
  Setting, Search, Download, Upload, Warning
}

const getIcon = (iconName: string) => {
  return iconMap[iconName] || Tools
}

// AI 推荐的操作
const aiRecommendedActions = computed(() => {
  return actions.value.filter(a => a.ai_recommended)
})

// 常规操作
const regularActions = computed(() => {
  return actions.value.filter(a => !a.ai_recommended)
})

// AI 推荐数量
const aiRecommendedCount = computed(() => aiRecommendedActions.value.length)

// 获取推荐操作
const fetchSuggestedActions = async () => {
  if (!props.toolResults?.length) {
    // 使用默认操作
    actions.value = getDefaultActions()
    return
  }

  loading.value = true
  try {
    const result = await getSuggestedActions({
      tool_results: props.toolResults
    })
    // 确保返回值是数组
    actions.value = Array.isArray(result) ? result : getDefaultActions()
  } catch (error) {
    console.error('获取推荐操作失败:', error)
    actions.value = getDefaultActions()
  } finally {
    loading.value = false
  }
}

// 默认操作列表
const getDefaultActions = (): ContextualAction[] => [
  { id: 'view_logs', label: '查看日志', icon: 'Document', highlight: false, ai_recommended: false, risk_level: 'low' },
  { id: 'check_status', label: '检查状态', icon: 'Monitor', highlight: false, ai_recommended: false, risk_level: 'low' },
  { id: 'restart_service', label: '重启服务', icon: 'Refresh', highlight: false, ai_recommended: false, risk_level: 'medium' },
  { id: 'scale_up', label: '扩容', icon: 'Plus', highlight: false, ai_recommended: false, risk_level: 'medium' }
]

// 监听工具结果变化
watch(() => props.toolResults, () => {
  fetchSuggestedActions()
}, { deep: true, immediate: true })

// 风险等级标签类型
const getRiskTagType = (level: string) => {
  switch (level) {
    case 'high': return 'danger'
    case 'medium': return 'warning'
    default: return 'success'
  }
}

// 风险等级文本
const getRiskText = (level: string) => {
  switch (level) {
    case 'high': return '高风险'
    case 'medium': return '中风险'
    default: return '低风险'
  }
}

// 按钮类型
const getButtonType = (action: ContextualAction) => {
  if (action.highlight) return 'primary'
  switch (action.risk_level) {
    case 'high': return 'danger'
    case 'medium': return 'warning'
    default: return 'default'
  }
}

// 处理操作点击
const handleAction = async (action: ContextualAction) => {
  // 高风险操作需要确认
  if (action.risk_level === 'high' || action.risk_level === 'medium') {
    try {
      await ElMessageBox.confirm(
        `确定要执行「${action.label}」操作吗？${action.description || ''}`,
        '操作确认',
        {
          confirmButtonText: '确定执行',
          cancelButtonText: '取消',
          type: action.risk_level === 'high' ? 'error' : 'warning'
        }
      )
    } catch {
      return
    }
  }

  loadingActions.value.add(action.id)

  try {
    emit('execute', action.command || action.label, action)
    ElMessage.success(`正在执行: ${action.label}`)
  } finally {
    setTimeout(() => {
      loadingActions.value.delete(action.id)
    }, 2000)
  }
}

onMounted(() => {
  fetchSuggestedActions()
})
</script>

<style scoped>
.contextual-actions-panel {
  background: #fff;
  border-radius: 12px;
  overflow: hidden;
  position: relative;
}

.panel-header {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 16px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: #fff;
}

.header-title {
  flex: 1;
  font-weight: 600;
  font-size: 14px;
}

.actions-content {
  padding: 12px;
}

.action-group {
  margin-bottom: 16px;
}

.action-group:last-child {
  margin-bottom: 0;
}

.group-label {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  font-weight: 600;
  color: #6b7280;
  margin-bottom: 10px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.ai-icon {
  color: #8b5cf6;
}

.actions-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.action-item {
  padding: 12px;
  border-radius: 8px;
  border: 1px solid #e5e7eb;
  cursor: pointer;
  transition: all 0.2s;
  background: #fff;
}

.action-item:hover {
  border-color: #3b82f6;
  box-shadow: 0 2px 8px rgba(59, 130, 246, 0.15);
}

.action-item.ai-recommended {
  border-color: #8b5cf6;
  background: linear-gradient(135deg, #faf5ff 0%, #f3e8ff 100%);
}

.action-item.ai-recommended:hover {
  box-shadow: 0 4px 12px rgba(139, 92, 246, 0.2);
}

.action-item.highlight {
  animation: pulse-border 2s infinite;
}

@keyframes pulse-border {
  0%, 100% { border-color: #8b5cf6; }
  50% { border-color: #c4b5fd; }
}

.action-item.risk-high {
  border-left: 3px solid #ef4444;
}

.action-item.risk-medium {
  border-left: 3px solid #f59e0b;
}

.action-item.risk-low {
  border-left: 3px solid #10b981;
}

.action-main {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 6px;
}

.action-icon {
  color: #6b7280;
}

.ai-recommended .action-icon {
  color: #8b5cf6;
}

.action-label {
  font-weight: 500;
  color: #1f2937;
  font-size: 14px;
}

.action-meta {
  display: flex;
  gap: 6px;
  margin-bottom: 4px;
}

.confidence-tag {
  font-size: 11px;
}

.action-description {
  font-size: 12px;
  color: #6b7280;
  line-height: 1.4;
}

.actions-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 8px;
}

.action-button {
  width: 100%;
  justify-content: flex-start;
}

.loading-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(255, 255, 255, 0.9);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 8px;
  color: #6b7280;
  font-size: 13px;
}
</style>
