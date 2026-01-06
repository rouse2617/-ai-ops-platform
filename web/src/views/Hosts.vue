<template>
  <div class="hosts-bento-page">
    <!-- 页面头部 -->
    <header class="page-header bento-card bento-card-interactive">
      <div class="header-content">
        <div class="header-icon">
          <el-icon><Monitor /></el-icon>
        </div>
        <div class="header-text">
          <h1>主机管理</h1>
          <p>管理您的服务器主机，支持批量操作</p>
        </div>
      </div>
      <div class="header-actions">
        <button class="bento-btn" @click="handleAdd">
          <el-icon><Plus /></el-icon>
          <span>添加主机</span>
        </button>
      </div>
    </header>

    <!-- 统计卡片 Bento Grid -->
    <div class="stats-bento-grid">
      <div class="bento-card stat-card stat-online" @click="statusFilter = 'online'">
        <div class="stat-icon-wrapper">
          <el-icon><CircleCheck /></el-icon>
        </div>
        <div class="stat-info">
          <span class="stat-value">{{ onlineCount }}</span>
          <span class="stat-label">在线主机</span>
        </div>
        <div class="stat-trend positive" v-if="onlineCount > 0">
          <el-icon><Top /></el-icon>
          <span>运行中</span>
        </div>
      </div>

      <div class="bento-card stat-card stat-offline" @click="statusFilter = 'offline'">
        <div class="stat-icon-wrapper">
          <el-icon><CircleClose /></el-icon>
        </div>
        <div class="stat-info">
          <span class="stat-value">{{ offlineCount }}</span>
          <span class="stat-label">离线主机</span>
        </div>
        <div class="stat-trend negative" v-if="offlineCount > 0">
          <el-icon><Bottom /></el-icon>
          <span>需关注</span>
        </div>
      </div>

      <div class="bento-card stat-card stat-warning" @click="statusFilter = 'unknown'">
        <div class="stat-icon-wrapper">
          <el-icon><Warning /></el-icon>
        </div>
        <div class="stat-info">
          <span class="stat-value">{{ unknownCount }}</span>
          <span class="stat-label">状态未知</span>
        </div>
      </div>

      <div class="bento-card stat-card stat-total" @click="statusFilter = ''">
        <div class="stat-icon-wrapper">
          <el-icon><DataBoard /></el-icon>
        </div>
        <div class="stat-info">
          <span class="stat-value">{{ hostStore.total }}</span>
          <span class="stat-label">总主机数</span>
        </div>
        <div class="stat-progress">
          <div class="progress-bar">
            <div class="progress-fill online" :style="{ width: onlinePercent + '%' }"></div>
            <div class="progress-fill offline" :style="{ width: offlinePercent + '%' }"></div>
          </div>
        </div>
      </div>
    </div>

    <!-- 工具栏 -->
    <div class="toolbar-card bento-card bento-card-interactive">
      <div class="toolbar-left">
        <div class="search-wrapper">
          <el-icon class="search-icon"><Search /></el-icon>
          <input
            v-model="keyword"
            type="text"
            class="bento-input search-input"
            placeholder="搜索主机名称、地址或标签..."
            @input="handleSearch"
          />
        </div>

        <div class="view-toggle">
          <button
            class="toggle-btn"
            :class="{ active: viewMode === 'table' }"
            @click="viewMode = 'table'"
            title="列表视图"
          >
            <el-icon><List /></el-icon>
          </button>
          <button
            class="toggle-btn"
            :class="{ active: viewMode === 'card' }"
            @click="viewMode = 'card'"
            title="卡片视图"
          >
            <el-icon><Grid /></el-icon>
          </button>
        </div>

        <div class="filter-chips">
          <button
            class="filter-chip"
            :class="{ active: statusFilter === '' }"
            @click="statusFilter = ''"
          >
            全部
          </button>
          <button
            class="filter-chip online"
            :class="{ active: statusFilter === 'online' }"
            @click="statusFilter = 'online'"
          >
            <span class="chip-dot"></span>
            在线
          </button>
          <button
            class="filter-chip offline"
            :class="{ active: statusFilter === 'offline' }"
            @click="statusFilter = 'offline'"
          >
            <span class="chip-dot"></span>
            离线
          </button>
        </div>
      </div>

      <div class="toolbar-right">
        <span v-if="selectedHosts.length > 0" class="selection-info">
          已选择 {{ selectedHosts.length }} 台
        </span>
        <button
          v-if="selectedHosts.length > 0"
          class="bento-btn bento-btn-secondary bento-btn-sm"
          @click="showImport = true"
        >
          <el-icon><Upload /></el-icon>
          <span>导入</span>
        </button>
        <button
          v-if="selectedHosts.length > 0"
          class="bento-btn bento-btn-sm"
          style="background: var(--bento-accent-pink)"
          @click="handleBatchDelete"
        >
          <el-icon><Delete /></el-icon>
          <span>删除</span>
        </button>
      </div>
    </div>

    <!-- 内容区域 -->
    <div class="content-area bento-card bento-card-interactive">
      <HostTable
        v-if="viewMode === 'table'"
        :hosts="filteredHosts"
        :total="filteredTotal"
        :loading="hostStore.loading"
        @edit="handleEdit"
        @delete="handleDelete"
        @test="handleTest"
        @selection-change="handleSelectionChange"
        @page-change="handlePageChange"
      />

      <HostGrid
        v-else
        :hosts="filteredHosts"
        :total="filteredTotal"
        :loading="hostStore.loading"
        @edit="handleEdit"
        @delete="handleDelete"
        @test="handleTest"
        @selection-change="handleSelectionChange"
        @page-change="handlePageChange"
      />
    </div>

    <!-- 弹窗 -->
    <HostForm
      v-model:visible="showForm"
      :host="currentHost"
      @submit="handleFormSubmit"
    />

    <HostImport
      v-model:visible="showImport"
      @submit="handleImportSubmit"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Search, Plus, Upload, Delete,
  CircleCheck, CircleClose, Warning, Monitor,
  List, Grid, Top, Bottom, DataBoard
} from '@element-plus/icons-vue'
import { useHostStore } from '@/stores/host'
import type { Host } from '@/api/host'
import HostTable from '@/components/host/HostTable.vue'
import HostGrid from '@/components/host/HostGrid.vue'
import HostForm from '@/components/host/HostForm.vue'
import HostImport from '@/components/host/HostImport.vue'

