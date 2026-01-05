<template>
  <div class="input-box" role="form" aria-label="Chat input">
    <div class="input-wrapper">
      <div class="input-container">
        <!-- Host Chips Display -->
        <div v-if="selectedHostChips.length > 0" class="host-chips" role="list" aria-label="Selected hosts">
          <el-tag
            v-for="host in selectedHostChips"
            :key="host.id"
            closable
            @close="removeHostChip(host.id)"
            class="host-chip"
            role="listitem"
            :aria-label="`Remove ${host.name}`"
          >
            {{ host.name }}
          </el-tag>
        </div>

        <!-- Main Input -->
        <el-input
          ref="inputRef"
          v-model="inputText"
          type="textarea"
          :rows="1"
          :autosize="{ minRows: 1, maxRows: 4 }"
          placeholder="输入问题... (/ 命令, @ 主机)"
          resize="none"
          @keydown="handleKeydown"
          @input="handleInput"
          :disabled="disabled"
          aria-label="Message input"
          role="textbox"
          aria-multiline="true"
        />

        <!-- Slash Command Dropdown -->
        <div
          v-if="showSlashMenu"
          class="dropdown-menu slash-menu"
          :style="menuPosition"
          role="listbox"
          aria-label="Available commands"
        >
          <div
            v-for="(cmd, index) in filteredCommands"
            :key="cmd.name"
            class="menu-item"
            :class="{ active: index === slashMenuIndex }"
            @click="selectSlashCommand(cmd)"
            @mouseenter="slashMenuIndex = index"
            role="option"
            :aria-selected="index === slashMenuIndex"
            :id="`slash-cmd-${index}`"
            tabindex="-1"
          >
            <div class="menu-item-content">
              <el-icon aria-hidden="true"><component :is="cmd.icon" /></el-icon>
              <div class="menu-item-text">
                <span class="command-name">{{ cmd.name }}</span>
                <span class="command-desc">{{ cmd.description }}</span>
              </div>
            </div>
          </div>
        </div>

        <!-- Host Mention Dropdown -->
        <div
          v-if="showHostMenu"
          class="dropdown-menu host-menu"
          :style="menuPosition"
          role="listbox"
          aria-label="Available hosts"
        >
          <div
            v-for="(host, index) in filteredHosts"
            :key="host.id"
            class="menu-item"
            :class="{ active: index === hostMenuIndex }"
            @click="selectHostMention(host)"
            @mouseenter="hostMenuIndex = index"
            role="option"
            :aria-selected="index === hostMenuIndex"
            :id="`host-mention-${index}`"
            tabindex="-1"
          >
            <div class="menu-item-content">
              <el-tag
                :type="host.status === 'online' ? 'success' : 'info'"
                size="small"
              >
                {{ host.status === 'online' ? '在线' : '离线' }}
              </el-tag>
              <div class="menu-item-text">
                <span class="host-name">{{ host.name }}</span>
                <span class="host-address">{{ host.host }}:{{ host.port }}</span>
              </div>
            </div>
          </div>
        </div>

        <!-- Slot Filling Dropdown -->
        <SlotFillingDropdown
          ref="slotFillingRef"
          :visible="showSlotFilling"
          :slots="slotSuggestions"
          :slot-type="slotType"
          :position="menuPosition"
          @select="selectSlot"
          @close="closeSlotFilling"
        />
      </div>

      <!-- Danger Confirmation Dialog -->
      <DangerConfirmDialog
        v-if="currentDangerOp"
        v-model:visible="showDangerDialog"
        :operation="currentDangerOp"
        :host-count="currentDangerOp.affectedHosts.length || 1"
        :estimated-impact="getCommandImpact(currentDangerOp.command)"
        @confirm="handleDangerConfirm"
        @cancel="handleDangerCancel"
      />

      <div class="input-actions">
        <el-tooltip content="清空对话" placement="top">
          <el-button
            :icon="Delete"
            circle
            size="small"
            @click="$emit('clear')"
            :disabled="disabled"
            aria-label="Clear conversation"
          />
        </el-tooltip>
        <el-button
          type="primary"
          :icon="Promotion"
          :loading="disabled"
          @click="handleSend"
          aria-label="Send message"
          :disabled="!canSend || disabled"
        >
          发送
        </el-button>
      </div>
    </div>
    <div class="input-hints">
      <span class="hint-text">Enter 发送 · Shift+Enter 换行 · ↑↓ 历史</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, nextTick, onMounted, onUnmounted } from 'vue'
