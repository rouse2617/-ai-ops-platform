<template>
  <div class="enhanced-tool-card" :class="statusClass">
    <!-- 工具头部 -->
    <div class="tool-header">
      <div class="tool-main-info">
        <div class="tool-icon-wrapper">
          <el-icon class="tool-icon" :size="20">
            <component :is="toolIcon" />
          </el-icon>
        </div>
        <div class="tool-meta">
          <div class="tool-name-row">
            <span class="tool-name">{{ toolCall.name }}</span>
            <el-tag :type="statusTagType" size="small" effect="plain">
              {{ statusText }}
            </el-tag>
          </div>
          <div class="tool-time" v-if="executionTime">
            <el-icon><Clock /></el-icon>
            <span>执行耗时: {{ executionTime }}</span>
          </div>
          <!-- 主机标签 -->
          <div class="host-tags" v-if="hasMultipleHosts">
            <HostTagBadge
              v-for="hostResult in toolCall.hostResults"
              :key="hostResult.hostId"
              :host-id="hostResult.hostId"
              :status="getHostStatus(hostResult)"
              :result="hostResult"
              @click="toggleHostResult(hostResult.hostId)"
            />
          </div>
        </div>
      </div>

      <!-- 快捷操作 -->
      <div class="tool-actions">
        <el-dropdown trigger="click" @command="handleAction">
          <el-button type="text" :icon="MoreFilled" circle />
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="copy" :icon="CopyDocument">
                复制结果
              </el-dropdown-item>
              <el-dropdown-item command="export" :icon="Download">
                导出为文件
              </el-dropdown-item>
              <el-dropdown-item command="rerun" :icon="RefreshRight">
                重新执行
              </el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </div>
    </div>

    <!-- AI 深度解读区块 -->
    <div class="ai-insight-section" v-if="semanticInsight || insightLoading">
      <div class="insight-header" @click="toggleInsight">
        <el-icon class="insight-icon"><MagicStick /></el-icon>
        <span class="insight-label">AI 深度解读</span>
        <el-tag
          v-if="semanticInsight"
          :type="riskTagType"
          size="small"
          effect="dark"
          class="risk-tag"
        >
          {{ riskLevelText }}
        </el-tag>
        <el-icon class="expand-arrow" :class="{ expanded: insightExpanded }">
          <ArrowRight />
        </el-icon>
      </div>
      <el-collapse-transition>
        <div v-show="insightExpanded" class="insight-content">
          <div v-if="insightLoading" class="insight-loading">
            <el-icon class="is-loading"><Loading /></el-icon>
            <span>正在分析...</span>
          </div>
          <template v-else-if="semanticInsight">
            <div class="insight-summary">
              <el-icon :class="['summary-icon', `risk-${semanticInsight.risk_level}`]">
                <component :is="riskIcon" />
              </el-icon>
              <span>{{ semanticInsight.summary }}</span>
            </div>
            <div class="insight-trend" v-if="semanticInsight.trend">
              <el-icon><TrendCharts /></el-icon>
              <span>{{ semanticInsight.trend }}</span>
            </div>
            <div class="insight-recommendation" v-if="semanticInsight.recommendation">
              <el-icon><Promotion /></el-icon>
              <span>{{ semanticInsight.recommendation }}</span>
            </div>
            <div class="insight-metrics" v-if="semanticInsight.key_metrics?.length">
              <div
                v-for="metric in semanticInsight.key_metrics"
                :key="metric.name"
                class="metric-item"
                :class="`metric-${metric.status}`"
              >
                <span class="metric-name">{{ metric.name }}</span>
                <span class="metric-value">{{ metric.value }}</span>
                <span class="metric-threshold" v-if="metric.threshold">/ {{ metric.threshold }}</span>
              </div>
            </div>
          </template>
        </div>
      </el-collapse-transition>
    </div>

    <!-- 历史关联分析卡片 -->
    <HistoricalCorrelationCard
      v-if="showHistoricalCorrelation"
      :host-id="hostId"
      :current-issue="currentIssueText"
      :issue-type="detectedIssueType"
      @apply-solution="handleApplySolution"
    />

    <!-- 进度条（执行中） -->
    <div class="tool-progress" v-if="toolCall.status === 'running'">
      <el-progress
        :percentage="progress"
        :indeterminate="true"
        :show-text="false"
        :stroke-width="3"
      />
    </div>

    <!-- 可折叠内容 -->
    <el-collapse-transition>
      <div v-show="isExpanded" class="tool-body">
        <!-- 参数展示 -->
        <div class="tool-section" v-if="toolCall.arguments">
          <div class="section-header" @click="toggleSection('args')">
            <el-icon class="section-icon" :class="{ expanded: expandedSections.args }">
              <ArrowRight />
            </el-icon>
            <span class="section-title">执行参数</span>
            <el-tag size="small" type="info" effect="plain">
              {{ argCount }} 个参数
            </el-tag>
          </div>
          <el-collapse-transition>
            <div v-show="expandedSections.args" class="section-content">
              <pre class="code-block"><code>{{ formattedArguments }}</code></pre>
              <div class="section-actions">
                <el-button
                  type="text"
                  size="small"
                  :icon="CopyDocument"
                  @click="copyToClipboard(formattedArguments, '参数')"
                >
                  复制
                </el-button>
              </div>
            </div>
          </el-collapse-transition>
        </div>

        <!-- 结果展示 -->
        <div class="tool-section" v-if="toolCall.result || hasMultipleHosts">
          <div class="section-header" @click="toggleSection('result')">
            <el-icon class="section-icon" :class="{ expanded: expandedSections.result }">
              <ArrowRight />
            </el-icon>
            <span class="section-title">执行结果</span>
            <div class="result-stats">
              <el-tag size="small" type="success" effect="plain" v-if="!hasMultipleHosts">
                {{ resultLineCount }} 行
              </el-tag>
              <el-tag size="small" type="info" effect="plain" v-if="resultSize && !hasMultipleHosts">
                {{ resultSize }}
              </el-tag>
              <el-tag size="small" type="info" effect="plain" v-if="hasMultipleHosts">
                {{ toolCall.hostResults?.length }} 台主机
              </el-tag>
            </div>
          </div>
          <el-collapse-transition>
            <div v-show="expandedSections.result" class="section-content">
              <!-- 多主机结果 -->
              <div v-if="hasMultipleHosts" class="multi-host-results">
                <div
                  v-for="hostResult in toolCall.hostResults"
                  :key="hostResult.hostId"
                  class="host-result-item"
                  :class="{ 'has-error': hostResult.exitCode !== 0 }"
                >
                  <div class="host-result-header" @click="toggleHostResult(hostResult.hostId)">
                    <HostTagBadge
                      :host-id="hostResult.hostId"
                      :status="getHostStatus(hostResult)"
                      :result="hostResult"
                      :show-details="false"
                    />
                    <el-icon class="expand-icon" :class="{ expanded: expandedHosts[hostResult.hostId] }">
                      <ArrowRight />
                    </el-icon>
                  </div>
                  <el-collapse-transition>
                    <div v-show="expandedHosts[hostResult.hostId]" class="host-result-content">
                      <pre class="code-block result-code"><code>{{ hostResult.output }}</code></pre>
                      <div v-if="hostResult.error" class="host-error">
                        <el-alert type="error" :closable="false" show-icon>
                          <template #title>错误信息</template>
                          {{ hostResult.error }}
                        </el-alert>
                      </div>
                      <div class="section-actions">
                        <el-button
                          type="text"
                          size="small"
                          :icon="CopyDocument"
                          @click="copyToClipboard(hostResult.output, `${hostResult.hostId} 结果`)"
                        >
                          复制
                        </el-button>
                      </div>
                    </div>
                  </el-collapse-transition>
                </div>
              </div>
              <!-- 单主机结果 -->
              <div v-else>
                <div v-if="isJsonResult" class="result-viewer">
                  <JsonViewer :data="parsedResult" />
                </div>
                <pre v-else class="code-block result-code"><code>{{ toolCall.result }}</code></pre>
                <div class="section-actions">
                  <el-button
                    type="text"
                    size="small"
                    :icon="CopyDocument"
                    @click="copyToClipboard(toolCall.result, '结果')"
                  >
                    复制结果
                  </el-button>
                  <el-button
                    type="text"
                    size="small"
                    :icon="Download"
                    @click="exportResult"
                  >
                    导出
                  </el-button>
                </div>
              </div>
            </div>
          </el-collapse-transition>
        </div>

        <!-- 错误信息 -->
        <div class="tool-section error-section" v-if="toolCall.error">
          <div class="section-header" @click="toggleSection('error')">
            <el-icon class="section-icon" :class="{ expanded: expandedSections.error }">
              <ArrowRight />
            </el-icon>
            <span class="section-title">错误信息</span>
            <el-tag size="small" type="danger" effect="plain">
              执行失败
            </el-tag>
          </div>
          <el-collapse-transition>
            <div v-show="expandedSections.error" class="section-content error-content">
              <pre class="code-block error-code"><code>{{ toolCall.error }}</code></pre>
            </div>
          </el-collapse-transition>
        </div>
      </div>
    </el-collapse-transition>

    <!-- 折叠按钮 -->
    <div class="collapse-trigger" @click="toggleExpand" v-if="toolCall.result || toolCall.error">
      <el-icon :class="{ expanded: isExpanded }">
        <ArrowDown />
      </el-icon>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, onMounted, onUnmounted, watch } from 'vue'
