<template>
  <el-dialog
    v-model="dialogVisible"
    title="风险警告"
    width="700px"
    :close-on-click-modal="false"
    class="risk-warning-dialog"
    @close="handleClose"
  >
    <div class="dialog-content">
      <el-alert
        v-if="criticalRisks.length > 0"
        type="error"
        :closable="false"
        show-icon
        class="alert-section"
      >
        <template #title>
          <span class="alert-title">发现 {{ criticalRisks.length }} 个严重风险</span>
        </template>
      </el-alert>

      <el-tabs v-model="activeTab" class="risk-tabs">
        <el-tab-pane label="全部" name="all">
          <div class="risks-list">
            <RiskItemCard
              v-for="risk in risks"
              :key="risk.id"
              :risk="risk"
              @view-details="handleViewDetails"
              @resolve="handleResolve"
            />
          </div>
        </el-tab-pane>

        <el-tab-pane :label="`严重 (${criticalRisks.length})`" name="critical">
          <div class="risks-list">
            <RiskItemCard
              v-for="risk in criticalRisks"
              :key="risk.id"
              :risk="risk"
              @view-details="handleViewDetails"
              @resolve="handleResolve"
            />
          </div>
        </el-tab-pane>

        <el-tab-pane :label="`高危 (${highRisks.length})`" name="high">
          <div class="risks-list">
            <RiskItemCard
              v-for="risk in highRisks"
              :key="risk.id"
              :risk="risk"
              @view-details="handleViewDetails"
              @resolve="handleResolve"
            />
          </div>
        </el-tab-pane>

        <el-tab-pane :label="`中等 (${mediumRisks.length})`" name="medium">
          <div class="risks-list">
            <RiskItemCard
              v-for="risk in mediumRisks"
              :key="risk.id"
              :risk="risk"
              @view-details="handleViewDetails"
              @resolve="handleResolve"
            />
          </div>
        </el-tab-pane>
      </el-tabs>
    </div>

    <template #footer>
      <div class="dialog-footer">
        <el-checkbox v-model="dontShowAgain">
          不再显示此类警告
        </el-checkbox>
        <div class="footer-actions">
          <el-button @click="handleClose">关闭</el-button>
          <el-button type="primary" @click="handleResolveAll">
            全部处理
          </el-button>
        </div>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'

export interface RiskItem {
  id: string
  type: 'security' | 'performance' | 'availability'
  severity: 'low' | 'medium' | 'high' | 'critical'
  title: string
  description: string
  affectedHosts: string[]
  recommendation: string
  timestamp: number
}

interface Props {
  visible: boolean
  risks: RiskItem[]
  autoShow?: boolean
}

interface Emits {
  (e: 'update:visible', value: boolean): void
  (e: 'view-details', risk: RiskItem): void
  (e: 'resolve', riskId: string): void
  (e: 'resolve-all'): void
}

const props = withDefaults(defineProps<Props>(), {
  autoShow: true
})

const emit = defineEmits<Emits>()

const dialogVisible = computed({
  get: () => props.visible,
  set: (val) => emit('update:visible', val)
})

const activeTab = ref('all')
const dontShowAgain = ref(false)

const criticalRisks = computed(() =>
  props.risks.filter(r => r.severity === 'critical')
)

const highRisks = computed(() =>
  props.risks.filter(r => r.severity === 'high')
)

const mediumRisks = computed(() =>
  props.risks.filter(r => r.severity === 'medium')
)

const handleViewDetails = (risk: RiskItem) => {
  emit('view-details', risk)
}

const handleResolve = (riskId: string) => {
  emit('resolve', riskId)
}

const handleResolveAll = () => {
  emit('resolve-all')
  handleClose()
}

const handleClose = () => {
  if (dontShowAgain.value) {
    localStorage.setItem('risk-warning-disabled', 'true')
  }
  emit('update:visible', false)
}
</script>

<script lang="ts">
import { defineComponent } from 'vue'