import { Promotion, Delete, TrendCharts, Document, CircleCheck, FolderOpened, Cpu, Operation } from '@element-plus/icons-vue'
import { SLASH_COMMANDS, type SlashCommand } from '@/types/chat-ui'
import { useHostStore } from '@/stores/host'
import type { Host } from '@/api/host'
import { useCommandHistory } from '@/composables/useCommandHistory'
import SlotFillingDropdown from './SlotFillingDropdown.vue'
import DangerConfirmDialog from './DangerConfirmDialog.vue'
import { analyzeMessage, type SlotSuggestion } from '@/api/slotFilling'
import { debounce } from 'lodash-es'
import { useDangerStore } from '@/stores/danger'
import { getCommandImpact } from '@/utils/dangerousCommands'

defineProps<{
  disabled: boolean
}>()

const emit = defineEmits<{
  send: [message: string]
  success: []
  clear: []
}>()

const hostStore = useHostStore()
const dangerStore = useDangerStore()
const inputRef = ref<any>()
const inputText = ref('')
const { addCommand, getPrevious, getNext, resetIndex } = useCommandHistory()

// Danger confirmation state
const showDangerDialog = ref(false)
const pendingMessage = ref('')
const currentDangerOp = ref<any>(null)

// Icon mapping for slash commands
const iconMap: Record<string, any> = {
  TrendCharts,
  Document,
  CircleCheck,
  FolderOpened,
  Cpu,
  Operation
}

// Enhanced slash commands with icon components
const slashCommands = ref<SlashCommand[]>(
  SLASH_COMMANDS.map(cmd => ({
    ...cmd,
    icon: iconMap[cmd.icon] || TrendCharts
  }))
)

// Host mention state
const mentionedHosts = ref<Map<string, Host>>(new Map())
const selectedHostChips = computed(() => Array.from(mentionedHosts.value.values()))

// Menu state
const showSlashMenu = ref(false)
const showHostMenu = ref(false)
const slashMenuIndex = ref(0)
const hostMenuIndex = ref(0)
const menuPosition = ref({ top: '0px', left: '0px' })

// Slot filling state
const showSlotFilling = ref(false)
const slotSuggestions = ref<SlotSuggestion[]>([])
const slotType = ref('')
const slotFillingRef = ref<any>()

// Computed properties
const filteredCommands = computed(() => {
  const match = inputText.value.match(/\/(\w*)$/)
  if (!match) return []
  const query = match[1].toLowerCase()
  return slashCommands.value.filter(cmd =>
    cmd.name.toLowerCase().startsWith(query) || cmd.description.toLowerCase().includes(query)
  )
})

const filteredHosts = computed(() => {
  const match = inputText.value.match(/@(\w*)$/)
  if (!match) return []
  const query = match[1].toLowerCase()
  return hostStore.allHosts.filter(host =>
    host.name.toLowerCase().includes(query) ||
    host.host.toLowerCase().includes(query)
  )
})

const canSend = computed(() => {
  return inputText.value.trim() && !showSlashMenu.value && !showHostMenu.value && !showSlotFilling.value
})

// 防抖分析消息
const analyzeMessageDebounced = debounce(async (message: string) => {
  if (!message.trim() || showSlashMenu.value || showHostMenu.value) {
    return
  }

  try {
    const response = await analyzeMessage({ message })
    if (response.needsSlotFilling) {
      slotSuggestions.value = response.suggestions
      slotType.value = response.slotType
      showSlotFilling.value = true
      updateMenuPosition()
    }
  } catch (error) {
    console.error('分析消息失败:', error)
  }
}, 800)

// Load hosts on mount
onMounted(async () => {
  await hostStore.loadAllHosts()
  document.addEventListener('click', handleClickOutside)
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
})

// Input handler
const handleInput = () => {
  const value = inputText.value

  // Check for slash command
  const slashMatch = value.match(/\/(\w*)$/)
  if (slashMatch && value.trim().startsWith('/')) {
    showSlashMenu.value = true
    showHostMenu.value = false
    slashMenuIndex.value = 0
    updateMenuPosition()
    return
  }

  // Check for host mention
  const hostMatch = value.match(/@(\w*)$/)
  if (hostMatch) {
    showHostMenu.value = true
    showSlashMenu.value = false
    hostMenuIndex.value = 0
    updateMenuPosition()
    return
  }

  // Hide menus if no match
  showSlashMenu.value = false
  showHostMenu.value = false

  // 分析是否需要参数补全
  analyzeMessageDebounced(value)
}

// Update menu position
const updateMenuPosition = () => {
  nextTick(() => {
    const textarea = inputRef.value?.$el?.querySelector('textarea')
    if (textarea) {
      const rect = textarea.getBoundingClientRect()
      menuPosition.value = {
        top: `${rect.bottom + 4}px`,
        left: `${rect.left}px`
      }
    }
  })
}

