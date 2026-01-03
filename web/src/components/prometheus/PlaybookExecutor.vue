<template>
  <div class="playbook-executor">
    <div class="executor-header">
      <h3>自愈排障剧本</h3>
      <p class="subtitle">一键操作选项</p>
    </div>

    <div class="playbook-list">
      <div v-for="playbook in playbooks" :key="playbook.id" class="playbook-card">
        <div class="playbook-info">
          <h4>{{ playbook.name }}</h4>
          <p class="trigger">触发条件: {{ playbook.trigger }}</p>
          <p class="steps-count">步骤数: {{ playbook.steps.length }}</p>
        </div>

        <button
          @click="executePlaybook(playbook.id)"
          :disabled="executing === playbook.id"
          class="execute-btn"
        >
          {{ executing === playbook.id ? '执行中...' : '执行' }}
        </button>
      </div>
    </div>

    <div v-if="error" class="error">{{ error }}</div>

    <div v-if="executionResult" class="execution-result">
      <div class="result-header" :class="executionResult.status">
        <span class="status">{{ executionResult.status }}</span>
      </div>

      <div class="steps-timeline">
        <div v-for="(step, idx) in executionResult.steps" :key="idx" class="step-item">
          <div class="step-header">
            <span class="step-number">{{ idx + 1 }}</span>
            <span class="step-name">{{ step.name }}</span>
            <span class="step-status" :class="step.status">{{ step.status }}</span>
          </div>

          <div v-if="step.output" class="step-output">
            <pre>{{ step.output }}</pre>
          </div>

          <div v-if="step.error" class="step-error">
            <pre>{{ step.error }}</pre>
          </div>
        </div>
      </div>

      <div class="result-footer">
        <p>执行时间: {{ formatDuration(executionResult.startTime, executionResult.endTime) }}</p>
      </div>
    </div>

    <div v-if="loading" class="loading">加载剧本中...</div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'

const playbooks = ref<any[]>([])
const executionResult = ref<any>(null)
const executing = ref<string | null>(null)
const error = ref('')
const loading = ref(false)

onMounted(async () => {
  await loadPlaybooks()
})

const loadPlaybooks = async () => {
  loading.value = true
  error.value = ''

  try {
    const response = await fetch('/api/prometheus/playbooks')
    if (!response.ok) throw new Error('加载失败')
    playbooks.value = await response.json()
  } catch (e) {
    error.value = e instanceof Error ? e.message : '加载失败'
  } finally {
    loading.value = false
  }
}

const executePlaybook = async (playbookId: string) => {
  executing.value = playbookId
  error.value = ''

  try {
    const response = await fetch(`/api/prometheus/playbooks/${playbookId}/execute`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ context: {} })
    })

    if (!response.ok) throw new Error('执行失败')
    executionResult.value = await response.json()
  } catch (e) {
    error.value = e instanceof Error ? e.message : '执行失败'
  } finally {
    executing.value = null
  }
}

const formatDuration = (start: number, end: number) => {
  if (!start || !end) return '未知'
  const duration = (end - start) / 1000
  return `${duration.toFixed(2)}秒`
}
</script>

<style scoped>
.playbook-executor {
  padding: 20px;
  border: 1px solid #ddd;
  border-radius: 8px;
  background-color: #fafafa;
}

.executor-header {
  margin-bottom: 20px;
}

.executor-header h3 {
  margin: 0 0 5px 0;
  color: #333;
}

.subtitle {
  margin: 0;
  color: #999;
  font-size: 12px;
}

.playbook-list {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 15px;
  margin-bottom: 20px;
}

.playbook-card {
  background-color: white;
  padding: 15px;
  border: 1px solid #ddd;
  border-radius: 4px;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
}

.playbook-info h4 {
  margin: 0 0 10px 0;
  color: #333;
}

.trigger,
.steps-count {
  margin: 5px 0;
  font-size: 12px;
  color: #666;
}

.execute-btn {
  padding: 8px 16px;
  background-color: #28a745;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  margin-top: 10px;
}

.execute-btn:disabled {
  background-color: #ccc;
  cursor: not-allowed;
}

.error {
  padding: 15px;
  background-color: #ffe7e7;
  color: #cc0000;
  border-radius: 4px;
  margin-bottom: 15px;
}

.loading {
  padding: 15px;
  background-color: #e7f3ff;
  color: #0066cc;
  border-radius: 4px;
}

.execution-result {
  background-color: white;
  padding: 15px;
  border-radius: 4px;
  border: 1px solid #ddd;
  margin-top: 20px;
}

.result-header {
  padding: 10px 15px;
  border-radius: 4px;
  margin-bottom: 15px;
  font-weight: 600;
}

.result-header.success {
  background-color: #e7ffe7;
  color: #00cc00;
}

.result-header.failed {
  background-color: #ffe7e7;
  color: #cc0000;
}

.result-header.running {
  background-color: #fff3cd;
  color: #ff9800;
}

.steps-timeline {
  display: flex;
  flex-direction: column;
  gap: 15px;
}

.step-item {
  border-left: 3px solid #ddd;
  padding-left: 15px;
  padding-top: 10px;
  padding-bottom: 10px;
}

.step-header {
  display: flex;
  gap: 10px;
  align-items: center;
  margin-bottom: 10px;
}

.step-number {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  background-color: #007bff;
  color: white;
  border-radius: 50%;
  font-size: 12px;
  font-weight: 600;
}

.step-name {
  font-weight: 600;
  color: #333;
}

.step-status {
  margin-left: auto;
  padding: 2px 8px;
  border-radius: 3px;
  font-size: 12px;
  font-weight: 600;
}

.step-status.success {
  background-color: #e7ffe7;
  color: #00cc00;
}

.step-status.failed {
  background-color: #ffe7e7;
  color: #cc0000;
}

.step-status.running {
  background-color: #fff3cd;
  color: #ff9800;
}

.step-output,
.step-error {
  margin-top: 10px;
}

.step-output pre {
  background-color: #f5f5f5;
  padding: 10px;
  border-radius: 4px;
  overflow-x: auto;
  margin: 0;
  font-size: 12px;
}

.step-error pre {
  background-color: #ffe7e7;
  color: #cc0000;
  padding: 10px;
  border-radius: 4px;
  overflow-x: auto;
  margin: 0;
  font-size: 12px;
}

.result-footer {
  margin-top: 15px;
  padding-top: 15px;
  border-top: 1px solid #ddd;
  color: #666;
  font-size: 12px;
}
</style>
