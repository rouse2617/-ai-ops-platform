<template>
  <div v-if="visible" class="code-correlation-card">
    <div class="card-header">
      <div class="header-icon">
        <svg width="24" height="24" viewBox="0 0 24 24" fill="none">
          <path d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" stroke="currentColor" stroke-width="2"/>
        </svg>
      </div>
      <div class="header-content">
        <h3>代码关联分析</h3>
        <span class="correlation-score" :class="scoreClass">
          关联度: {{ (correlation.correlation_score * 100).toFixed(0) }}%
        </span>
      </div>
      <button class="close-btn" @click="close">×</button>
    </div>

    <div class="card-body">
      <div class="alert-info">
        <strong>{{ correlation.alert_name }}</strong>
        <span class="service-tag">{{ correlation.service }}</span>
      </div>

      <div class="message">
        {{ correlation.message }}
      </div>

      <div v-if="correlation.suspicious_files.length > 0" class="suspicious-files">
        <h4>可疑文件变更</h4>
        <div
          v-for="file in correlation.suspicious_files"
          :key="file.file_path"
          class="file-item"
          @click="viewDiff(file)"
        >
          <div class="file-header">
            <span class="file-path">{{ file.file_path }}</span>
            <span class="file-score">{{ (file.score * 100).toFixed(0) }}%</span>
          </div>
          <div class="file-meta">
            <span class="commit-hash">{{ file.commit_hash.substring(0, 7) }}</span>
            <span class="reason">{{ file.reason }}</span>
          </div>
        </div>
      </div>

      <div v-if="correlation.commits.length > 0" class="commits-section">
        <h4>相关提交 ({{ correlation.commits.length }})</h4>
        <div
          v-for="commit in correlation.commits.slice(0, 3)"
          :key="commit.hash"
          class="commit-item"
        >
          <div class="commit-header">
            <span class="commit-author">{{ commit.author }}</span>
            <span class="commit-time">{{ formatTime(commit.timestamp) }}</span>
          </div>
          <div class="commit-message">{{ commit.message }}</div>
          <div class="commit-files">
            {{ commit.changed_files.length }} 个文件变更
          </div>
        </div>
      </div>

      <div class="actions">
        <button class="btn-primary" @click="viewAllDiffs">查看所有改动</button>
        <button class="btn-secondary" @click="ignoreCorrelation">忽略</button>
      </div>
    </div>

    <DiffViewer
      v-if="showDiffViewer"
      :file="selectedFile"
      :repository="correlation.service"
      @close="showDiffViewer = false"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import DiffViewer from './DiffViewer.vue'

interface SuspiciousFile {
  file_path: string
  commit_hash: string
  change_type: string
  score: number
  reason: string
  diff_url: string
}

interface Commit {
  hash: string
  author: string
  email: string
  message: string
  timestamp: string
  changed_files: string[]
  repository: string
  branch: string
}

interface CodeCorrelation {
  alert_id: string
  alert_name: string
  service: string
  correlation_score: number
  message: string
  commits: Commit[]
  suspicious_files: SuspiciousFile[]
  time_window: {
    alert_time: string
    search_start: string
    search_end: string
  }
  timestamp: string
}

const props = defineProps<{
  correlation: CodeCorrelation
}>()

const emit = defineEmits<{
  close: []
  ignore: [alertId: string]
}>()

const visible = ref(true)
const showDiffViewer = ref(false)
const selectedFile = ref<SuspiciousFile | null>(null)

const scoreClass = computed(() => {
  const score = props.correlation.correlation_score
  if (score >= 0.7) return 'score-high'
  if (score >= 0.5) return 'score-medium'
  return 'score-low'
})

const formatTime = (timestamp: string) => {
  const date = new Date(timestamp)
  const now = new Date()
  const diff = now.getTime() - date.getTime()
  const minutes = Math.floor(diff / 60000)

  if (minutes < 1) return '刚刚'
  if (minutes < 60) return `${minutes}分钟前`
  const hours = Math.floor(minutes / 60)
  if (hours < 24) return `${hours}小时前`
  return date.toLocaleString('zh-CN')
}

