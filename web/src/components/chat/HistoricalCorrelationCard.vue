<template>
  <div class="historical-correlation-card" v-if="correlation && correlation.confidence > 0">
    <div class="card-header">
      <el-icon class="history-icon"><Clock /></el-icon>
      <span class="card-title">历史关联分析</span>
      <el-tag size="small" type="info" effect="plain">
        {{ Math.round(correlation.confidence * 100) }}% 匹配度
      </el-tag>
    </div>

    <div class="card-content">
      <div class="correlation-item" v-if="correlation.similar_incident">
        <el-icon class="item-icon"><Warning /></el-icon>
        <div class="item-content">
          <span class="item-label">相似问题</span>
          <span class="item-value">{{ correlation.similar_incident }}</span>
          <span class="item-time" v-if="correlation.occurred_at">
            <el-icon><Calendar /></el-icon>
            {{ correlation.occurred_at }}
          </span>
        </div>
      </div>

      <div class="correlation-item resolution" v-if="correlation.resolution">
        <el-icon class="item-icon"><CircleCheck /></el-icon>
        <div class="item-content">
          <span class="item-label">历史解决方案</span>
          <span class="item-value">{{ correlation.resolution }}</span>
        </div>
      </div>

      <div class="action-hint">
        <el-button type="primary" size="small" text @click="handleApplySolution">
          <el-icon><MagicStick /></el-icon>
          尝试此方案
        </el-button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'
import { Clock, Warning, Calendar, CircleCheck, MagicStick } from '@element-plus/icons-vue'
import { getHistoricalCorrelation, type HistoricalCorrelation } from '@/api/analysis'

interface Props {
  hostId?: string
  currentIssue?: string
  issueType?: string
}

const props = defineProps<Props>()

const emit = defineEmits<{
  applySolution: [resolution: string]
}>()

const correlation = ref<HistoricalCorrelation | null>(null)
const loading = ref(false)

const fetchCorrelation = async () => {
  if (!props.hostId || !props.currentIssue) return

  loading.value = true
  try {
    const result = await getHistoricalCorrelation({
      host_id: props.hostId,
      current_issue: props.currentIssue,
      issue_type: props.issueType
    })
    correlation.value = result
  } catch (error) {
    console.error('获取历史关联失败:', error)
  } finally {
    loading.value = false
  }
}

const handleApplySolution = () => {
  if (correlation.value?.resolution) {
    emit('applySolution', correlation.value.resolution)
  }
}

watch(() => [props.hostId, props.currentIssue], () => {
  fetchCorrelation()
}, { immediate: true })
</script>

<style scoped>
.historical-correlation-card {
  background: linear-gradient(135deg, #fefce8 0%, #fef9c3 100%);
  border: 1px solid #fde047;
  border-radius: 10px;
  padding: 14px;
  margin: 12px 0;
}

.card-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
}

.history-icon {
  color: #ca8a04;
}

.card-title {
  flex: 1;
  font-weight: 600;
  font-size: 13px;
  color: #854d0e;
}

.card-content {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.correlation-item {
  display: flex;
  gap: 10px;
  padding: 10px;
  background: rgba(255, 255, 255, 0.6);
  border-radius: 8px;
}

.correlation-item.resolution {
  background: rgba(34, 197, 94, 0.1);
  border: 1px solid rgba(34, 197, 94, 0.3);
}

.item-icon {
  flex-shrink: 0;
  color: #ca8a04;
  margin-top: 2px;
}

.resolution .item-icon {
  color: #16a34a;
}

.item-content {
  display: flex;
  flex-direction: column;
  gap: 4px;
  flex: 1;
}

.item-label {
  font-size: 11px;
  color: #92400e;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.resolution .item-label {
  color: #166534;
}

.item-value {
  font-size: 13px;
  color: #1f2937;
  line-height: 1.5;
}

.item-time {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 11px;
  color: #6b7280;
  margin-top: 2px;
}

.action-hint {
  display: flex;
  justify-content: flex-end;
  margin-top: 4px;
}
</style>
