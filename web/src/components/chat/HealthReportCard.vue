<template>
  <el-card class="health-report-card" :class="`score-${scoreLevel}`">
    <template #header>
      <div class="card-header">
        <div class="header-left">
          <el-icon class="host-icon" :size="20"><Monitor /></el-icon>
          <span class="host-id">{{ hostId }}</span>
        </div>
        <div class="header-right">
          <el-tag :type="scoreTagType" size="large">
            健康度: {{ score }}
          </el-tag>
        </div>
      </div>
    </template>

    <div class="card-body">
      <div class="score-display">
        <div class="score-circle" :style="scoreCircleStyle">
          <span class="score-value">{{ score }}</span>
          <span class="score-label">分</span>
        </div>
        <div class="score-description">
          <span class="score-level-text">{{ scoreLevelText }}</span>
          <span class="last-update">
            更新于 {{ formatTime(lastUpdate) }}
          </span>
        </div>
      </div>

      <div class="metrics-list">
        <div
          v-for="metric in metrics"
          :key="metric.name"
          class="metric-item"
          :class="`status-${metric.status}`"
        >
          <div class="metric-header">
            <span class="metric-name">{{ metric.name }}</span>
            <el-icon class="metric-status-icon" :size="16">
              <component :is="getStatusIcon(metric.status)" />
            </el-icon>
          </div>

          <div class="metric-value">
            <span class="value">{{ formatMetricValue(metric.value) }}</span>
            <span v-if="metric.trend !== 0" class="trend" :class="getTrendClass(metric.trend)">
              <el-icon :size="12">
                <component :is="metric.trend > 0 ? Top : Bottom" />
              </el-icon>
              {{ Math.abs(metric.trend).toFixed(1) }}%
            </span>
          </div>

          <el-progress
            :percentage="getMetricPercentage(metric)"
            :status="getProgressStatus(metric.status)"
            :show-text="false"
            :stroke-width="6"
          />
        </div>
      </div>

      <div v-if="recommendations && recommendations.length > 0" class="recommendations">
        <div class="recommendations-header">
          <el-icon><Memo /></el-icon>
          <span>优化建议</span>
        </div>
        <ul class="recommendations-list">
          <li v-for="(rec, index) in recommendations" :key="index">
            {{ rec }}
          </li>
        </ul>
      </div>
    </div>

    <template #footer>
      <div class="card-footer">
        <el-button size="small" @click="handleRefresh" :loading="refreshing">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
        <el-button size="small" type="primary" @click="handleViewDetails">
          查看详情
        </el-button>
      </div>
    </template>
  </el-card>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import {
  Monitor,
  CircleCheckFilled,
  WarningFilled,
  CircleCloseFilled,
  Top,
  Bottom,
  Memo,
  Refresh
} from '@element-plus/icons-vue'

export interface HealthMetric {
  name: string
  value: number
  status: 'good' | 'warning' | 'critical'
  trend: number
}

interface Props {
  hostId: string
  score: number
  metrics: HealthMetric[]
  lastUpdate: number
  recommendations?: string[]
}

interface Emits {
  (e: 'refresh'): void
  (e: 'view-details'): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

const refreshing = ref(false)

const scoreLevel = computed(() => {
  if (props.score >= 90) return 'excellent'
  if (props.score >= 75) return 'good'
  if (props.score >= 60) return 'warning'
  return 'critical'
})

const scoreLevelText = computed(() => {
  const textMap = {
    excellent: '优秀',
    good: '良好',
    warning: '需关注',
    critical: '异常'
  }
  return textMap[scoreLevel.value]
})

const scoreTagType = computed(() => {
  const typeMap = {
    excellent: 'success',
    good: 'success',
    warning: 'warning',
    critical: 'danger'
  }
  return typeMap[scoreLevel.value] as 'success' | 'warning' | 'danger'
})

const scoreCircleStyle = computed(() => {
  const colorMap = {
    excellent: '#67c23a',
    good: '#95d475',
    warning: '#e6a23c',
    critical: '#f56c6c'
  }
  const color = colorMap[scoreLevel.value]

  return {
    background: `conic-gradient(${color} ${props.score * 3.6}deg, #f0f0f0 0deg)`
  }
})

const getStatusIcon = (status: HealthMetric['status']) => {
  const iconMap = {
    good: CircleCheckFilled,
    warning: WarningFilled,
    critical: CircleCloseFilled
  }
  return iconMap[status]
}

const getTrendClass = (trend: number) => {
  return trend > 0 ? 'trend-up' : 'trend-down'
}

const getMetricPercentage = (metric: HealthMetric): number => {
  return Math.min(metric.value, 100)
}

const getProgressStatus = (status: HealthMetric['status']) => {
  const statusMap = {
    good: 'success',
    warning: 'warning',
    critical: 'exception'
  }
  return statusMap[status] as 'success' | 'warning' | 'exception'
}

const formatMetricValue = (value: number): string => {
  return `${value.toFixed(1)}%`
}

const formatTime = (timestamp: number): string => {
  const now = Date.now()
  const diff = now - timestamp

  if (diff < 60000) return '刚刚'
  if (diff < 3600000) return `${Math.floor(diff / 60000)}分钟前`
  if (diff < 86400000) return `${Math.floor(diff / 3600000)}小时前`

  return new Date(timestamp).toLocaleString('zh-CN')
}

const handleRefresh = async () => {
  refreshing.value = true
  try {
    emit('refresh')
  } finally {
    setTimeout(() => {
      refreshing.value = false
    }, 500)
  }
}

const handleViewDetails = () => {
  emit('view-details')
}
</script>

<style scoped lang="scss">
.health-report-card {
  border-radius: 12px;
  overflow: hidden;
  transition: all 0.3s ease;

  &:hover {
    box-shadow: 0 4px 16px rgba(0, 0, 0, 0.12);
  }

  &.score-excellent {
    border-top: 4px solid #67c23a;
  }

  &.score-good {
    border-top: 4px solid #95d475;
  }

  &.score-warning {
    border-top: 4px solid #e6a23c;
  }

  &.score-critical {
    border-top: 4px solid #f56c6c;
  }

  :deep(.el-card__header) {
    padding: 16px 20px;
    background: #f7f8fa;
    border-bottom: 1px solid #e4e7ed;
  }

  :deep(.el-card__body) {
    padding: 20px;
  }

  :deep(.el-card__footer) {
    padding: 12px 20px;
    background: #f7f8fa;
    border-top: 1px solid #e4e7ed;
  }
}

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;

