<template>
  <div class="mcp-status-indicator">
    <el-tooltip :content="tooltipContent" placement="right">
      <div class="status-dot-wrapper">
        <div
          class="status-dot"
          :class="statusClass"
        />
        <div v-if="!collapsed" class="status-info">
          <span class="status-text">{{ statusText }}</span>
          <span v-if="latency > 0" class="latency">{{ latency }}ms</span>
        </div>
      </div>
    </el-tooltip>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useSystemStore } from '@/stores/system'

defineProps<{
  collapsed?: boolean
}>()

const systemStore = useSystemStore()

const statusClass = computed(() => {
  const status = systemStore.overallConnectionStatus
  return {
    'status-connected': status === 'connected',
    'status-partial': status === 'partial',
    'status-disconnected': status === 'disconnected',
    'status-error': status === 'error'
  }
})

const statusText = computed(() => {
  const status = systemStore.overallConnectionStatus
  const map = {
    connected: 'MCP 已连接',
    partial: 'MCP 部分连接',
    disconnected: 'MCP 未连接',
    error: 'MCP 连接错误'
  }
  return map[status] || 'MCP 未知'
})

const latency = computed(() => systemStore.averageLatency)

const tooltipContent = computed(() => {
  const connections = systemStore.mcpConnections
  if (connections.length === 0) return 'MCP 服务未连接'

  return connections.map(c =>
    `${c.name}: ${c.status} (${c.latency}ms)`
  ).join('\n')
})
</script>

<style scoped>
.mcp-status-indicator {
  padding: var(--spacing-2) var(--spacing-3);
}

.status-dot-wrapper {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  position: relative;
  flex-shrink: 0;
}

.status-dot::before {
  content: '';
  position: absolute;
  inset: -2px;
  border-radius: 50%;
  opacity: 0;
  animation: pulse 2s ease-in-out infinite;
}

.status-connected {
  background: #10b981;
}

.status-connected::before {
  background: #10b981;
}

.status-partial {
  background: #f59e0b;
}

.status-partial::before {
  background: #f59e0b;
}

.status-disconnected {
  background: #6b7280;
}

.status-error {
  background: #ef4444;
}

.status-error::before {
  background: #ef4444;
}

.status-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
}

.status-text {
  font-size: var(--text-xs);
  color: var(--sidebar-text);
  white-space: nowrap;
}

.latency {
  font-size: 10px;
  color: var(--sidebar-text);
  opacity: 0.6;
}

@keyframes pulse {
  0%, 100% {
    opacity: 0;
    transform: scale(1);
  }
  50% {
    opacity: 0.3;
    transform: scale(1.5);
  }
}
</style>
