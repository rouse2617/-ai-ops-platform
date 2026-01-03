<template>
  <div class="promql-editor">
    <div class="editor-header">
      <span class="editor-label">PromQL 查询</span>
      <el-button-group size="small">
        <el-button @click="handleFormat">
          <el-icon><DocumentCopy /></el-icon>
          格式化
        </el-button>
        <el-button @click="handleValidate">
          <el-icon><CircleCheck /></el-icon>
          验证
        </el-button>
        <el-button type="primary" @click="handleExecute" :loading="executing">
          <el-icon><CaretRight /></el-icon>
          执行
        </el-button>
      </el-button-group>
    </div>

    <div class="editor-container">
      <textarea
        ref="editorRef"
        v-model="localQuery"
        class="query-input"
        :style="{ height: `${height}px` }"
        placeholder="输入 PromQL 查询，例如: rate(http_requests_total[5m])"
        @keydown="handleKeydown"
      />

      <div v-if="showSuggestions && filteredSuggestions.length > 0" class="suggestions-dropdown">
        <div
          v-for="(suggestion, index) in filteredSuggestions"
          :key="index"
          class="suggestion-item"
          :class="{ active: selectedSuggestionIndex === index }"
          @click="applySuggestion(suggestion)"
          @mouseenter="selectedSuggestionIndex = index"
        >
          <el-icon class="suggestion-icon"><Document /></el-icon>
          <span class="suggestion-text">{{ suggestion }}</span>
        </div>
      </div>
    </div>

    <div v-if="validationError" class="validation-error">
      <el-alert type="error" :closable="false" show-icon>
        {{ validationError }}
      </el-alert>
    </div>

    <div class="editor-footer">
      <div class="query-templates">
        <el-dropdown @command="handleTemplateSelect">
          <el-button size="small">
            快捷查询
            <el-icon class="el-icon--right"><ArrowDown /></el-icon>
          </el-button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item
                v-for="template in queryTemplates"
                :key="template.name"
                :command="template.query"
              >
                {{ template.name }}
              </el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </div>

      <div v-if="recentQueries.length > 0" class="query-history">
        <el-dropdown @command="handleHistorySelect">
          <el-button size="small">
            <el-icon><Clock /></el-icon>
            历史查询
          </el-button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item
                v-for="(query, index) in recentQueries"
                :key="index"
                :command="query"
              >
                {{ query }}
              </el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import {
  DocumentCopy,
  CircleCheck,
  CaretRight,
  Document,
  ArrowDown,
  Clock
} from '@element-plus/icons-vue'
import { usePrometheusStore } from '@/stores/prometheus'

interface Props {
  modelValue: string
  suggestions?: string[]
  validation?: boolean
  height?: number
}

interface Emits {
  (e: 'update:modelValue', value: string): void
  (e: 'execute', query: string): void
}

const props = withDefaults(defineProps<Props>(), {
  suggestions: () => [],
  validation: true,
  height: 120
})

const emit = defineEmits<Emits>()

const prometheusStore = usePrometheusStore()

const editorRef = ref<HTMLTextAreaElement>()
const localQuery = ref(props.modelValue)
const executing = ref(false)
const validationError = ref<string>()
const showSuggestions = ref(false)
const selectedSuggestionIndex = ref(0)

const queryTemplates = [
  { name: 'CPU 使用率', query: 'rate(node_cpu_seconds_total{mode="user"}[5m])' },
  { name: '内存使用率', query: '(node_memory_MemTotal_bytes - node_memory_MemAvailable_bytes) / node_memory_MemTotal_bytes * 100' },
  { name: '磁盘使用率', query: '(node_filesystem_size_bytes - node_filesystem_free_bytes) / node_filesystem_size_bytes * 100' },
  { name: '网络流量', query: 'rate(node_network_receive_bytes_total[5m])' },
  { name: 'HTTP 请求率', query: 'rate(http_requests_total[5m])' },
  { name: 'HTTP 错误率', query: 'rate(http_requests_total{status=~"5.."}[5m]) / rate(http_requests_total[5m])' }
]

const recentQueries = computed(() => prometheusStore.recentQueries)

const filteredSuggestions = computed(() => {
  if (!localQuery.value || props.suggestions.length === 0) return []

  const query = localQuery.value.toLowerCase()
  return props.suggestions.filter(s => s.toLowerCase().includes(query)).slice(0, 10)
})

