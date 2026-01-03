<template>
  <el-tooltip
    :content="tooltipContent"
    placement="top"
    :disabled="!showDetails"
  >
    <div
      class="host-tag-badge"
      :class="[`status-${status}`, { clickable: showDetails }]"
      @click="handleClick"
    >
      <el-icon class="status-icon" :size="14">
        <component :is="statusIcon" />
      </el-icon>
      <span class="host-id">{{ hostId }}</span>
      <el-badge
        v-if="result"
        :value="result.exitCode"
        :type="badgeType"
        class="exit-code-badge"
      />
    </div>
  </el-tooltip>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import {
  SuccessFilled,
  CircleCloseFilled,
  WarningFilled,
  Loading
} from '@element-plus/icons-vue'

export interface ExecutionResult {
  exitCode: number
  output: string
  error?: string
  duration: number
  timestamp: number
}

interface Props {
  hostId: string
  status: 'success' | 'error' | 'warning' | 'running'
  result?: ExecutionResult
  showDetails?: boolean
}

interface Emits {
  (e: 'click', hostId: string): void
  (e: 'show-detail', result: ExecutionResult): void
}

const props = withDefaults(defineProps<Props>(), {
  showDetails: true
})

const emit = defineEmits<Emits>()

const statusIcon = computed(() => {
  const iconMap = {
    success: SuccessFilled,
    error: CircleCloseFilled,
    warning: WarningFilled,
    running: Loading
  }
  return iconMap[props.status]
})

const badgeType = computed(() => {
  if (!props.result) return 'info'
  return props.result.exitCode === 0 ? 'success' : 'danger'
})

const tooltipContent = computed(() => {
  if (!props.result) return `主机: ${props.hostId}`

  const lines = [
    `主机: ${props.hostId}`,
    `退出码: ${props.result.exitCode}`,
    `耗时: ${props.result.duration}ms`,
    `时间: ${new Date(props.result.timestamp).toLocaleString()}`
  ]

  if (props.result.error) {
    lines.push(`错误: ${props.result.error}`)
  }

  return lines.join('\n')
})

const handleClick = () => {
  if (!props.showDetails) return

  emit('click', props.hostId)
  if (props.result) {
    emit('show-detail', props.result)
  }
}
</script>

<style scoped lang="scss">
.host-tag-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 500;
  transition: all 0.2s ease;
  user-select: none;

  &.clickable {
    cursor: pointer;

    &:hover {
      transform: translateY(-2px);
      box-shadow: 0 4px 8px rgba(0, 0, 0, 0.1);
    }
  }

  .status-icon {
    flex-shrink: 0;
  }

  .host-id {
    font-family: 'Consolas', 'Monaco', monospace;
    white-space: nowrap;
  }

  .exit-code-badge {
    :deep(.el-badge__content) {
      font-size: 10px;
      height: 16px;
      line-height: 16px;
      padding: 0 4px;
    }
  }

  &.status-success {
    background: rgba(103, 194, 58, 0.1);
    color: #67c23a;
    border: 1px solid #67c23a;

    .status-icon {
      color: #67c23a;
    }
  }

  &.status-error {
    background: rgba(245, 108, 108, 0.1);
    color: #f56c6c;
    border: 1px solid #f56c6c;

    .status-icon {
      color: #f56c6c;
    }
  }

  &.status-warning {
    background: rgba(230, 162, 60, 0.1);
    color: #e6a23c;
    border: 1px solid #e6a23c;

    .status-icon {
      color: #e6a23c;
    }
  }

  &.status-running {
    background: rgba(64, 158, 255, 0.1);
    color: #409eff;
    border: 1px solid #409eff;

    .status-icon {
      color: #409eff;
      animation: spin 1s linear infinite;
    }
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
  .host-tag-badge {
    font-size: 11px;
    padding: 4px 8px;
    gap: 4px;

    .status-icon {
      font-size: 12px;
    }
  }
}
</style>
