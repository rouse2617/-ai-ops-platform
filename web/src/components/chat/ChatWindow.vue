<template>
  <div class="chat-bento-container">
    <!-- 左侧会话列表 Bento Card -->
    <aside class="session-panel bento-card bento-card-interactive" :class="{ collapsed: sidebarCollapsed }">
      <div class="session-header">
        <template v-if="!sidebarCollapsed">
          <span class="session-title">对话</span>
          <button class="bento-btn-icon bento-btn-ghost" @click="handleNewSession" title="新建对话">
            <el-icon><Plus /></el-icon>
          </button>
        </template>
        <button class="bento-btn-icon bento-btn-ghost" @click="toggleSidebar" :title="sidebarCollapsed ? '展开' : '收起'">
          <el-icon><component :is="sidebarCollapsed ? Expand : Fold" /></el-icon>
        </button>
      </div>

      <!-- 展开状态 -->
      <div v-if="!sidebarCollapsed" class="session-list">
        <div
          v-for="session in sessions"
          :key="session.id"
          class="session-item"
          :class="{ active: session.id === currentSessionId }"
          @click="handleSessionSelect(session.id)"
        >
          <div class="session-item-icon">
            <el-icon><ChatDotRound /></el-icon>
          </div>
          <div class="session-item-content">
            <span class="session-item-title">{{ session.title || '未命名对话' }}</span>
          </div>
          <button class="session-delete" @click.stop="handleDeleteSession(session.id)">
            <el-icon><Delete /></el-icon>
          </button>
        </div>
        <div v-if="sessions.length === 0" class="empty-state">
          <el-icon class="empty-icon"><ChatDotRound /></el-icon>
          <p>暂无对话</p>
          <button class="bento-btn bento-btn-sm" @click="handleNewSession">开始新对话</button>
        </div>
      </div>

      <!-- 折叠状态 -->
      <div v-else class="session-icons">
        <button class="session-icon-btn" @click="handleNewSession" title="新建对话">
          <el-icon><Plus /></el-icon>
        </button>
        <div
          v-for="session in sessions.slice(0, 6)"
          :key="session.id"
          class="session-icon-btn"
          :class="{ active: session.id === currentSessionId }"
          @click="handleSessionSelect(session.id)"
          :title="session.title || '未命名对话'"
        >
          <el-icon><ChatDotRound /></el-icon>
        </div>
      </div>
    </aside>

    <!-- 主聊天区域 -->
    <main class="chat-main">
      <!-- 顶部信息栏 -->
      <header class="chat-header bento-card bento-card-interactive">
        <div class="header-left">
          <div class="chat-title-wrapper">
            <div class="chat-title-icon">
              <el-icon><ChatDotRound /></el-icon>
            </div>
            <div class="chat-title-text">
              <h1>{{ currentSessionTitle }}</h1>
              <span class="chat-subtitle">AI 智能运维助手</span>
            </div>
          </div>
        </div>
        <div class="header-right">
          <div class="host-selector-wrapper">
            <div class="host-label">
              <el-icon><Monitor /></el-icon>
              <span>目标主机</span>
            </div>
            <HostSelector v-model="selectedHostIds" />
            <span v-if="selectedHostIds.length > 0" class="bento-badge">{{ selectedHostIds.length }}</span>
          </div>
        </div>
      </header>

      <!-- 消息区域 -->
      <div class="message-container">
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

      <!-- 输入区域 -->
      <footer class="input-container">
        <div class="input-wrapper bento-card bento-card-interactive">
          <InputBox
            ref="inputBoxRef"
            :disabled="isLoading"
            @send="handleSend"
            @clear="handleClearMessages"
          />
        </div>
      </footer>
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

const sidebarCollapsed = ref(false)

const toggleSidebar = () => {
  sidebarCollapsed.value = !sidebarCollapsed.value
}

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
.chat-bento-container {
  display: flex;
  height: 100%;
  gap: var(--bento-gap);
  padding: var(--bento-gap);
  background: var(--bento-bg);
}

/* 会话面板 */
.session-panel {
  width: 260px;
  display: flex;
  flex-direction: column;
  padding: 0;
  overflow: hidden;
  transition: width 0.3s cubic-bezier(0.25, 0.46, 0.45, 0.94);
}

.session-panel.collapsed {
  width: 64px;
}

.session-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px;
  border-bottom: 1px solid var(--bento-border);
}

.session-panel.collapsed .session-header {
  justify-content: center;
  padding: 16px 12px;
}

.session-title {
  font-size: 15px;
  font-weight: 600;
  color: var(--color-text-primary);
}

.session-list {
  flex: 1;
  overflow-y: auto;
  padding: 12px;
}

