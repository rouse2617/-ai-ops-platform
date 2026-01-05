<template>
  <div class="server-cards-container" :class="{ 'single-card': servers.length === 1 }">
    <div
      v-for="server in servers"
      :key="server.hostname"
      class="server-card"
      :class="getCardClass(server)"
    >
      <!-- 卡片头部 -->
      <div class="card-header">
        <div class="server-info">
          <span class="server-name">{{ server.hostname }}</span>
          <el-tag :type="getStatusType(server.status)" size="small" effect="plain">
            {{ getStatusText(server.status) }}
          </el-tag>
        </div>
        <el-button
          v-if="server.hasDetails"
          text
          size="small"
          @click="toggleDetails(server.hostname)"
        >
          {{ expandedServers.has(server.hostname) ? '收起' : '详情' }}
          <el-icon>
            <ArrowUp v-if="expandedServers.has(server.hostname)" />
            <ArrowDown v-else />
          </el-icon>
        </el-button>
      </div>

      <!-- 核心指标 -->
      <div class="metrics-grid">
        <div class="metric-item" v-if="server.cpu !== undefined">
          <span class="metric-label">CPU</span>
          <div class="metric-value-row">
            <el-progress
              :percentage="server.cpu"
              :stroke-width="6"
              :color="getProgressColor(server.cpu)"
              :show-text="false"
              class="metric-progress"
            />
            <span class="metric-value" :class="getValueClass(server.cpu)">{{ server.cpu }}%</span>
          </div>
        </div>

        <div class="metric-item" v-if="server.memory !== undefined">
          <span class="metric-label">内存</span>
          <div class="metric-value-row">
            <el-progress
              :percentage="server.memory"
              :stroke-width="6"
              :color="getProgressColor(server.memory)"
              :show-text="false"
              class="metric-progress"
            />
            <span class="metric-value" :class="getValueClass(server.memory)">{{ server.memory }}%</span>
          </div>
        </div>

        <div class="metric-item" v-if="server.disk !== undefined">
          <span class="metric-label">磁盘</span>
          <div class="metric-value-row">
            <el-progress
              :percentage="server.disk"
              :stroke-width="6"
              :color="getProgressColor(server.disk)"
              :show-text="false"
              class="metric-progress"
            />
            <span class="metric-value" :class="getValueClass(server.disk)">{{ server.disk }}%</span>
          </div>
        </div>

        <div class="metric-item" v-if="server.load !== undefined">
          <span class="metric-label">负载</span>
          <span class="metric-value metric-text">{{ server.load }}</span>
        </div>
      </div>

      <!-- 可折叠详情 -->
      <el-collapse-transition>
        <div v-if="expandedServers.has(server.hostname) && server.details" class="card-details">
          <el-divider />
          <div class="details-section" v-if="server.details.topProcesses?.length">
            <span class="details-title">Top 进程</span>
            <div class="process-list">
              <div v-for="proc in server.details.topProcesses" :key="proc.pid" class="process-item">
                <span class="proc-name">{{ proc.name }}</span>
                <span class="proc-cpu">{{ proc.cpu }}%</span>
              </div>
            </div>
          </div>
          <div class="details-section" v-if="server.details.uptime">
            <span class="details-title">运行时间</span>
            <span class="details-value">{{ server.details.uptime }}</span>
          </div>
        </div>
      </el-collapse-transition>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { ArrowDown, ArrowUp } from '@element-plus/icons-vue'

export interface ServerProcess {
  pid: number
  name: string
  cpu: number
}

export interface ServerDetails {
  topProcesses?: ServerProcess[]
  uptime?: string
}

export interface ServerStatus {
  hostname: string
  status: 'normal' | 'warning' | 'error'
  cpu?: number
  memory?: number
  disk?: number
  load?: string
  hasDetails?: boolean
  details?: ServerDetails
}

defineProps<{
  servers: ServerStatus[]
}>()

const expandedServers = ref<Set<string>>(new Set())

const toggleDetails = (hostname: string) => {
  if (expandedServers.value.has(hostname)) {
    expandedServers.value.delete(hostname)
  } else {
    expandedServers.value.add(hostname)
  }
}

const getCardClass = (server: ServerStatus) => ({
  'status-normal': server.status === 'normal',
  'status-warning': server.status === 'warning',
  'status-error': server.status === 'error'
})

const getStatusType = (status: string) => {
  switch (status) {
    case 'normal': return 'success'
    case 'warning': return 'warning'
    case 'error': return 'danger'
    default: return 'info'
  }
}

const getStatusText = (status: string) => {
  switch (status) {
    case 'normal': return '正常'
    case 'warning': return '警告'
    case 'error': return '异常'
    default: return '未知'
  }
}

const getProgressColor = (value: number) => {
  if (value >= 90) return '#f56c6c'
  if (value >= 70) return '#e6a23c'
  return '#67c23a'
}

const getValueClass = (value: number) => ({
  'value-danger': value >= 90,
  'value-warning': value >= 70 && value < 90,
  'value-normal': value < 70
})
</script>

<style scoped>
.server-cards-container {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: 12px;
  margin: 12px 0;
}

.server-cards-container.single-card {
  grid-template-columns: 1fr;
  max-width: 400px;
}

.server-card {
  background: #fff;
  border: 1px solid #e4e7ed;
  border-radius: 8px;
  padding: 12px 16px;
  transition: all 0.2s ease;
}

.server-card:hover {
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.08);
}

/* 状态边框指示 */
.server-card.status-warning {
  border-left: 3px solid #e6a23c;
  background: linear-gradient(90deg, rgba(230, 162, 60, 0.05) 0%, #fff 20%);
}

.server-card.status-error {
  border-left: 3px solid #f56c6c;
  background: linear-gradient(90deg, rgba(245, 108, 108, 0.05) 0%, #fff 20%);
}

.server-card.status-normal {
  border-left: 3px solid #67c23a;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.server-info {
  display: flex;
  align-items: center;
  gap: 8px;
}

.server-name {
  font-weight: 600;
  font-size: 14px;
  color: #303133;
}

.metrics-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 10px;
}

.metric-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.metric-label {
  font-size: 12px;
  color: #909399;
}

.metric-value-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.metric-progress {
  flex: 1;
  min-width: 60px;
}

.metric-value {
  font-size: 13px;
  font-weight: 500;
  min-width: 40px;
  text-align: right;
}

.metric-text {
  color: #606266;
}

.value-normal {
  color: #67c23a;
}

.value-warning {
  color: #e6a23c;
}

.value-danger {
  color: #f56c6c;
}

.card-details {
  margin-top: 8px;
}

.card-details :deep(.el-divider) {
  margin: 8px 0;
}

.details-section {
  display: flex;
  flex-direction: column;
  gap: 4px;
  margin-bottom: 8px;
}

.details-title {
  font-size: 12px;
  color: #909399;
  font-weight: 500;
}

.details-value {
  font-size: 13px;
  color: #606266;
}

.process-list {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.process-item {
  display: flex;
  justify-content: space-between;
  font-size: 12px;
  padding: 2px 0;
}

.proc-name {
  color: #606266;
  max-width: 150px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.proc-cpu {
  color: #909399;
}

/* 响应式 */
@media (max-width: 600px) {
  .server-cards-container {
    grid-template-columns: 1fr;
  }

  .metrics-grid {
    grid-template-columns: 1fr;
  }
}
</style>
