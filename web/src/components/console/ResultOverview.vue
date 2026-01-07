<template>
  <div class="result-overview">
    <!-- 状态统计仪表盘 -->
    <div class="status-dashboard">
      <div class="stat-item success" @click="statusFilter = 'success'" :class="{ active: statusFilter === 'success' }">
        <div class="stat-value">{{ successCount }}</div>
        <div class="stat-label">成功</div>
      </div>
      <div class="stat-item error" @click="statusFilter = 'error'" :class="{ active: statusFilter === 'error' }">
        <div class="stat-value">{{ errorCount }}</div>
        <div class="stat-label">失败</div>
      </div>
      <div class="stat-item total" @click="statusFilter = ''" :class="{ active: statusFilter === '' }">
        <div class="stat-value">{{ results.length }}</div>
        <div class="stat-label">总计</div>
      </div>
      <div class="stat-item elapsed">
        <div class="stat-value">{{ averageElapsed }}</div>
        <div class="stat-label">平均耗时</div>
      </div>
    </div>

    <!-- 双栏布局 -->
    <div class="split-view">
      <!-- 左侧：主机列表 -->
      <div class="host-list-panel">
        <div class="panel-header">
          <el-button
            v-if="errorCount > 0"
            type="danger"
            size="small"
            @click="handleRetryAllFailed"
          >
            重试失败 ({{ errorCount }})
          </el-button>
          <span class="result-count">{{ filteredResults.length }} 条</span>
        </div>
        <div class="host-list">
          <div
            v-for="(result, index) in filteredResults"
            :key="index"
            class="host-item"
            :class="{ selected: selectedResult === result, error: result.status === 'error' }"
            @click="selectResult(result)"
          >
            <span class="status-dot" :class="result.status"></span>
            <span class="host-name">{{ result.host }}</span>
            <span class="elapsed-time">{{ result.elapsed }}</span>
          </div>
        </div>
      </div>

      <!-- 右侧：详情面板 -->
      <div class="detail-panel">
        <template v-if="selectedResult">
          <div class="detail-header">
            <el-tag :type="selectedResult.status === 'success' ? 'success' : 'danger'" size="large">
              {{ selectedResult.host }}
            </el-tag>
            <div class="detail-actions">
              <el-button size="small" @click="handleCopy(selectedResult)">复制</el-button>
              <el-button size="small" @click="handleExport(selectedResult)">导出</el-button>
              <el-button v-if="selectedResult.status === 'error'" type="danger" size="small" @click="handleRetry(selectedResult)">重试</el-button>
            </div>
          </div>
          <div class="detail-content">
            <!-- 错误信息 -->
            <div v-if="selectedResult.error" class="error-block">
              <div class="error-label">错误信息</div>
              <pre class="error-text">{{ selectedResult.error }}</pre>
            </div>
            <!-- 执行结果 -->
            <div v-if="selectedResult.result" class="result-block">
              <pre class="result-text">{{ formatResult(selectedResult.result) }}</pre>
            </div>
          </div>
        </template>
        <div v-else class="empty-detail">
          <el-icon :size="48"><Document /></el-icon>
          <p>选择左侧主机查看详情</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { Document } from '@element-plus/icons-vue'
import type { BatchExecuteResult } from '@/api/operations'
import { useConsoleStore } from '@/stores/console'

const consoleStore = useConsoleStore()

const statusFilter = ref('')
const selectedResult = ref<BatchExecuteResult | null>(null)

const results = computed(() => consoleStore.executionResults)

// 默认显示失败项（如果有失败的话）
onMounted(() => {
  if (results.value.some(r => r.status === 'error')) {
    statusFilter.value = 'error'
  }
})

// 监听结果变化
watch(() => results.value, (newResults) => {
  if (newResults.length > 0 && !selectedResult.value) {
    // 自动选中第一个失败项或第一项
    const firstError = newResults.find(r => r.status === 'error')
    selectedResult.value = firstError || newResults[0]
  }
}, { deep: true, immediate: true })

const filteredResults = computed(() => {
  if (!statusFilter.value) {
    const sorted = [...results.value]
    return sorted.sort((a, b) => {
      if (a.status === 'error' && b.status !== 'error') return -1
      if (a.status !== 'error' && b.status === 'error') return 1
      return 0
    })
  }
  return results.value.filter(r => r.status === statusFilter.value)
})

const successCount = computed(() => results.value.filter(r => r.status === 'success').length)
const errorCount = computed(() => results.value.filter(r => r.status === 'error').length)

const averageElapsed = computed(() => {
  if (results.value.length === 0) return '0ms'
  const total = results.value.reduce((sum, r) => sum + parseElapsed(r.elapsed), 0)
  return formatElapsed(total / results.value.length)
})

function parseElapsed(elapsed: string): number {
  if (!elapsed) return 0
  const match = elapsed.match(/(\d+\.?\d*)(ms|s|m)/)
  if (!match) return 0
  const value = parseFloat(match[1])
  const unit = match[2]
  return unit === 'ms' ? value : unit === 's' ? value * 1000 : value * 60000
}

function formatElapsed(ms: number): string {
  if (ms < 1000) return `${Math.round(ms)}ms`
  if (ms < 60000) return `${(ms / 1000).toFixed(2)}s`
  return `${(ms / 60000).toFixed(2)}m`
}

const selectResult = (result: BatchExecuteResult) => {
  selectedResult.value = result
}

const formatResult = (result: any): string => {
  if (typeof result === 'string') return result
  if (result.log) return result.log
  if (result.output) return result.output
  if (result.cpu) return result.cpu
  return JSON.stringify(result, null, 2)
}