.session-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  margin-bottom: 8px;
  border-radius: var(--bento-radius-sm);
  cursor: pointer;
  transition: all 0.2s;
  background: transparent;
}

.session-item:hover {
  background: var(--color-bg-secondary);
}

.session-item.active {
  background: rgba(0, 113, 227, 0.08);
}

.session-item-icon {
  width: 36px;
  height: 36px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--color-bg-secondary);
  color: var(--color-text-secondary);
  flex-shrink: 0;
  transition: all 0.2s;
}

.session-item.active .session-item-icon {
  background: var(--bento-accent-blue);
  color: white;
}

.session-item-content {
  flex: 1;
  min-width: 0;
}

.session-item-title {
  font-size: 14px;
  font-weight: 500;
  color: var(--color-text-primary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  display: block;
}

.session-delete {
  opacity: 0;
  width: 28px;
  height: 28px;
  border-radius: 8px;
  border: none;
  background: transparent;
  color: var(--color-text-tertiary);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
}

.session-item:hover .session-delete {
  opacity: 1;
}

.session-delete:hover {
  background: rgba(255, 55, 95, 0.1);
  color: var(--bento-accent-pink);
}

/* 空状态 */
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 40px 20px;
  text-align: center;
}

.empty-icon {
  font-size: 48px;
  color: var(--color-text-tertiary);
  margin-bottom: 16px;
}

.empty-state p {
  color: var(--color-text-secondary);
  margin-bottom: 16px;
}

/* 折叠状态图标 */
.session-icons {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 12px 8px;
  gap: 8px;
}

.session-icon-btn {
  width: 44px;
  height: 44px;
  border-radius: 12px;
  border: none;
  background: transparent;
  color: var(--color-text-secondary);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
  font-size: 18px;
}

.session-icon-btn:hover {
  background: var(--color-bg-secondary);
  color: var(--color-text-primary);
}

.session-icon-btn.active {
  background: rgba(0, 113, 227, 0.1);
  color: var(--bento-accent-blue);
}

/* 主聊天区域 */
.chat-main {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: var(--bento-gap);
  min-width: 0;
}

/* 头部 */
.chat-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 24px;
  flex-shrink: 0;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 16px;
}

.chat-title-wrapper {
  display: flex;
  align-items: center;
  gap: 14px;
}

.chat-title-icon {
  width: 44px;
  height: 44px;
  border-radius: 12px;
  background: var(--bento-gradient-ocean);
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-size: 20px;
}

.chat-title-text h1 {
  font-size: 17px;
  font-weight: 600;
  color: var(--color-text-primary);
  margin: 0;
  line-height: 1.3;
}

.chat-subtitle {
  font-size: 13px;
  color: var(--color-text-secondary);
}

.header-right {
  display: flex;
  align-items: center;
  gap: 16px;
}

.host-selector-wrapper {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 8px 16px;
  background: var(--color-bg-secondary);
  border-radius: 12px;
}

.host-label {
  display: flex;
  align-items: center;
  gap: 6px;
  color: var(--color-text-secondary);
  font-size: 13px;
  white-space: nowrap;
}

/* 消息区域 */
.message-container {
  flex: 1;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  background: var(--bento-card-bg);
  border-radius: var(--bento-radius);
  border: 1px solid var(--bento-border);
}

.message-container :deep(.message-list-container) {
  max-width: 900px;
  margin: 0 auto;
  width: 100%;
  padding: 24px;
}

/* 输入区域 */
.input-container {
  flex-shrink: 0;
}

.input-wrapper {
  padding: 16px 20px;
}

/* 响应式 */
@media (max-width: 900px) {
  .chat-bento-container {
    padding: 12px;
    gap: 12px;
  }

  .session-panel {
    width: 64px;
  }

  .session-panel .session-header {
    justify-content: center;
    padding: 12px 8px;
  }

  .session-panel .session-list {
    display: none;
  }

  .session-panel .session-icons {
    display: flex;
  }

  .chat-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
    padding: 16px;
  }

  .header-right {
    width: 100%;
  }

  .host-selector-wrapper {
    flex: 1;
  }
}

@media (max-width: 600px) {
  .session-panel {
    display: none;
  }

  .chat-bento-container {
    padding: 8px;
  }
}

/* 暗色模式适配 */
[data-theme="dark"] .session-item:hover {
  background: var(--color-surface-hover);
}

[data-theme="dark"] .session-item.active {
  background: rgba(0, 113, 227, 0.15);
}

[data-theme="dark"] .session-icon-btn:hover {
  background: var(--color-surface-hover);
}

[data-theme="dark"] .host-selector-wrapper {
  background: var(--color-surface);
}
</style>
