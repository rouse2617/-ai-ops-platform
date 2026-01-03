<template>
  <div class="context-panel" :class="{ 'dark-mode': isDark }">
    <!-- Header with View Tabs -->
    <div class="panel-header">
      <el-tabs v-model="activeView" class="panel-tabs">
        <el-tab-pane label="监控" name="monitor">
          <template #label>
            <span class="tab-label">
              <el-icon><Monitor /></el-icon>
              监控
            </span>
          </template>
        </el-tab-pane>
        <el-tab-pane label="日志" name="logs">
          <template #label>
            <span class="tab-label">
              <el-icon><Document /></el-icon>
              日志
            </span>
          </template>
        </el-tab-pane>
        <el-tab-pane label="文件" name="files">
          <template #label>
            <span class="tab-label">
              <el-icon><Folder /></el-icon>
              文件
            </span>
          </template>
        </el-tab-pane>
      </el-tabs>
      <el-switch
        v-model="localSilentMode"
        @change="handleSilentModeToggle"
        active-text="静默"
        :active-icon="MuteNotification"
        :inactive-icon="Bell"
        size="small"
      />
    </div>

    <!-- View Content -->
    <div class="panel-content">
      <!-- Monitor View -->
      <div v-show="activeView === 'monitor'" class="view-container">
        <div
          v-for="host in displayHosts"
          :key="host.hostId"
          class="host-metrics"
          :class="{ 'pulse-warning': isAbnormal(host), 'expanded': isAbnormal(host) || isExpanded(host.hostId) }"
        >
          <div class="host-header" @click="toggleExpand(host.hostId)">
            <span class="host-name">{{ host.hostName }}</span>
            <div class="host-status">
              <el-tag :type="getStatusType(host)" size="small" effect="dark">
                {{ host.status }}
              </el-tag>
            </div>
          </div>

          <transition name="expand">
            <div v-show="isAbnormal(host) || isExpanded(host.hostId)" class="metrics-details">
              <div class="metric-item">
                <div class="metric-label">CPU Usage</div>
                <div class="metric-value">
                  <el-progress
                    :percentage="host.cpu"
                    :status="getProgressStatus(host.cpu)"
                    :show-text="true"
                    :stroke-width="8"
                  />
                </div>
              </div>

              <div class="metric-item">
                <div class="metric-label">Memory Usage</div>
                <div class="metric-value">
                  <el-progress
                    :percentage="host.memory"
                    :status="getProgressStatus(host.memory)"
                    :show-text="true"
                    :stroke-width="8"
                  />
                </div>
              </div>

              <div class="metric-item">
                <div class="metric-label">Disk Usage</div>
                <div class="metric-value">
                  <el-progress
                    :percentage="host.disk"
                    :status="getProgressStatus(host.disk)"
                    :show-text="true"
                    :stroke-width="8"
                  />
                </div>
              </div>

              <div v-if="isAbnormal(host)" class="alert-indicator">
                <el-alert
                  :title="getAlertMessage(host)"
                  type="warning"
                  :closable="false"
                  show-icon
                />
              </div>

              <div class="last-updated">
                <span class="update-label">Last updated:</span>
                <span class="update-time">{{ formatTimestamp(host.lastUpdated) }}</span>
              </div>
            </div>
          </transition>
        </div>

        <div v-if="displayHosts.length === 0" class="empty-state">
          <el-empty description="No monitoring data available" :image-size="80" />
        </div>
      </div>

      <!-- Logs View -->
      <div v-show="activeView === 'logs'" class="view-container">
        <div class="logs-container">
          <div class="logs-header">
            <el-select v-model="logLevel" size="small" placeholder="日志级别">
              <el-option label="全部" value="all" />
              <el-option label="错误" value="error" />
              <el-option label="警告" value="warning" />
              <el-option label="信息" value="info" />
            </el-select>
          </div>
          <div class="logs-list">
            <div
              v-for="(log, index) in filteredLogs"
              :key="index"
              class="log-item"
              :class="`log-${log.level}`"
            >
              <span class="log-time">{{ log.time }}</span>
              <el-tag :type="getLogTagType(log.level)" size="small">{{ log.level }}</el-tag>
              <span class="log-message">{{ log.message }}</span>
            </div>
          </div>
          <div v-if="filteredLogs.length === 0" class="empty-state">
            <el-empty description="暂无日志数据" :image-size="80" />
          </div>
        </div>
      </div>

      <!-- Files View -->
      <div v-show="activeView === 'files'" class="view-container">
        <div v-if="displayRelatedFiles.length > 0" class="related-files">
          <div class="section-title">Related Files ({{ displayRelatedFiles.length }})</div>
          <div class="files-list">
            <div
              v-for="file in displayRelatedFiles"
              :key="file.path"
              class="file-item"
              :class="{ 'has-highlights': file.highlightLines && file.highlightLines.length > 0 }"
              @click="handleFileClick(file)"
              title="Click to preview file"
            >
              <el-icon><Document /></el-icon>
              <div class="file-info">
                <span class="file-path">{{ file.path }}</span>
                <span v-if="file.highlightLines && file.highlightLines.length > 0" class="highlight-info">
                  Lines: {{ file.highlightLines.join(', ') }}
                </span>
              </div>
              <el-icon class="arrow-icon"><ArrowRight /></el-icon>
            </div>
          </div>
        </div>
        <div v-else class="empty-state">
          <el-empty description="暂无相关文件" :image-size="80" />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { Bell, MuteNotification, Document, ArrowRight, Monitor, Folder } from '@element-plus/icons-vue'
