<template>
  <div class="session-health-badge" :class="`health-${healthLevel}`">
    <el-tooltip :content="tooltipContent" placement="right">
      <div class="health-score">
        <el-icon class="health-icon">
          <component :is="healthIcon" />
        </el-icon>
        <span class="score-text">{{ score }}</span>
      </div>
    </el-tooltip>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { SuccessFilled, WarningFilled, CircleCloseFilled } from '@element-plus/icons-vue'
import { useSystemStore } from '@/stores/system'

const props = defineProps<{
  sessionId: string
}>()

const systemStore = useSystemStore()

const health = computed(() => systemStore.getSessionHealth(props.sessionId))

const score = computed(() => health.value?.score ?? 0)

const healthLevel = computed(() => {
  const s = score.value
  if (s >= 80) return 'excellent'
  if (s >= 60) return 'good'
  if (s >= 40) return 'warning'
  return 'poor'
})

const healthIcon = computed(() => {
  const level = healthLevel.value
  if (level === 'excellent' || level === 'good') return SuccessFilled
  if (level === 'warning') return WarningFilled
  return CircleCloseFilled
})

const tooltipContent = computed(() => {
  if (!health.value) return '暂无健康数据'

  const { metrics } = health.value
  return `响应时间: ${metrics.responseTime}ms\n错误率: ${(metrics.errorRate * 100).toFixed(1)}%\n工具成功率: ${(metrics.toolSuccessRate * 100).toFixed(1)}%`
})
</script>

<style scoped>
.session-health-badge {
  display: inline-flex;
  align-items: center;
  padding: 2px 6px;
  border-radius: var(--radius-md);
  font-size: 11px;
  font-weight: var(--font-medium);
}

.health-score {
  display: flex;
  align-items: center;
  gap: 3px;
}

.health-icon {
  font-size: 12px;
}

.score-text {
  line-height: 1;
}

.health-excellent {
  background: rgba(16, 185, 129, 0.15);
  color: #10b981;
}

.health-good {
  background: rgba(59, 130, 246, 0.15);
  color: #3b82f6;
}

.health-warning {
  background: rgba(245, 158, 11, 0.15);
  color: #f59e0b;
}

.health-poor {
  background: rgba(239, 68, 68, 0.15);
  color: #ef4444;
}
</style>
