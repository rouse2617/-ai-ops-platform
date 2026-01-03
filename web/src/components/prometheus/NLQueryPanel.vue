<template>
  <div class="nl-query-panel">
    <div class="query-input-section">
      <el-input
        v-model="queryText"
        placeholder="用自然语言描述你想查询的指标，例如：查看所有主机的 CPU 使用率"
        :prefix-icon="Search"
        size="large"
        @keyup.enter="executeQuery"
        clearable
      >
        <template #append>
          <el-button :loading="loading" @click="executeQuery">
            <el-icon><Search /></el-icon>
            查询
          </el-button>
        </template>
      </el-input>

      <div class="suggestions" v-if="showSuggestions && suggestions.length > 0">
        <span class="suggestions-label">快捷查询：</span>
        <el-tag
          v-for="(suggestion, index) in suggestions"
          :key="index"
          class="suggestion-tag"
          effect="plain"
          @click="selectSuggestion(suggestion)"
        >
          {{ suggestion }}
        </el-tag>
      </div>
    </div>

    <div class="query-result" v-if="result">
      <div class="result-header">
        <div class="promql-section">
          <span class="label">PromQL:</span>
          <code class="promql">{{ result.promql }}</code>
          <el-button size="small" text @click="copyPromQL">
            <el-icon><CopyDocument /></el-icon>
          </el-button>
        </div>
        <div class="explanation" v-if="result.explanation">
          <el-icon><InfoFilled /></el-icon>
          {{ result.explanation }}
        </div>
      </div>

      <div class="chart-container" v-if="result.data && result.data.length > 0">
        <div class="chart-header">
          <span>{{ result.chart_type === 'gauge' ? '当前值' : '趋势图' }}</span>
          <el-select v-model="selectedTimeRange" size="small" @change="refreshQuery">
            <el-option label="1小时" value="1h" />
            <el-option label="3小时" value="3h" />
            <el-option label="24小时" value="24h" />
            <el-option label="7天" value="7d" />
          </el-select>
        </div>

        <!-- 仪表盘类型 -->
        <div v-if="result.chart_type === 'gauge'" class="gauge-container">
          <div v-for="(item, index) in result.data" :key="index" class="gauge-item">
            <el-progress
              type="dashboard"
              :percentage="parseFloat(item.value) || 0"
              :color="getGaugeColor"
              :width="120"
            >
              <template #default="{ percentage }">
                <span class="gauge-value">{{ percentage.toFixed(1) }}%</span>
                <span class="gauge-label">{{ getMetricLabel(item.metric) }}</span>
              </template>
            </el-progress>
          </div>
        </div>

        <!-- 表格类型 -->
        <div v-else-if="result.chart_type === 'table'" class="table-container">
          <el-table :data="result.data" stripe size="small">
            <el-table-column label="实例" prop="metric.instance" />
            <el-table-column label="值">
              <template #default="{ row }">
                {{ formatValue(row.value) }}
              </template>
            </el-table-column>
          </el-table>
        </div>

        <!-- 折线图类型 -->
        <div v-else class="line-chart-placeholder">
          <div class="chart-data">
            <div v-for="(series, index) in result.data" :key="index" class="series-item">
              <span class="series-label">{{ getMetricLabel(series.metric) }}</span>
              <span class="series-value" v-if="series.value">
                当前: {{ formatValue(series.value) }}
              </span>
              <span class="series-points" v-if="series.values">
                {{ series.values.length }} 个数据点
              </span>
            </div>
          </div>
          <p class="chart-hint">
            <el-icon><TrendCharts /></el-icon>
            图表将在右侧监控面板中渲染
          </p>
        </div>
      </div>

      <div class="no-data" v-else-if="!loading">
        <el-empty description="暂无数据" :image-size="80" />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Search, CopyDocument, InfoFilled, TrendCharts } from '@element-plus/icons-vue'
import { request } from '@/api/request'

interface QueryResult {
  query: string
  promql: string
  explanation: string
  data: Array<{
    metric: Record<string, string>
    value?: [number, string]
    values?: Array<[number, string]>
  }>
  chart_type: string
  time_range: string
}

const queryText = ref('')
const loading = ref(false)
const result = ref<QueryResult | null>(null)
const suggestions = ref<string[]>([])
const showSuggestions = ref(true)
const selectedTimeRange = ref('1h')

const emit = defineEmits<{
  (e: 'query-result', result: QueryResult): void
  (e: 'render-chart', config: { promql: string; type: string; timeRange: string }): void
}>()