const handleRetryAllFailed = async () => {
  const failedHosts = results.value.filter(r => r.status === 'error').map(r => r.host)
  if (failedHosts.length === 0 || !consoleStore.currentOperation) return

  try {
    await consoleStore.executeOperation(consoleStore.currentOperation, failedHosts, consoleStore.operationParams)
    ElMessage.success(`重试 ${failedHosts.length} 个主机`)
    statusFilter.value = ''
  } catch (error: any) {
    ElMessage.error(error.message || '重试失败')
  }
}

const handleCopy = (result: BatchExecuteResult) => {
  const text = result.error || formatResult(result.result)
  navigator.clipboard.writeText(text).then(() => ElMessage.success('已复制'))
}

const handleExport = (result: BatchExecuteResult) => {
  const content = result.error || formatResult(result.result)
  const blob = new Blob([content], { type: 'text/plain' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `${result.host}_result.txt`
  a.click()
  URL.revokeObjectURL(url)
}

const handleRetry = async (result: BatchExecuteResult) => {
  if (!consoleStore.currentOperation) return
  try {
    await consoleStore.executeOperation(consoleStore.currentOperation, [result.host], consoleStore.operationParams)
    ElMessage.success('重试成功')
  } catch (error: any) {
    ElMessage.error(error.message || '重试失败')
  }
}

const emit = defineEmits<{
  'view-detail': [result: BatchExecuteResult]
}>()
</script>

<style scoped>
.result-overview {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.status-dashboard {
  display: flex;
  gap: 16px;
  padding: 16px;
  background: linear-gradient(to right, #f5f7fa, #fff);
  border-bottom: 1px solid var(--el-border-color);
}

.stat-item {
  flex: 1;
  padding: 12px;
  background: #fff;
  border-radius: 8px;
  text-align: center;
  border: 1px solid var(--el-border-color);
  transition: all 0.3s;
}

.stat-item:hover {
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.stat-item.success {
  border-left: 4px solid #67c23a;
}

.stat-item.error {
  border-left: 4px solid #f56c6c;
}

.stat-item.total {
  border-left: 4px solid #409eff;
}

.stat-item.elapsed {
  border-left: 4px solid #e6a23c;
}

.stat-value {
  font-size: 24px;
  font-weight: 600;
  margin-bottom: 4px;
}

.stat-item.success .stat-value {
  color: #67c23a;
}

.stat-item.error .stat-value {
  color: #f56c6c;
}

.stat-item.total .stat-value {
  color: #409eff;
}

.stat-item.elapsed .stat-value {
  color: #e6a23c;
}

.stat-label {
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

/* 双栏布局 */
.split-view {
  flex: 1;
  display: flex;
  gap: 1px;
  background: var(--el-border-color);
  overflow: hidden;
}

.host-list-panel {
  width: 280px;
  min-width: 200px;
  background: #fff;
  display: flex;
  flex-direction: column;
}

.panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 12px;
  border-bottom: 1px solid var(--el-border-color);
  background: #fafafa;
}

.result-count {
  color: var(--el-text-color-secondary);
  font-size: 12px;
}

.host-list {
  flex: 1;
  overflow-y: auto;
}

.host-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 12px;
  cursor: pointer;
  border-bottom: 1px solid var(--el-border-color-lighter);
  transition: background 0.2s;
}

.host-item:hover {
  background: #f5f7fa;
}

.host-item.selected {
  background: #ecf5ff;
  border-left: 3px solid #409eff;
}

.host-item.error {
  background: #fef0f0;
}

.host-item.error.selected {
  background: #fde2e2;
  border-left-color: #f56c6c;
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}

.status-dot.success {
  background: #67c23a;
}

.status-dot.error {
  background: #f56c6c;
}

.host-name {
  flex: 1;
  font-size: 13px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.elapsed-time {
  font-size: 11px;
  color: var(--el-text-color-secondary);
}

/* 详情面板 */
.detail-panel {
  flex: 1;
  background: #fff;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.detail-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  border-bottom: 1px solid var(--el-border-color);
  background: #fafafa;
}

.detail-actions {
  display: flex;
  gap: 8px;
}

.detail-content {
  flex: 1;
  overflow-y: auto;
  padding: 16px;
}

.error-block {
  margin-bottom: 16px;
}

.error-label {
  font-size: 12px;
  font-weight: 600;
  color: #f56c6c;
  margin-bottom: 8px;
}

.error-text {
  background: #2d2d2d;
  color: #ff6b6b;
  padding: 12px 16px;
  border-radius: 6px;
  font-family: 'Consolas', 'Monaco', monospace;
  font-size: 13px;
  line-height: 1.5;
  overflow-x: auto;
  white-space: pre-wrap;
  word-break: break-all;
  margin: 0;
}

.result-block {
  flex: 1;
}

.result-text {
  background: #1e1e1e;
  color: #d4d4d4;
  padding: 12px 16px;
  border-radius: 6px;
  font-family: 'Consolas', 'Monaco', monospace;
  font-size: 13px;
  line-height: 1.5;
  overflow-x: auto;
  white-space: pre-wrap;
  word-break: break-all;
  margin: 0;
  min-height: 200px;
}

.empty-detail {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: var(--el-text-color-secondary);
}

.empty-detail p {
  margin-top: 12px;
  font-size: 14px;
}

/* 可点击的统计卡片 */
.stat-item.success,
.stat-item.error,
.stat-item.total {
  cursor: pointer;
}

.stat-item.active {
  box-shadow: 0 0 0 2px var(--el-color-primary);
}
</style>