import type { HostMetrics, FileReference } from '@/types/chat-ui'
import { useMetricsStore } from '@/stores/metrics'

// Props
interface Props {
  hosts?: HostMetrics[]
  relatedFiles?: FileReference[]
  silentMode?: boolean
  maxHosts?: number
  userMessage?: string
}

const props = withDefaults(defineProps<Props>(), {
  hosts: () => [],
  relatedFiles: () => [],
  silentMode: false,
  maxHosts: 5,
  userMessage: ''
})

// Emits
const emit = defineEmits<{
  silentModeToggle: [value: boolean]
  fileClick: [file: FileReference]
}>()

// Store
const metricsStore = useMetricsStore()

// Local state
const localSilentMode = ref(props.silentMode)
const expandedHosts = ref<Set<string>>(new Set())
const activeView = ref<'monitor' | 'logs' | 'files'>('monitor')
const logLevel = ref('all')

// Mock logs data
const mockLogs = ref([
  { time: '10:23:45', level: 'error', message: 'Connection timeout to database' },
  { time: '10:22:30', level: 'warning', message: 'High memory usage detected' },
  { time: '10:21:15', level: 'info', message: 'Service started successfully' },
  { time: '10:20:00', level: 'error', message: 'Failed to load configuration file' }
])

// Computed
const isDark = computed(() => {
  return document.documentElement.classList.contains('dark')
})

const displayHosts = computed(() => {
  const hostsToDisplay = metricsStore.hostList.length > 0
    ? metricsStore.hostList
    : props.hosts

  return hostsToDisplay.slice(0, props.maxHosts)
})

const displayRelatedFiles = computed(() => {
  return props.relatedFiles.slice(0, 3)
})

const filteredLogs = computed(() => {
  if (logLevel.value === 'all') return mockLogs.value
  return mockLogs.value.filter(log => log.level === logLevel.value)
})

// Intent recognition - auto switch view based on keywords
const detectIntent = (message: string) => {
  const lowerMsg = message.toLowerCase()

  // Log keywords
  if (lowerMsg.includes('报错') || lowerMsg.includes('错误') || lowerMsg.includes('error') ||
      lowerMsg.includes('日志') || lowerMsg.includes('log')) {
    activeView.value = 'logs'
    return
  }

  // Monitor keywords
  if (lowerMsg.includes('cpu') || lowerMsg.includes('内存') || lowerMsg.includes('memory') ||
      lowerMsg.includes('磁盘') || lowerMsg.includes('disk') || lowerMsg.includes('监控') ||
      lowerMsg.includes('性能') || lowerMsg.includes('为什么高')) {
    activeView.value = 'monitor'
    return
  }

  // File keywords
  if (lowerMsg.includes('文件') || lowerMsg.includes('file') || lowerMsg.includes('代码') ||
      lowerMsg.includes('配置')) {
    activeView.value = 'files'
    return
  }
}

// Watch for prop changes
watch(() => props.silentMode, (newValue) => {
  localSilentMode.value = newValue
})

// Watch user message for intent detection
watch(() => props.userMessage, (newMessage) => {
  if (newMessage) {
    detectIntent(newMessage)
  }
})

