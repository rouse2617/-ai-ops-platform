<template>
  <div class="collapsible-content">
    <div v-if="!isCollapsed || !shouldCollapse" class="content-wrapper">
      <pre v-if="syntax" class="code-block" :class="`language-${syntax}`"><code>{{ content }}</code></pre>
      <div v-else class="text-content">{{ content }}</div>
    </div>

    <div v-else class="content-wrapper collapsed">
      <pre v-if="syntax" class="code-block" :class="`language-${syntax}`"><code>{{ collapsedContent }}</code></pre>
      <div v-else class="text-content">{{ collapsedContent }}</div>
      <div class="ellipsis-indicator">
        <el-icon><MoreFilled /></el-icon>
        <span>省略 {{ omittedLines }} 行</span>
        <el-icon><MoreFilled /></el-icon>
      </div>
    </div>

    <div v-if="showExpandButton" class="expand-controls">
      <el-button
        type="primary"
        text
        @click="toggleCollapse"
        class="expand-button"
      >
        <el-icon><component :is="isCollapsed ? ArrowDown : ArrowUp" /></el-icon>
        {{ isCollapsed ? '展开全部' : '收起' }}
        <span v-if="isCollapsed && lineCount" class="line-count">
          (共 {{ lineCount }} 行)
        </span>
      </el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { ArrowDown, ArrowUp, MoreFilled } from '@element-plus/icons-vue'

interface Props {
  content: string
  maxLines?: number
  headLines?: number
  tailLines?: number
  expandable?: boolean
  syntax?: string
}

interface Emits {
  (e: 'expand'): void
  (e: 'collapse'): void
}

const props = withDefaults(defineProps<Props>(), {
  maxLines: 50,
  headLines: 10,
  tailLines: 10,
  expandable: true
})

const emit = defineEmits<Emits>()

const isCollapsed = ref(true)

const lines = computed(() => {
  if (!props.content) return []
  return props.content.split('\n')
})

const lineCount = computed(() => lines.value.length)

const shouldCollapse = computed(() => {
  return props.expandable && lineCount.value > props.maxLines
})

const omittedLines = computed(() => {
  if (!shouldCollapse.value) return 0
  return lineCount.value - props.headLines - props.tailLines
})

const collapsedContent = computed(() => {
  if (!shouldCollapse.value) return props.content

  const headContent = lines.value.slice(0, props.headLines).join('\n')
  const tailContent = lines.value.slice(-props.tailLines).join('\n')

  return `${headContent}\n\n${tailContent}`
})

const showExpandButton = computed(() => {
  return props.expandable && shouldCollapse.value
})

const toggleCollapse = () => {
  isCollapsed.value = !isCollapsed.value
  emit(isCollapsed.value ? 'collapse' : 'expand')
}
</script>

<style scoped lang="scss">
.collapsible-content {
  position: relative;
  background: #f7f8fa;
  border-radius: 8px;
  padding: 12px;

  .content-wrapper {
    position: relative;

    &.collapsed {
      .ellipsis-indicator {
        display: flex;
        align-items: center;
        justify-content: center;
        gap: 8px;
        padding: 16px 0;
        margin: 8px 0;
        color: #909399;
        font-size: 13px;
        border-top: 1px dashed #dcdfe6;
        border-bottom: 1px dashed #dcdfe6;
        background: linear-gradient(to bottom, transparent, rgba(0, 0, 0, 0.02), transparent);
      }
    }

    .code-block {
      margin: 0;
      padding: 12px;
      background: #282c34;
      color: #abb2bf;
      border-radius: 6px;
      font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
      font-size: 13px;
      line-height: 1.6;
      overflow-x: auto;

      code {
        display: block;
        white-space: pre;
      }
    }

    .text-content {
      color: #303133;
      font-size: 14px;
      line-height: 1.8;
      white-space: pre-wrap;
      word-break: break-word;
    }
  }

  .expand-controls {
    margin-top: 12px;
    text-align: center;
    border-top: 1px solid #e4e7ed;
    padding-top: 12px;

    .expand-button {
      display: inline-flex;
      align-items: center;
      gap: 6px;
      font-size: 14px;

      .line-count {
        color: #909399;
        font-size: 12px;
      }
    }
  }
}

@media (max-width: 768px) {
  .collapsible-content {
    .content-wrapper {
      .code-block {
        font-size: 12px;
        padding: 8px;
      }

      .text-content {
        font-size: 13px;
      }
    }
  }
}
</style>
