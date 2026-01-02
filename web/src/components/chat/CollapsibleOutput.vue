<template>
  <div class="collapsible-output">
    <div class="output-header">
      <div class="output-info">
        <el-tag :type="exitCode === 0 ? 'success' : 'danger'" size="small">
          {{ exitCode === 0 ? '成功' : `失败 (${exitCode})` }}
        </el-tag>
        <span class="meta">耗时: {{ duration }}ms</span>
        <span class="meta">大小: {{ formatSize(outputSize) }}</span>
      </div>
      <div class="output-actions">
        <el-button size="small" @click="copyOutput">
          <el-icon><CopyDocument /></el-icon>
          复制
        </el-button>
        <el-button size="small" @click="saveAsFile">
          <el-icon><Download /></el-icon>
          存为文件
        </el-button>
      </div>
    </div>

    <div class="output-content">
      <pre><code>{{ displayedOutput }}</code></pre>
    </div>

    <div v-if="totalLines > previewLines" class="output-footer">
      <el-button text @click="expanded = !expanded">
        <el-icon><ArrowDown :class="{ expanded }" /></el-icon>
        {{ expanded ? '收起' : `展开全部 (共 ${totalLines} 行)` }}
      </el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { CopyDocument, Download, ArrowDown } from '@element-plus/icons-vue'

interface Props {
  output: string
  exitCode: number
  duration: number
  previewLines?: number
}

const props = withDefaults(defineProps<Props>(), {
  previewLines: 5
})

const expanded = ref(false)

const lines = computed(() => props.output.split('\n'))
const totalLines = computed(() => lines.value.length)
const outputSize = computed(() => new Blob([props.output]).size)

const displayedOutput = computed(() => {
  if (expanded.value || totalLines.value <= props.previewLines) {
    return props.output
  }
  return lines.value.slice(0, props.previewLines).join('\n') + '\n...'
})

const formatSize = (bytes: number): string => {
  if (bytes < 1024) return `${bytes}B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)}KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)}MB`
}

const copyOutput = async () => {
  try {
    await navigator.clipboard.writeText(props.output)
    ElMessage.success('输出已复制')
  } catch {
    ElMessage.error('复制失败')
  }
}

const saveAsFile = () => {
  const blob = new Blob([props.output], { type: 'text/plain' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `output-${Date.now()}.txt`
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
  URL.revokeObjectURL(url)
  ElMessage.success('文件已保存')
}
</script>

<style scoped lang="scss">
.collapsible-output {
  border: 1px solid #dcdfe6;
  border-radius: 8px;
  margin: 12px 0;
  background: #fff;
}

.output-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  border-bottom: 1px solid #ebeef5;
  background: #f5f7fa;
  border-radius: 8px 8px 0 0;

  .output-info {
    display: flex;
    align-items: center;
    gap: 12px;

    .meta {
      font-size: 13px;
      color: #909399;
    }
  }

  .output-actions {
    display: flex;
    gap: 8px;
  }
}

.output-content {
  padding: 16px;
  background: #fafafa;
  overflow: auto;
  max-height: 400px;

  pre {
    margin: 0;
    font-family: 'Consolas', 'Monaco', monospace;
    font-size: 13px;
    line-height: 1.6;
    color: #303133;
    white-space: pre-wrap;
    word-break: break-all;

    code {
      font-family: inherit;
    }
  }
}

.output-footer {
  padding: 8px 16px;
  text-align: center;
  border-top: 1px solid #ebeef5;

  .el-icon {
    transition: transform 0.3s;

    &.expanded {
      transform: rotate(180deg);
    }
  }
}
</style>
