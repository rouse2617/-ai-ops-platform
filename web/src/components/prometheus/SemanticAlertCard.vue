<template>
  <div class="semantic-alert-card" :class="[`severity-${severity}`]">
    <div class="alert-header">
      <div class="alert-icon">
        <el-icon :size="24">
          <Warning v-if="severity === 'critical' || severity === 'high'" />
          <InfoFilled v-else />
        </el-icon>
      </div>
      <div class="alert-title">
        <h4>{{ title }}</h4>
        <el-tag :type="severityType" size="small">{{ severityText }}</el-tag>
      </div>
      <div class="confidence" v-if="confidence">
        <span>置信度: {{ (confidence * 100).toFixed(0) }}%</span>
      </div>
    </div>

    <div class="alert-body">
      <div class="summary" v-if="summary">
        <p>{{ summary }}</p>
      </div>

      <el-collapse v-model="activeCollapse">
        <el-collapse-item title="详细分析" name="description" v-if="description">
          <div class="description" v-html="renderedDescription"></div>
        </el-collapse-item>

        <el-collapse-item title="可能的根因" name="rootCause" v-if="rootCause">
          <div class="root-cause">
            <el-icon><Search /></el-icon>
            <span>{{ rootCause }}</span>
          </div>
        </el-collapse-item>

        <el-collapse-item title="相关指标" name="metrics" v-if="relatedMetrics && Object.keys(relatedMetrics).length > 0">
          <div class="metrics-grid">
            <div v-for="(value, key) in relatedMetrics" :key="key" class="metric-item">
              <span class="metric-name">{{ formatMetricName(key) }}</span>
              <span class="metric-value">{{ formatMetricValue(value) }}</span>
            </div>
          </div>
        </el-collapse-item>
      </el-collapse>

      <div class="suggestion" v-if="suggestion">
        <el-icon><Promotion /></el-icon>
        <span>{{ suggestion }}</span>
      </div>
    </div>

    <div class="alert-actions" v-if="actions && actions.length > 0">
      <el-button
        v-for="action in actions"
        :key="action.id"
        :type="action.dangerous ? 'danger' : 'primary'"
        size="small"
        @click="$emit('action', action)"
      >
        {{ action.label }}
      </el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { marked } from 'marked'
import { Warning, InfoFilled, Search, Promotion } from '@element-plus/icons-vue'

interface Props {
  title?: string
  summary?: string
  description?: string
  rootCause?: string
  suggestion?: string
  severity?: string
  confidence?: number
  relatedMetrics?: Record<string, number>
  actions?: Array<{ id: string; label: string; dangerous?: boolean }>
}

const props = withDefaults(defineProps<Props>(), {
  severity: 'medium',
  confidence: 0.8
})

defineEmits<{
  (e: 'action', action: { id: string; label: string }): void
}>()

const activeCollapse = ref(['description'])

const severityType = computed(() => {
  switch (props.severity) {
    case 'critical': return 'danger'
    case 'high': return 'warning'
    case 'medium': return 'info'
    default: return 'success'
  }
})

const severityText = computed(() => {
  switch (props.severity) {
    case 'critical': return '严重'
    case 'high': return '高'
    case 'medium': return '中'
    default: return '低'
  }
})

const renderedDescription = computed(() => {
  if (!props.description) return ''
  try {
    return marked(props.description)
  } catch {
    return props.description
  }
})

function formatMetricName(name: string): string {
  return name.replace(/_/g, ' ').replace(/\b\w/g, l => l.toUpperCase())
}

function formatMetricValue(value: number): string {
  if (value > 1000000) return (value / 1000000).toFixed(2) + 'M'
  if (value > 1000) return (value / 1000).toFixed(2) + 'K'
  return value.toFixed(2)
}
</script>

<style scoped>
.semantic-alert-card {
  background: #fff;
  border-radius: 12px;
  border-left: 4px solid #409eff;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
  padding: 16px;
  margin: 12px 0;
}

.semantic-alert-card.severity-critical {
  border-left-color: #f56c6c;
  background: linear-gradient(135deg, #fff 0%, #fef0f0 100%);
}

.semantic-alert-card.severity-high {
  border-left-color: #e6a23c;
  background: linear-gradient(135deg, #fff 0%, #fdf6ec 100%);
}

.semantic-alert-card.severity-medium {
  border-left-color: #409eff;
}

.semantic-alert-card.severity-low {
  border-left-color: #67c23a;
}

.alert-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 12px;
}

.alert-icon {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.severity-critical .alert-icon {
  background: rgba(245, 108, 108, 0.1);
  color: #f56c6c;
}

.severity-high .alert-icon {
  background: rgba(230, 162, 60, 0.1);
  color: #e6a23c;
}

.severity-medium .alert-icon {
  background: rgba(64, 158, 255, 0.1);
  color: #409eff;
}

.severity-low .alert-icon {
  background: rgba(103, 194, 58, 0.1);
  color: #67c23a;
}

.alert-title {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 8px;
}

.alert-title h4 {
  margin: 0;
  font-size: 16px;
  color: #303133;
}

.confidence {
  font-size: 12px;
  color: #909399;
}

.alert-body {
  margin-bottom: 12px;
}

.summary {
  font-size: 14px;
  color: #606266;
  line-height: 1.6;
  margin-bottom: 12px;
  padding: 12px;
  background: rgba(0, 0, 0, 0.02);
  border-radius: 8px;
}

.summary p {
  margin: 0;
}

.description {
  font-size: 14px;
  color: #606266;
  line-height: 1.6;
}

.root-cause, .suggestion {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  padding: 12px;
  border-radius: 8px;
  font-size: 14px;
}

.root-cause {
  background: rgba(64, 158, 255, 0.05);
  color: #409eff;
}

.suggestion {
  background: rgba(103, 194, 58, 0.05);
  color: #67c23a;
  margin-top: 12px;
}

.metrics-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(150px, 1fr));
  gap: 12px;
}

.metric-item {
  display: flex;
  flex-direction: column;
  padding: 8px 12px;
  background: #f5f7fa;
  border-radius: 6px;
}

.metric-name {
  font-size: 12px;
  color: #909399;
}

.metric-value {
  font-size: 16px;
  font-weight: 600;
  color: #303133;
}

.alert-actions {
  display: flex;
  gap: 8px;
  padding-top: 12px;
  border-top: 1px solid #ebeef5;
}
</style>
