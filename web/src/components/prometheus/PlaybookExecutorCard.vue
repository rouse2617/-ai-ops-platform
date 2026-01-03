<template>
  <div class="playbook-executor">
    <div class="playbook-header">
      <div class="playbook-icon">
        <el-icon :size="24"><Tools /></el-icon>
      </div>
      <div class="playbook-info">
        <h3>{{ playbook.name }}</h3>
        <p>{{ playbook.description }}</p>
      </div>
      <el-tag type="warning" v-if="execution?.status === 'waiting_approval'">
        等待确认
      </el-tag>
    </div>

    <div class="execution-steps" v-if="execution">
      <div class="step-timeline">
        <div
          v-for="(step, index) in execution.steps"
          :key="index"
          class="step-item"
          :class="[`status-${step.status}`]"
        >
          <div class="step-indicator">
            <el-icon v-if="step.status === 'completed'"><Check /></el-icon>
            <el-icon v-else-if="step.status === 'failed'"><Close /></el-icon>
            <el-icon v-else-if="step.status === 'running'" class="is-loading"><Loading /></el-icon>
            <span v-else>{{ index + 1 }}</span>
          </div>
          <div class="step-content">
            <div class="step-name">{{ step.step_name }}</div>
            <div class="step-output" v-if="step.output">
              <template v-if="typeof step.output === 'object'">
                <code v-if="step.output.command">{{ step.output.command }}</code>
                <p v-if="step.output.description">{{ step.output.description }}</p>
              </template>
              <pre v-else>{{ formatOutput(step.output) }}</pre>
            </div>
            <div class="step-error" v-if="step.error">
              <el-icon><Warning /></el-icon>
              {{ step.error }}
            </div>
          </div>
        </div>
      </div>
    </div>

    <div class="playbook-summary" v-if="execution?.summary">
      <div class="summary-content" v-html="renderedSummary"></div>
    </div>

    <div class="playbook-actions" v-if="execution?.actions && execution.actions.length > 0">
      <p class="actions-title">请选择要执行的操作：</p>
      <div class="action-buttons">
        <el-button
          v-for="action in execution.actions"
          :key="action.id"
          :type="action.dangerous ? 'danger' : 'primary'"
          :plain="!action.dangerous"
          @click="executeAction(action)"
          :loading="executingAction === action.id"
        >
          <el-icon v-if="action.dangerous"><Warning /></el-icon>
          {{ action.label }}
        </el-button>
        <el-button @click="$emit('dismiss')">稍后处理</el-button>
      </div>
      <p class="action-hint" v-if="selectedAction?.dangerous">
        <el-icon><Warning /></el-icon>
        此操作可能影响系统运行，请确认后执行
      </p>
    </div>

    <div class="playbook-footer">
      <span class="instance">实例: {{ execution?.instance }}</span>
      <span class="time">{{ formatTime(execution?.start_time) }}</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { marked } from 'marked'
import { ElMessageBox, ElMessage } from 'element-plus'
import { Tools, Check, Close, Loading, Warning } from '@element-plus/icons-vue'
import { request } from '@/api/request'

interface PlaybookStep {
  name: string
  type: string
  command?: string
  query?: string
  description: string
  requires_approval?: boolean
  dangerous?: boolean
}

interface Playbook {
  id: string
  name: string
  description: string
  steps: PlaybookStep[]
}

interface StepExecution {
  step_name: string
  status: string
  output?: unknown
  error?: string
}

interface QuickAction {
  id: string
  label: string
  command?: string
  description?: string
  dangerous?: boolean
}

interface PlaybookExecution {
  id: string
  playbook_id: string
  instance: string
  status: string
  steps: StepExecution[]
  actions: QuickAction[]
  summary: string
  start_time: string
}

interface Props {
  playbook: Playbook
  execution?: PlaybookExecution
}

const props = defineProps<Props>()

const emit = defineEmits<{
  (e: 'action-executed', action: QuickAction, result: unknown): void
  (e: 'dismiss'): void
}>()

const executingAction = ref<string | null>(null)
const selectedAction = ref<QuickAction | null>(null)

const renderedSummary = computed(() => {
  if (!props.execution?.summary) return ''
  try {
    return marked(props.execution.summary)
  } catch {
    return props.execution.summary
  }
})

function formatOutput(output: unknown): string {
  if (typeof output === 'string') return output
  return JSON.stringify(output, null, 2)
}

function formatTime(time?: string): string {
  if (!time) return ''
  return new Date(time).toLocaleString('zh-CN')
}

