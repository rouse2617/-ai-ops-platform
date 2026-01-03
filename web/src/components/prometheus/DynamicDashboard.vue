<template>
  <div class="dynamic-dashboard">
    <div class="dashboard-header">
      <h3>动态看板</h3>
      <p class="subtitle">根据对话上下文动态显示相关指标</p>
    </div>

    <div class="dashboard-controls">
      <div class="control-group">
        <label>选择模板</label>
        <select v-model="selectedTemplate" @change="loadTemplate">
          <option value="">-- 选择模板 --</option>
          <option value="system_overview">系统概览</option>
          <option value="application_performance">应用性能</option>
          <option value="business_metrics">业务指标</option>
          <option value="alerts">告警</option>
        </select>
      </div>

      <button @click="generateDashboard" class="generate-btn">
        生成看板
      </button>

      <button @click="refreshDashboard" class="refresh-btn">
        刷新
      </button>
    </div>

    <div v-if="loading" class="loading">加载中...</div>
    <div v-if="error" class="error">{{ error }}</div>

    <div v-if="dashboard" class="dashboard-grid">
      <div v-for="panel in dashboard.panels" :key="panel.id" class="panel-container">
        <div class="panel-header">
          <h4>{{ panel.title }}</h4>
          <span class="panel-type">{{ panel.type }}</span>
        </div>

        <div class="panel-content">
          <div v-if="panel.type === 'stat'" class="stat-panel">
            <div class="stat-value">{{ formatValue(panel.data?.value) }}</div>
            <div class="stat-label">{{ panel.data?.label }}</div>
          </div>

          <div v-else-if="panel.type === 'graph'" class="graph-panel">
            <canvas :ref="`chart-${panel.id}`"></canvas>
          </div>

          <div v-else-if="panel.type === 'table'" class="table-panel">
            <table>
              <thead>
                <tr>
                  <th v-for="col in panel.data?.columns" :key="col">{{ col }}</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="(row, idx) in panel.data?.rows" :key="idx">
                  <td v-for="(cell, cidx) in row" :key="cidx">{{ cell }}</td>
                </tr>
              </tbody>
            </table>
          </div>

          <div v-else-if="panel.type === 'heatmap'" class="heatmap-panel">
            <canvas :ref="`heatmap-${panel.id}`"></canvas>
          </div>
        </div>

        <div class="panel-footer">
          <span class="query">{{ panel.query }}</span>
        </div>
      </div>
    </div>

    <div v-if="recommendedPanels.length > 0" class="recommended-panels">
      <h4>推荐面板</h4>
      <div class="panels-grid">
        <div v-for="panel in recommendedPanels" :key="panel.id" class="recommended-item">
          <h5>{{ panel.title }}</h5>
          <p>{{ panel.query }}</p>
          <button @click="addPanel(panel)">添加</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'

const selectedTemplate = ref('')
const dashboard = ref<any>(null)
const recommendedPanels = ref<any[]>([])
const loading = ref(false)
const error = ref('')

onMounted(async () => {
  await loadRecommendedPanels()
})

const loadTemplate = async () => {
  if (!selectedTemplate.value) return

  loading.value = true
  error.value = ''

  try {
    const response = await fetch(
      `/api/prometheus/dashboards/template/${selectedTemplate.value}`
    )
    if (!response.ok) throw new Error('加载模板失败')
    const template = await response.json()
    dashboard.value = {
      panels: template.map((panel: any) => ({
        ...panel,
        data: {}
      }))
    }
  } catch (e) {
    error.value = e instanceof Error ? e.message : '加载失败'
  } finally {
    loading.value = false
  }
}

const generateDashboard = async () => {
  loading.value = true
  error.value = ''

  try {
    const response = await fetch('/api/prometheus/dashboards', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        name: `Dashboard-${Date.now()}`,
        panels: []
      })
    })

    if (!response.ok) throw new Error('生成失败')
    dashboard.value = await response.json()
  } catch (e) {
    error.value = e instanceof Error ? e.message : '生成失败'
  } finally {
    loading.value = false
  }
}