import { ElMessage } from 'element-plus'
import type { ToolCall, HostResult } from '@/api/chat'
import { getSemanticInsight, type SemanticInsight } from '@/api/analysis'
import {
  Clock, MoreFilled,
  CopyDocument, Download, RefreshRight, ArrowRight, ArrowDown,
  Monitor, Tools, Document, Warning, MagicStick, Loading,
  CircleCheck, WarningFilled, CircleClose, TrendCharts, Promotion
} from '@element-plus/icons-vue'
import JsonViewer from './JsonViewer.vue'
import HostTagBadge from './HostTagBadge.vue'
import HistoricalCorrelationCard from './HistoricalCorrelationCard.vue'

const props = defineProps<{
  toolCall: ToolCall
  hostId?: string
}>()

const emit = defineEmits<{
  rerun: [toolCall: ToolCall]
  applySolution: [solution: string]
}>()

// 折叠状态
const isExpanded = ref(true)
const expandedSections = ref({
  args: true,
  result: true,
  error: true
})
const expandedHosts = ref<Record<string, boolean>>({})

// AI 深度解读状态
const semanticInsight = ref<SemanticInsight | null>(null)
const insightLoading = ref(false)
const insightExpanded = ref(true)

// 历史关联分析
const showHistoricalCorrelation = computed(() => {
  // 当检测到问题时显示历史关联
  return semanticInsight.value?.risk_level === 'warning' ||
         semanticInsight.value?.risk_level === 'critical' ||
         props.toolCall.error
})

