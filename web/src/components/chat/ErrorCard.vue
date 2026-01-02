<template>
  <div class="error-card">
    <div class="error-header">
      <el-icon class="error-icon"><CircleClose /></el-icon>
      <span class="error-title">执行异常</span>
    </div>

    <div class="error-message">{{ errorMessage }}</div>

    <div v-if="stackTrace" class="stack-section">
      <div class="stack-toggle" @click="stackExpanded = !stackExpanded">
        <el-icon><ArrowRight :class="{ expanded: stackExpanded }" /></el-icon>
        <span>堆栈信息</span>
      </div>
      <div v-show="stackExpanded" class="stack-content">
        <pre><code class="language-javascript" v-html="highlightedStack"></code></pre>
      </div>
    </div>

    <div class="error-actions">
      <el-button size="small" type="primary" @click="diagnoseSource">
        <el-icon><Search /></el-icon>
        诊断源码
      </el-button>
      <el-button size="small" @click="viewLogs">
        <el-icon><Document /></el-icon>
        查看日志
      </el-button>
      <el-button size="small" @click="copyError">
        <el-icon><CopyDocument /></el-icon>
        复制错误
      </el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { CircleClose, ArrowRight, Search, Document, CopyDocument } from '@element-plus/icons-vue'
import hljs from 'highlight.js/lib/core'
import javascript from 'highlight.js/lib/languages/javascript'
import 'highlight.js/styles/github-dark.css'

hljs.registerLanguage('javascript', javascript)

interface Props {
  errorMessage: string
  stackTrace?: string
}

const props = defineProps<Props>()
const emit = defineEmits<{
  diagnose: []
  viewLogs: []
}>()

const stackExpanded = ref(false)

const highlightedStack = computed(() => {
  if (!props.stackTrace) return ''
  return hljs.highlight(props.stackTrace, { language: 'javascript' }).value
})

const diagnoseSource = () => {
  emit('diagnose')
}

const viewLogs = () => {
  emit('viewLogs')
}

const copyError = async () => {
  const text = `Error: ${props.errorMessage}\n\n${props.stackTrace || ''}`
  try {
    await navigator.clipboard.writeText(text)
    ElMessage.success('错误信息已复制')
  } catch {
    ElMessage.error('复制失败')
  }
}
</script>

<style scoped lang="scss">
.error-card {
  border: 2px solid #f56c6c;
  border-radius: 8px;
  padding: 16px;
  background: #fef0f0;
  margin: 12px 0;
}

.error-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;

  .error-icon {
    color: #f56c6c;
    font-size: 20px;
  }

  .error-title {
    font-weight: 600;
    color: #f56c6c;
    font-size: 16px;
  }
}

.error-message {
  color: #c45656;
  font-family: 'Consolas', 'Monaco', monospace;
  font-size: 14px;
  padding: 12px;
  background: #fff;
  border-radius: 4px;
  margin-bottom: 12px;
  word-break: break-word;
}

.stack-section {
  margin-bottom: 12px;

  .stack-toggle {
    display: flex;
    align-items: center;
    gap: 6px;
    cursor: pointer;
    padding: 8px;
    background: #fff;
    border-radius: 4px;
    user-select: none;

    &:hover {
      background: #f5f5f5;
    }

    .el-icon {
      transition: transform 0.3s;

      &.expanded {
        transform: rotate(90deg);
      }
    }

    span {
      font-size: 14px;
      color: #606266;
    }
  }

  .stack-content {
    margin-top: 8px;
    background: #1e1e1e;
    border-radius: 4px;
    overflow: auto;
    max-height: 300px;

    pre {
      margin: 0;
      padding: 12px;

      code {
        font-family: 'Consolas', 'Monaco', monospace;
        font-size: 13px;
        line-height: 1.5;
      }
    }
  }
}

.error-actions {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}
</style>