const refreshDashboard = async () => {
  if (!dashboard.value?.id) return

  loading.value = true
  error.value = ''

  try {
    const response = await fetch(`/api/prometheus/dashboards/${dashboard.value.id}`)
    if (!response.ok) throw new Error('刷新失败')
    dashboard.value = await response.json()
  } catch (e) {
    error.value = e instanceof Error ? e.message : '刷新失败'
  } finally {
    loading.value = false
  }
}

const loadRecommendedPanels = async () => {
  try {
    const response = await fetch('/api/prometheus/dashboards/recommend-panels')
    if (!response.ok) throw new Error('加载推荐失败')
    recommendedPanels.value = await response.json()
  } catch (e) {
    console.error('加载推荐面板失败:', e)
  }
}

const addPanel = (panel: any) => {
  if (!dashboard.value) {
    dashboard.value = { panels: [] }
  }
  dashboard.value.panels.push(panel)
}

const formatValue = (value: any) => {
  if (typeof value === 'number') {
    return value.toFixed(2)
  }
  return value
}
</script>

<style scoped>
.dynamic-dashboard {
  padding: 20px;
  border: 1px solid #ddd;
  border-radius: 8px;
  background-color: #fafafa;
}

.dashboard-header {
  margin-bottom: 20px;
}

.dashboard-header h3 {
  margin: 0 0 5px 0;
  color: #333;
}

.subtitle {
  margin: 0;
  color: #999;
  font-size: 12px;
}

.dashboard-controls {
  display: flex;
  gap: 15px;
  margin-bottom: 20px;
  flex-wrap: wrap;
}

.control-group {
  display: flex;
  flex-direction: column;
  gap: 5px;
}

.control-group label {
  font-weight: 600;
  font-size: 14px;
}

.control-group select {
  padding: 8px 12px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 14px;
}

.generate-btn,
.refresh-btn {
  padding: 8px 16px;
  background-color: #007bff;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  align-self: flex-end;
}

.refresh-btn {
  background-color: #6c757d;
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

.dashboard-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(400px, 1fr));
  gap: 20px;
  margin-bottom: 30px;
}

.panel-container {
  background-color: white;
  border: 1px solid #ddd;
  border-radius: 8px;
  overflow: hidden;
}

.panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 15px;
  border-bottom: 1px solid #ddd;
  background-color: #f9f9f9;
}

.panel-header h4 {
  margin: 0;
  color: #333;
}

.panel-type {
  font-size: 12px;
  color: #999;
  background-color: #f0f0f0;
  padding: 2px 8px;
  border-radius: 3px;
}

.panel-content {
  padding: 15px;
  min-height: 200px;
}

.stat-panel {
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  height: 100%;
}

.stat-value {
  font-size: 48px;
  font-weight: 600;
  color: #007bff;
}

.stat-label {
  font-size: 14px;
  color: #666;
  margin-top: 10px;
}

.graph-panel,
.heatmap-panel {
  height: 250px;
}

.table-panel {
  overflow-x: auto;
}

.table-panel table {
  width: 100%;
  border-collapse: collapse;
  font-size: 12px;
}

.table-panel th,
.table-panel td {
  padding: 8px;
  text-align: left;
  border-bottom: 1px solid #ddd;
}

.table-panel th {
  background-color: #f9f9f9;
  font-weight: 600;
}

.panel-footer {
  padding: 10px 15px;
  border-top: 1px solid #ddd;
  background-color: #f9f9f9;
  font-size: 12px;
  color: #666;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.query {
  font-family: monospace;
}

.recommended-panels {
  margin-top: 30px;
  padding: 20px;
  background-color: white;
  border: 1px solid #ddd;
  border-radius: 8px;
}

.recommended-panels h4 {
  margin-top: 0;
  color: #333;
}

.panels-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
  gap: 15px;
}

.recommended-item {
  padding: 15px;
  border: 1px solid #ddd;
  border-radius: 4px;
  background-color: #f9f9f9;
}

.recommended-item h5 {
  margin: 0 0 10px 0;
  color: #333;
}

.recommended-item p {
  margin: 0 0 10px 0;
  font-size: 12px;
  color: #666;
  font-family: monospace;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.recommended-item button {
  padding: 6px 12px;
  background-color: #28a745;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 12px;
}
</style>