  .header-left {
    display: flex;
    align-items: center;
    gap: 8px;

    .host-icon {
      color: #409eff;
    }

    .host-id {
      font-weight: 600;
      font-size: 15px;
      color: #303133;
      font-family: 'Consolas', 'Monaco', monospace;
    }
  }
}

.card-body {
  .score-display {
    display: flex;
    align-items: center;
    gap: 24px;
    margin-bottom: 24px;
    padding-bottom: 20px;
    border-bottom: 1px solid #f0f0f0;

    .score-circle {
      position: relative;
      width: 100px;
      height: 100px;
      border-radius: 50%;
      display: flex;
      flex-direction: column;
      align-items: center;
      justify-content: center;

      &::before {
        content: '';
        position: absolute;
        inset: 8px;
        background: white;
        border-radius: 50%;
      }

      .score-value {
        position: relative;
        z-index: 1;
        font-size: 32px;
        font-weight: 700;
        color: #303133;
      }

      .score-label {
        position: relative;
        z-index: 1;
        font-size: 14px;
        color: #909399;
      }
    }

    .score-description {
      flex: 1;
      display: flex;
      flex-direction: column;
      gap: 8px;

      .score-level-text {
        font-size: 20px;
        font-weight: 600;
        color: #303133;
      }

      .last-update {
        font-size: 13px;
        color: #909399;
      }
    }
  }

  .metrics-list {
    display: flex;
    flex-direction: column;
    gap: 16px;
    margin-bottom: 20px;

    .metric-item {
      .metric-header {
        display: flex;
        align-items: center;
        justify-content: space-between;
        margin-bottom: 8px;

        .metric-name {
          font-size: 14px;
          font-weight: 500;
          color: #606266;
        }

        .metric-status-icon {
          &.status-good {
            color: #67c23a;
          }

          &.status-warning {
            color: #e6a23c;
          }

          &.status-critical {
            color: #f56c6c;
          }
        }
      }

      .metric-value {
        display: flex;
        align-items: center;
        gap: 8px;
        margin-bottom: 8px;

        .value {
          font-size: 18px;
          font-weight: 600;
          color: #303133;
        }

        .trend {
          display: flex;
          align-items: center;
          gap: 2px;
          font-size: 12px;
          font-weight: 500;

          &.trend-up {
            color: #f56c6c;
          }

          &.trend-down {
            color: #67c23a;
          }
        }
      }
    }
  }

  .recommendations {
    background: #fff9e6;
    border: 1px solid #e6a23c;
    border-radius: 8px;
    padding: 16px;

    .recommendations-header {
      display: flex;
      align-items: center;
      gap: 8px;
      margin-bottom: 12px;
      font-weight: 600;
      color: #606266;

      .el-icon {
        color: #e6a23c;
      }
    }

    .recommendations-list {
      margin: 0;
      padding-left: 20px;

      li {
        color: #606266;
        font-size: 13px;
        line-height: 1.8;
        margin-bottom: 4px;

        &:last-child {
          margin-bottom: 0;
        }
      }
    }
  }
}

.card-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

@media (max-width: 768px) {
  .health-report-card {
    :deep(.el-card__body) {
      padding: 16px;
    }

    .card-body {
      .score-display {
        flex-direction: column;
        align-items: flex-start;
        gap: 16px;

        .score-circle {
          width: 80px;
          height: 80px;

          .score-value {
            font-size: 24px;
          }
        }
      }
    }
  }
}
</style>
