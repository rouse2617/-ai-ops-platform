<template>
  <div class="console-page">
    <!-- 页面头部 - 改进 -->
    <div class="page-header">
      <div class="header-content">
        <h2 class="page-title">批量操作控制台</h2>
        <p class="page-description">选择多个主机节点，执行批量操作，查看执行结果</p>
      </div>

      <!-- 快速统计 -->
      <div class="quick-stats">
        <div class="stat-item">
          <span class="stat-label">已选择</span>
          <span class="stat-value selected">{{ selectedCount }}</span>
        </div>
        <div class="stat-divider"></div>
        <div class="stat-item">
          <span class="stat-label">在线</span>
          <span class="stat-value online">{{ onlineCount }}</span>
        </div>
        <div class="stat-divider"></div>
        <div class="stat-item">
          <span class="stat-label">离线</span>
          <span class="stat-value offline">{{ offlineCount }}</span>
        </div>
      </div>

      <!-- 移动端右侧面板切换按钮 -->
      <el-button
        v-if="isMobile"
        :icon="rightPanelVisible ? 'Hide' : 'View'"
        @click="rightPanelVisible = !rightPanelVisible"
        class="mobile-panel-toggle"
      >
        {{ rightPanelVisible ? '隐藏' : '显示' }}结果
      </el-button>
    </div>

    <!-- 可调整大小的面板 -->
    <ResizablePanels
      :initial-left-width="380"
      :min-left-width="280"
      :max-left-width="600"
      height="calc(100vh - 200px)"
      :responsive="true"
    >
      <template #left>
        <div class="left-panel">
          <!-- 主机树头部 -->
          <div class="panel-header">
            <h3 class="panel-title">目标主机</h3>
            <div class="panel-actions">
              <el-button size="small" @click="selectAll">全选</el-button>
              <el-button size="small" @click="clearSelection">清空</el-button>
              <el-button size="small" @click="invertSelection">反选</el-button>
            </div>
          </div>

          <!-- 主机树 -->
          <div class="panel-content">
            <HostTree />
          </div>

          <!-- 主机树底部 -->
          <div class="panel-footer">
            <div class="selection-count">
              已选择: <strong>{{ selectedCount }}</strong>
            </div>
          </div>
        </div>
      </template>

      <template #right>
        <div v-if="!isMobile || rightPanelVisible" class="right-panel-container">
          <ResizableVerticalPanels
            :initial-top-height="200"
            :min-top-height="150"
            :max-top-height="350"
            height="100%"
          >
            <template #top>
              <div class="operation-panel-wrapper">
                <div class="panel-header">
                  <h3 class="panel-title">操作配置</h3>
                </div>
                <div class="panel-content">
                  <OperationPanel />
                </div>
              </div>
            </template>

            <template #bottom>
              <div class="result-panel-wrapper">
                <!-- 标签页 - 存储监控始终可用 -->
                <el-tabs v-model="resultTab" class="result-tabs" key="result-tabs">
                  <!-- 存储监控 - 始终可用 -->
                  <el-tab-pane name="storage" label="存储监控">
                    <template #label>
                      <div class="tab-label">
                        <el-icon><FolderOpened /></el-icon>
                        <span>存储监控</span>
                      </div>
                    </template>
                    <StorageMonitor />
                  </el-tab-pane>

                  <!-- 有结果时显示的其他标签页 -->
                  <template v-if="hasResults">
                    <el-tab-pane name="overview" label="结果概览">
                      <template #label>
                        <div class="tab-label">
                          <el-icon><DataBoard /></el-icon>
                          <span>结果概览</span>
                          <el-badge :value="resultCount" />
                        </div>
                      </template>
                      <ResultOverview @view-detail="handleViewDetail" />
                    </el-tab-pane>

                    <el-tab-pane name="detail" label="详细结果">
                      <template #label>
                        <div class="tab-label">
                          <el-icon><Document /></el-icon>
                          <span>详细结果</span>
                        </div>
                      </template>
                      <ResultDetail @retry="handleRetry" />
                    </el-tab-pane>

                    <el-tab-pane name="analysis" label="AI 分析">
                      <template #label>
                        <div class="tab-label">
                          <el-icon><ChatDotRound /></el-icon>
                          <span>AI 分析</span>
                        </div>
                      </template>
                      <AIAnalysis />
                    </el-tab-pane>
                  </template>

                  <!-- 无结果时显示提示 -->
                  <el-tab-pane v-else name="empty" label="等待操作">
                    <template #label>
                      <div class="tab-label tab-label-disabled">
                        <el-icon><Operation /></el-icon>
                        <span>等待操作</span>
                      </div>
                    </template>
                    <div class="empty-state">
                      <div class="empty-illustration">
                        <el-icon :size="64" color="#d1d5db"><Operation /></el-icon>
                      </div>
                      <h3 class="empty-title">选择主机并执行操作</h3>
                      <p class="empty-description">
                        在左侧选择目标主机，配置操作参数，点击执行按钮查看结果
                      </p>
                      <div class="empty-tips">
                        <div class="tip-item">
                          <el-icon><Check /></el-icon>
                          <span>支持批量查看日志、执行命令</span>
                        </div>
                        <div class="tip-item">
                          <el-icon><Check /></el-icon>
                          <span>AI 智能分析执行结果</span>
                        </div>
                        <div class="tip-item">
                          <el-icon><Check /></el-icon>
                          <span>失败操作支持一键重试</span>
                        </div>
                      </div>
                    </div>
                  </el-tab-pane>
                </el-tabs>
              </div>
            </template>
          </ResizableVerticalPanels>
        </div>
      </template>
    </ResizablePanels>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { ElMessage } from 'element-plus'
