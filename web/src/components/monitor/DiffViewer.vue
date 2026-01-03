<template>
  <div v-if="visible" class="diff-viewer-modal" @click.self="close">
    <div class="diff-viewer-container">
      <div class="diff-header">
        <div class="file-info">
          <h3>{{ file.file_path }}</h3>
          <div class="meta">
            <span class="commit">{{ file.commit_hash.substring(0, 7) }}</span>
            <span class="repo">{{ repository }}</span>
          </div>
        </div>
        <button class="close-btn" @click="close">×</button>
      </div>

      <div class="diff-content">
        <div v-if="loading" class="loading">加载中...</div>
        <div v-else-if="error" class="error">{{ error }}</div>
        <pre v-else class="diff-text">{{ diffContent }}</pre>
      </div>

      <div class="diff-actions">
        <button class="btn-secondary" @click="copyDiff">复制 Diff</button>
        <button class="btn-primary" @click="openInGit">在 Git 中查看</button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getDiff } from '@/api/monitor'

interface SuspiciousFile {
  file_path: string
  commit_hash: string
  change_type: string
  score: number
  reason: string
  diff_url: string
}

const props = defineProps<{
  file: SuspiciousFile
  repository: string
}>()

const emit = defineEmits<{
  close: []
}>()

const visible = ref(true)
const loading = ref(true)
const error = ref('')
const diffContent = ref('')

onMounted(async () => {
  try {
    const response = await getDiff(props.repository, props.file.commit_hash, props.file.file_path)
    diffContent.value = response.data.diff
  } catch (err: any) {
    error.value = err.message || '加载 Diff 失败'
  } finally {
    loading.value = false
  }
})

const copyDiff = () => {
  navigator.clipboard.writeText(diffContent.value)
}

const openInGit = () => {
  if (props.file.diff_url) {
    window.open(props.file.diff_url, '_blank')
  }
}

const close = () => {
  visible.value = false
  emit('close')
}
</script>

<style scoped>
.diff-viewer-modal {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 2000;
}

.diff-viewer-container {
  width: 90%;
  max-width: 1200px;
  height: 80vh;
  background: white;
  border-radius: 12px;
  display: flex;
  flex-direction: column;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
}

.diff-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px;
  border-bottom: 1px solid #e5e7eb;
}

.file-info h3 {
  margin: 0 0 8px 0;
  font-size: 18px;
  font-weight: 600;
  color: #111827;
  font-family: monospace;
}

.meta {
  display: flex;
  gap: 12px;
  font-size: 13px;
}

.commit,
.repo {
  padding: 4px 8px;
  border-radius: 4px;
  font-family: monospace;
}

.commit {
  background: #e5e7eb;
  color: #374151;
}

.repo {
  background: #e0e7ff;
  color: #4f46e5;
}

.close-btn {
  width: 36px;
  height: 36px;
  border: none;
  background: transparent;
  font-size: 28px;
  color: #6b7280;
  cursor: pointer;
  border-radius: 4px;
}

.close-btn:hover {
  background: #f3f4f6;
}

.diff-content {
  flex: 1;
  overflow: auto;
  padding: 20px;
  background: #f9fafb;
}

.loading,
.error {
  text-align: center;
  padding: 40px;
  color: #6b7280;
}

.error {
  color: #dc2626;
}

.diff-text {
  font-family: 'Consolas', 'Monaco', monospace;
  font-size: 13px;
  line-height: 1.6;
  margin: 0;
  white-space: pre-wrap;
  word-wrap: break-word;
}

.diff-actions {
  display: flex;
  gap: 12px;
  padding: 20px;
  border-top: 1px solid #e5e7eb;
  justify-content: flex-end;
}

.btn-primary,
.btn-secondary {
  padding: 10px 20px;
  border: none;
  border-radius: 6px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-primary {
  background: #4f46e5;
  color: white;
}

.btn-primary:hover {
  background: #4338ca;
}

.btn-secondary {
  background: #f3f4f6;
  color: #374151;
}

.btn-secondary:hover {
  background: #e5e7eb;
}
</style>
