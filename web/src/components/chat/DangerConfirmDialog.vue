<template>
  <el-dialog
    v-model="dialogVisible"
    :title="dialogTitle"
    width="600px"
    :close-on-click-modal="false"
    :close-on-press-escape="false"
    class="danger-confirm-dialog"
    @close="handleCancel"
  >
    <template #header>
      <div class="dialog-header">
        <el-icon class="warning-icon" :size="24"><WarningFilled /></el-icon>
        <span class="header-title">{{ dialogTitle }}</span>
      </div>
    </template>

    <div class="dialog-content">
      <el-alert
        :type="operation.riskLevel === 'critical' ? 'error' : 'warning'"
        :closable="false"
        show-icon
      >
        <template #title>
          <span class="alert-title">
            {{ operation.riskLevel === 'critical' ? '极高风险操作' : '高风险操作' }}
          </span>
        </template>
      </el-alert>

      <div class="operation-info">
        <div class="info-item">
          <span class="label">执行命令:</span>
          <el-tag type="danger" effect="dark" class="command-tag">
            {{ operation.command }}
          </el-tag>
        </div>

        <div class="info-item">
          <span class="label">影响主机:</span>
          <el-tag type="warning">{{ hostCount }} 台</el-tag>
        </div>

        <div class="info-item">
          <span class="label">预估影响:</span>
          <span class="impact-text">{{ estimatedImpact }}</span>
        </div>

        <div v-if="operation.reason" class="info-item">
          <span class="label">风险原因:</span>
          <span class="reason-text">{{ operation.reason }}</span>
        </div>
      </div>

      <div v-if="operation.affectedHosts.length > 0" class="affected-hosts">
        <div class="hosts-header">
          <span class="label">受影响主机列表:</span>
        </div>
        <div class="hosts-list">
          <el-tag
            v-for="host in operation.affectedHosts"
            :key="host"
            type="info"
            size="small"
            class="host-tag"
          >
            {{ host }}
          </el-tag>
        </div>
      </div>

      <div class="confirmation-input">
        <el-checkbox v-model="understood" size="large">
          我已理解此操作的风险，并确认执行
        </el-checkbox>
      </div>

      <div v-if="operation.riskLevel === 'critical'" class="critical-warning">
        <el-input
          v-model="confirmText"
          placeholder="请输入 'CONFIRM' 以确认执行"
          clearable
        />
      </div>
    </div>

    <template #footer>
      <div class="dialog-footer">
        <div v-if="countdown > 0" class="countdown-notice">
          {{ countdown }}秒后可确认
        </div>
        <el-button @click="handleCancel" size="large">
          取消 (ESC)
        </el-button>
        <el-button
          type="danger"
          @click="handleConfirm"
          :disabled="!canConfirm"
          :loading="confirming"
          size="large"
        >
          确认执行
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch, onUnmounted } from 'vue'
import { WarningFilled } from '@element-plus/icons-vue'

export interface DangerOperation {
  id: string
  command: string
  type: 'delete' | 'restart' | 'stop' | 'modify'
  riskLevel: 'high' | 'critical'
  affectedHosts: string[]
  reason?: string
}

interface Props {
  visible: boolean
  operation: DangerOperation
  hostCount: number
  estimatedImpact: string
}

