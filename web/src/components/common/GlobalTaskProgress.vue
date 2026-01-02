<template>
  <teleport to="body">
    <transition-group name="task-slide" tag="div" class="global-task-container">
      <div
        v-for="task in activeTasks"
        :key="task.id"
        class="task-progress-item"
        :class="`task-${task.status}`"
      >
        <div class="task-header">
          <span class="task-title">{{ task.title }}</span>
          <el-icon
            class="task-close"
            @click="handleClose(task.id)"
          >
            <Close />
          </el-icon>
        </div>
        <div class="task-progress-bar">
          <div
            class="task-progress-fill"
            :style="{ width: `${task.progress}%` }"
          />
        </div>
        <div class="task-footer">
          <span class="task-percentage">{{ task.progress }}%</span>
          <span v-if="task.estimatedTime" class="task-eta">
            预计 {{ formatTime(task.estimatedTime) }}
          </span>
        </div>
      </div>
    </transition-group>
  </teleport>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { Close } from '@element-plus/icons-vue'
import { useSystemStore } from '@/stores/system'

const systemStore = useSystemStore()

const activeTasks = computed(() =>
  systemStore.taskProgress.filter(t => t.status === 'running' || t.status === 'paused')
)

function handleClose(taskId: string) {
  systemStore.removeTaskProgress(taskId)
}

function formatTime(ms: number): string {
  const seconds = Math.floor(ms / 1000)
  if (seconds < 60) return `${seconds}秒`
  const minutes = Math.floor(seconds / 60)
  return `${minutes}分钟`
}
</script>

<style scoped>
.global-task-container {
  position: fixed;
  top: 80px;
  right: 20px;
  z-index: 3000;
  display: flex;
  flex-direction: column;
  gap: var(--spacing-2);
  max-width: 320px;
  pointer-events: none;
}

.task-progress-item {
  background: var(--bg-elevated);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-lg);
  padding: var(--spacing-3);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  pointer-events: auto;
  backdrop-filter: blur(10px);
}

.task-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: var(--spacing-2);
}

.task-title {
  font-size: var(--text-sm);
  font-weight: var(--font-medium);
  color: var(--text-primary);
}

.task-close {
  cursor: pointer;
  color: var(--text-secondary);
  transition: color var(--transition-fast);
}

.task-close:hover {
  color: var(--text-primary);
}

.task-progress-bar {
  height: 4px;
  background: var(--bg-secondary);
  border-radius: 2px;
  overflow: hidden;
  margin-bottom: var(--spacing-2);
}

.task-progress-fill {
  height: 100%;
  background: linear-gradient(90deg, #3b82f6, #60a5fa);
  transition: width 0.3s ease;
  position: relative;
}

.task-progress-fill::after {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: linear-gradient(
    90deg,
    transparent,
    rgba(255, 255, 255, 0.3),
    transparent
  );
  animation: shimmer 1.5s infinite;
}

.task-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: var(--text-xs);
  color: var(--text-secondary);
}

.task-running .task-progress-fill {
  background: linear-gradient(90deg, #3b82f6, #60a5fa);
}

.task-paused .task-progress-fill {
  background: linear-gradient(90deg, #f59e0b, #fbbf24);
}

.task-error .task-progress-fill {
  background: linear-gradient(90deg, #ef4444, #f87171);
}

@keyframes shimmer {
  0% {
    transform: translateX(-100%);
  }
  100% {
    transform: translateX(100%);
  }
}

.task-slide-enter-active,
.task-slide-leave-active {
  transition: all 0.3s ease;
}

.task-slide-enter-from {
  opacity: 0;
  transform: translateX(100%);
}

.task-slide-leave-to {
  opacity: 0;
  transform: translateX(100%) scale(0.8);
}

@media (max-width: 768px) {
  .global-task-container {
    right: 10px;
    left: 10px;
    max-width: none;
  }
}
</style>
