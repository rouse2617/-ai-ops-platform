<template>
  <div
    class="trend-warning-bubble"
    :class="[`severity-${severity}`, `trend-${trend}`]"
    :style="bubbleStyle"
  >
    <div class="bubble-content">
      <div class="bubble-header">
        <el-icon class="warning-icon" :size="20">
          <component :is="severityIcon" />
        </el-icon>
        <span class="metric-name">{{ metric }}</span>
      </div>

      <div class="bubble-body">
        <div class="value-display">
          <span class="current-value">{{ formatValue(currentValue) }}</span>
          <el-icon class="trend-icon" :size="16">
            <component :is="trendIcon" />
          </el-icon>
        </div>

        <div class="threshold-info">
          <span class="threshold-label">阈值:</span>
          <span class="threshold-value">{{ formatValue(threshold) }}</span>
        </div>

        <div v-if="changePercent" class="change-info">
          <span :class="['change-percent', trend]">
            {{ changePercent > 0 ? '+' : '' }}{{ changePercent.toFixed(1) }}%
          </span>
        </div>
      </div>

      <div class="bubble-footer">
        <el-button size="small" text @click="handleViewDetails">
          查看详情
        </el-button>
        <el-button size="small" text @click="handleDismiss">
          忽略
        </el-button>
      </div>
    </div>

    <div class="bubble-arrow" />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import {
  WarningFilled,
  InfoFilled,
  CircleCloseFilled,
  TrendCharts,
  Bottom,
  Top
} from '@element-plus/icons-vue'

interface Props {
  metric: string
  currentValue: number
  threshold: number
  trend: 'up' | 'down' | 'stable'
  severity: 'info' | 'warning' | 'critical'
  position: { x: number; y: number }
  changePercent?: number
}

interface Emits {
  (e: 'view-details', metric: string): void
  (e: 'dismiss'): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

const bubbleStyle = computed(() => ({
  left: `${props.position.x}px`,
  top: `${props.position.y}px`
}))

const severityIcon = computed(() => {
  const iconMap = {
    info: InfoFilled,
    warning: WarningFilled,
    critical: CircleCloseFilled
  }
  return iconMap[props.severity]
})

const trendIcon = computed(() => {
  const iconMap = {
    up: Top,
    down: Bottom,
    stable: TrendCharts
  }
  return iconMap[props.trend]
})

const formatValue = (value: number): string => {
  if (value >= 1000000) {
    return `${(value / 1000000).toFixed(2)}M`
  }
  if (value >= 1000) {
    return `${(value / 1000).toFixed(2)}K`
  }
  return value.toFixed(2)
}

const handleViewDetails = () => {
  emit('view-details', props.metric)
}

const handleDismiss = () => {
  emit('dismiss')
}
</script>

<style scoped lang="scss">
.trend-warning-bubble {
  position: absolute;
  z-index: 1000;
  min-width: 240px;
  background: white;
  border-radius: 12px;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.15);
  animation: bubble-appear 0.3s ease-out;
  transform-origin: bottom center;

  &.severity-info {
    border: 2px solid #409eff;

    .warning-icon {
      color: #409eff;
    }
  }

  &.severity-warning {
    border: 2px solid #e6a23c;

    .warning-icon {
      color: #e6a23c;
    }
  }

  &.severity-critical {
    border: 2px solid #f56c6c;
    animation: bubble-appear 0.3s ease-out, pulse 2s infinite;

    .warning-icon {
      color: #f56c6c;
    }
  }

  .bubble-content {
    padding: 16px;

    .bubble-header {
      display: flex;
      align-items: center;
      gap: 8px;
      margin-bottom: 12px;

      .metric-name {
        font-weight: 600;
        color: #303133;
        font-size: 14px;
      }
    }

    .bubble-body {
      margin-bottom: 12px;

      .value-display {
        display: flex;
        align-items: center;
        gap: 8px;
        margin-bottom: 8px;

        .current-value {
          font-size: 24px;
          font-weight: 700;
          color: #303133;
        }

        .trend-icon {
          &.trend-up {
            color: #f56c6c;
          }

          &.trend-down {
            color: #67c23a;
          }

          &.trend-stable {
            color: #909399;
          }
        }
      }

      .threshold-info {
        display: flex;
        align-items: center;
        gap: 8px;
        font-size: 13px;
        color: #606266;
        margin-bottom: 4px;

        .threshold-label {
          color: #909399;
        }

        .threshold-value {
          font-weight: 600;
        }
      }

      .change-info {
        .change-percent {
          font-size: 14px;
          font-weight: 600;

          &.up {
            color: #f56c6c;
          }

          &.down {
            color: #67c23a;
          }

          &.stable {
            color: #909399;
          }
        }
      }
    }

    .bubble-footer {
      display: flex;
      justify-content: flex-end;
      gap: 8px;
      padding-top: 8px;
      border-top: 1px solid #f0f0f0;
    }
  }

  .bubble-arrow {
    position: absolute;
    bottom: -10px;
    left: 50%;
    transform: translateX(-50%);
    width: 0;
    height: 0;
    border-left: 10px solid transparent;
    border-right: 10px solid transparent;
    border-top: 10px solid white;

    &::before {
      content: '';
      position: absolute;
      bottom: 2px;
      left: -11px;
      width: 0;
      height: 0;
      border-left: 11px solid transparent;
      border-right: 11px solid transparent;
      border-top: 11px solid currentColor;
    }
  }

  &.severity-info .bubble-arrow::before {
    color: #409eff;
  }

  &.severity-warning .bubble-arrow::before {
    color: #e6a23c;
  }

  &.severity-critical .bubble-arrow::before {
    color: #f56c6c;
  }
}

@keyframes bubble-appear {
  from {
    opacity: 0;
    transform: scale(0.8) translateY(10px);
  }
  to {
    opacity: 1;
    transform: scale(1) translateY(0);
  }
}

@keyframes pulse {
  0%, 100% {
    box-shadow: 0 8px 24px rgba(0, 0, 0, 0.15);
  }
  50% {
    box-shadow: 0 8px 24px rgba(245, 108, 108, 0.4);
  }
}

@media (max-width: 768px) {
  .trend-warning-bubble {
    min-width: 200px;

    .bubble-content {
      padding: 12px;

      .bubble-body {
        .value-display {
          .current-value {
            font-size: 20px;
          }
        }
      }
    }
  }
}
</style>