// Select slash command
const selectSlashCommand = (command: SlashCommand) => {
  const value = inputText.value
  const match = value.match(/^(.*)\/\w*$/)
  if (match) {
    inputText.value = match[1] + command.template
  }
  showSlashMenu.value = false
  inputRef.value?.focus()
}

// Select host mention
const selectHostMention = (host: Host) => {
  const value = inputText.value
  const match = value.match(/^(.*)@\w*$/)
  if (match) {
    inputText.value = match[1]
    // Add to mentioned hosts
    mentionedHosts.value.set(host.id, host)
  }
  showHostMenu.value = false
  inputRef.value?.focus()
}

// Remove host chip
const removeHostChip = (hostId: string) => {
  mentionedHosts.value.delete(hostId)
}

// Select slot
const selectSlot = (slot: SlotSuggestion) => {
  inputText.value = `${inputText.value} ${slot.value}`
  showSlotFilling.value = false
  inputRef.value?.focus()
}

// Close slot filling
const closeSlotFilling = () => {
  showSlotFilling.value = false
}

// Keyboard navigation
const handleKeydown = (e: KeyboardEvent) => {
  // Handle slot filling navigation
  if (showSlotFilling.value) {
    if (e.key === 'ArrowDown') {
      e.preventDefault()
      slotFillingRef.value?.navigateDown()
    } else if (e.key === 'ArrowUp') {
      e.preventDefault()
      slotFillingRef.value?.navigateUp()
    } else if (e.key === 'Enter') {
      e.preventDefault()
      slotFillingRef.value?.selectActive()
    } else if (e.key === 'Escape') {
      showSlotFilling.value = false
    }
    return
  }

  // Handle menu navigation
  if (showSlashMenu.value) {
    if (e.key === 'ArrowDown') {
      e.preventDefault()
      slashMenuIndex.value = (slashMenuIndex.value + 1) % filteredCommands.value.length
    } else if (e.key === 'ArrowUp') {
      e.preventDefault()
      slashMenuIndex.value = (slashMenuIndex.value - 1 + filteredCommands.value.length) % filteredCommands.value.length
    } else if (e.key === 'Enter') {
      e.preventDefault()
      const selected = filteredCommands.value[slashMenuIndex.value]
      if (selected) {
        selectSlashCommand(selected)
      }
    } else if (e.key === 'Escape') {
      showSlashMenu.value = false
    }
    return
  }

  if (showHostMenu.value) {
    if (e.key === 'ArrowDown') {
      e.preventDefault()
      hostMenuIndex.value = (hostMenuIndex.value + 1) % filteredHosts.value.length
    } else if (e.key === 'ArrowUp') {
      e.preventDefault()
      hostMenuIndex.value = (hostMenuIndex.value - 1 + filteredHosts.value.length) % filteredHosts.value.length
    } else if (e.key === 'Enter') {
      e.preventDefault()
      const selected = filteredHosts.value[hostMenuIndex.value]
      if (selected) {
        selectHostMention(selected)
      }
    } else if (e.key === 'Escape') {
      showHostMenu.value = false
    }
    return
  }

  // Handle command history navigation
  if (e.key === 'ArrowUp' && !e.shiftKey) {
    e.preventDefault()
    const prev = getPrevious()
    if (prev !== null) {
      inputText.value = prev
    }
    return
  }

  if (e.key === 'ArrowDown' && !e.shiftKey) {
    e.preventDefault()
    const next = getNext()
    if (next !== null) {
      inputText.value = next
    }
    return
  }

  // Handle send
  if (e.key === 'Enter' && !e.shiftKey) {
    e.preventDefault()
    handleSend()
  }
}

// Send message
const handleSend = async () => {
  if (!canSend.value) return

  // Include mentioned hosts in the message
  let message = inputText.value.trim()

  if (selectedHostChips.value.length > 0) {
    const hostNames = selectedHostChips.value.map(h => h.name).join(', ')
    message = `[目标主机: ${hostNames}]\n${message}`
  }

  // Check for dangerous commands
  const { level, reason, type } = dangerStore.checkDangerLevel(message)

  if (level === 'high' || level === 'critical') {
    // Show confirmation dialog
    pendingMessage.value = message
    currentDangerOp.value = {
      id: `op-${Date.now()}`,
      command: message,
      type: type || 'modify',
      riskLevel: level,
      affectedHosts: selectedHostChips.value.map(h => h.name),
      reason
    }
    showDangerDialog.value = true
    return
  }

  // Safe to send
  sendMessage(message)
}