interface Emits {
  (e: 'update:visible', value: boolean): void
  (e: 'confirm', operation: DangerOperation): void
  (e: 'cancel'): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

const dialogVisible = computed({
  get: () => props.visible,
  set: (val) => emit('update:visible', val)
})

const understood = ref(false)
const confirmText = ref('')
const confirming = ref(false)
const countdown = ref(5)
let countdownTimer: number | null = null

const dialogTitle = computed(() => {
  const typeMap = {
    delete: '删除操作确认',
    restart: '重启操作确认',
    stop: '停止操作确认',
    modify: '修改操作确认'
  }
  return typeMap[props.operation.type] || '危险操作确认'
})

const canConfirm = computed(() => {
  if (countdown.value > 0) return false
  if (!understood.value) return false
  if (props.operation.riskLevel === 'critical') {
    return confirmText.value.toUpperCase() === 'CONFIRM'
  }
  return true
})

const startCountdown = () => {
  countdown.value = 5
  countdownTimer = window.setInterval(() => {
    countdown.value--
    if (countdown.value <= 0 && countdownTimer) {
      clearInterval(countdownTimer)
      countdownTimer = null
    }
  }, 1000)
}

const stopCountdown = () => {
  if (countdownTimer) {
    clearInterval(countdownTimer)
    countdownTimer = null
  }
}

const handleConfirm = async () => {
  if (!canConfirm.value) return

  confirming.value = true
  try {
    emit('confirm', props.operation)
    resetForm()
  } finally {
    confirming.value = false
  }
}

const handleCancel = () => {
  emit('cancel')
  resetForm()
}

const handleKeydown = (e: KeyboardEvent) => {
  if (e.key === 'Escape') {
    handleCancel()
  }
}

const resetForm = () => {
  understood.value = false
  confirmText.value = ''
  stopCountdown()
}

watch(() => props.visible, (val) => {
  if (val) {
    startCountdown()
    document.addEventListener('keydown', handleKeydown)
  } else {
    resetForm()
    document.removeEventListener('keydown', handleKeydown)
  }
})

onUnmounted(() => {
  stopCountdown()
  document.removeEventListener('keydown', handleKeydown)
})
</script>

<style scoped lang="scss">
.danger-confirm-dialog {
  :deep(.el-dialog__header) {
    background: linear-gradient(135deg, #f56c6c 0%, #c71585 100%);
    padding: 20px 24px;
    margin: 0;
  }

  :deep(.el-dialog__body) {
    padding: 24px;
  }

  :deep(.el-dialog__footer) {
    padding: 16px 24px;
    border-top: 1px solid #f0f0f0;
  }
}

.dialog-header {
  display: flex;
  align-items: center;
  gap: 12px;
  color: white;

  .warning-icon {
    animation: pulse 2s infinite;
  }

  .header-title {
    font-size: 18px;
    font-weight: 600;
  }
}

@keyframes pulse {
  0%, 100% {
    opacity: 1;
  }
  50% {
    opacity: 0.6;
  }
}

.dialog-content {
  display: flex;
  flex-direction: column;
  gap: 20px;

  .alert-title {
    font-weight: 600;
    font-size: 15px;
  }
}

.operation-info {
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 16px;
  background: #f7f8fa;
  border-radius: 8px;

  .info-item {
    display: flex;
    align-items: center;
    gap: 12px;

    .label {
      font-weight: 500;
      color: #606266;
      min-width: 80px;
    }

    .command-tag {
      font-family: 'Consolas', 'Monaco', monospace;
      font-size: 13px;
    }

    .impact-text,
    .reason-text {
      color: #303133;
      flex: 1;
    }
  }
}

.affected-hosts {
  .hosts-header {
    margin-bottom: 12px;

    .label {
      font-weight: 500;
      color: #606266;
    }
  }

  .hosts-list {
    max-height: 200px;
    overflow-y: auto;
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    padding: 12px;
    background: #f7f8fa;
    border-radius: 8px;

    .host-tag {
      font-family: 'Consolas', 'Monaco', monospace;
    }
  }
}

.confirmation-input {
  padding: 16px;
  background: #fff9e6;
  border: 1px solid #e6a23c;
  border-radius: 8px;

  :deep(.el-checkbox__label) {
    font-weight: 500;
    color: #606266;
  }
}

.critical-warning {
  padding: 16px;
  background: #fef0f0;
  border: 2px solid #f56c6c;
  border-radius: 8px;
}

.dialog-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;

  .countdown-notice {
    color: #e6a23c;
    font-weight: 500;
    font-size: 14px;
  }
}

@media (max-width: 768px) {
  .danger-confirm-dialog {
    :deep(.el-dialog) {
      width: 90% !important;
      margin: 0 auto;
    }
  }

  .operation-info {
    .info-item {
      flex-direction: column;
      align-items: flex-start;
      gap: 8px;

      .label {
        min-width: auto;
      }
    }
  }
}
</style>