const RiskItemCard = defineComponent({
  name: 'RiskItemCard',
  props: {
    risk: {
      type: Object as () => RiskItem,
      required: true
    }
  },
  emits: ['view-details', 'resolve'],
  setup(props, { emit }) {
    const severityConfig = {
      low: { color: '#67c23a', label: '低' },
      medium: { color: '#e6a23c', label: '中' },
      high: { color: '#f56c6c', label: '高' },
      critical: { color: '#c71585', label: '严重' }
    }

    const typeConfig = {
      security: { icon: 'Lock', label: '安全' },
      performance: { icon: 'TrendCharts', label: '性能' },
      availability: { icon: 'Connection', label: '可用性' }
    }

    return {
      severityConfig,
      typeConfig,
      handleViewDetails: () => emit('view-details', props.risk),
      handleResolve: () => emit('resolve', props.risk.id)
    }
  },
  template: `
    <div class="risk-item-card" :class="'severity-' + risk.severity">
      <div class="risk-header">
        <div class="risk-title-section">
          <el-tag :color="severityConfig[risk.severity].color" effect="dark" size="small">
            {{ severityConfig[risk.severity].label }}
          </el-tag>
          <el-tag type="info" size="small">
            {{ typeConfig[risk.type].label }}
          </el-tag>
          <span class="risk-title">{{ risk.title }}</span>
        </div>
        <span class="risk-time">{{ new Date(risk.timestamp).toLocaleString('zh-CN') }}</span>
      </div>

      <div class="risk-body">
        <p class="risk-description">{{ risk.description }}</p>

        <div v-if="risk.affectedHosts.length > 0" class="affected-hosts">
          <span class="label">影响主机:</span>
          <el-tag
            v-for="host in risk.affectedHosts.slice(0, 3)"
            :key="host"
            size="small"
            type="info"
          >
            {{ host }}
          </el-tag>
          <span v-if="risk.affectedHosts.length > 3" class="more-hosts">
            +{{ risk.affectedHosts.length - 3 }} 台
          </span>
        </div>

        <div class="recommendation">
          <span class="label">建议:</span>
          <span class="recommendation-text">{{ risk.recommendation }}</span>
        </div>
      </div>

      <div class="risk-footer">
        <el-button size="small" @click="handleViewDetails">查看详情</el-button>
        <el-button size="small" type="primary" @click="handleResolve">标记已处理</el-button>
      </div>
    </div>
  `
})

export default RiskItemCard
</script>

<style scoped lang="scss">
.risk-warning-dialog {
  :deep(.el-dialog__body) {
    padding: 0;
  }

  .dialog-content {
    .alert-section {
      margin: 20px 20px 0;

      .alert-title {
        font-weight: 600;
        font-size: 15px;
      }
    }

    .risk-tabs {
      padding: 20px;

      .risks-list {
        display: flex;
        flex-direction: column;
        gap: 16px;
        max-height: 500px;
        overflow-y: auto;
      }
    }
  }

  .dialog-footer {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 16px 20px;
    border-top: 1px solid #f0f0f0;

    .footer-actions {
      display: flex;
      gap: 12px;
    }
  }
}

.risk-item-card {
  background: white;
  border-radius: 8px;
  border: 1px solid #e4e7ed;
  padding: 16px;
  transition: all 0.2s ease;

  &:hover {
    box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
  }

  &.severity-critical {
    border-left: 4px solid #c71585;
  }

  &.severity-high {
    border-left: 4px solid #f56c6c;
  }

  &.severity-medium {
    border-left: 4px solid #e6a23c;
  }

  &.severity-low {
    border-left: 4px solid #67c23a;
  }

  .risk-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 12px;

    .risk-title-section {
      display: flex;
      align-items: center;
      gap: 8px;
      flex: 1;

      .risk-title {
        font-weight: 600;
        color: #303133;
        font-size: 15px;
      }
    }

    .risk-time {
      font-size: 12px;
      color: #909399;
      white-space: nowrap;
    }
  }

  .risk-body {
    margin-bottom: 12px;

    .risk-description {
      color: #606266;
      font-size: 14px;
      line-height: 1.6;
      margin-bottom: 12px;
    }

    .affected-hosts,
    .recommendation {
      display: flex;
      align-items: center;
      gap: 8px;
      margin-bottom: 8px;
      flex-wrap: wrap;

      .label {
        font-weight: 500;
        color: #909399;
        font-size: 13px;
      }

      .more-hosts {
        font-size: 12px;
        color: #909399;
      }
    }

    .recommendation {
      .recommendation-text {
        color: #409eff;
        font-size: 13px;
      }
    }
  }

  .risk-footer {
    display: flex;
    justify-content: flex-end;
    gap: 8px;
    padding-top: 12px;
    border-top: 1px solid #f0f0f0;
  }
}

@media (max-width: 768px) {
  .risk-warning-dialog {
    :deep(.el-dialog) {
      width: 90% !important;
    }

    .dialog-footer {
      flex-direction: column;
      align-items: stretch;
      gap: 12px;

      .footer-actions {
        width: 100%;
        justify-content: stretch;

        .el-button {
          flex: 1;
        }
      }
    }
  }

  .risk-item-card {
    .risk-header {
      flex-direction: column;
      align-items: flex-start;
      gap: 8px;
    }
  }
}
</style>