const sendMessage = (message: string) => {
  // Add to command history
  addCommand(inputText.value.trim())

  emit('send', message)

  // Clear input and reset state
  inputText.value = ''
  mentionedHosts.value.clear()
  showSlotFilling.value = false
  resetIndex()
}

const handleDangerConfirm = () => {
  showDangerDialog.value = false
  if (pendingMessage.value) {
    sendMessage(pendingMessage.value)
    pendingMessage.value = ''
    currentDangerOp.value = null
  }
}

const handleDangerCancel = () => {
  showDangerDialog.value = false
  pendingMessage.value = ''
  currentDangerOp.value = null
}

// Clear input function
function clearInput() {
  inputText.value = ''
  mentionedHosts.value.clear()
  showSlashMenu.value = false
  showHostMenu.value = false
  showSlotFilling.value = false
  resetIndex()
}

defineExpose({
  clearInput
})

// Close menus when clicking outside
const handleClickOutside = (e: MouseEvent) => {
  const target = e.target as Node
  const inputContainer = inputRef.value?.$el
  const dropdownMenus = document.querySelectorAll('.dropdown-menu')

  let clickedInsideMenu = false
  dropdownMenus.forEach(menu => {
    if (menu.contains(target)) {
      clickedInsideMenu = true
    }
  })

  if (!clickedInsideMenu && inputContainer && !inputContainer.contains(target)) {
    showSlashMenu.value = false
    showHostMenu.value = false
    showSlotFilling.value = false
  }
}
</script>

<style scoped>
.input-box {
  padding: 12px 16px;
  background: #fff;
  position: relative;
}

.input-wrapper {
  display: flex;
  gap: 12px;
  align-items: flex-end;
  position: relative;
}

.input-container {
  flex: 1;
  min-width: 0;
  position: relative;
}

.host-chips {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  padding: 8px 0;
  margin-bottom: 4px;
}

.host-chip {
  user-select: none;
  transition: all 0.2s ease;
}

.host-chip:hover {
  transform: translateY(-1px);
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.1);
}

.input-container :deep(.el-textarea__inner) {
  padding: 10px 12px;
  font-size: 14px;
  line-height: 1.4;
  border-radius: 8px;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.04);
  transition: all 0.3s ease;
}

.input-container :deep(.el-textarea__inner:focus) {
  box-shadow: 0 2px 12px rgba(64, 158, 255, 0.2);
}

/* Dropdown Menu Styles */
.dropdown-menu {
  position: fixed;
  background: #fff;
  border: 1px solid #e4e7ed;
  border-radius: 8px;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.12);
  max-height: 300px;
  overflow-y: auto;
  z-index: 2000;
  padding: 6px 0;
  animation: dropdownFadeIn 0.2s ease;
}

@keyframes dropdownFadeIn {
  from {
    opacity: 0;
    transform: translateY(-8px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.menu-item {
  padding: 10px 16px;
  cursor: pointer;
  transition: all 0.2s ease;
  user-select: none;
}

.menu-item:hover,
.menu-item.active {
  background: #f5f7fa;
}

.menu-item.active {
  background: #ecf5ff;
}

.menu-item-content {
  display: flex;
  align-items: center;
  gap: 12px;
}

.menu-item-text {
  display: flex;
  flex-direction: column;
  gap: 2px;
  flex: 1;
  min-width: 0;
}

.command-name {
  font-weight: 500;
  color: #303133;
  font-size: 14px;
}

.command-desc {
  font-size: 12px;
  color: #909399;
}

.host-name {
  font-weight: 500;
  color: #303133;
  font-size: 14px;
}

.host-address {
  font-size: 12px;
  color: #909399;
}

/* Slash Command Menu Specifics */
.slash-menu {
  min-width: 280px;
}

.slash-menu .el-icon {
  font-size: 20px;
  color: #409eff;
}

/* Host Mention Menu Specifics */
.host-menu {
  min-width: 320px;
}

.input-actions {
  display: flex;
  gap: 8px;
  flex-shrink: 0;
}

.input-hints {
  display: flex;
  justify-content: center;
  margin-top: 6px;
}

.hint-text {
  font-size: 11px;
  color: #c0c4cc;
  letter-spacing: 0.3px;
}

/* Scrollbar styling for dropdown menus */
.dropdown-menu::-webkit-scrollbar {
  width: 6px;
}

.dropdown-menu::-webkit-scrollbar-track {
  background: #f5f7fa;
  border-radius: 3px;
}

.dropdown-menu::-webkit-scrollbar-thumb {
  background: #dcdfe6;
  border-radius: 3px;
}

.dropdown-menu::-webkit-scrollbar-thumb:hover {
  background: #c0c4cc;
}
</style>
