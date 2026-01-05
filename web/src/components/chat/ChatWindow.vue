<template>
  <div class="chat-window-container" :class="{ 'sidebar-collapsed': sidebarCollapsed }">
    <!-- 可折叠侧边栏 -->
    <aside class="session-sidebar">
      <div class="sidebar-header">
        <template v-if="!sidebarCollapsed">
          <span class="sidebar-title">对话</span>
          <el-button :icon="Plus" size="small" circle @click="handleNewSession" title="新建对话" />
        </template>
        <el-button
          :icon="sidebarCollapsed ? Expand : Fold"
          size="small"
          circle
          @click="toggleSidebar"
          :title="sidebarCollapsed ? '展开' : '收起'"
        />
      </div>

      <!-- 展开状态：显示会话列表 -->
      <div v-if="!sidebarCollapsed" class="session-list">
        <div
          v-for="session in sessions"
          :key="session.id"
          class="session-item"
          :class="{ active: session.id === currentSessionId }"
          @click="handleSessionSelect(session.id)"
        >
          <el-icon class="session-icon"><ChatDotRound /></el-icon>
          <span class="session-title">{{ session.title || '未命名对话' }}</span>
          <el-icon class="session-delete" @click.stop="handleDeleteSession(session.id)"><Delete /></el-icon>
        </div>
        <div v-if="sessions.length === 0" class="empty-sessions">
          <p>暂无对话</p>
        </div>
      </div>

      <!-- 折叠状态：只显示图标 -->
      <div v-else class="session-icons">
        <el-button :icon="Plus" circle @click="handleNewSession" title="新建对话" />
        <div
          v-for="session in sessions.slice(0, 5)"
          :key="session.id"
          class="session-icon-item"
          :class="{ active: session.id === currentSessionId }"
          @click="handleSessionSelect(session.id)"
          :title="session.title || '未命名对话'"
        >
          <el-icon><ChatDotRound /></el-icon>
        </div>
      </div>
    </aside>

    <!-- 聊天主区域 -->
    <main class="chat-panel">
      <div class="chat-header">
        <span class="current-title">{{ currentSessionTitle }}</span>
        <div class="host-section">
          <el-icon class="host-icon"><Monitor /></el-icon>
          <HostSelector v-model="selectedHostIds" />
          <el-badge v-if="selectedHostIds.length > 0" :value="selectedHostIds.length" />
        </div>
      </div>

      <div class="message-area">
        <MessageList
          :messages="messages"
          :is-loading="isLoading"
          :thinking-steps="thinkingSteps"
          :current-thinking-status="currentThinkingStatus"
          :session-loading="sessionLoading"
          ref="messageListRef"
          @send-message="handleSendFromList"
          @regenerate="handleRegenerate"
          @tool-rerun="handleToolRerun"
        />
        <PrometheusIntegration
          ref="prometheusRef"
          :message-content="lastUserMessage"
          @send-message="handleSendFromList"
          @execute-action="handlePrometheusAction"
        />
      </div>

      <div class="input-area">
        <InputBox
          ref="inputBoxRef"
          :disabled="isLoading"
          @send="handleSend"
          @clear="handleClearMessages"
        />
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, onMounted, watch } from 'vue'
import { useChatStore } from '@/stores/chat'
import { usePanelSync } from '@/composables/usePanelSync'
import type { Message } from '@/api/chat'
import { ChatDotRound, Plus, Delete, Monitor, Expand, Fold } from '@element-plus/icons-vue'
import { ElMessageBox, ElMessage } from 'element-plus'
import MessageList from './MessageList.vue'
import InputBox from './InputBox.vue'
import HostSelector from './HostSelector.vue'
import PrometheusIntegration from './PrometheusIntegration.vue'

const chatStore = useChatStore()
const { syncToHost } = usePanelSync()
const messageListRef = ref()
const inputBoxRef = ref()
const prometheusRef = ref()

// 侧边栏折叠状态
const sidebarCollapsed = ref(false)

const toggleSidebar = () => {
  sidebarCollapsed.value = !sidebarCollapsed.value
}

// Chat store state
const messages = computed(() => chatStore.messages)
const isLoading = computed(() => chatStore.isLoading)
const sessions = computed(() => chatStore.sessions)
const thinkingSteps = computed(() => chatStore.thinkingSteps)
const currentThinkingStatus = computed(() => chatStore.currentThinkingStatus)
const sessionLoading = computed(() => chatStore.sessionLoading)
const currentSessionId = computed(() => chatStore.currentSessionId)

const selectedHostIds = computed({
  get: () => chatStore.selectedHostIds,
  set: (val) => chatStore.setSelectedHosts(val)
})

const currentSessionTitle = computed(() => {
  const session = sessions.value.find(s => s.id === chatStore.currentSessionId)
  return session?.title || '新对话'
})

const lastUserMessage = computed(() => {
  const userMessages = messages.value.filter(m => m.role === 'user')
  return userMessages.length > 0 ? userMessages[userMessages.length - 1].content : ''
})

const handlePrometheusAction = async (action: { id: string; command?: string }) => {
  if (action.command) {
    await chatStore.sendMessage(`执行命令: ${action.command}`)
  }
}

