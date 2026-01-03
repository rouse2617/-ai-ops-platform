<template>
  <div class="prometheus-query-builder">
    <div class="query-input-section">
      <div class="input-group">
        <label>自然语言查询</label>
        <input
          v-model="nlQuery"
          type="text"
          placeholder="例如：CPU 使用率超过 80%"
          @keyup.enter="generatePromQL"
        />
        <button @click="generatePromQL" :disabled="!nlQuery.trim()">
          生成 PromQL
        </button>
      </div>

      <div v-if="generatedQuery" class="generated-query">
        <div class="query-display">
          <label>生成的 PromQL</label>
          <code>{{ generatedQuery.query }}</code>
          <p class="explanation">{{ generatedQuery.explanation }}</p>
        </div>
        <button @click="executeQuery" class="execute-btn">执行查询</button>
      </div>
    </div>

    <div v-if="loading" class="loading">加载中...</div>
    <div v-if="error" class="error">{{ error }}</div>

    <div v-if="queryResult" class="query-result">
      <h3>查询结果</h3>
      <div class="result-chart">
        <canvas ref="chartCanvas"></canvas>
      </div>
      <div class="result-table">
        <table>
          <thead>
            <tr>
              <th>时间</th>
              <th>值</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(point, idx) in queryResult.data" :key="idx">
              <td>{{ formatTime(point.timestamp) }}</td>
              <td>{{ point.value }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { usePrometheus } from '@/composables/usePrometheus'

const { executeQuery: executePrometheusQuery } = usePrometheus()

const nlQuery = ref('')
const generatedQuery = ref<any>(null)
const queryResult = ref<any>(null)
const loading = ref(false)
const error = ref('')
const chartCanvas = ref<HTMLCanvasElement>()

const generatePromQL = async () => {
  if (!nlQuery.value.trim()) return

  loading.value = true
  error.value = ''

  try {
    const response = await fetch('/api/prometheus/generate-promql', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ question: nlQuery.value })
    })

    if (!response.ok) throw new Error('生成失败')
    generatedQuery.value = await response.json()
  } catch (e) {
    error.value = e instanceof Error ? e.message : '生成失败'
  } finally {
    loading.value = false
  }
}

const executeQuery = async () => {
  if (!generatedQuery.value?.query) return

  loading.value = true
  error.value = ''

  try {
    const result = await executePrometheusQuery(generatedQuery.value.query)
    if (result) {
      queryResult.value = result
      renderChart()
    }
  } catch (e) {
    error.value = e instanceof Error ? e.message : '执行失败'
  } finally {
    loading.value = false
  }
}

const renderChart = () => {
  if (!chartCanvas.value || !queryResult.value?.data) return
  // 使用 Chart.js 或其他图表库渲染
}

const formatTime = (timestamp: number) => {
  return new Date(timestamp * 1000).toLocaleString()
}
</script>

<style scoped>
.prometheus-query-builder {
  padding: 20px;
  border: 1px solid #ddd;
  border-radius: 8px;
}

.query-input-section {
  margin-bottom: 20px;
}

.input-group {
  display: flex;
  flex-direction: column;
  gap: 10px;
  margin-bottom: 15px;
}

.input-group label {
  font-weight: 600;
  color: #333;
}

.input-group input {
  padding: 10px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 14px;
}

.input-group button {
  padding: 10px 20px;
  background-color: #007bff;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
}

.input-group button:disabled {
  background-color: #ccc;
  cursor: not-allowed;
}

.generated-query {
  background-color: #f5f5f5;
  padding: 15px;
  border-radius: 4px;
  margin-bottom: 15px;
}

.query-display {
  margin-bottom: 10px;
}

.query-display label {
  display: block;
  font-weight: 600;
  margin-bottom: 5px;
}

.query-display code {
  display: block;
  background-color: #fff;
  padding: 10px;
  border-radius: 4px;
  border: 1px solid #ddd;
  overflow-x: auto;
  font-family: monospace;
  margin-bottom: 10px;
}

.explanation {
  color: #666;
  font-size: 12px;
  margin: 0;
}

.execute-btn {
  padding: 8px 16px;
  background-color: #28a745;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
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

.query-result {
  margin-top: 20px;
}

.query-result h3 {
  margin-top: 0;
}

.result-chart {
  margin-bottom: 20px;
  height: 300px;
}

.result-table {
  overflow-x: auto;
}

.result-table table {
  width: 100%;
  border-collapse: collapse;
}

.result-table th,
.result-table td {
  padding: 10px;
  text-align: left;
  border-bottom: 1px solid #ddd;
}

.result-table th {
  background-color: #f5f5f5;
  font-weight: 600;
}
</style>