const handleKeydown = (e: KeyboardEvent) => {
  if (showSuggestions.value && filteredSuggestions.value.length > 0) {
    if (e.key === 'ArrowDown') {
      e.preventDefault()
      selectedSuggestionIndex.value = Math.min(
        selectedSuggestionIndex.value + 1,
        filteredSuggestions.value.length - 1
      )
    } else if (e.key === 'ArrowUp') {
      e.preventDefault()
      selectedSuggestionIndex.value = Math.max(selectedSuggestionIndex.value - 1, 0)
    } else if (e.key === 'Enter' && e.ctrlKey) {
      e.preventDefault()
      handleExecute()
    } else if (e.key === 'Tab' || e.key === 'Enter') {
      e.preventDefault()
      applySuggestion(filteredSuggestions.value[selectedSuggestionIndex.value])
    } else if (e.key === 'Escape') {
      showSuggestions.value = false
    }
  } else if (e.key === 'Enter' && e.ctrlKey) {
    e.preventDefault()
    handleExecute()
  }
}

const applySuggestion = (suggestion: string) => {
  localQuery.value = suggestion
  showSuggestions.value = false
  selectedSuggestionIndex.value = 0
}

const handleFormat = () => {
  // 简单的格式化：添加换行和缩进
  localQuery.value = localQuery.value
    .replace(/\s+/g, ' ')
    .replace(/\(/g, '(\n  ')
    .replace(/\)/g, '\n)')
    .trim()
}

const handleValidate = () => {
  validationError.value = undefined

  if (!localQuery.value.trim()) {
    validationError.value = '查询不能为空'
    return
  }

  // 基本语法检查
  const openParens = (localQuery.value.match(/\(/g) || []).length
  const closeParens = (localQuery.value.match(/\)/g) || []).length

  if (openParens !== closeParens) {
    validationError.value = '括号不匹配'
    return
  }

  // 检查是否包含有效的指标名称
  if (!/[a-zA-Z_:][a-zA-Z0-9_:]*/.test(localQuery.value)) {
    validationError.value = '无效的指标名称'
    return
  }
}

const handleExecute = async () => {
  if (props.validation) {
    handleValidate()
    if (validationError.value) return
  }

  executing.value = true
  try {
    emit('execute', localQuery.value)
    await prometheusStore.executeQuery(localQuery.value)
  } finally {
    executing.value = false
  }
}

const handleTemplateSelect = (query: string) => {
  localQuery.value = query
  emit('update:modelValue', query)
}

const handleHistorySelect = (query: string) => {
  localQuery.value = query
  emit('update:modelValue', query)
}

watch(localQuery, (value) => {
  emit('update:modelValue', value)
  validationError.value = undefined

  // 显示建议
  if (value.length > 2) {
    showSuggestions.value = true
    selectedSuggestionIndex.value = 0
  } else {
    showSuggestions.value = false
  }
})

watch(() => props.modelValue, (value) => {
  if (value !== localQuery.value) {
    localQuery.value = value
  }
})
</script>

<style scoped lang="scss">
.promql-editor {
  background: white;
  border-radius: 8px;
  border: 1px solid #dcdfe6;
  overflow: hidden;

  .editor-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 12px 16px;
    background: #f5f7fa;
    border-bottom: 1px solid #dcdfe6;

    .editor-label {
      font-weight: 600;
      color: #303133;
    }
  }

  .editor-container {
    position: relative;

    .query-input {
      width: 100%;
      padding: 12px 16px;
      border: none;
      outline: none;
      font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
      font-size: 14px;
      line-height: 1.6;
      resize: vertical;
      background: #fafafa;

      &::placeholder {
        color: #c0c4cc;
      }
    }

    .suggestions-dropdown {
      position: absolute;
      top: 100%;
      left: 0;
      right: 0;
      max-height: 200px;
      overflow-y: auto;
      background: white;
      border: 1px solid #dcdfe6;
      border-top: none;
      box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
      z-index: 100;

      .suggestion-item {
        display: flex;
        align-items: center;
        gap: 8px;
        padding: 8px 16px;
        cursor: pointer;
        transition: background 0.2s;

        &:hover,
        &.active {
          background: #f5f7fa;
        }

        .suggestion-icon {
          color: #909399;
        }

        .suggestion-text {
          font-family: 'Consolas', 'Monaco', monospace;
          font-size: 13px;
          color: #303133;
        }
      }
    }
  }

  .validation-error {
    padding: 12px 16px;
    border-top: 1px solid #dcdfe6;
  }

  .editor-footer {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 12px 16px;
    background: #f5f7fa;
    border-top: 1px solid #dcdfe6;
  }
}

@media (max-width: 768px) {
  .promql-editor {
    .editor-header {
      flex-direction: column;
      align-items: flex-start;
      gap: 12px;
    }

    .editor-footer {
      flex-direction: column;
      align-items: stretch;
    }
  }
}
</style>