// Methods
const handleSilentModeToggle = (value: boolean) => {
  emit('silentModeToggle', value)
}

const handleFileClick = (file: FileReference) => {
  emit('fileClick', file)
}

const isAbnormal = (host: HostMetrics): boolean => {
  return (
    host.cpu > 80 ||
    host.memory > 85 ||
    host.status === 'warning' ||
    host.status === 'critical'
  )
}

const getStatusType = (host: HostMetrics): 'success' | 'warning' | 'danger' | 'info' => {
  if (host.cpu > 80 || host.memory > 85 || host.status === 'critical') {
    return 'danger'
  }
  if (host.status === 'warning') {
    return 'warning'
  }
  if (host.status === 'normal') {
    return 'success'
  }
  return 'info'
}

const getProgressStatus = (percentage: number): 'success' | 'exception' | 'warning' => {
  if (percentage >= 90) {
    return 'exception'
  } else if (percentage >= 80) {
    return 'warning'
  }
  return 'success'
}

const getAlertMessage = (host: HostMetrics): string => {
  const alerts: string[] = []
  if (host.cpu > 80) {
    alerts.push(`CPU usage is ${host.cpu.toFixed(1)}%`)
  }
  if (host.memory > 85) {
    alerts.push(`Memory usage is ${host.memory.toFixed(1)}%`)
  }
  if (host.disk > 90) {
    alerts.push(`Disk usage is ${host.disk.toFixed(1)}%`)
  }
  return alerts.join('; ')
}

const getLogTagType = (level: string): 'success' | 'warning' | 'danger' | 'info' => {
  switch (level) {
    case 'error': return 'danger'
    case 'warning': return 'warning'
    case 'info': return 'info'
    default: return 'info'
  }
}

const isExpanded = (hostId: string): boolean => {
  return expandedHosts.value.has(hostId)
}

const toggleExpand = (hostId: string) => {
  if (expandedHosts.value.has(hostId)) {
    expandedHosts.value.delete(hostId)
  } else {
    expandedHosts.value.add(hostId)
  }
}

const formatTimestamp = (timestamp: number): string => {
  const date = new Date(timestamp)
  const now = new Date()
  const diff = now.getTime() - date.getTime()

  if (diff < 60000) {
    return 'just now'
  } else if (diff < 3600000) {
    const minutes = Math.floor(diff / 60000)
    return `${minutes}m ago`
  } else if (diff < 86400000) {
    const hours = Math.floor(diff / 3600000)
    return `${hours}h ago`
  } else {
    return date.toLocaleDateString()
  }
}

// Subscribe to hosts from props on mount
onMounted(() => {
  props.hosts.forEach(host => {
    metricsStore.subscribe(host.hostId)
  })
})

onUnmounted(() => {
  props.hosts.forEach(host => {
    metricsStore.unsubscribe(host.hostId)
  })
})
</script>

<style scoped lang="scss">
.context-panel {
  background: var(--el-bg-color);
  border-radius: 8px;
  padding: 16px;
  height: 100%;
  overflow: hidden;
  border: 1px solid var(--el-border-color);
  transition: all 0.3s ease;
  display: flex;
  flex-direction: column;

  &.dark-mode {
    background: #1a1a1a;
    border-color: #333;
  }
}

.panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
  padding-bottom: 12px;
  border-bottom: 1px solid var(--el-border-color);
  flex-shrink: 0;
}

.panel-tabs {
  flex: 1;

  :deep(.el-tabs__header) {
    margin: 0;
  }

  :deep(.el-tabs__nav-wrap::after) {
    display: none;
  }
}

.tab-label {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 13px;
}

.panel-content {
  flex: 1;
  overflow-y: auto;
  min-height: 0;
}

.view-container {
  height: 100%;
}

.host-metrics {
  margin-bottom: 12px;
  border-radius: 6px;
  border: 1px solid var(--el-border-color);
  overflow: hidden;
  transition: all 0.3s ease;

  &.pulse-warning {
    animation: pulse-border 2s ease-in-out infinite;
    border-color: var(--el-color-warning);
  }

  &.expanded {
    box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
  }
}