const currentIssueText = computed(() => {
  if (props.toolCall.error) return props.toolCall.error
  if (semanticInsight.value?.summary) return semanticInsight.value.summary
  return props.toolCall.result || ''
})

const detectedIssueType = computed(() => {
  const name = props.toolCall.name.toLowerCase()
  if (name.includes('disk')) return 'disk'
  if (name.includes('memory')) return 'memory'
  if (name.includes('cpu')) return 'cpu'
  return undefined
})

const handleApplySolution = (solution: string) => {
  emit('applySolution', solution)
}

// 获取语义化解读
const fetchSemanticInsight = async () => {
  if (!props.toolCall.result || props.toolCall.status !== 'success') return

  // 只对特定工具类型获取解读
  const analyzableTools = ['check_disk', 'check_memory', 'check_cpu', 'execute_command', 'query_log']
  if (!analyzableTools.some(t => props.toolCall.name.toLowerCase().includes(t))) return

  insightLoading.value = true
  try {
    const result = await getSemanticInsight({
      tool_name: props.toolCall.name,
      tool_result: props.toolCall.result,
      host_id: props.hostId
    })
    semanticInsight.value = result
  } catch (error) {
    console.error('获取语义化解读失败:', error)
  } finally {
    insightLoading.value = false
  }
}

// 监听工具状态变化
watch(() => props.toolCall.status, (newStatus) => {
  if (newStatus === 'success') {
    fetchSemanticInsight()
  }
}, { immediate: true })

const toggleInsight = () => {
  insightExpanded.value = !insightExpanded.value
}

// 风险等级相关计算属性
const riskTagType = computed(() => {
  switch (semanticInsight.value?.risk_level) {
    case 'critical': return 'danger'
    case 'warning': return 'warning'
    default: return 'success'
  }
})

