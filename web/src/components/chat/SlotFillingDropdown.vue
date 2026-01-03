<template>
  <div
    v-if="visible && slots.length > 0"
    class="slot-filling-dropdown"
    :style="position"
    role="listbox"
    aria-label="Parameter suggestions"
  >
    <div class="dropdown-header">
      <el-icon class="header-icon"><InfoFilled /></el-icon>
      <span class="header-text">{{ headerText }}</span>
    </div>

    <div class="slot-list">
      <div
        v-for="(slot, index) in slots"
        :key="slot.value"
        class="slot-item"
        :class="{ active: index === activeIndex }"
        @click="selectSlot(slot)"
        @mouseenter="activeIndex = index"
        role="option"
        :aria-selected="index === activeIndex"
        tabindex="-1"
      >
        <div class="slot-content">
          <el-icon class="slot-icon" :style="{ color: slot.iconColor }">
            <component :is="slot.icon" />
          </el-icon>
          <div class="slot-text">
            <span class="slot-label">{{ slot.label }}</span>
            <span v-if="slot.description" class="slot-desc">{{ slot.description }}</span>
          </div>
          <el-tag v-if="slot.tag" size="small" :type="slot.tagType">{{ slot.tag }}</el-tag>
        </div>
      </div>
    </div>

    <div class="dropdown-footer">
      <span class="footer-hint">↑↓ 选择 · Enter 确认 · Esc 取消</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { InfoFilled, Service, Files, Operation, Setting } from '@element-plus/icons-vue'

export interface SlotOption {
  value: string
  label: string
  description?: string
  icon?: any
  iconColor?: string
  tag?: string
  tagType?: 'success' | 'info' | 'warning' | 'danger'
}

interface Props {
  visible: boolean
  slots: SlotOption[]
  slotType: string
  position?: { top: string; left: string }
}

const props = withDefaults(defineProps<Props>(), {
  position: () => ({ top: '0px', left: '0px' })
})

const emit = defineEmits<{
  select: [slot: SlotOption]
  close: []
}>()

const activeIndex = ref(0)

const headerText = computed(() => {
  const typeMap: Record<string, string> = {
    service: '请选择要操作的服务',
    file: '请选择文件',
    process: '请选择进程',
    port: '请选择端口',
    user: '请选择用户',
    default: '请选择参数'
  }
  return typeMap[props.slotType] || typeMap.default
})

const selectSlot = (slot: SlotOption) => {
  emit('select', slot)
}

const navigateUp = () => {
  if (activeIndex.value > 0) {
    activeIndex.value--
  }
}

const navigateDown = () => {
  if (activeIndex.value < props.slots.length - 1) {
    activeIndex.value++
  }
}

const selectActive = () => {
  if (props.slots[activeIndex.value]) {
    selectSlot(props.slots[activeIndex.value])
  }
}

const close = () => {
  emit('close')
}

defineExpose({
  navigateUp,
  navigateDown,
  selectActive,
  close
})
</script>

<style scoped>
.slot-filling-dropdown {
  position: fixed;
  background: #fff;
  border: 1px solid #e4e7ed;
  border-radius: 8px;
  box-shadow: 0 6px 20px rgba(0, 0, 0, 0.15);
  min-width: 320px;
  max-width: 480px;
  max-height: 400px;
  z-index: 2001;
  animation: slideUp 0.2s ease;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

@keyframes slideUp {
  from {
    opacity: 0;
    transform: translateY(8px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.dropdown-header {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 16px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: #fff;
  font-weight: 500;
  font-size: 14px;
}

.header-icon {
  font-size: 16px;
}

.header-text {
  flex: 1;
}

.slot-list {
  flex: 1;
  overflow-y: auto;
  padding: 6px 0;
}

.slot-item {
  padding: 10px 16px;
  cursor: pointer;
  transition: all 0.2s ease;
  user-select: none;
}

.slot-item:hover,
.slot-item.active {
  background: #f5f7fa;
}

.slot-item.active {
  background: #ecf5ff;
  border-left: 3px solid #409eff;
}

.slot-content {
  display: flex;
  align-items: center;
  gap: 12px;
}

.slot-icon {
  font-size: 20px;
  flex-shrink: 0;
}

.slot-text {
  display: flex;
  flex-direction: column;
  gap: 2px;
  flex: 1;
  min-width: 0;
}

.slot-label {
  font-weight: 500;
  color: #303133;
  font-size: 14px;
}

.slot-desc {
  font-size: 12px;
  color: #909399;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.dropdown-footer {
  padding: 8px 16px;
  background: #f5f7fa;
  border-top: 1px solid #e4e7ed;
}

.footer-hint {
  font-size: 11px;
  color: #909399;
}

.slot-list::-webkit-scrollbar {
  width: 6px;
}

.slot-list::-webkit-scrollbar-track {
  background: #f5f7fa;
}

.slot-list::-webkit-scrollbar-thumb {
  background: #dcdfe6;
  border-radius: 3px;
}

.slot-list::-webkit-scrollbar-thumb:hover {
  background: #c0c4cc;
}
</style>
