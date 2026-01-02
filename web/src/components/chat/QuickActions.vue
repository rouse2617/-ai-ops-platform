<template>
  <div class="quick-actions">
    <div class="actions-grid">
      <el-button
        v-for="action in actions"
        :key="action.id"
        :type="action.type"
        :icon="action.icon"
        :loading="loadingActions.has(action.id)"
        :disabled="loadingActions.size > 0"
        @click="handleAction(action)"
      >
        {{ action.label }}
      </el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { ElMessageBox, ElMessage } from 'element-plus'
import {
  RefreshRight,
  ScaleToOriginal,
  Delete,
  Setting
} from '@element-plus/icons-vue'

interface QuickAction {
  id: string
  label: string
  type: 'primary' | 'success' | 'warning' | 'danger' | 'info'
  icon: any
  riskLevel: 'low' | 'medium' | 'high'
  command: string
  confirmMessage?: string
}

interface Props {
  hostId: string
}

const props = defineProps<Props>()

const emit = defineEmits<{
  execute: [command: string]
}>()

const loadingActions = ref<Set<string>>(new Set())

const actions: QuickAction[] = [
  {
    id: 'restart',
    label: 'Restart Service',
    type: 'warning',
    icon: RefreshRight,
    riskLevel: 'high',
    command: 'systemctl restart app',
    confirmMessage: 'This will restart the service. Continue?'
  },
  {
    id: 'scale',
    label: 'Scale Up',
    type: 'primary',
    icon: ScaleToOriginal,
    riskLevel: 'medium',
    command: 'kubectl scale deployment app --replicas=3',
    confirmMessage: 'Scale deployment to 3 replicas?'
  },
  {
    id: 'clear-cache',
    label: 'Clear Cache',
    type: 'info',
    icon: Delete,
    riskLevel: 'low',
    command: 'redis-cli FLUSHDB'
  },
  {
    id: 'check-config',
    label: 'Check Config',
    type: 'success',
    icon: Setting,
    riskLevel: 'low',
    command: 'nginx -t'
  }
]

async function handleAction(action: QuickAction) {
  try {
    // High risk actions require confirmation
    if (action.riskLevel === 'high' && action.confirmMessage) {
      await ElMessageBox.confirm(
        action.confirmMessage,
        'Confirm Action',
        {
          confirmButtonText: 'Execute',
          cancelButtonText: 'Cancel',
          type: 'warning',
          distinguishCancelAndClose: true
        }
      )
    }

    loadingActions.value.add(action.id)

    // Emit execute event
    emit('execute', action.command)

    // Simulate execution
    await new Promise(resolve => setTimeout(resolve, 2000))

    ElMessage.success(`${action.label} completed successfully`)
  } catch (error) {
    if (error !== 'cancel' && error !== 'close') {
      ElMessage.error(`${action.label} failed: ${error}`)
    }
  } finally {
    loadingActions.value.delete(action.id)
  }
}
</script>

<style scoped>
.quick-actions {
  padding: 16px;
  background: #f8fafc;
  border-radius: 8px;
}

.actions-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(140px, 1fr));
  gap: 12px;
}

.el-button {
  width: 100%;
}
</style>