const hostStore = useHostStore()

const keyword = ref('')
const statusFilter = ref('')
const viewMode = ref<'table' | 'card'>('card')
const showForm = ref(false)
const showImport = ref(false)
const currentHost = ref<Host | null>(null)
const selectedHosts = ref<Host[]>([])

const onlineCount = computed(() =>
  hostStore.hosts.filter(h => h.status === 'online').length
)
const offlineCount = computed(() =>
  hostStore.hosts.filter(h => h.status === 'offline').length
)
const unknownCount = computed(() =>
  hostStore.hosts.filter(h => !h.status || h.status === 'unknown').length
)

const onlinePercent = computed(() =>
  hostStore.total > 0 ? (onlineCount.value / hostStore.total) * 100 : 0
)
const offlinePercent = computed(() =>
  hostStore.total > 0 ? (offlineCount.value / hostStore.total) * 100 : 0
)

const filteredHosts = computed(() => {
  let hosts = hostStore.hosts

  if (statusFilter.value) {
    hosts = hosts.filter(h => h.status === statusFilter.value)
  }

  if (keyword.value) {
    const kw = keyword.value.toLowerCase()
    hosts = hosts.filter(h =>
      h.name.toLowerCase().includes(kw) ||
      h.host.toLowerCase().includes(kw) ||
      (h.tags || []).some(tag => tag.toLowerCase().includes(kw))
    )
  }

  return hosts
})