import {
  DataBoard, Document, ChatDotRound, Operation,
  Check, FolderOpened
} from '@element-plus/icons-vue'
import HostTree from '@/components/console/HostTree.vue'
import OperationPanel from '@/components/console/OperationPanel.vue'
import ResultOverview from '@/components/console/ResultOverview.vue'
import ResultDetail from '@/components/console/ResultDetail.vue'
import AIAnalysis from '@/components/console/AIAnalysis.vue'
import StorageMonitor from '@/components/console/StorageMonitor.vue'
import ResizablePanels from '@/components/common/ResizablePanels.vue'
import ResizableVerticalPanels from '@/components/common/ResizableVerticalPanels.vue'
import { useConsoleStore } from '@/stores/console'
import { useHostStore } from '@/stores/host'
import type { BatchExecuteResult } from '@/api/operations'

const consoleStore = useConsoleStore()
const hostStore = useHostStore()

const resultTab = ref('storage')
const isMobile = ref(false)
const rightPanelVisible = ref(true)

// 检测移动端
const checkMobile = () => {
  isMobile.value = window.innerWidth < 1024
  if (isMobile.value) {
    rightPanelVisible.value = false
  }
}

onMounted(() => {
  checkMobile()
  window.addEventListener('resize', checkMobile)
})

onUnmounted(() => {
  window.removeEventListener('resize', checkMobile)
})

const hasResults = computed(() => consoleStore.executionResults.length > 0)
const selectedCount = computed(() => consoleStore.selectedHosts.length)
const onlineCount = computed(() =>
  hostStore.allHosts.filter(h => h.status === 'online').length
)
const offlineCount = computed(() =>
  hostStore.allHosts.filter(h => h.status === 'offline').length
)
const resultCount = computed(() => consoleStore.executionResults.length)

const selectAll = () => {
  const allHostIds = hostStore.allHosts.map(h => h.id || h.name)
  consoleStore.selectHosts(allHostIds)
}

const clearSelection = () => {
  consoleStore.selectHosts([])
}

const invertSelection = () => {
  const allIds = hostStore.allHosts.map(h => h.id || h.name)
  const selected = consoleStore.selectedHosts
  const newSelected = allIds.filter(id => !selected.includes(id))
  consoleStore.selectHosts(newSelected)
}

const handleViewDetail = (_result: BatchExecuteResult) => {
  resultTab.value = 'detail'
}

const handleRetry = async (result: BatchExecuteResult) => {
  if (!consoleStore.currentOperation) {
    ElMessage.warning('无法重试：未知操作类型')
    return
  }

  try {
    await consoleStore.executeOperation(
      consoleStore.currentOperation,
      [result.host],
      consoleStore.operationParams
    )
    ElMessage.success('重试成功')
    // 重试后切换到 AI 分析标签
    resultTab.value = 'analysis'
  } catch (error: any) {
    ElMessage.error(error.message || '重试失败')
  }
}

