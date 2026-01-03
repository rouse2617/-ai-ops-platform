<template>
  <div class="semantic-translator">
    <div class="translator-header">
      <h3>故障语义化翻译</h3>
      <p class="subtitle">告警翻译工具</p>
    </div>

    <div class="translator-form">
      <div class="form-group">
        <label>指标名称</label>
        <input v-model="metricName" type="text" placeholder="例如：node_cpu_seconds_total" />
      </div>

      <div class="form-group">
        <label>指标值</label>
        <input v-model.number="metricValue" type="number" placeholder="输入指标值" />
      </div>

      <button @click="translateMetric" :disabled="!metricName || metricValue === null">
        翻译
      </button>
    </div>

    <div v-if="loading" class="loading">翻译中...</div>
    <div v-if="error" class="error">{{ error }}</div>

    <div v-if="translation" class="translation-result">
      <div class="result-section">
        <h4>{{ translation.title }}</h4>
        <p class="description">{{ translation.description }}</p>
      </div>

      <div class="result-section">
        <h4>根本原因</h4>
        <p>{{ translation.rootCause }}</p>
      </div>

      <div class="result-section">
        <h4>影响分析</h4>
        <p>{{ translation.impact }}</p>
      </div>

      <div class="result-section">
        <h4>建议措施</h4>
        <ul>
          <li v-for="(suggestion, idx) in translation.suggestions" :key="idx">
            {{ suggestion }}
          </li>
        </ul>
      </div>

      <div class="alert-description">
        <h4>告警描述</h4>
        <div class="alert-box">
          <pre>{{ alertDescription }}</pre>
        </div>
        <button @click="copyToClipboard" class="copy-btn">复制</button>
      </div>

      <div class="recommendations">
        <h4>处理建议</h4>
        <div v-for="(rec, idx) in recommendations" :key="idx" class="recommendation-item">
          <span class="rec-number">{{ idx + 1 }}</span>
          <span class="rec-text">{{ rec }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'

const metricName = ref('')
const metricValue = ref<number | null>(null)
const translation = ref<any>(null)
const loading = ref(false)
const error = ref('')

const alertDescription = computed(() => {
  if (!translation.value) return ''
  return `告警: ${translation.value.title}
严重级别: 高
当前值: ${metricValue.value}
描述: ${translation.value.description}
根本原因: ${translation.value.rootCause}
影响: ${translation.value.impact}`
})

const recommendations = computed(() => {
  if (!translation.value) return []
  return translation.value.suggestions || []
})

const translateMetric = async () => {
  if (!metricName.value || metricValue.value === null) return

  loading.value = true
  error.value = ''

  try {
    const response = await fetch('/api/prometheus/translate', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        metric_name: metricName.value,
        value: metricValue.value,
        labels: {}
      })
    })

    if (!response.ok) throw new Error('翻译失败')
    translation.value = await response.json()
  } catch (e) {
    error.value = e instanceof Error ? e.message : '翻译失败'
  } finally {
    loading.value = false
  }
}

const copyToClipboard = () => {
  navigator.clipboard.writeText(alertDescription.value)
  alert('已复制到剪贴板')
}
</script>

<style scoped>
.semantic-translator {
  padding: 20px;
  border: 1px solid #ddd;
  border-radius: 8px;
  background-color: #fafafa;
}

.translator-header {
  margin-bottom: 20px;
}

.translator-header h3 {
  margin: 0 0 5px 0;
  color: #333;
}

.subtitle {
  margin: 0;
  color: #999;
  font-size: 12px;
}

.translator-form {
  display: flex;
  gap: 15px;
  margin-bottom: 20px;
  flex-wrap: wrap;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 5px;
  flex: 1;
  min-width: 200px;
}

.form-group label {
  font-weight: 600;
  font-size: 14px;
  color: #333;
}

.form-group input {
  padding: 8px 12px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 14px;
}

.translator-form button {
  padding: 8px 20px;
  background-color: #007bff;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  align-self: flex-end;
}

.translator-form button:disabled {
  background-color: #ccc;
  cursor: not-allowed;
}

.loading,
.error {
  padding: 15px;
  border-radius: 4px;
  margin-bottom: 15px;
}

.loading {
  background-color: #e7f3ff;
  color: #0066cc;
}

.error {
  background-color: #ffe7e7;
  color: #cc0000;
}

.translation-result {
  background-color: white;
  padding: 20px;
  border-radius: 4px;
  border: 1px solid #ddd;
}

.result-section {
  margin-bottom: 20px;
  padding-bottom: 20px;
  border-bottom: 1px solid #eee;
}

.result-section:last-child {
  border-bottom: none;
}

.result-section h4 {
  margin: 0 0 10px 0;
  color: #333;
  font-size: 16px;
}

.result-section p {
  margin: 0;
  color: #666;
  line-height: 1.6;
}

.result-section ul {
  margin: 0;
  padding-left: 20px;
}

.result-section li {
  margin: 8px 0;
  color: #666;
  line-height: 1.6;
}

.alert-description {
  margin-top: 20px;
}

.alert-description h4 {
  margin: 0 0 10px 0;
  color: #333;
}

.alert-box {
  background-color: #f5f5f5;
  padding: 15px;
  border-radius: 4px;
  border: 1px solid #ddd;
  margin-bottom: 10px;
}

.alert-box pre {
  margin: 0;
  font-family: monospace;
  font-size: 12px;
  color: #333;
  white-space: pre-wrap;
  word-wrap: break-word;
}

.copy-btn {
  padding: 6px 12px;
  background-color: #6c757d;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 12px;
}

.recommendations {
  margin-top: 20px;
}

.recommendations h4 {
  margin: 0 0 15px 0;
  color: #333;
}

.recommendation-item {
  display: flex;
  gap: 15px;
  padding: 12px;
  background-color: #f9f9f9;
  border-left: 3px solid #007bff;
  border-radius: 4px;
  margin-bottom: 10px;
}

.rec-number {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  background-color: #007bff;
  color: white;
  border-radius: 50%;
  font-weight: 600;
  font-size: 12px;
  flex-shrink: 0;
}

.rec-text {
  color: #666;
  line-height: 1.6;
}
</style>