const viewDiff = (file: SuspiciousFile) => {
  selectedFile.value = file
  showDiffViewer.value = true
}

const viewAllDiffs = () => {
  if (props.correlation.suspicious_files.length > 0) {
    viewDiff(props.correlation.suspicious_files[0])
  }
}

const ignoreCorrelation = () => {
  emit('ignore', props.correlation.alert_id)
  close()
}

const close = () => {
  visible.value = false
  emit('close')
}
</script>

<style scoped>
.code-correlation-card {
  position: fixed;
  top: 80px;
  right: 20px;
  width: 480px;
  background: white;
  border-radius: 12px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.12);
  z-index: 1000;
  animation: slideIn 0.3s ease-out;
}

@keyframes slideIn {
  from {
    transform: translateX(100%);
    opacity: 0;
  }
  to {
    transform: translateX(0);
    opacity: 1;
  }
}

.card-header {
  display: flex;
  align-items: center;
  padding: 16px;
  border-bottom: 1px solid #e5e7eb;
}

.header-icon {
  width: 40px;
  height: 40px;
  background: #fef3c7;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #f59e0b;
  margin-right: 12px;
}

.header-content {
  flex: 1;
}

.header-content h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: #111827;
}

.correlation-score {
  font-size: 12px;
  padding: 2px 8px;
  border-radius: 4px;
  font-weight: 500;
}

.score-high {
  background: #fee2e2;
  color: #dc2626;
}

.score-medium {
  background: #fef3c7;
  color: #f59e0b;
}

.score-low {
  background: #dbeafe;
  color: #3b82f6;
}

.close-btn {
  width: 32px;
  height: 32px;
  border: none;
  background: transparent;
  font-size: 24px;
  color: #6b7280;
  cursor: pointer;
  border-radius: 4px;
}

.close-btn:hover {
  background: #f3f4f6;
}

.card-body {
  padding: 16px;
  max-height: 600px;
  overflow-y: auto;
}

.alert-info {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
}

.service-tag {
  padding: 2px 8px;
  background: #e0e7ff;
  color: #4f46e5;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 500;
}

.message {
  padding: 12px;
  background: #fef3c7;
  border-left: 3px solid #f59e0b;
  border-radius: 4px;
  margin-bottom: 16px;
  line-height: 1.6;
  color: #78350f;
}

.suspicious-files h4,
.commits-section h4 {
  font-size: 14px;
  font-weight: 600;
  color: #374151;
  margin: 16px 0 8px 0;
}

.file-item {
  padding: 12px;
  background: #f9fafb;
  border: 1px solid #e5e7eb;
  border-radius: 6px;
  margin-bottom: 8px;
  cursor: pointer;
  transition: all 0.2s;
}

.file-item:hover {
  background: #f3f4f6;
  border-color: #d1d5db;
}

.file-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 4px;
}

.file-path {
  font-family: monospace;
  font-size: 13px;
  color: #111827;
  font-weight: 500;
}

.file-score {
  font-size: 12px;
  color: #f59e0b;
  font-weight: 600;
}

.file-meta {
  display: flex;
  gap: 12px;
  font-size: 12px;
  color: #6b7280;
}

.commit-hash {
  font-family: monospace;
  background: #e5e7eb;
  padding: 2px 6px;
  border-radius: 3px;
}

.commit-item {
  padding: 10px;
  background: #f9fafb;
  border-radius: 6px;
  margin-bottom: 8px;
}

.commit-header {
  display: flex;
  justify-content: space-between;
  margin-bottom: 4px;
}

.commit-author {
  font-weight: 500;
  font-size: 13px;
  color: #111827;
}

.commit-time {
  font-size: 12px;
  color: #6b7280;
}

.commit-message {
  font-size: 13px;
  color: #374151;
  margin-bottom: 4px;
}

.commit-files {
  font-size: 12px;
  color: #6b7280;
}

.actions {
  display: flex;
  gap: 8px;
  margin-top: 16px;
  padding-top: 16px;
  border-top: 1px solid #e5e7eb;
}

.btn-primary,
.btn-secondary {
  flex: 1;
  padding: 10px;
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
