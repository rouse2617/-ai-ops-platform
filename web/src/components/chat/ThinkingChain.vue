<template>
  <div class="thinking-chain" :class="{ 'show-timeline': showTimeline }">
    <div class="chain-header">
      <el-icon class="header-icon"><BrainFilled /></el-icon>
      <span class="header-title">思考过程</span>
      <el-tag v-if="currentStep !== undefined" type="primary" size="small">
        步骤 {{ currentStep }} / {{ steps.length }}
      </el-tag>
    </div>

    <div class="steps-container">
      <div
        v-for="(step, index) in steps"
        :key="step.id"
        class="step-item"
        :class="[
          `status-${step.status}`,
          { active: currentStep === step.step, completed: step.status === 'completed' }
        ]"
      >
        <div class="step-icon">
          <el-icon v-if="step.status === 'completed'" :size="16">
            <CircleCheckFilled />
          </el-icon>
          <el-icon v-else-if="isActiveStep(step)" :size="16" class="spinning">
            <Loading />
          </el-icon>
          <span v-else class="step-number">{{ step.step }}</span>
        </div>

        <div class="step-content">
          <div class="step-header">
            <span class="step-title">{{ getStepTitle(step.status) }}</span>
            <span v-if="step.duration" class="step-duration">
              {{ formatDuration(step.duration) }}
            </span>
          </div>

          <div v-if="step.content" class="step-description">
            {{ step.content }}
          </div>

          <div v-if="step.toolCalls && step.toolCalls.length > 0" class="tool-calls">
            <div
              v-for="(tool, toolIndex) in step.toolCalls"
              :key="toolIndex"
              class="tool-call-item"
              :class="`tool-status-${tool.status}`"
            >
              <div class="tool-header">
                <el-icon class="tool-icon"><Tools /></el-icon>
                <span class="tool-name">{{ tool.tool }}</span>
                <el-tag :type="getToolTagType(tool.status)" size="small">
                  {{ getToolStatusText(tool.status) }}
                </el-tag>
              </div>

              <div v-if="tool.params" class="tool-params">
                <el-text size="small" type="info">参数:</el-text>
                <pre class="params-code">{{ formatParams(tool.params) }}</pre>
              </div>

              <div v-if="tool.result" class="tool-result">
                <el-text size="small" type="success">结果:</el-text>
                <CollapsibleContent
                  :content="tool.result"
                  :max-lines="5"
                  syntax="json"
                />
              </div>

              <div v-if="tool.error" class="tool-error">
                <el-alert type="error" :closable="false" show-icon>
                  {{ tool.error }}
                </el-alert>
              </div>
            </div>
          </div>

          <div v-if="showTimeline" class="step-timestamp">
            {{ formatTimestamp(step.timestamp) }}
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import {
  BrainFilled,
  CircleCheckFilled,
  Loading,
  Tools
} from '@element-plus/icons-vue'
import CollapsibleContent from './CollapsibleContent.vue'

export interface ToolCallInfo {
  id: string
  tool: string
  params: string
  status: 'pending' | 'running' | 'success' | 'error'
  result?: string
  error?: string
}

export interface ThinkingStep {
  id: string
  step: number
  status: 'calling_llm' | 'executing_tools' | 'analyzing_results' | 'completed'
  content: string
  timestamp: number
  duration?: number
  toolCalls?: ToolCallInfo[]
}

interface Props {
  steps: ThinkingStep[]
  currentStep?: number
  showTimeline?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  showTimeline: true
})

const isActiveStep = (step: ThinkingStep) => {
  return props.currentStep === step.step && step.status !== 'completed'
}

const getStepTitle = (status: ThinkingStep['status']) => {
  const titleMap = {
    calling_llm: '调用 AI 模型',
    executing_tools: '执行工具',
    analyzing_results: '分析结果',
    completed: '完成'
  }
  return titleMap[status]
}

const getToolTagType = (status: ToolCallInfo['status']) => {
  const typeMap = {
    pending: 'info',
    running: 'primary',
    success: 'success',
    error: 'danger'
  }
  return typeMap[status] as 'info' | 'primary' | 'success' | 'danger'
}