const filteredTotal = computed(() => filteredHosts.value.length)

onMounted(() => {
  loadData()
})

const loadData = (page = 1, pageSize = 10) => {
  hostStore.loadHosts({ page, pageSize, keyword: keyword.value })
}

const handleSearch = () => {
  loadData()
}

const handleAdd = () => {
  currentHost.value = null
  showForm.value = true
}

const handleEdit = (host: Host) => {
  currentHost.value = host
  showForm.value = true
}

const handleDelete = async (host: Host) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除主机 "${host.name}" 吗？`,
      '提示',
      { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
    )
    await hostStore.removeHost(host.id)
    ElMessage.success('删除成功')
  } catch {}
}

const handleTest = async (host: Host) => {
  try {
    ElMessage.info('正在测试连接...')
    const result = await hostStore.testConnection(host.id)
    if (result.success) {
      ElMessage.success('连接成功')
    } else {
      ElMessage.error(result.message || '连接失败')
    }
  } catch {
    ElMessage.error('测试连接失败')
  }
}

const handleSelectionChange = (hosts: Host[]) => {
  selectedHosts.value = hosts
}

const handleBatchDelete = async () => {
  if (selectedHosts.value.length === 0) return

  try {
    await ElMessageBox.confirm(
      `确定要删除选中的 ${selectedHosts.value.length} 台主机吗？`,
      '提示',
      { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
    )
    for (const host of selectedHosts.value) {
      await hostStore.removeHost(host.id)
    }
    ElMessage.success('批量删除成功')
    selectedHosts.value = []
  } catch {}
}

const handlePageChange = (page: number, pageSize: number) => {
  loadData(page, pageSize)
}

const handleFormSubmit = async (data: Omit<Host, 'id' | 'createdAt' | 'updatedAt'>) => {
  try {
    if (currentHost.value) {
      await hostStore.editHost(currentHost.value.id, data)
      ElMessage.success('更新成功')
    } else {
      await hostStore.addHost(data)
      ElMessage.success('添加成功')
    }
  } catch {}
}

const handleImportSubmit = async (hosts: Omit<Host, 'id' | 'createdAt' | 'updatedAt'>[]) => {
  try {
    const result = await hostStore.batchImport(hosts)
    ElMessage.success(`成功导入 ${result.success} 台主机${result.failed > 0 ? `，失败 ${result.failed} 台` : ''}`)
  } catch {}
}
</script>

<style scoped>
.hosts-bento-page {
  display: flex;
  flex-direction: column;
  gap: var(--bento-gap);
  padding: var(--bento-gap);
  background: var(--bento-bg);
  min-height: 100%;
}

/* 页面头部 */
.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 24px;
}

.header-content {
  display: flex;
  align-items: center;
  gap: 16px;
}

.header-icon {
  width: 56px;
  height: 56px;
  border-radius: 16px;
  background: var(--bento-gradient-purple);
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-size: 28px;
}

.header-text h1 {
  font-size: 24px;
  font-weight: 700;
  color: var(--color-text-primary);
  margin: 0 0 4px 0;
}

.header-text p {
  font-size: 14px;
  color: var(--color-text-secondary);
  margin: 0;
}

/* 统计卡片网格 */
.stats-bento-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: var(--bento-gap);
}

.stat-card {
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 16px;
  cursor: pointer;
}

.stat-icon-wrapper {
  width: 48px;
  height: 48px;
  border-radius: 14px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24px;
}

.stat-online .stat-icon-wrapper {
  background: rgba(48, 209, 88, 0.15);
  color: var(--bento-accent-green);
}

.stat-offline .stat-icon-wrapper {
  background: rgba(255, 55, 95, 0.15);
  color: var(--bento-accent-pink);
}

.stat-warning .stat-icon-wrapper {
  background: rgba(255, 149, 0, 0.15);
  color: var(--bento-accent-orange);
}

.stat-total .stat-icon-wrapper {
  background: rgba(0, 113, 227, 0.15);
  color: var(--bento-accent-blue);
}

.stat-info {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.stat-value {
  font-size: 36px;
  font-weight: 700;
  color: var(--color-text-primary);
  line-height: 1;
}

.stat-label {
  font-size: 13px;
  color: var(--color-text-secondary);
}

.stat-trend {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  font-weight: 500;
}

.stat-trend.positive {
  color: var(--bento-accent-green);
}

.stat-trend.negative {
  color: var(--bento-accent-pink);
}

.stat-progress {
  margin-top: auto;
}

.progress-bar {
  height: 6px;
  background: var(--color-bg-secondary);
  border-radius: 3px;
  overflow: hidden;
  display: flex;
}

.progress-fill {
  height: 100%;
  transition: width 0.3s ease;
}

.progress-fill.online {
  background: var(--bento-accent-green);
}

.progress-fill.offline {
  background: var(--bento-accent-pink);
}

/* 工具栏 */
.toolbar-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 20px;
  gap: 16px;
  flex-wrap: wrap;
}

.toolbar-left {
  display: flex;
  align-items: center;
  gap: 16px;
  flex: 1;
}

.search-wrapper {
  position: relative;
  width: 280px;
}

.search-icon {
  position: absolute;
  left: 14px;
  top: 50%;
  transform: translateY(-50%);
  color: var(--color-text-tertiary);
  font-size: 16px;
}

.search-input {
  padding-left: 42px;
  width: 100%;
}

.view-toggle {
  display: flex;
  background: var(--color-bg-secondary);
  border-radius: 10px;
  padding: 4px;
}

.toggle-btn {
  width: 36px;
  height: 36px;
  border: none;
  background: transparent;
  border-radius: 8px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--color-text-secondary);
  transition: all 0.2s;
}

.toggle-btn:hover {
  color: var(--color-text-primary);
}

.toggle-btn.active {
  background: var(--bento-card-bg);
  color: var(--bento-accent-blue);
  box-shadow: var(--bento-shadow);
}

.filter-chips {
  display: flex;
  gap: 8px;
}

.filter-chip {
  padding: 8px 16px;
  border: none;
  background: var(--color-bg-secondary);
  border-radius: 20px;
  font-size: 13px;
  font-weight: 500;
  color: var(--color-text-secondary);
  cursor: pointer;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  gap: 6px;
}

.filter-chip:hover {
  background: var(--color-bg-tertiary);
}

.filter-chip.active {
  background: var(--bento-accent-blue);
  color: white;
}

.filter-chip.online.active {
  background: var(--bento-accent-green);
}

.filter-chip.offline.active {
  background: var(--bento-accent-pink);
}

.chip-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: currentColor;
}

.toolbar-right {
  display: flex;
  align-items: center;
  gap: 12px;
}

.selection-info {
  font-size: 13px;
  color: var(--color-text-secondary);
  padding: 8px 12px;
  background: var(--color-bg-secondary);
  border-radius: 8px;
}

/* 内容区域 */
.content-area {
  flex: 1;
  padding: 0;
  overflow: hidden;
}

/* 响应式 */
@media (max-width: 1200px) {
  .stats-bento-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 768px) {
  .stats-bento-grid {
    grid-template-columns: 1fr;
  }

  .toolbar-card {
    flex-direction: column;
    align-items: stretch;
  }

  .toolbar-left {
    flex-direction: column;
    align-items: stretch;
  }

  .search-wrapper {
    width: 100%;
  }

  .filter-chips {
    flex-wrap: wrap;
  }

  .page-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 16px;
  }
}

/* 暗色模式 */
[data-theme="dark"] .toggle-btn.active {
  background: var(--color-surface);
}
</style>
