<template>
  <div class="anomaly-detector">
    <div class="detector-header">
      <h3>动态异常检测</h3>
      <p class="subtitle">基于基线计算（均值±标准差）</p>
    </div>

    <div class="detector-form">
      <div class="form-group">
        <label>指标名称</label>
        <input v-model="metricName" type="text" placeholder="例如：node_cpu_seconds_total" />
      </div>

      <div class="form-group">
        <label>当前值</label>
        <input v-model.number="currentValue" type="number" placeholder="输入当前值" />
      </div>

      <button @click="detectAnomaly" :disabled="!metricName || currentValue === null">
        检测异常
      </button>
    </div>

    <div v-if="loading" class="loading">检测中...</div>
    <div v-if="error" class="error">{{ error }}</div>

    <div v-if="anomalyResult" class="anomaly-result">
      <div class="result-header" :class="anomalyResult.isAnomaly ? 'anomaly' : 'normal'">
        <span class="status-badge">
          {{ anomalyResult.isAnomaly ? '异常' : '正常' }}
        </span>
        <span class="severity">
          严重程度: {{ (anomalyResult.severity * 100).toFixed(0) }}%
        </span>
      </div>

      <div class="result-details">
        <div class="detail-item">
          <label>偏差</label>
          <span>{{ anomalyResult.deviation.toFixed(2) }}%</span>
        </div>

        <div class="detail-item">
          <label>建议</label>
          <span>{{ anomalyResult.recommendation }}</span>
        </div>
      </div>

      <div class="baseline-chart">
        <canvas ref="baselineCanvas"></canvas>
      </div>

      <div class="anomaly-points" v-if="anomalyPoints.length > 0">
        <h4>异常点标记</h4>
        <div class="points-list">
          <div v-for="(point, idx) in anomalyPoints" :key="idx" class="point-item">
            <span class="time">{{ formatTime(point.timestamp) }}</span>
            <span class="value">{{ point.value.toFixed(2) }}</span>
            <span class="deviation">偏差: {{ point.deviation.toFixed(2) }}%</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'

const metricName = ref('')
const currentValue = ref<number | null>(null)
const anomalyResult = ref<any>(null)
const anomalyPoints = ref<any[]>([])
const loading = ref(false)
const error = ref('')
const baselineCanvas = ref<HTMLCanvasElement>()

const detectAnomaly = async () => {
  if (!metricName.value || currentValue.value === null) return

  loading.value = true
  error.value = ''

  try {
    const response = await fetch('/api/prometheus/detect-anomaly', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        metric_name: metricName.value,
        value: currentValue.value
      })
    })

    if (!response.ok) throw new Error('检测失败')
    anomalyResult.value = await response.json()

    // 模拟异常点数据
    if (anomalyResult.value.isAnomaly) {
      anomalyPoints.value = [
        {
          timestamp: Date.now() / 1000,
          value: currentValue.value,
          deviation: anomalyResult.value.deviation
        }
      ]
    }
  } catch (e) {
    error.value = e instanceof Error ? e.message : '检测失败'
  } finally {
    loading.value = false
  }
}

const formatTime = (timestamp: number) => {
  return new Date(timestamp * 1000).toLocaleString()
}
</script>

<style scoped>
.anomaly-detector {
  padding: 20px;
  border: 1px solid #ddd;
  border-radius: 8px;
  background-color: #fafafa;
}

.detector-header {
  margin-bottom: 20px;
}

.detector-header h3 {
  margin: 0 0 5px 0;
  color: #333;
}

.subtitle {
  margin: 0;
  color: #999;
  font-size: 12px;
}

.detector-form {
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

.detector-form button {
  padding: 8px 20px;
  background-color: #007bff;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  align-self: flex-end;
}

.detector-form button:disabled {
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

.anomaly-result {
  background-color: white;
  padding: 15px;
  border-radius: 4px;
  border: 1px solid #ddd;
}

.result-header {
  display: flex;
  gap: 15px;
  padding: 15px;
  border-radius: 4px;
  margin-bottom: 15px;
  align-items: center;
}

.result-header.anomaly {
  background-color: #ffe7e7;
  border: 1px solid #ffcccc;
}

.result-header.normal {
  background-color: #e7ffe7;
  border: 1px solid #ccffcc;
}

.status-badge {
  font-weight: 600;
  font-size: 16px;
}

.severity {
  margin-left: auto;
  color: #666;
}

.result-details {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 15px;
  margin-bottom: 15px;
}

.detail-item {
  display: flex;
  flex-direction: column;
  gap: 5px;
}

.detail-item label {
  font-weight: 600;
  font-size: 12px;
  color: #999;
  text-transform: uppercase;
}

.detail-item span {
  font-size: 16px;
  color: #333;
}

.baseline-chart {
  height: 250px;
  margin-bottom: 15px;
  border: 1px solid #ddd;
  border-radius: 4px;
  padding: 10px;
}

.anomaly-points {
  margin-top: 15px;
}

.anomaly-points h4 {
  margin: 0 0 10px 0;
  color: #333;
}

.points-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.point-item {
  display: flex;
  gap: 15px;
  padding: 10px;
  background-color: #fff3cd;
  border-radius: 4px;
  border-left: 3px solid #ffc107;
}

.point-item .time {
  flex: 1;
  font-size: 12px;
  color: #666;
}

.point-item .value {
  font-weight: 600;
  color: #333;
}

.point-item .deviation {
  color: #cc0000;
  font-weight: 600;
}
</style>
