<template>
  <div class="action-cards">
    <div class="action-cards-header">
      <h3>建议步骤</h3>
      <div class="action-buttons">
        <el-button size="small" @click="executeAll" :disabled="isExecuting || allCompleted">
          批量执行
        </el-button>
        <el-button size="small" @click="reset" :disabled="isExecuting">
          重置
        </el-button>
      </div>
    </div>

    <div class="action-list">
      <div
        v-for="(action, index) in actions"
        :key="index"
        class="action-item"
        :class="action.status"
      >
        <div class="action-status-icon">
          <el-icon v-if="action.status === 'pending'"><Clock /></el-icon>
          <el-icon v-else-if="action.status === 'running'" class="rotating"><Loading /></el-icon>
          <el-icon v-else-if="action.status === 'completed'" class="success"><CircleCheck /></el-icon>
          <el-icon v-else-if="action.status === 'error'" class="error"><CircleClose /></el-icon>
        </div>

        <div class="action-content">
          <div class="action-title">{{ action.title }}</div>
          <div v-if="action.description" class="action-description">{{ action.description }}</div>
          <div v-if="action.command" class="action-command">
            <code>{{ action.command }}</code>
          </div>
        </div>

        <el-button
          v-if="action.status === 'pending' || action.status === 'error'"
          size="small"
          type="primary"
          @click="executeAction(index)"
          :disabled="isExecuting"
        >
          执行
        </el-button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { Clock, Loading, CircleCheck, CircleClose } from '@element-plus/icons-vue'

export interface ActionStep {
  title: string
  description?: string
  command?: string
  status: 'pending' | 'running' | 'completed' | 'error'
}

const props = defineProps<{
  steps: Array<{ title: string; description?: string; command?: string }>
}>()

const emit = defineEmits<{
  execute: [command: string, index: number]
}>()

const actions = ref<ActionStep[]>(
  props.steps.map(step => ({ ...step, status: 'pending' as const }))
)

const isExecuting = ref(false)

const allCompleted = computed(() =>
  actions.value.every(a => a.status === 'completed')
)

const executeAction = async (index: number) => {
  const action = actions.value[index]
  if (!action.command) return

  action.status = 'running'
  isExecuting.value = true

  try {
    emit('execute', action.command, index)
  } catch (error) {
    action.status = 'error'
  } finally {
    isExecuting.value = false
  }
}

const executeAll = async () => {
  isExecuting.value = true

  for (let i = 0; i < actions.value.length; i++) {
    const action = actions.value[i]
    if (action.status === 'pending' && action.command) {
      action.status = 'running'
      emit('execute', action.command, i)
      await new Promise(resolve => setTimeout(resolve, 500))
    }
  }

  isExecuting.value = false
}

const reset = () => {
  actions.value.forEach(action => {
    action.status = 'pending'
  })
}

const updateStatus = (index: number, status: ActionStep['status']) => {
  if (actions.value[index]) {
    actions.value[index].status = status
  }
}

defineExpose({
  updateStatus
})
</script>

<style scoped>
.action-cards {
  background: #f8fafc;
  border-radius: 8px;
  padding: 16px;
  margin: 12px 0;
}

.action-cards-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.action-cards-header h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: #303133;
}

.action-buttons {
  display: flex;
  gap: 8px;
}

.action-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.action-item {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 12px;
  background: #fff;
  border-radius: 6px;
  border: 1px solid #e4e7ed;
  transition: all 0.2s;
}

.action-item.running {
  border-color: #409eff;
  background: #ecf5ff;
}

.action-item.completed {
  border-color: #67c23a;
  background: #f0f9ff;
}

.action-item.error {
  border-color: #f56c6c;
  background: #fef0f0;
}

.action-status-icon {
  flex-shrink: 0;
  font-size: 20px;
  margin-top: 2px;
}

.action-status-icon .rotating {
  animation: rotate 1s linear infinite;
}

.action-status-icon .success {
  color: #67c23a;
}

.action-status-icon .error {
  color: #f56c6c;
}

@keyframes rotate {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

.action-content {
  flex: 1;
  min-width: 0;
}

.action-title {
  font-weight: 500;
  color: #303133;
  margin-bottom: 4px;
}

.action-description {
  font-size: 13px;
  color: #606266;
  margin-bottom: 8px;
}

.action-command {
  background: #f5f7fa;
  padding: 8px 12px;
  border-radius: 4px;
  border-left: 3px solid #409eff;
}

.action-command code {
  font-family: 'Consolas', 'Monaco', monospace;
  font-size: 13px;
  color: #303133;
}
</style>
