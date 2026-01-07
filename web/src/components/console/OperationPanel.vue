<template>
  <div class="operation-panel">
    <el-tabs v-model="activeTab" @tab-change="handleTabChange">
      <el-tab-pane label="查看日志" name="query_log">
        <el-form :model="logForm" label-width="100px">
          <el-form-item label="日志类型">
            <el-select v-model="logForm.log_type" placeholder="选择日志类型">
              <el-option label="Nginx" value="nginx" />
              <el-option label="应用日志" value="app" />
              <el-option label="系统日志" value="system" />
              <el-option label="自定义" value="custom" />
            </el-select>
          </el-form-item>
          <el-form-item label="日志路径" v-if="logForm.log_type === 'custom'">
            <el-input v-model="logForm.path" placeholder="请输入日志路径" />
          </el-form-item>
          <el-form-item label="查看行数">
            <el-input-number v-model="logForm.lines" :min="1" :max="10000" />
          </el-form-item>
          <el-form-item label="关键词">
            <el-input v-model="logForm.keyword" placeholder="过滤关键词（可选）" />
          </el-form-item>
          <el-form-item label="日志级别">
            <el-select v-model="logForm.level" placeholder="选择级别（可选）" clearable>
              <el-option label="错误" value="error" />
              <el-option label="警告" value="warn" />
              <el-option label="信息" value="info" />
            </el-select>
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="handleExecuteLog" :loading="executing">
              执行查询
            </el-button>
          </el-form-item>
        </el-form>
      </el-tab-pane>

      <el-tab-pane label="执行命令" name="run_command">
        <el-form :model="commandForm" label-width="80px" class="compact-form">
          <el-form-item label="命令">
            <el-input
              v-model="commandForm.command"
              type="textarea"
              :autosize="{ minRows: 2, maxRows: 6 }"
              placeholder="请输入要执行的命令"
              :class="{ 'dangerous-command': isDangerous, 'critical-command': isCritical }"
            />
            <div v-if="dangerWarning" class="danger-warning">
              <el-icon><Warning /></el-icon>
              <span>{{ dangerWarning }}</span>
            </div>
          </el-form-item>

          <!-- 快捷命令 -->
          <el-form-item label="快捷命令">
            <div class="quick-commands">
              <el-tag
                v-for="cmd in quickCommands"
                :key="cmd.label"
                class="quick-cmd-tag"
                @click="fillCommand(cmd.command)"
                effect="plain"
              >
                {{ cmd.label }}
              </el-tag>
            </div>
          </el-form-item>

          <el-form-item label="超时">
            <div class="timeout-inline">
              <el-slider
                v-model="commandForm.timeout"
                :min="5"
                :max="300"
                :step="5"
                style="flex: 1"
              />
              <span class="timeout-value">{{ commandForm.timeout }}s</span>
              <el-button
                type="primary"
                @click="handleExecuteCommand"
                :loading="executing"
                :class="{ 'danger-btn': isCritical }"
              >
                {{ isCritical ? '执行危险命令' : '执行' }}
              </el-button>
            </div>
          </el-form-item>
        </el-form>
      </el-tab-pane>

      <el-tab-pane label="系统信息" name="system">
        <div class="system-actions">
          <el-button type="primary" @click="handleCheckCPU" :loading="executing">
            检查CPU
          </el-button>
          <el-button type="primary" @click="handleCheckMemory" :loading="executing">
            检查内存
          </el-button>
          <el-button type="primary" @click="handleCheckDisk" :loading="executing">
            检查磁盘
          </el-button>
        </div>
      </el-tab-pane>
    </el-tabs>

    <!-- 高危操作确认弹窗 -->
    <ConfirmationDialog
      v-model:visible="confirmDialogVisible"
      :request="confirmRequest"
      @confirm="handleConfirmExecution"
      @cancel="handleCancelExecution"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { Warning } from '@element-plus/icons-vue'
import { useConsoleStore } from '@/stores/console'
import {
  isDangerousCommand,
  isCriticalCommand,
  getDangerousCommandWarning,
  highlightDangerousKeywords,
  getCommandRiskLevel,
  getCommandImpact
} from '@/utils/dangerousCommands'
import ConfirmationDialog from '@/components/common/ConfirmationDialog.vue'
import type { ConfirmationRequest } from '@/types/chat-ui'

const consoleStore = useConsoleStore()

const activeTab = ref('query_log')
const executing = ref(false)
const confirmDialogVisible = ref(false)
const confirmRequest = ref<ConfirmationRequest | null>(null)
const pendingCommand = ref<{ command: string; timeout: number } | null>(null)

// 快捷命令
const quickCommands = [
  { label: 'Top', command: 'top -bn1 | head -20' },
  { label: '磁盘', command: 'df -h' },
  { label: '内存', command: 'free -h' },
  { label: '进程', command: 'ps aux --sort=-%cpu | head -10' },
  { label: '网络', command: 'netstat -tlnp' },
  { label: '日志', command: 'tail -100 /var/log/syslog' }
]

const fillCommand = (cmd: string) => {
  commandForm.value.command = cmd
}

const logForm = ref({
  log_type: 'nginx',
  path: '',
  lines: 100,
  keyword: '',
  level: ''
})

const commandForm = ref({
  command: '',
  timeout: 30
})

// 检测命令是否危险
const isDangerous = computed(() => {
  return commandForm.value.command ? isDangerousCommand(commandForm.value.command) : false
})

const isCritical = computed(() => {
  return commandForm.value.command ? isCriticalCommand(commandForm.value.command) : false
})

const dangerWarning = computed(() => {
  return commandForm.value.command ? getDangerousCommandWarning(commandForm.value.command) : ''
})

const highlightedCommand = computed(() => {
  return commandForm.value.command ? highlightDangerousKeywords(commandForm.value.command) : ''
})

