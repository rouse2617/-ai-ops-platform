<template>
  <div class="prometheus-integration">
    <!-- 语义化告警卡片 -->
    <SemanticAlertCard
      v-if="semanticAlert"
      :title="semanticAlert.title"
      :summary="semanticAlert.summary"
      :description="semanticAlert.description"
      :root-cause="semanticAlert.rootCause"
      :suggestion="semanticAlert.suggestion"
      :severity="semanticAlert.severity"
      :confidence="semanticAlert.confidence"
      :related-metrics="semanticAlert.relatedMetrics"
      :actions="semanticAlert.actions"
      @action="handleAlertAction"
    />

    <!-- 自然语言查询面板 -->
    <NLQueryPanel
      v-if="showNLQuery"
      :initial-query="nlQueryText"
      @query-result="handleQueryResult"
      @close="showNLQuery = false"
    />

    <!-- Playbook 执行卡片 -->
    <PlaybookExecutorCard
      v-if="activePlaybook"
      :playbook="activePlaybook"
      :execution="playbookExecution"
      @execute="executePlaybook"
      @cancel="cancelPlaybook"
    />

    <!-- 异常检测结果 -->
    <div v-if="anomalyResult" class="anomaly-result-card">
      <div class="anomaly-header">
        <el-icon :class="['anomaly-icon', `severity-${anomalyResult.severity > 0.7 ? 'high' : anomalyResult.severity > 0.4 ? 'medium' : 'low'}`]">
          <WarningFilled v-if="anomalyResult.is_anomaly" />
          <CircleCheck v-else />
        </el-icon>
        <span class="anomaly-title">{{ anomalyResult.is_anomaly ? '检测到异常' : '指标正常' }}</span>
      </div>
      <div class="anomaly-body">
        <div class="anomaly-metric">
          <span class="label">偏离度:</span>
          <span class="value">{{ (anomalyResult.deviation * 100).toFixed(1) }}%</span>
        </div>
        <div class="anomaly-metric">
          <span class="label">严重程度:</span>
          <el-progress
            :percentage="anomalyResult.severity * 100"
            :color="getSeverityColor(anomalyResult.severity)"
            :stroke-width="8"
          />
        </div>
        <div class="anomaly-recommendation" v-if="anomalyResult.recommendation">
          <el-icon><Promotion /></el-icon>
          <span>{{ anomalyResult.recommendation }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, defineAsyncComponent } from 'vue'
import { WarningFilled, CircleCheck, Promotion } from '@element-plus/icons-vue'
import type { AnomalyResult, Playbook, PlaybookExecution } from '@/api/prometheus'

// 懒加载 Prometheus 组件
const SemanticAlertCard = defineAsyncComponent(() => import('@/components/prometheus/SemanticAlertCard.vue'))
const NLQueryPanel = defineAsyncComponent(() => import('@/components/prometheus/NLQueryPanel.vue'))
const PlaybookExecutorCard = defineAsyncComponent(() => import('@/components/prometheus/PlaybookExecutorCard.vue'))

interface SemanticAlertData {
  title: string
  summary: string
  description: string
  rootCause: string
  suggestion: string
  severity: string
  confidence: number
  relatedMetrics?: Record<string, number>
  actions?: Array<{ id: string; label: string; dangerous?: boolean }>
}

const props = defineProps<{
  messageContent?: string
  toolResults?: unknown[]
}>()

const emit = defineEmits<{
  'send-message': [message: string]
  'execute-action': [action: { id: string; command?: string }]
}>()

// 状态
const semanticAlert = ref<SemanticAlertData | null>(null)
const showNLQuery = ref(false)
const nlQueryText = ref('')
const activePlaybook = ref<Playbook | null>(null)
const playbookExecution = ref<PlaybookExecution | null>(null)
const anomalyResult = ref<AnomalyResult | null>(null)

// 监听消息内容，检测 Prometheus 相关意图
watch(() => props.messageContent, (content) => {
  if (!content) return
  detectPrometheusIntent(content)
})

// 检测 Prometheus 相关意图
function detectPrometheusIntent(content: string) {
  const lowerContent = content.toLowerCase()

  // 检测自然语言查询意图
  const queryPatterns = [
    /查询|查看|显示|获取/,
    /cpu|内存|磁盘|网络|负载/,
    /使用率|占用|流量|连接数/,
    /过去|最近|小时|分钟/
  ]

  const matchCount = queryPatterns.filter(p => p.test(lowerContent)).length
  if (matchCount >= 2) {
    nlQueryText.value = content
    showNLQuery.value = true
  }
}

// 处理告警操作
function handleAlertAction(action: { id: string; label: string; command?: string }) {
  if (action.command) {
    emit('execute-action', action)
  } else {
    emit('send-message', `执行操作: ${action.label}`)
  }
}

// 处理查询结果
function handleQueryResult(result: { promql: string; data: unknown[] }) {
  emit('send-message', `查询完成: ${result.promql}`)
  showNLQuery.value = false
}

// 执行 Playbook
function executePlaybook(playbookId: string) {
  emit('send-message', `执行排障剧本: ${playbookId}`)
}

// 取消 Playbook
function cancelPlaybook() {
  activePlaybook.value = null
  playbookExecution.value = null
}

// 获取严重程度颜色
function getSeverityColor(severity: number): string {
  if (severity > 0.7) return '#f56c6c'
  if (severity > 0.4) return '#e6a23c'
  return '#67c23a'
}

// 暴露方法供父组件调用
defineExpose({
  showSemanticAlert: (data: SemanticAlertData) => {
    semanticAlert.value = data
  },
  showNLQueryPanel: (query?: string) => {
    nlQueryText.value = query || ''
    showNLQuery.value = true
  },
  showPlaybook: (playbook: Playbook) => {
    activePlaybook.value = playbook
  },
  showAnomalyResult: (result: AnomalyResult) => {
    anomalyResult.value = result
  },
  clearAll: () => {
    semanticAlert.value = null
    showNLQuery.value = false
    activePlaybook.value = null
    anomalyResult.value = null
  }
})
</script>

<style scoped>
.prometheus-integration {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.anomaly-result-card {
  background: #fff;
  border-radius: 12px;
  border: 1px solid #e8e8e8;
  padding: 16px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
}

.anomaly-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
}

.anomaly-icon {
  font-size: 24px;
}

.anomaly-icon.severity-high {
  color: #f56c6c;
}

.anomaly-icon.severity-medium {
  color: #e6a23c;
}

.anomaly-icon.severity-low {
  color: #67c23a;
}

.anomaly-title {
  font-size: 16px;
  font-weight: 600;
  color: #303133;
}

.anomaly-body {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.anomaly-metric {
  display: flex;
  align-items: center;
  gap: 8px;
}

.anomaly-metric .label {
  font-size: 13px;
  color: #909399;
  min-width: 70px;
}

.anomaly-metric .value {
  font-size: 14px;
  font-weight: 600;
  color: #303133;
}

.anomaly-recommendation {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 12px;
  background: rgba(103, 194, 58, 0.1);
  border-radius: 8px;
  font-size: 13px;
  color: #67c23a;
}
</style>