onMounted(() => {
  fetchSuggestions()
})

async function fetchSuggestions() {
  try {
    const res = await request.get<{ suggestions: string[] }>('/prometheus/semantic/suggestions')
    suggestions.value = res.suggestions || []
  } catch (e) {
    console.error('获取建议失败', e)
  }
}

async function executeQuery() {
  if (!queryText.value.trim()) {
    ElMessage.warning('请输入查询内容')
    return
  }

  loading.value = true
  showSuggestions.value = false

  try {
    const res = await request.post<QueryResult>('/prometheus/semantic/query', {
      query: queryText.value
    })
    result.value = res
    selectedTimeRange.value = res.time_range || '1h'

    emit('query-result', res)

    // 触发图表渲染
    if (res.promql) {
      emit('render-chart', {
        promql: res.promql,
        type: res.chart_type,
        timeRange: res.time_range
      })
    }
  } catch (e) {
    ElMessage.error('查询失败')
    console.error(e)
  } finally {
    loading.value = false
  }
}

function selectSuggestion(suggestion: string) {
  queryText.value = suggestion
  executeQuery()
}

function copyPromQL() {
  if (result.value?.promql) {
    navigator.clipboard.writeText(result.value.promql)
    ElMessage.success('已复制 PromQL')
  }
}

function refreshQuery() {
  if (result.value) {
    executeQuery()
  }
}

function getMetricLabel(metric: Record<string, string>): string {
  return metric?.instance || metric?.job || '未知'
}

function formatValue(value: unknown): string {
  if (Array.isArray(value) && value.length >= 2) {
    const num = parseFloat(value[1] as string)
    if (num > 1000000) return (num / 1000000).toFixed(2) + 'M'
    if (num > 1000) return (num / 1000).toFixed(2) + 'K'
    return num.toFixed(2)
  }
  if (typeof value === 'string') {
    const num = parseFloat(value)
    if (!isNaN(num)) return num.toFixed(2)
  }
  return String(value)
}

function getGaugeColor(percentage: number): string {
  if (percentage >= 90) return '#f56c6c'
  if (percentage >= 70) return '#e6a23c'
  return '#67c23a'
}
</script>

<style scoped>
.nl-query-panel {
  padding: 16px;
  background: #fff;
  border-radius: 8px;
}

.query-input-section {
  margin-bottom: 16px;
}

.suggestions {
  margin-top: 12px;
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
}

.suggestions-label {
  font-size: 13px;
  color: #909399;
}

.suggestion-tag {
  cursor: pointer;
  transition: all 0.2s;
}

.suggestion-tag:hover {
  background: #409eff;
  color: #fff;
  border-color: #409eff;
}

.query-result {
  border-top: 1px solid #ebeef5;
  padding-top: 16px;
}

.result-header {
  margin-bottom: 16px;
}

.promql-section {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.promql-section .label {
  font-size: 13px;
  color: #909399;
}

.promql {
  flex: 1;
  padding: 8px 12px;
  background: #f5f7fa;
  border-radius: 4px;
  font-family: monospace;
  font-size: 13px;
  color: #303133;
  overflow-x: auto;
}

.explanation {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  color: #606266;
}

.chart-container {
  background: #fafafa;
  border-radius: 8px;
  padding: 16px;
}

.chart-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
  font-weight: 500;
}

.gauge-container {
  display: flex;
  flex-wrap: wrap;
  gap: 24px;
  justify-content: center;
}

.gauge-item {
  text-align: center;
}

.gauge-value {
  display: block;
  font-size: 20px;
  font-weight: 600;
  color: #303133;
}

.gauge-label {
  display: block;
  font-size: 12px;
  color: #909399;
  margin-top: 4px;
}

.table-container {
  max-height: 300px;
  overflow-y: auto;
}

.line-chart-placeholder {
  text-align: center;
  padding: 24px;
}

.chart-data {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  justify-content: center;
  margin-bottom: 16px;
}

.series-item {
  padding: 8px 16px;
  background: #fff;
  border-radius: 6px;
  border: 1px solid #ebeef5;
}

.series-label {
  font-size: 13px;
  color: #606266;
  margin-right: 8px;
}

.series-value, .series-points {
  font-size: 14px;
  font-weight: 500;
  color: #303133;
}

.chart-hint {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  color: #909399;
  font-size: 13px;
}

.no-data {
  padding: 40px;
}
</style>