// 超时滑块标记
const timeoutMarks = computed(() => ({
  5: '5s',
  30: '30s',
  60: '1m',
  120: '2m',
  300: '5m'
}))

const handleTabChange = (tab: string) => {
  // 切换标签时清空表单
  if (tab === 'query_log') {
    logForm.value = {
      log_type: 'nginx',
      path: '',
      lines: 100,
      keyword: '',
      level: ''
    }
  } else if (tab === 'run_command') {
    commandForm.value = {
      command: '',
      timeout: 30
    }
  }
}

const executeOperation = async (
  operation: 'query_log' | 'run_command' | 'check_cpu' | 'check_memory' | 'check_disk',
  params?: Record<string, any>
) => {
  if (consoleStore.selectedHosts.length === 0) {
    ElMessage.warning('请先选择主机')
    return
  }

  executing.value = true
  try {
    await consoleStore.executeOperation(operation, consoleStore.selectedHosts, params)
    ElMessage.success('执行成功')
  } catch (error: any) {
    ElMessage.error(error.message || '执行失败')
  } finally {
    executing.value = false
  }
}

const handleExecuteLog = () => {
  const params: Record<string, any> = {
    log_type: logForm.value.log_type,
    lines: logForm.value.lines
  }
  if (logForm.value.path) {
    params.path = logForm.value.path
  }
  if (logForm.value.keyword) {
    params.keyword = logForm.value.keyword
  }
  if (logForm.value.level) {
    params.level = logForm.value.level
  }
  executeOperation('query_log', params)
}

const handleExecuteCommand = async () => {
  if (!commandForm.value.command.trim()) {
    ElMessage.warning('请输入命令')
    return
  }

  const command = commandForm.value.command.trim()
  const riskLevel = getCommandRiskLevel(command)

  // 低风险命令直接执行
  if (riskLevel === 'low') {
    executeOperation('run_command', {
      command: command,
      timeout: commandForm.value.timeout
    })
    return
  }

  // 中高风险命令显示确认弹窗
  pendingCommand.value = {
    command: command,
    timeout: commandForm.value.timeout
  }

  confirmRequest.value = {
    id: `cmd-${Date.now()}`,
    command: command,
    riskLevel: riskLevel,
    affectedHosts: consoleStore.selectedHosts,
    requiresTyping: riskLevel === 'critical',
    estimatedImpact: getCommandImpact(command)
  }

  confirmDialogVisible.value = true
}

const handleConfirmExecution = () => {
  if (!pendingCommand.value) return

  executeOperation('run_command', {
    command: pendingCommand.value.command,
    timeout: pendingCommand.value.timeout
  })

  pendingCommand.value = null
}

const handleCancelExecution = () => {
  pendingCommand.value = null
  ElMessage.info('已取消命令执行')
}

const handleCheckCPU = () => {
  executeOperation('check_cpu')
}

const handleCheckMemory = () => {
  executeOperation('check_memory')
}

const handleCheckDisk = () => {
  executeOperation('check_disk')
}
</script>

<style scoped>
.operation-panel {
  padding: 12px 16px;
}

.compact-form :deep(.el-form-item) {
  margin-bottom: 12px;
}

.compact-form :deep(.el-form-item__label) {
  font-size: 13px;
}

/* 快捷命令 */
.quick-commands {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.quick-cmd-tag {
  cursor: pointer;
  transition: all 0.2s;
  background: #ecf5ff;
  border-color: #d9ecff;
  color: #409eff;
}

.quick-cmd-tag:hover {
  background: #409eff;
  color: #fff;
  transform: translateY(-1px);
}

/* 超时和执行按钮内联 */
.timeout-inline {
  display: flex;
  align-items: center;
  gap: 12px;
  width: 100%;
}

.timeout-inline .timeout-value {
  min-width: 40px;
  font-size: 13px;
  font-weight: 500;
  color: #409eff;
}

.danger-btn {
  background: #f56c6c !important;
  border-color: #f56c6c !important;
}

.system-actions {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
}

.dangerous-command :deep(.el-textarea__inner) {
  border-color: #e6a23c;
}

.critical-command :deep(.el-textarea__inner) {
  border-color: #f56c6c;
  background-color: #fef0f0;
}

.danger-warning {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 8px;
  padding: 8px 12px;
  background-color: #fdf6ec;
  border: 1px solid #e6a23c;
  border-radius: 4px;
  color: #e6a23c;
  font-size: 14px;
}

.critical-command ~ .danger-warning {
  background-color: #fef0f0;
  border-color: #f56c6c;
  color: #f56c6c;
}

.command-preview {
  margin-top: 8px;
  padding: 12px 16px;
  background: linear-gradient(135deg, #1a1a2e 0%, #16213e 100%);
  border: 1px solid #0f3460;
  border-radius: 6px;
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
  font-size: 13px;
  box-shadow: inset 0 1px 3px rgba(0, 0, 0, 0.3);
}

.preview-label {
  color: #00ff9d;
  font-size: 11px;
  margin-right: 8px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  font-weight: 600;
}

.preview-content {
  display: inline;
  color: #e0e0e0;
}

.preview-content :deep(.danger-keyword) {
  color: #ff6b6b;
  font-weight: 600;
}

.preview-content :deep(.warning-keyword) {
  color: #ffd93d;
  font-weight: 600;
}

/* 超时控制 */
.timeout-control {
  width: 100%;
  padding: 0 10px;
}

.timeout-value {
  display: block;
  text-align: center;
  margin-top: 8px;
  font-size: 13px;
  font-weight: 600;
  color: var(--el-color-primary);
}

.timeout-control :deep(.el-slider__marks-text) {
  font-size: 11px;
}
</style>