const getToolStatusText = (status: ToolCallInfo['status']) => {
  const textMap = {
    pending: '等待中',
    running: '执行中',
    success: '成功',
    error: '失败'
  }
  return textMap[status]
}

const formatDuration = (duration: number) => {
  if (duration < 1000) return `${duration}ms`
  return `${(duration / 1000).toFixed(2)}s`
}

const formatTimestamp = (timestamp: number) => {
  return new Date(timestamp).toLocaleTimeString('zh-CN')
}

const formatParams = (params: string) => {
  try {
    const parsed = JSON.parse(params)
    return JSON.stringify(parsed, null, 2)
  } catch {
    return params
  }
}
</script>

<style scoped lang="scss">
.thinking-chain {
  background: #f7f8fa;
  border-radius: 12px;
  padding: 20px;
  margin: 16px 0;

  .chain-header {
    display: flex;
    align-items: center;
    gap: 12px;
    margin-bottom: 20px;
    padding-bottom: 16px;
    border-bottom: 2px solid #e4e7ed;

    .header-icon {
      color: #409eff;
      font-size: 20px;
    }

    .header-title {
      font-size: 16px;
      font-weight: 600;
      color: #303133;
      flex: 1;
    }
  }
}

.steps-container {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.step-item {
  position: relative;
  display: flex;
  gap: 16px;
  padding-left: 40px;

  &::before {
    content: '';
    position: absolute;
    left: 16px;
    top: 32px;
    bottom: -16px;
    width: 2px;
    background: #dcdfe6;
  }

  &:last-child::before {
    display: none;
  }

  &.active::before {
    background: #409eff;
  }

  &.completed::before {
    background: #67c23a;
  }

  .step-icon {
    position: absolute;
    left: 0;
    top: 0;
    width: 32px;
    height: 32px;
    border-radius: 50%;
    background: white;
    border: 2px solid #dcdfe6;
    display: flex;
    align-items: center;
    justify-content: center;
    font-weight: 600;
    color: #909399;
    transition: all 0.3s ease;

    .spinning {
      animation: spin 1s linear infinite;
    }
  }

  &.active .step-icon {
    background: #409eff;
    border-color: #409eff;
    color: white;
  }

  &.completed .step-icon {
    background: #67c23a;
    border-color: #67c23a;
    color: white;
  }

  .step-content {
    flex: 1;
    min-width: 0;

    .step-header {
      display: flex;
      align-items: center;
      justify-content: space-between;
      margin-bottom: 8px;

      .step-title {
        font-weight: 600;
        color: #303133;
        font-size: 14px;
      }

      .step-duration {
        font-size: 12px;
        color: #909399;
      }
    }

    .step-description {
      color: #606266;
      font-size: 14px;
      line-height: 1.6;
      margin-bottom: 12px;
    }

    .step-timestamp {
      font-size: 12px;
      color: #c0c4cc;
      margin-top: 8px;
    }
  }
}

.tool-calls {
  display: flex;
  flex-direction: column;
  gap: 12px;
  margin-top: 12px;
}

.tool-call-item {
  background: white;
  border-radius: 8px;
  padding: 12px;
  border: 1px solid #e4e7ed;
  transition: all 0.2s ease;

  &:hover {
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
  }

  .tool-header {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 8px;

    .tool-icon {
      color: #409eff;
    }

    .tool-name {
      font-weight: 500;
      color: #303133;
      font-family: 'Consolas', 'Monaco', monospace;
      flex: 1;
    }
  }

  .tool-params,
  .tool-result {
    margin-top: 8px;

    .params-code {
      background: #f5f7fa;
      padding: 8px;
      border-radius: 4px;
      font-size: 12px;
      font-family: 'Consolas', 'Monaco', monospace;
      overflow-x: auto;
      margin-top: 4px;
    }
  }

  .tool-error {
    margin-top: 8px;
  }
}

@keyframes spin {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

@media (max-width: 768px) {
  .thinking-chain {
    padding: 16px;

    .chain-header {
      flex-wrap: wrap;
      gap: 8px;
    }
  }

  .step-item {
    padding-left: 32px;

    .step-icon {
      width: 24px;
      height: 24px;
      font-size: 12px;
    }

    &::before {
      left: 12px;
    }
  }
}
</style>