const riskLevelText = computed(() => {
  switch (semanticInsight.value?.risk_level) {
    case 'critical': return '需立即处理'
    case 'warning': return '需关注'
    default: return '正常'
  }
})

const riskIcon = computed(() => {
  switch (semanticInsight.value?.risk_level) {
    case 'critical': return CircleClose
    case 'warning': return WarningFilled
    default: return CircleCheck
  }
})

// 多主机支持
const hasMultipleHosts = computed(() => {
  return !!props.toolCall.hostResults && props.toolCall.hostResults.length > 0
})

const getHostStatus = (hostResult: HostResult): 'success' | 'error' | 'warning' | 'running' => {
  if (hostResult.exitCode === 0) return 'success'
  if (hostResult.error) return 'error'
  return 'warning'
}

const toggleHostResult = (hostId: string) => {
  expandedHosts.value[hostId] = !expandedHosts.value[hostId]
}

// 执行时间追踪
const startTime = ref<number>(0)
const endTime = ref<number>(0)
const progress = ref(0)

// 进度动画
let progressTimer: number | null = null

onMounted(() => {
  if (props.toolCall.status === 'running') {
    startTime.value = Date.now()
    startProgressAnimation()
  } else if (props.toolCall.status === 'success' || props.toolCall.status === 'error') {
    endTime.value = Date.now()
    // 假设平均执行时间 2-5 秒
    startTime.value = endTime.value - 3000
  }
})

onUnmounted(() => {
  if (progressTimer) {
    clearInterval(progressTimer)
  }
})

const startProgressAnimation = () => {
  progress.value = 0
  progressTimer = setInterval(() => {
    if (progress.value < 90) {
      progress.value += Math.random() * 10
    }
  }, 500) as unknown as number
}

// 当工具完成时
if (props.toolCall.status !== 'running') {
  progress.value = 100
  if (progressTimer) {
    clearInterval(progressTimer)
  }
}

const statusClass = computed(() => `status-${props.toolCall.status || 'pending'}`)

const toolIcon = computed(() => {
  const name = props.toolCall.name.toLowerCase()
  if (name.includes('execute') || name.includes('command')) return Tools
  if (name.includes('monitor') || name.includes('status')) return Monitor
  if (name.includes('log') || name.includes('file')) return Document
  return Warning
})

const statusTagType = computed(() => {
  switch (props.toolCall.status) {
    case 'running': return 'warning'
    case 'success': return 'success'
    case 'error': return 'danger'
    default: return 'info'
  }
})

const statusText = computed(() => {
  switch (props.toolCall.status) {
    case 'running': return '执行中...'
    case 'success': return '执行成功'
    case 'error': return '执行失败'
    default: return '等待中'
  }
})

const executionTime = computed(() => {
  if (!startTime.value) return null
  const end = endTime.value || Date.now()
  const duration = end - startTime.value
  if (duration < 1000) return `${duration}ms`
  return `${(duration / 1000).toFixed(2)}s`
})

const formattedArguments = computed(() => {
  try {
    const args = JSON.parse(props.toolCall.arguments)
    return JSON.stringify(args, null, 2)
  } catch {
    return props.toolCall.arguments
  }
})

const argCount = computed(() => {
  try {
    const args = JSON.parse(props.toolCall.arguments)
    return Object.keys(args).length
  } catch {
    return 1
  }
})

const resultLineCount = computed(() => {
  if (!props.toolCall.result) return 0
  return props.toolCall.result.split('\n').length
})

const resultSize = computed(() => {
  if (!props.toolCall.result) return null
  const bytes = new Blob([props.toolCall.result]).size
  if (bytes < 1024) return `${bytes}B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)}KB`
  return `${(bytes / (1024 * 1024)).toFixed(2)}MB`
})

const isJsonResult = computed(() => {
  if (!props.toolCall.result) return false
  try {
    JSON.parse(props.toolCall.result)
    return true
  } catch {
    return false
  }
})

const parsedResult = computed(() => {
  if (!isJsonResult.value) return null
  try {
    return JSON.parse(props.toolCall.result!)
  } catch {
    return null
  }
})

const toggleExpand = () => {
  isExpanded.value = !isExpanded.value
}