onMounted(async () => {
  await chatStore.loadSessions()
  if (!chatStore.currentSessionId) {
    await chatStore.newSession()
  }
})

watch(() => messages.value, (newMessages) => {
  if (newMessages.length === 0) return
  const lastMessage = newMessages[newMessages.length - 1]
  if (lastMessage.role === 'assistant' && selectedHostIds.value.length > 0) {
    syncToHost(selectedHostIds.value[0])
  }
}, { deep: true })

const handleSessionSelect = async (sessionId: string) => {
  await chatStore.switchSession(sessionId)
}

const handleNewSession = async () => {
  await chatStore.newSession()
}

const handleDeleteSession = async (sessionId: string) => {
  try {
    await ElMessageBox.confirm('确定要删除这个对话吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    await chatStore.removeSession(sessionId)
  } catch {}
}

const handleSend = async (message: string) => {
  try {
    await chatStore.sendMessage(message)
    inputBoxRef.value?.clearInput()
  } catch (error) {
    console.error('发送消息失败:', error)
  }
}

const handleSendFromList = async (message: string) => {
  try {
    await chatStore.sendMessage(message)
  } catch (error) {
    console.error('发送消息失败:', error)
  }
}

const handleRegenerate = async (message: Message) => {
  try {
    const messageIndex = messages.value.findIndex(m => m === message)
    if (messageIndex > 0) {
      const userMessage = messages.value[messageIndex - 1]
      if (userMessage.role === 'user') {
        messages.value.splice(messageIndex, 1)
        await chatStore.sendMessage(userMessage.content)
      }
    }
  } catch (error) {
    console.error('重新生成失败:', error)
    ElMessage.error('重新生成失败，请稍后重试')
  }
}

const handleToolRerun = async (toolCall: any) => {
  try {
    ElMessage.info(`正在重新执行工具: ${toolCall.name}`)
  } catch (error) {
    console.error('工具重新执行失败:', error)
    ElMessage.error('工具重新执行失败，请稍后重试')
  }
}

const handleClearMessages = async () => {
  try {
    await ElMessageBox.confirm('确定要清空当前对话吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    chatStore.clearMessages()
  } catch {}
}
</script>

<style scoped>
.chat-window-container {
  display: flex;
  height: 100%;
  width: 100%;
  background: #fff;
}

/* 侧边栏 - 浅色系 */
.session-sidebar {
  width: 200px;
  background: #f7f8fa;
  border-right: 1px solid #e8e8e8;
  display: flex;
  flex-direction: column;
  transition: width 0.2s ease;
  flex-shrink: 0;
}

.sidebar-collapsed .session-sidebar {
  width: 50px;
}

.sidebar-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px;
  border-bottom: 1px solid #e8e8e8;
  gap: 8px;
}

.sidebar-collapsed .sidebar-header {
  justify-content: center;
  padding: 12px 8px;
}

.sidebar-title {
  font-size: 14px;
  font-weight: 600;
  color: #303133;
}

.session-list {
  flex: 1;
  overflow-y: auto;
  padding: 8px;
}

.session-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 10px;
  margin-bottom: 4px;
  border-radius: 6px;
  cursor: pointer;
  color: #606266;
  transition: all 0.15s;
}

.session-item:hover {
  background: #ebeef5;
  color: #303133;
}

.session-item.active {
  background: #ecf5ff;
  color: #409eff;
}

.session-icon {
  flex-shrink: 0;
  font-size: 14px;
}

.session-title {
  flex: 1;
  font-size: 13px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.session-delete {
  opacity: 0;
  font-size: 14px;
  color: #909399;
  transition: opacity 0.15s;
}

.session-item:hover .session-delete {
  opacity: 1;
}

.session-delete:hover {
  color: #f56c6c;
}

.empty-sessions {
  text-align: center;
  color: #909399;
  font-size: 12px;
  padding: 20px;
}

/* 折叠状态图标列表 */
.session-icons {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 8px 0;
  gap: 8px;
}

.session-icon-item {
  width: 36px;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 8px;
  cursor: pointer;
  color: #606266;
  transition: all 0.15s;
}

.session-icon-item:hover {
  background: #ebeef5;
  color: #303133;
}

.session-icon-item.active {
  background: #ecf5ff;
  color: #409eff;
}

/* 聊天面板 */
.chat-panel {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  min-width: 0;
}

.chat-header {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 12px 20px;
  border-bottom: 1px solid #f0f0f0;
  flex-shrink: 0;
}

.current-title {
  font-size: 15px;
  font-weight: 500;
  color: #303133;
}

.host-section {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 8px;
  max-width: 400px;
  padding: 6px 12px;
  border: 1px solid #dcdfe6;
  border-radius: 8px;
}

.host-icon {
  color: #909399;
}

.message-area {
  flex: 1;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.message-area :deep(.message-list-container) {
  max-width: 900px;
  margin: 0 auto;
  width: 100%;
}

.input-area {
  background: #fff;
  flex-shrink: 0;
  box-shadow: 0 -2px 10px rgba(0, 0, 0, 0.03);
}
</style>
