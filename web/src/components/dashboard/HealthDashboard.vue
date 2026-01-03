<template>
  <div class="health-dashboard">
    <!-- 整体健康分 -->
    <div class="overall-health">
      <div class="score-circle" :class="scoreLevel">
        <svg viewBox="0 0 100 100">
          <circle class="bg" cx="50" cy="50" r="45" />
          <circle
            class="progress"
            cx="50"
            cy="50"
            r="45"
            :stroke-dasharray="`${scoreProgress} 283`"
          />
        </svg>
        <div class="score-content">
          <span class="score-value">{{ overallScore }}</span>
          <span class="score-label">健康分</span>
        </div>
      </div>
      <div class="score-info">
        <h3>系统健康度</h3>
        <el-tag :type="levelTagType" size="large">{{ levelText }}</el-tag>
        <p class="update-time">更新于 {{ lastUpdateTime }}</p>
      </div>
    </div>

    <!-- 主机对比表格 -->
    <div class="hosts-comparison">
      <div class="section-header">
        <h4>主机健康对比</h4>
        <el-button type="primary" size="small" :loading="loading" @click="runHealthCheck">
          <el-icon><Refresh /></el-icon>
          开始体检
        </el-button>
      </div>

      <el-table :data="hosts" stripe size="small" v-loading="loading">
        <el-table-column prop="host_name" label="主机" width="150" />
        <el-table-column label="健康分" width="100">
          <template #default="{ row }">
            <div class="score-badge" :class="row.score_level">
              {{ row.score }}
            </div>
          </template>
        </el-table-column>
        <el-table-column label="CPU" width="120">
          <template #default="{ row }">
            <el-progress
              :percentage="row.cpu"
              :color="getProgressColor(row.cpu)"
              :stroke-width="8"
              :show-text="true"
            />
          </template>
        </el-table-column>
        <el-table-column label="内存" width="120">
          <template #default="{ row }">
            <el-progress
              :percentage="row.memory"
              :color="getProgressColor(row.memory)"
              :stroke-width="8"
              :show-text="true"
            />
          </template>
        </el-table-column>
        <el-table-column label="磁盘" width="120">
          <template #default="{ row }">
            <el-progress
              :percentage="row.disk"
              :color="getProgressColor(row.disk)"
              :stroke-width="8"
              :show-text="true"
            />
          </template>
        </el-table-column>
        <el-table-column label="问题数" width="80" align="center">
          <template #default="{ row }">
            <el-badge :value="row.issue_count" :type="row.issue_count > 0 ? 'danger' : 'success'" />
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.status)" size="small">
              {{ getStatusText(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <!-- 异常主机提示 -->
    <div v-if="anomalies.length > 0" class="anomaly-alert">
      <el-alert type="warning" :closable="false" show-icon>
        <template #title>
          发现 {{ anomalies.length }} 台异常主机
        </template>
        <template #default>
          <div class="anomaly-hosts">
            <el-tag v-for="host in anomalies" :key="host" type="danger" size="small">
              {{ host }}
            </el-tag>
          </div>
        </template>
      </el-alert>
    </div>

    <!-- 资源占用最高 -->
    <div class="highlights">
      <div class="highlight-item" v-if="highestCPU">
        <span class="label">CPU 最高</span>
        <span class="value">{{ highestCPU }}</span>
      </div>
      <div class="highlight-item" v-if="highestMem">
        <span class="label">内存最高</span>
        <span class="value">{{ highestMem }}</span>
      </div>
      <div class="highlight-item" v-if="highestDisk">
        <span class="label">磁盘最高</span>
        <span class="value">{{ highestDisk }}</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { Refresh } from '@element-plus/icons-vue'
import { request } from '@/api/request'

interface HostHealthSummary {
  host_id: string
  host_name: string
  score: number
  score_level: string
  cpu: number
  memory: number
  disk: number
  status: string
  issue_count: number
}

interface HostComparison {
  hosts: HostHealthSummary[]
  highest_cpu: string
  highest_mem: string
  highest_disk: string
  anomalies: string[]
  avg_score: number
  checked_at: string
}

const loading = ref(false)
const overallScore = ref(0)
const scoreLevel = ref('good')
const hosts = ref<HostHealthSummary[]>([])
const anomalies = ref<string[]>([])
const highestCPU = ref('')
const highestMem = ref('')
const highestDisk = ref('')
const lastUpdateTime = ref('')

const scoreProgress = computed(() => {
  return (overallScore.value / 100) * 283
})

const levelTagType = computed(() => {
  switch (scoreLevel.value) {
    case 'excellent': return 'success'
    case 'good': return ''
    case 'warning': return 'warning'
    default: return 'danger'
  }
})

const levelText = computed(() => {
  switch (scoreLevel.value) {
    case 'excellent': return '优秀'
    case 'good': return '良好'
    case 'warning': return '需关注'
    default: return '异常'
  }
})

function getProgressColor(value: number): string {
  if (value >= 90) return '#f56c6c'
  if (value >= 80) return '#e6a23c'
  if (value >= 60) return '#409eff'
  return '#67c23a'
}

function getStatusType(status: string): string {
  switch (status) {
    case 'healthy': return 'success'
    case 'warning': return 'warning'
    case 'critical': return 'danger'
    default: return 'info'
  }
}

function getStatusText(status: string): string {
  switch (status) {
    case 'healthy': return '健康'
    case 'warning': return '警告'
    case 'critical': return '严重'
    default: return '未知'
  }
}

async function fetchOverallHealth() {
  try {
    const res = await request.get<{ score: number; level: string }>('/health/overall')
    overallScore.value = res.score
    scoreLevel.value = res.level
  } catch (e) {
    console.error('获取整体健康分失败', e)
  }
}

async function runHealthCheck() {
  loading.value = true
  try {
    const res = await request.post<HostComparison>('/health/compare', { host_ids: [] })
    hosts.value = res.hosts || []
    anomalies.value = res.anomalies || []
    highestCPU.value = res.highest_cpu || ''
    highestMem.value = res.highest_mem || ''
    highestDisk.value = res.highest_disk || ''
    overallScore.value = res.avg_score || 0

    // 更新等级
    if (overallScore.value >= 90) scoreLevel.value = 'excellent'
    else if (overallScore.value >= 75) scoreLevel.value = 'good'
    else if (overallScore.value >= 60) scoreLevel.value = 'warning'
    else scoreLevel.value = 'critical'

    lastUpdateTime.value = new Date().toLocaleTimeString('zh-CN')
  } catch (e) {
    console.error('健康检查失败', e)
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchOverallHealth()
})
</script>

<style scoped>
.health-dashboard {
  padding: 16px;
  background: #fff;
  border-radius: 8px;
}

.overall-health {
  display: flex;
  align-items: center;
  gap: 24px;
  padding: 20px;
  background: linear-gradient(135deg, #f5f7fa 0%, #e4e7ed 100%);
  border-radius: 12px;
  margin-bottom: 20px;
}

.score-circle {
  position: relative;
  width: 120px;
  height: 120px;
}

.score-circle svg {
  transform: rotate(-90deg);
  width: 100%;
  height: 100%;
}

.score-circle circle {
  fill: none;
  stroke-width: 8;
  stroke-linecap: round;
}

.score-circle .bg {
  stroke: #e4e7ed;
}

.score-circle .progress {
  stroke: #67c23a;
  transition: stroke-dasharray 0.5s ease;
}

.score-circle.excellent .progress { stroke: #67c23a; }
.score-circle.good .progress { stroke: #409eff; }
.score-circle.warning .progress { stroke: #e6a23c; }
.score-circle.critical .progress { stroke: #f56c6c; }

.score-content {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  text-align: center;
}

.score-value {
  font-size: 32px;
  font-weight: 700;
  color: #303133;
}

.score-label {
  display: block;
  font-size: 12px;
  color: #909399;
}

.score-info h3 {
  margin: 0 0 8px 0;
  font-size: 18px;
  color: #303133;
}

.update-time {
  margin: 8px 0 0 0;
  font-size: 12px;
  color: #909399;
}

.hosts-comparison {
  margin-bottom: 20px;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.section-header h4 {
  margin: 0;
  font-size: 16px;
  color: #303133;
}

.score-badge {
  display: inline-block;
  padding: 4px 12px;
  border-radius: 12px;
  font-weight: 600;
  font-size: 14px;
}

.score-badge.excellent {
  background: #f0f9eb;
  color: #67c23a;
}

.score-badge.good {
  background: #ecf5ff;
  color: #409eff;
}

.score-badge.warning {
  background: #fdf6ec;
  color: #e6a23c;
}

.score-badge.critical {
  background: #fef0f0;
  color: #f56c6c;
}

.anomaly-alert {
  margin-bottom: 20px;
}

.anomaly-hosts {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-top: 8px;
}

.highlights {
  display: flex;
  gap: 16px;
}

.highlight-item {
  flex: 1;
  padding: 12px;
  background: #f5f7fa;
  border-radius: 8px;
  text-align: center;
}

.highlight-item .label {
  display: block;
  font-size: 12px;
  color: #909399;
  margin-bottom: 4px;
}

.highlight-item .value {
  font-size: 14px;
  font-weight: 600;
  color: #303133;
}
</style>