async function executeAction(action: QuickAction) {
  selectedAction.value = action

  if (action.dangerous) {
    try {
      await ElMessageBox.confirm(
        `确定要执行 "${action.label}" 吗？\n\n${action.description || '此操作可能影响系统运行'}`,
        '确认执行',
        {
          confirmButtonText: '确定执行',
          cancelButtonText: '取消',
          type: 'warning'
        }
      )
    } catch {
      selectedAction.value = null
      return
    }
  }

  executingAction.value = action.id

  try {
    // 这里调用后端执行操作
    const result = await request.post('/prometheus/semantic/playbook/action', {
      execution_id: props.execution?.id,
      action_id: action.id,
      command: action.command
    })

    ElMessage.success(`${action.label} 执行成功`)
    emit('action-executed', action, result)
  } catch (e) {
    ElMessage.error(`执行失败: ${e}`)
  } finally {
    executingAction.value = null
    selectedAction.value = null
  }
}
</script>

<style scoped>
.playbook-executor {
  background: #fff;
  border-radius: 12px;
  border: 1px solid #e6a23c;
  overflow: hidden;
}

.playbook-header {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 16px;
  background: linear-gradient(135deg, #fffbeb 0%, #fef3c7 100%);
  border-bottom: 1px solid #fcd34d;
}

.playbook-icon {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  background: rgba(245, 158, 11, 0.1);
  display: flex;
  align-items: center;
  justify-content: center;
  color: #f59e0b;
}

.playbook-info {
  flex: 1;
}

.playbook-info h3 {
  margin: 0 0 4px 0;
  font-size: 16px;
  color: #303133;
}

.playbook-info p {
  margin: 0;
  font-size: 13px;
  color: #606266;
}

.execution-steps {
  padding: 16px;
  border-bottom: 1px solid #ebeef5;
}

.step-timeline {
  position: relative;
}

.step-item {
  display: flex;
  gap: 12px;
  padding-bottom: 16px;
  position: relative;
}

.step-item:not(:last-child)::before {
  content: '';
  position: absolute;
  left: 15px;
  top: 32px;
  bottom: 0;
  width: 2px;
  background: #e4e7ed;
}

.step-item.status-completed:not(:last-child)::before {
  background: #67c23a;
}

.step-indicator {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  background: #f5f7fa;
  border: 2px solid #e4e7ed;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 14px;
  color: #909399;
  flex-shrink: 0;
  z-index: 1;
}

.status-completed .step-indicator {
  background: #67c23a;
  border-color: #67c23a;
  color: #fff;
}

.status-failed .step-indicator {
  background: #f56c6c;
  border-color: #f56c6c;
  color: #fff;
}

.status-running .step-indicator {
  background: #409eff;
  border-color: #409eff;
  color: #fff;
}

.step-content {
  flex: 1;
  padding-top: 4px;
}

.step-name {
  font-weight: 500;
  color: #303133;
  margin-bottom: 8px;
}

.step-output {
  background: #f5f7fa;
  border-radius: 6px;
  padding: 12px;
  font-size: 13px;
}

.step-output code {
  display: block;
  font-family: monospace;
  color: #303133;
  margin-bottom: 8px;
}

.step-output p {
  margin: 0;
  color: #606266;
}

.step-output pre {
  margin: 0;
  white-space: pre-wrap;
  word-break: break-all;
  font-size: 12px;
}

.step-error {
  display: flex;
  align-items: center;
  gap: 6px;
  color: #f56c6c;
  font-size: 13px;
  margin-top: 8px;
}

.playbook-summary {
  padding: 16px;
  border-bottom: 1px solid #ebeef5;
}

.summary-content {
  font-size: 14px;
  line-height: 1.6;
  color: #606266;
}

.summary-content :deep(strong) {
  color: #303133;
}

.playbook-actions {
  padding: 16px;
  background: #fafafa;
}

.actions-title {
  margin: 0 0 12px 0;
  font-size: 14px;
  color: #606266;
}

.action-buttons {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.action-hint {
  display: flex;
  align-items: center;
  gap: 6px;
  margin: 12px 0 0 0;
  font-size: 12px;
  color: #e6a23c;
}

.playbook-footer {
  display: flex;
  justify-content: space-between;
  padding: 12px 16px;
  font-size: 12px;
  color: #909399;
  background: #fafafa;
}

.is-loading {
  animation: rotating 1s linear infinite;
}

@keyframes rotating {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}
</style>