const toggleSection = (section: 'args' | 'result' | 'error') => {
  expandedSections.value[section] = !expandedSections.value[section]
}

const copyToClipboard = async (text: string, label: string) => {
  try {
    await navigator.clipboard.writeText(text)
    ElMessage.success(`${label}已复制到剪贴板`)
  } catch {
    ElMessage.error('复制失败')
  }
}

const exportResult = () => {
  if (!props.toolCall.result) return

  const blob = new Blob([props.toolCall.result], { type: 'text/plain' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `${props.toolCall.name}_${Date.now()}.txt`
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
  URL.revokeObjectURL(url)

  ElMessage.success('文件已导出')
}

const handleAction = (command: string) => {
  switch (command) {
    case 'copy':
      if (props.toolCall.result) {
        copyToClipboard(props.toolCall.result, '结果')
      }
      break
    case 'export':
      exportResult()
      break
    case 'rerun':
      emit('rerun', props.toolCall)
      break
  }
}
</script>

<style scoped>
.enhanced-tool-card {
  background: #fff;
  border-radius: 12px;
  border: 1px solid #e8e8e8;
  margin: 12px 0;
  overflow: hidden;
  transition: all 0.3s ease;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
}

.enhanced-tool-card:hover {
  box-shadow: 0 6px 20px rgba(0, 0, 0, 0.1);
  transform: translateY(-1px);
}

.enhanced-tool-card.status-running {
  border-left: 4px solid #f59e0b;
  box-shadow: 0 2px 12px rgba(245, 158, 11, 0.15);
}

.enhanced-tool-card.status-success {
  border-left: 4px solid #10b981;
  box-shadow: 0 2px 12px rgba(16, 185, 129, 0.1);
}

.enhanced-tool-card.status-error {
  border-left: 4px solid #ef4444;
  box-shadow: 0 2px 12px rgba(239, 68, 68, 0.1);
}

.tool-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 16px;
  background: #fff;
  border-bottom: 1px solid #f0f0f0;
}

.tool-main-info {
  display: flex;
  align-items: center;
  gap: 12px;
  flex: 1;
  min-width: 0;
}

.tool-icon-wrapper {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 44px;
  height: 44px;
  background: linear-gradient(135deg, #3b82f6, #2563eb);
  border-radius: 10px;
  color: #fff;
  flex-shrink: 0;
}

.status-running .tool-icon-wrapper {
  background: linear-gradient(135deg, #f59e0b, #d97706);
  animation: pulse 2s ease-in-out infinite;
}

.status-success .tool-icon-wrapper {
  background: linear-gradient(135deg, #10b981, #059669);
}

.status-error .tool-icon-wrapper {
  background: linear-gradient(135deg, #ef4444, #dc2626);
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.8; }
}

.tool-meta {
  flex: 1;
  min-width: 0;
}

.tool-name-row {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 4px;
}

.tool-name {
  font-weight: 600;
  color: #1f2937;
  font-size: 14px;
}

.tool-time {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  color: #6b7280;
}

.tool-progress {
  padding: 0 16px;
  background: #fff;
}

.tool-body {
  padding: 12px 16px;
  background: #f9fafb;
}

.tool-section {
  margin-bottom: 12px;
  border-radius: 6px;
  background: #fff;
  border: 1px solid #f0f0f0;
  overflow: hidden;
}

.tool-section:last-child {
  margin-bottom: 0;
}

.section-header {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 12px;
  background: #f9fafb;
  cursor: pointer;
  user-select: none;
  transition: background 0.2s;
}

.section-header:hover {
  background: #f3f4f6;
}

.section-icon {
  transition: transform 0.3s;
}

.section-icon.expanded {
  transform: rotate(90deg);
}

.section-title {
  flex: 1;
  font-weight: 500;
  color: #374151;
  font-size: 13px;
}

.result-stats {
  display: flex;
  gap: 6px;
}

.section-content {
  padding: 12px;
  border-top: 1px solid #f0f0f0;
}

.code-block {
  background: #1e293b;
  border-radius: 6px;
  padding: 12px;
  margin: 0 0 8px 0;
  overflow-x: auto;
  max-height: 300px;
  overflow-y: auto;
}

.code-block code {
  color: #e2e8f0;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 12px;
  line-height: 1.6;
  white-space: pre;
}

.result-code {
  background: #f0fdf4;
  border: 1px solid #bbf7d0;
}

.result-code code {
  color: #166534;
}

.error-section {
  border-color: #fecaca;
}

.error-content {
  background: #fef2f2;
}

.error-code {
  background: #fee2e2;
  border-color: #fca5a5;
}

.error-code code {
  color: #991b1b;
}

.section-actions {
  display: flex;
  gap: 8px;
  justify-content: flex-end;
}

.collapse-trigger {
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 8px;
  background: #fff;
  cursor: pointer;
  transition: all 0.3s;
}

.collapse-trigger:hover {
  background: #f9fafb;
}

.collapse-trigger .el-icon {
  transition: transform 0.3s;
  color: #6b7280;
}

.collapse-trigger .el-icon.expanded {
  transform: rotate(180deg);
}

.result-viewer {
  max-height: 400px;
  overflow-y: auto;
}

.host-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 8px;
}

.multi-host-results {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.host-result-item {
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  overflow: hidden;
  background: #fff;
  transition: all 0.2s;
}

.host-result-item.has-error {
  border-color: #fca5a5;
  background: #fef2f2;
}

.host-result-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px;
  cursor: pointer;
  background: #f9fafb;
  transition: background 0.2s;
}

.host-result-header:hover {
  background: #f3f4f6;
}

.host-result-item.has-error .host-result-header {
  background: #fee2e2;
}

.host-result-item.has-error .host-result-header:hover {
  background: #fecaca;
}

.expand-icon {
  transition: transform 0.3s;
  color: #6b7280;
}

.expand-icon.expanded {
  transform: rotate(90deg);
}

.host-result-content {
  padding: 12px;
  border-top: 1px solid #e5e7eb;
}

.host-result-item.has-error .host-result-content {
  border-top-color: #fca5a5;
}

.host-error {
  margin-top: 12px;
}

/* AI 深度解读样式 */
.ai-insight-section {
  border-top: 1px solid #f0f0f0;
  background: linear-gradient(135deg, #f0f9ff 0%, #e0f2fe 100%);
}

.insight-header {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 16px;
  cursor: pointer;
  transition: background 0.2s;
}

.insight-header:hover {
  background: rgba(59, 130, 246, 0.05);
}

.insight-icon {
  color: #3b82f6;
}

.insight-label {
  font-size: 13px;
  font-weight: 600;
  color: #1e40af;
  flex: 1;
}

.risk-tag {
  margin-left: auto;
}

.expand-arrow {
  transition: transform 0.3s;
  color: #6b7280;
}

.expand-arrow.expanded {
  transform: rotate(90deg);
}

.insight-content {
  padding: 12px 16px;
  border-top: 1px solid rgba(59, 130, 246, 0.1);
}

.insight-loading {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #6b7280;
  font-size: 13px;
}

.insight-summary {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  font-size: 14px;
  color: #1f2937;
  line-height: 1.5;
  margin-bottom: 8px;
}

.summary-icon {
  flex-shrink: 0;
  margin-top: 2px;
}

.summary-icon.risk-normal {
  color: #10b981;
}

.summary-icon.risk-warning {
  color: #f59e0b;
}

.summary-icon.risk-critical {
  color: #ef4444;
}

.insight-trend,
.insight-recommendation {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: #6b7280;
  margin-top: 6px;
  padding: 6px 10px;
  background: rgba(255, 255, 255, 0.6);
  border-radius: 6px;
}

.insight-trend .el-icon {
  color: #8b5cf6;
}

.insight-recommendation .el-icon {
  color: #10b981;
}

.insight-metrics {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 10px;
}

.metric-item {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 4px 10px;
  border-radius: 4px;
  font-size: 12px;
  background: #fff;
  border: 1px solid #e5e7eb;
}

.metric-item.metric-normal {
  border-color: #10b981;
  background: #f0fdf4;
}

.metric-item.metric-warning {
  border-color: #f59e0b;
  background: #fffbeb;
}

.metric-item.metric-critical {
  border-color: #ef4444;
  background: #fef2f2;
}

.metric-name {
  color: #6b7280;
}

.metric-value {
  font-weight: 600;
  color: #1f2937;
}

.metric-threshold {
  color: #9ca3af;
  font-size: 11px;
}
</style>