// 监听结果变化，自动切换标签
watch(hasResults, (has, had) => {
  // 当从无结果变为有结果时，自动切换到 AI 分析
  if (has && !had && resultTab.value === 'storage') {
    resultTab.value = 'analysis'
  } else if (has && resultTab.value === 'empty') {
    resultTab.value = 'overview'
  } else if (!has && resultTab.value !== 'storage') {
    resultTab.value = 'storage'
  }
})
</script>

<style scoped>
.console-page {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-6);
  height: 100%;
  padding: var(--spacing-5);
  background: var(--color-gray-50);
}

/* 页面头部 */
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: var(--spacing-6);
  flex-wrap: wrap;
}

.header-content {
  flex: 1;
  min-width: 200px;
}

.page-title {
  font-size: var(--text-3xl);
  font-weight: var(--font-semibold);
  color: var(--color-gray-800);
  margin: 0 0 var(--spacing-1) 0;
}

.page-description {
  font-size: var(--text-sm);
  color: var(--color-gray-500);
  margin: 0;
}

/* 快速统计 */
.quick-stats {
  display: flex;
  align-items: center;
  gap: var(--spacing-4);
  background: #fff;
  padding: var(--spacing-4);
  border-radius: var(--radius-xl);
  box-shadow: var(--shadow-sm);
  border: 1px solid var(--color-gray-200);
  flex-shrink: 0;
}

.stat-item {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
}

.stat-label {
  font-size: var(--text-sm);
  color: var(--color-gray-500);
}

.stat-value {
  font-size: var(--text-lg);
  font-weight: var(--font-semibold);
  min-width: 24px;
  text-align: center;
}

.stat-value.selected { color: var(--color-primary); }
.stat-value.online { color: #22c55e; }
.stat-value.offline { color: #ef4444; }

.stat-divider {
  width: 1px;
  height: 24px;
  background: var(--color-gray-200);
}

.mobile-panel-toggle {
  flex-shrink: 0;
}

/* 左侧面板 */
.left-panel {
  height: 100%;
  display: flex;
  flex-direction: column;
  background: #fff;
  border-radius: var(--radius-xl);
  border: 1px solid var(--color-gray-200);
  overflow: hidden;
}

.panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--spacing-4);
  border-bottom: 1px solid var(--color-gray-200);
  background: var(--color-gray-50);
  flex-wrap: wrap;
  gap: var(--spacing-2);
}

.panel-title {
  font-size: var(--text-base);
  font-weight: var(--font-semibold);
  color: var(--color-gray-800);
  margin: 0;
}

.panel-actions {
  display: flex;
  gap: var(--spacing-2);
  flex-wrap: wrap;
}

.panel-content {
  flex: 1;
  overflow: hidden;
}

.panel-footer {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: var(--spacing-4);
  border-top: 1px solid var(--color-gray-200);
  background: var(--color-gray-50);
}

.selection-count {
  font-size: var(--text-sm);
  color: var(--color-gray-600);
}

/* 右侧面板容器 */
.right-panel-container {
  height: 100%;
  display: flex;
  flex-direction: column;
}

/* 操作面板 */
.operation-panel-wrapper,
.result-panel-wrapper {
  height: 100%;
  display: flex;
  flex-direction: column;
  background: #fff;
  border-radius: var(--radius-xl);
  border: 1px solid var(--color-gray-200);
  overflow: hidden;
}

.result-tabs {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.result-tabs :deep(.el-tabs__header) {
  margin: 0;
  padding: 0 var(--spacing-4);
  background: var(--color-gray-50);
  border-bottom: 1px solid var(--color-gray-200);
}

.result-tabs :deep(.el-tabs__content) {
  flex: 1;
  overflow: hidden;
}

.result-tabs :deep(.el-tab-pane) {
  height: 100%;
  overflow: auto;
}

.tab-label {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
}

.tab-label-disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* 空状态 */
.empty-state {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: var(--spacing-6);
  text-align: center;
}

.empty-illustration {
  margin-bottom: var(--spacing-4);
  opacity: 0.4;
}

.empty-title {
  font-size: var(--text-base);
  font-weight: var(--font-semibold);
  color: var(--color-gray-600);
  margin: 0 0 var(--spacing-2) 0;
}

.empty-description {
  font-size: var(--text-sm);
  color: var(--color-gray-400);
  margin: 0 0 var(--spacing-4) 0;
  max-width: 300px;
}

.empty-tips {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-2);
  align-items: flex-start;
}