@keyframes pulse-border {
  0%, 100% {
    border-color: var(--el-color-warning);
    box-shadow: 0 0 0 0 rgba(var(--el-color-warning-rgb), 0.4);
  }
  50% {
    border-color: var(--el-color-warning);
    box-shadow: 0 0 0 8px rgba(var(--el-color-warning-rgb), 0);
  }
}

.host-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 12px;
  background: var(--el-fill-color-light);
  cursor: pointer;
  user-select: none;
  transition: background 0.2s;

  &:hover {
    background: var(--el-fill-color);
  }
}

.host-name {
  font-size: 13px;
  font-weight: 500;
  color: var(--el-text-color-primary);
}

.expand-enter-active,
.expand-leave-active {
  transition: all 0.3s ease;
  max-height: 500px;
  opacity: 1;
  overflow: hidden;
}

.expand-enter-from,
.expand-leave-to {
  max-height: 0;
  opacity: 0;
}

.metrics-details {
  padding: 12px;
  background: var(--el-bg-color);
  border-top: 1px solid var(--el-border-color);
}

.metric-item {
  margin-bottom: 12px;

  &:last-child {
    margin-bottom: 0;
  }
}

.metric-label {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  margin-bottom: 6px;
}

.last-updated {
  margin-top: 12px;
  padding-top: 12px;
  border-top: 1px solid var(--el-border-color-lighter);
  font-size: 11px;
  color: var(--el-text-color-secondary);
  display: flex;
  justify-content: space-between;
}

.alert-indicator {
  margin-top: 12px;
  padding-top: 12px;
  border-top: 1px solid var(--el-border-color-lighter);

  :deep(.el-alert) {
    --el-alert-padding: 8px 12px;
  }

  :deep(.el-alert__title) {
    font-size: 12px;
  }
}

// Logs View
.logs-container {
  display: flex;
  flex-direction: column;
  height: 100%;
}

.logs-header {
  margin-bottom: 12px;
}

.logs-list {
  flex: 1;
  overflow-y: auto;
}

.log-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  margin-bottom: 8px;
  background: var(--el-fill-color-light);
  border-radius: 4px;
  font-size: 12px;
  border-left: 3px solid transparent;

  &.log-error {
    border-left-color: var(--el-color-danger);
  }

  &.log-warning {
    border-left-color: var(--el-color-warning);
  }

  &.log-info {
    border-left-color: var(--el-color-info);
  }
}

.log-time {
  color: var(--el-text-color-secondary);
  font-family: monospace;
  flex-shrink: 0;
}

.log-message {
  flex: 1;
  color: var(--el-text-color-regular);
}

// Files View
.related-files {
  margin-top: 0;
  padding-top: 0;
  border-top: none;
}

.section-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--el-text-color-primary);
  margin-bottom: 12px;
}

.files-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.file-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  background: var(--el-fill-color-light);
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.2s;

  &:hover {
    background: var(--el-fill-color);
    transform: translateX(4px);
  }

  &.has-highlights {
    border-left: 3px solid var(--el-color-warning);
  }

  .el-icon {
    font-size: 16px;
    color: var(--el-color-primary);
  }
}

.file-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.file-path {
  flex: 1;
  font-size: 12px;
  color: var(--el-text-color-regular);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.highlight-info {
  font-size: 11px;
  color: var(--el-color-warning);
  font-weight: 500;
}

.arrow-icon {
  font-size: 14px;
  color: var(--el-text-color-secondary);
}

.empty-state {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 200px;
}

// Dark mode
.dark-mode {
  .host-header {
    background: #2a2a2a;

    &:hover {
      background: #333;
    }
  }

  .metrics-details {
    background: #1a1a1a;
  }

  .file-item, .log-item {
    background: #2a2a2a;

    &:hover {
      background: #333;
    }
  }
}

// Scrollbar
.context-panel::-webkit-scrollbar,
.panel-content::-webkit-scrollbar,
.logs-list::-webkit-scrollbar {
  width: 6px;
}

.context-panel::-webkit-scrollbar-track,
.panel-content::-webkit-scrollbar-track,
.logs-list::-webkit-scrollbar-track {
  background: transparent;
}

.context-panel::-webkit-scrollbar-thumb,
.panel-content::-webkit-scrollbar-thumb,
.logs-list::-webkit-scrollbar-thumb {
  background: var(--el-border-color-darker);
  border-radius: 3px;

  &:hover {
    background: var(--el-border-color-dark);
  }
}
</style>
