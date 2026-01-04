<template>
  <div class="mcp-settings">
    <div class="header">
      <h2>MCP 工具配置</h2>
      <el-button type="primary" @click="showAddDialog = true">
        <el-icon><Plus /></el-icon>
        添加服务器
      </el-button>
    </div>

    <!-- 统计信息 -->
    <el-row :gutter="20" class="stats-row">
      <el-col :span="8">
        <el-card shadow="hover">
          <el-statistic title="MCP 服务器" :value="stats.clients" />
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card shadow="hover">
          <el-statistic title="已注册工具" :value="stats.adapters" />
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card shadow="hover">
          <el-statistic title="健康服务器" :value="healthyCount" />
        </el-card>
      </el-col>
    </el-row>

    <!-- 服务器列表 -->
    <el-card class="server-list" v-loading="loading">
      <template #header>
        <div class="card-header">
          <span>MCP 服务器列表</span>
          <el-button text @click="loadData">
            <el-icon><Refresh /></el-icon>
            刷新
          </el-button>
        </div>
      </template>

      <el-collapse v-model="activeServers" v-if="servers.length > 0">
        <el-collapse-item v-for="server in servers" :key="server.name" :name="server.name">
          <template #title>
            <div class="server-header">
              <span class="server-name">{{ server.name }}</span>
              <el-tag :type="server.status === 'healthy' ? 'success' : 'danger'" size="small">
                {{ server.status === 'healthy' ? '健康' : '异常' }}
              </el-tag>
              <el-tag type="info" size="small">
                {{ stats.tools_by_server?.[server.name] || 0 }} 个工具
              </el-tag>
            </div>
          </template>

          <!-- 工具列表 -->
          <div class="tools-section">
            <div class="tools-header">
              <span>工具列表</span>
              <el-button type="danger" size="small" text @click.stop="handleRemove(server.name)">
                <el-icon><Delete /></el-icon>
                删除服务器
              </el-button>
            </div>

            <div v-if="toolsByServer[server.name]?.length > 0" class="tools-list">
              <el-card v-for="tool in toolsByServer[server.name]" :key="tool.name" class="tool-card" shadow="hover">
                <div class="tool-header">
                  <span class="tool-name">{{ tool.name }}</span>
                  <el-tag size="small">MCP</el-tag>
                </div>
                <div class="tool-desc">{{ tool.description }}</div>
                <div class="tool-params" v-if="tool.inputSchema?.properties">
                  <div class="params-title">参数:</div>
                  <div v-for="(prop, key) in tool.inputSchema.properties" :key="key" class="param-item">
                    <code>{{ key }}</code>
                    <span class="param-type">({{ prop.type }})</span>
                    <span v-if="tool.inputSchema.required?.includes(key)" class="required">*必填</span>
                    <span class="param-desc">{{ prop.description }}</span>
                  </div>
                </div>
              </el-card>
            </div>
            <el-empty v-else description="暂无工具" :image-size="60" />
          </div>
        </el-collapse-item>
      </el-collapse>

      <el-empty v-else description="暂无 MCP 服务器，点击右上角添加" />
    </el-card>

    <!-- 添加服务器对话框 -->
    <el-dialog v-model="showAddDialog" title="添加 MCP 服务器" width="500px">
      <el-form :model="addForm" :rules="rules" ref="formRef" label-width="100px">
        <el-form-item label="名称" prop="name">
          <el-input v-model="addForm.name" placeholder="如: my-tools" />
        </el-form-item>
        <el-form-item label="URL" prop="url">
          <el-input v-model="addForm.url" placeholder="如: http://localhost:3001" />
        </el-form-item>
        <el-form-item label="超时(秒)" prop="timeout">
          <el-input-number v-model="addForm.timeout" :min="5" :max="300" />
        </el-form-item>
      </el-form>
      <div class="dialog-tip">
        <el-alert type="info" :closable="false" show-icon>
          <template #title>
            MCP 服务器需要实现以下接口：
          </template>
          <ul>
            <li><code>GET /health</code> - 健康检查</li>
            <li><code>GET /tools</code> - 返回工具列表</li>
            <li><code>POST /tools/{name}</code> - 调用工具</li>
          </ul>
        </el-alert>
      </div>
      <template #footer>
        <el-button @click="showAddDialog = false">取消</el-button>
        <el-button type="primary" @click="handleAdd" :loading="adding">添加</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance } from 'element-plus'
import { Plus, Refresh, Delete } from '@element-plus/icons-vue'
import {
  getMCPServers, getMCPStats, addMCPServer, removeMCPServer, getMCPServerTools,
  type MCPServer, type MCPStats, type MCPTool
} from '@/api/mcp'