.tip-item {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  font-size: var(--text-xs);
  color: var(--color-gray-500);
}

.tip-item .el-icon {
  color: #22c55e;
  flex-shrink: 0;
  font-size: 12px;
}

/* ===== 响应式设计 - Responsive Design ===== */

/* 平板设备 (768px - 1023px) */
@media (max-width: 1023px) {
  .console-page {
    padding: var(--spacing-3);
    gap: var(--spacing-4);
  }

  .page-header {
    flex-direction: column;
    align-items: flex-start;
    gap: var(--spacing-4);
  }

  .header-content {
    width: 100%;
  }

  .page-title {
    font-size: var(--text-2xl);
  }

  .quick-stats {
    width: 100%;
    justify-content: space-around;
    padding: var(--spacing-3);
  }

  .stat-divider {
    display: none;
  }

  .panel-header {
    padding: var(--spacing-3);
  }

  .panel-actions {
    width: 100%;
  }

  .panel-actions :deep(.el-button) {
    flex: 1;
    min-width: 60px;
  }
}

/* 手机设备 (< 768px) */
@media (max-width: 767px) {
  .console-page {
    padding: var(--spacing-2);
    gap: var(--spacing-3);
    height: auto;
  }

  .page-header {
    flex-direction: column;
    align-items: stretch;
    gap: var(--spacing-3);
  }

  .header-content {
    width: 100%;
  }

  .page-title {
    font-size: var(--text-xl);
    margin-bottom: var(--spacing-1);
  }

  .page-description {
    font-size: var(--text-xs);
  }

  .quick-stats {
    width: 100%;
    padding: var(--spacing-2);
    gap: var(--spacing-2);
  }

  .stat-item {
    flex: 1;
    flex-direction: column;
    gap: var(--spacing-1);
  }

  .stat-label {
    font-size: var(--text-xs);
  }

  .stat-value {
    font-size: var(--text-base);
  }

  .stat-divider {
    display: none;
  }

  .mobile-panel-toggle {
    width: 100%;
  }

  .panel-header {
    padding: var(--spacing-2);
    gap: var(--spacing-1);
  }

  .panel-title {
    font-size: var(--text-sm);
  }

  .panel-actions {
    width: 100%;
    gap: var(--spacing-1);
  }

  .panel-actions :deep(.el-button) {
    flex: 1;
    font-size: var(--text-xs);
    padding: 4px 8px;
  }

  .panel-footer {
    padding: var(--spacing-2);
  }

  .selection-count {
    font-size: var(--text-xs);
  }

  .empty-state {
    padding: var(--spacing-4);
  }

  .empty-illustration {
    margin-bottom: var(--spacing-3);
  }

  .empty-title {
    font-size: var(--text-base);
    margin-bottom: var(--spacing-2);
  }

  .empty-description {
    font-size: var(--text-sm);
    margin-bottom: var(--spacing-4);
  }

  .tab-label {
    gap: var(--spacing-1);
  }

  .tab-label span {
    display: none;
  }

  .tab-label :deep(.el-icon) {
    margin: 0;
  }

  .result-tabs :deep(.el-tabs__header) {
    padding: 0 var(--spacing-2);
  }
}

/* 超小屏幕 (< 480px) */
@media (max-width: 479px) {
  .console-page {
    padding: var(--spacing-1);
    gap: var(--spacing-2);
  }

  .page-title {
    font-size: var(--text-lg);
  }

  .quick-stats {
    padding: var(--spacing-1);
    gap: var(--spacing-1);
  }

  .stat-item {
    gap: var(--spacing-1);
  }

  .stat-label {
    display: none;
  }

  .panel-header {
    padding: var(--spacing-1);
  }

  .panel-actions :deep(.el-button) {
    padding: 2px 6px;
    font-size: 11px;
  }
}
</style>
