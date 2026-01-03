<template>
  <div class="alert-panel">
    <div class="alert-header">
      <h3>实时告警 ({{ activeCount }})</h3>
      <el-button size="small" @click="clearAll" :disabled="activeCount === 0">
        清空
      </el-button>
    </div>

    <div class="connection-status">
      <el-tag :type="isConnected ? 'success' : 'danger'" size="small">
        {{ isConnected ? '已连接' : '未连接' }}
      </el-tag>
    </div>

    <div class="alert-list">
      <div
        v-for="alert in activeAlerts"
        :key="alert.id"
        :class="['alert-item', `alert-${alert.level}`]"
      >
        <div class="alert-content">
          <div class="alert-title">
            <el-icon class="alert-icon">
              <Warning v-if="alert.level === 'critical' || alert.level === 'high'" />
              <InfoFilled v-else />
            </el-icon>
            <span>{{ alert.title }}</span>
            <el-tag :type="getLevelType(alert.level)" size="small">
              {{ alert.level }}
            </el-tag>
          </div>

          <div class="alert-message">{{ alert.message }}</div>

          <div class="alert-metrics" v-if="alert.metrics">
            <el-tag
              v-for="(value, key) in alert.metrics"
              :key="key"
              size="small"
              class="metric-tag"
            >
              {{ key }}: {{ formatMetric(value) }}
            </el-tag>
          </div>

          <div class="alert-suggestions" v-if="alert.suggestions?.length">
            <div class="suggestions-title">建议措施：</div>
            <ul>
              <li v-for="(suggestion, idx) in alert.suggestions" :key="idx">
                {{ suggestion }}
              </li>
            </ul>
          </div>

          <div class="alert-actions" v-if="alert.actions?.length">
            <el-button
              v-for="action in alert.actions"
              :key="action.id"
              size="small"
              :type="action.dangerous ? 'danger' : 'primary'"
              @click="executeAction(alert, action)"
            >
              {{ action.label }}
            </el-button>
          </div>

          <div class="alert-footer">
            <span class="alert-time">{{ formatTime(alert.created_at) }}</span>
            <div class="alert-controls">
              <el-button size="small" text @click="resolveAlert(alert.id)">
                标记已解决
              </el-button>
              <el-button size="small" text @click="ignoreAlert(alert.id)">
                忽略
              </el-button>
            </div>
          </div>
        </div>
      </div>

      <el-empty v-if="activeCount === 0" description="暂无告警" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { Warning, InfoFilled } from '@element-plus/icons-vue'
import { useMonitorAlerts } from '@/composables/useMonitorAlerts'
import { ElMessage } from 'element-plus'

const {
  alerts,
  isConnected,
  resolveAlert: resolve,
  ignoreAlert: ignore,
  clearAlerts,
  getActiveAlerts
} = useMonitorAlerts()

const activeAlerts = computed(() => getActiveAlerts())
const activeCount = computed(() => activeAlerts.value.length)

const getLevelType = (level: string) => {
  const map: Record<string, any> = {
    critical: 'danger',
    high: 'warning',
    medium: 'warning',
    low: 'info'
  }
  return map[level] || 'info'
}

const formatMetric = (value: any) => {
  if (typeof value === 'number') {
    return value.toFixed(1) + '%'
  }
  return value
}

const formatTime = (time: string) => {
  const date = new Date(time)
  const now = new Date()
  const diff = now.getTime() - date.getTime()

  if (diff < 60000) return '刚刚'
  if (diff < 3600000) return Math.floor(diff / 60000) + '分钟前'
  if (diff < 86400000) return Math.floor(diff / 3600000) + '小时前'
  return date.toLocaleString()
}

const executeAction = (alert: any, action: any) => {
  // 触发命令执行
  window.dispatchEvent(new CustomEvent('execute-quick-action', {
    detail: {
      host: alert.host_name,
      command: action.command,
      label: action.label
    }
  }))
  ElMessage.success(`正在执行: ${action.label}`)
}

const resolveAlert = (id: string) => {
  resolve(id)
  ElMessage.success('告警已标记为已解决')
}

const ignoreAlert = (id: string) => {
  ignore(id)
  ElMessage.info('告警已忽略')
}

const clearAll = () => {
  clearAlerts()
  ElMessage.success('已清空所有告警')
}
</script>

<style scoped>
.alert-panel {
  height: 100%;
  display: flex;
  flex-direction: column;
  background: var(--el-bg-color);
}

.alert-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px;
  border-bottom: 1px solid var(--el-border-color);
}

.alert-header h3 {
  margin: 0;
  font-size: 16px;
}

.connection-status {
  padding: 8px 16px;
  background: var(--el-fill-color-light);
}

.alert-list {
  flex: 1;
  overflow-y: auto;
  padding: 16px;
}

.alert-item {
  margin-bottom: 16px;
  border-radius: 8px;
  border-left: 4px solid;
  background: var(--el-bg-color-overlay);
  transition: all 0.3s;
}

.alert-item:hover {
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
}

.alert-critical {
  border-left-color: var(--el-color-danger);
}

.alert-high {
  border-left-color: var(--el-color-warning);
}

.alert-medium {
  border-left-color: var(--el-color-primary);
}

.alert-low {
  border-left-color: var(--el-color-info);
}

.alert-content {
  padding: 16px;
}

.alert-title {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
  font-weight: 600;
}

.alert-icon {
  font-size: 18px;
}

.alert-message {
  margin-bottom: 12px;
  color: var(--el-text-color-regular);
}

.alert-metrics {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-bottom: 12px;
}

.metric-tag {
  font-family: monospace;
}

.alert-suggestions {
  margin-bottom: 12px;
  padding: 12px;
  background: var(--el-fill-color-light);
  border-radius: 4px;
}

.suggestions-title {
  font-weight: 600;
  margin-bottom: 8px;
}

.alert-suggestions ul {
  margin: 0;
  padding-left: 20px;
}

.alert-suggestions li {
  margin-bottom: 4px;
}

.alert-actions {
  display: flex;
  gap: 8px;
  margin-bottom: 12px;
}

.alert-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-top: 12px;
  border-top: 1px solid var(--el-border-color-lighter);
}

.alert-time {
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.alert-controls {
  display: flex;
  gap: 8px;
}
</style>
