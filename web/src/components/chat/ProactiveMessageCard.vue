<template>
  <div class="proactive-message" :class="[`priority-${priority}`, `type-${type}`]">
    <div class="proactive-header">
      <div class="proactive-icon">
        <el-icon :size="20">
          <component :is="iconComponent" />
        </el-icon>
      </div>
      <div class="proactive-title">
        <span class="title-text">{{ title }}</span>
        <el-tag :type="priorityTagType" size="small" effect="plain">
          {{ priorityLabel }}
        </el-tag>
      </div>
      <div class="proactive-badge">
        <span class="monitoring-badge">
          <span class="pulse"></span>
          监控中
        </span>
      </div>
    </div>

    <div class="proactive-content">
      <div class="content-text" v-html="renderedContent"></div>

      <div v-if="hosts && hosts.length > 0" class="hosts-list">
        <el-tag v-for="host in hosts" :key="host" size="small" type="info">
          {{ host }}
        </el-tag>
      </div>

      <div v-if="metrics && Object.keys(metrics).length > 0" class="metrics-summary">
        <div v-for="(value, key) in metrics" :key="key" class="metric-item">
          <span class="metric-label">{{ key }}</span>
          <span class="metric-value">{{ formatMetricValue(value) }}</span>
        </div>
      </div>
    </div>

    <div v-if="actions && actions.length > 0" class="proactive-actions">
      <el-button
        v-for="action in actions"
        :key="action.id"
        :type="action.dangerous ? 'danger' : 'primary'"
        size="small"
        @click="handleAction(action)"
      >
        {{ action.label }}
      </el-button>
    </div>

    <div class="proactive-footer">
      <span class="timestamp">{{ formattedTime }}</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { marked } from 'marked'
import {
  Sunny,
  Warning,
  Monitor,
  Tools,
  Bell
} from '@element-plus/icons-vue'
import type { QuickAction } from '@/api/chat'

interface Props {
  type: string
  title: string
  content: string
  hosts?: string[]
  metrics?: Record<string, unknown>
  actions?: QuickAction[]
  priority: string
  timestamp?: number
}

const props = withDefaults(defineProps<Props>(), {
  priority: 'medium',
  hosts: () => [],
  actions: () => []
})

const emit = defineEmits<{
  (e: 'action', action: QuickAction): void
}>()

const iconComponent = computed(() => {
  switch (props.type) {
    case 'morning_report': return Sunny
    case 'anomaly_alert': return Warning
    case 'health_summary': return Monitor
    case 'auto_fix': return Tools
    default: return Bell
  }
})

const priorityTagType = computed(() => {
  switch (props.priority) {
    case 'critical': return 'danger'
    case 'high': return 'warning'
    case 'medium': return 'info'
    default: return 'success'
  }
})

const priorityLabel = computed(() => {
  switch (props.priority) {
    case 'critical': return '紧急'
    case 'high': return '重要'
    case 'medium': return '一般'
    default: return '低'
  }
})

const renderedContent = computed(() => {
  try {
    return marked(props.content)
  } catch {
    return props.content
  }
})

const formattedTime = computed(() => {
  if (!props.timestamp) return ''
  const date = new Date(props.timestamp * 1000)
  return date.toLocaleString('zh-CN', {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
})

function formatMetricValue(value: unknown): string {
  if (typeof value === 'number') {
    return value.toFixed(1)
  }
  return String(value)
}

function handleAction(action: QuickAction) {
  emit('action', action)
}
</script>

<style scoped>
.proactive-message {
  background: linear-gradient(135deg, #f0f9ff 0%, #e0f2fe 100%);
  border: 1px solid #bae6fd;
  border-radius: 12px;
  padding: 16px;
  margin: 12px 0;
  max-width: 600px;
}

.proactive-message.priority-critical {
  background: linear-gradient(135deg, #fef2f2 0%, #fee2e2 100%);
  border-color: #fca5a5;
}

.proactive-message.priority-high {
  background: linear-gradient(135deg, #fffbeb 0%, #fef3c7 100%);
  border-color: #fcd34d;
}

.proactive-message.type-morning_report {
  background: linear-gradient(135deg, #fefce8 0%, #fef9c3 100%);
  border-color: #fde047;
}

.proactive-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 12px;
}

.proactive-icon {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background: rgba(59, 130, 246, 0.1);
  display: flex;
  align-items: center;
  justify-content: center;
  color: #3b82f6;
}

.priority-critical .proactive-icon {
  background: rgba(239, 68, 68, 0.1);
  color: #ef4444;
}

.priority-high .proactive-icon {
  background: rgba(245, 158, 11, 0.1);
  color: #f59e0b;
}

.type-morning_report .proactive-icon {
  background: rgba(234, 179, 8, 0.1);
  color: #eab308;
}

.proactive-title {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 8px;
}

.title-text {
  font-weight: 600;
  font-size: 15px;
  color: #1e293b;
}

.proactive-badge {
  display: flex;
  align-items: center;
}

.monitoring-badge {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: #10b981;
  background: rgba(16, 185, 129, 0.1);
  padding: 4px 10px;
  border-radius: 12px;
}

.pulse {
  width: 8px;
  height: 8px;
  background: #10b981;
  border-radius: 50%;
  animation: pulse 2s infinite;
}

@keyframes pulse {
  0% {
    box-shadow: 0 0 0 0 rgba(16, 185, 129, 0.4);
  }
  70% {
    box-shadow: 0 0 0 6px rgba(16, 185, 129, 0);
  }
  100% {
    box-shadow: 0 0 0 0 rgba(16, 185, 129, 0);
  }
}

.proactive-content {
  margin-bottom: 12px;
}

.content-text {
  font-size: 14px;
  line-height: 1.6;
  color: #475569;
}

.content-text :deep(p) {
  margin: 0 0 8px 0;
}

.content-text :deep(p:last-child) {
  margin-bottom: 0;
}

.hosts-list {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-top: 10px;
}

.metrics-summary {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(100px, 1fr));
  gap: 8px;
  margin-top: 12px;
  padding: 10px;
  background: rgba(255, 255, 255, 0.6);
  border-radius: 8px;
}

.metric-item {
  display: flex;
  flex-direction: column;
  align-items: center;
}

.metric-label {
  font-size: 11px;
  color: #64748b;
  text-transform: uppercase;
}

.metric-value {
  font-size: 16px;
  font-weight: 600;
  color: #1e293b;
}

.proactive-actions {
  display: flex;
  gap: 8px;
  margin-bottom: 12px;
}

.proactive-footer {
  display: flex;
  justify-content: flex-end;
}

.timestamp {
  font-size: 12px;
  color: #94a3b8;
}
</style>