const loading = ref(false)
const adding = ref(false)
const showAddDialog = ref(false)
const servers = ref<MCPServer[]>([])
const stats = ref<MCPStats>({ clients: 0, adapters: 0, tools_by_server: {} })
const toolsByServer = ref<Record<string, MCPTool[]>>({})
const activeServers = ref<string[]>([])
const formRef = ref<FormInstance>()

const addForm = ref({
  name: '',
  url: '',
  timeout: 30
})

const rules = {
  name: [{ required: true, message: '请输入名称', trigger: 'blur' }],
  url: [{ required: true, message: '请输入 URL', trigger: 'blur' }]
}

const healthyCount = computed(() => servers.value.filter(s => s.status === 'healthy').length)

async function loadData() {
  loading.value = true
  try {
    const [serversRes, statsRes] = await Promise.all([
      getMCPServers(),
      getMCPStats()
    ])

    servers.value = serversRes?.servers || []
    stats.value = statsRes || { clients: 0, adapters: 0, tools_by_server: {} }

    // 为每个服务器加载工具列表
    const toolsMap: Record<string, any[]> = {}
    for (const server of servers.value) {
      try {
        const toolsRes = await getMCPServerTools(server.name)
        toolsMap[server.name] = toolsRes?.tools || []
      } catch (e) {
        console.error(`Failed to load tools for ${server.name}:`, e)
        toolsMap[server.name] = []
      }
    }
    toolsByServer.value = toolsMap

    // 默认展开第一个服务器
    if (servers.value.length > 0 && activeServers.value.length === 0) {
      activeServers.value = [servers.value[0].name]
    }
  } catch (e: any) {
    console.error('loadData error:', e)
    ElMessage.error('加载失败: ' + e.message)
  } finally {
    loading.value = false
  }
}

async function handleAdd() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  adding.value = true
  try {
    await addMCPServer(addForm.value)
    ElMessage.success('添加成功')
    showAddDialog.value = false
    addForm.value = { name: '', url: '', timeout: 30 }
    loadData()
  } catch (e: any) {
    ElMessage.error('添加失败: ' + e.message)
  } finally {
    adding.value = false
  }
}

async function handleRemove(name: string) {
  try {
    await ElMessageBox.confirm(`确定要删除 MCP 服务器 "${name}" 吗？相关工具也会被移除。`, '确认删除', { type: 'warning' })
    await removeMCPServer(name)
    ElMessage.success('删除成功')
    loadData()
  } catch (e: any) {
    if (e !== 'cancel') ElMessage.error('删除失败: ' + e.message)
  }
}

onMounted(loadData)
</script>

<style scoped>
.mcp-settings {
  padding: 20px;
  max-width: 1200px;
  margin: 0 auto;
}
.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}
.header h2 {
  margin: 0;
}
.stats-row {
  margin-bottom: 20px;
}
.server-list {
  margin-top: 20px;
}
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.server-header {
  display: flex;
  align-items: center;
  gap: 10px;
}
.server-name {
  font-weight: 600;
  font-size: 15px;
}
.tools-section {
  padding: 10px 0;
}
.tools-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 15px;
  font-weight: 500;
  color: #606266;
}
.tools-list {
  display: grid;
  gap: 12px;
}
.tool-card {
  border-left: 3px solid #409eff;
}
.tool-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}
.tool-name {
  font-weight: 600;
  color: #303133;
  font-family: monospace;
}
.tool-desc {
  color: #606266;
  font-size: 13px;
  margin-bottom: 10px;
  line-height: 1.5;
}
.tool-params {
  background: #f5f7fa;
  padding: 10px;
  border-radius: 4px;
  font-size: 12px;
}
.params-title {
  font-weight: 500;
  margin-bottom: 6px;
  color: #909399;
}
.param-item {
  margin: 4px 0;
  display: flex;
  align-items: center;
  gap: 6px;
  flex-wrap: wrap;
}
.param-item code {
  background: #e6f7ff;
  padding: 2px 6px;
  border-radius: 3px;
  color: #1890ff;
}
.param-type {
  color: #909399;
  font-size: 11px;
}
.required {
  color: #f56c6c;
  font-size: 11px;
}
.param-desc {
  color: #606266;
}
.dialog-tip {
  margin-top: 15px;
}
.dialog-tip ul {
  margin: 8px 0 0 0;
  padding-left: 20px;
}
.dialog-tip li {
  margin: 4px 0;
}
.dialog-tip code {
  background: #f0f0f0;
  padding: 2px 4px;
  border-radius: 3px;
}
</style>
