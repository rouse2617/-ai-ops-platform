<template>
  <div class="feedback-buttons" v-if="show">
    <el-tooltip content="回答有帮助" placement="top">
      <el-button
        :type="rating === 1 ? 'success' : 'default'"
        :icon="Select"
        circle
        size="small"
        @click="handleFeedback(1)"
      />
    </el-tooltip>
    <el-tooltip content="回答需改进" placement="top">
      <el-button
        :type="rating === -1 ? 'danger' : 'default'"
        :icon="CloseBold"
        circle
        size="small"
        @click="handleFeedback(-1)"
      />
    </el-tooltip>
    <span v-if="rating !== 0" class="feedback-text">
      {{ rating === 1 ? '感谢反馈！' : '我们会改进' }}
    </span>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { Select, CloseBold } from '@element-plus/icons-vue'
import { submitFeedback } from '@/api/feedback'
import { ElMessage } from 'element-plus'

const props = defineProps<{
  messageId: string
  sessionId?: string
  question?: string
  answer?: string
}>()

const rating = ref(0)

const show = computed(() => props.messageId && props.answer)

const handleFeedback = async (value: number) => {
  if (rating.value === value) return

  try {
    await submitFeedback({
      message_id: props.messageId,
      session_id: props.sessionId || '',
      question: props.question || '',
      answer: props.answer || '',
      rating: value
    })
    rating.value = value
  } catch (error) {
    ElMessage.error('反馈提交失败')
  }
}
</script>

<style scoped>
.feedback-buttons {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 8px;
  opacity: 0.6;
  transition: opacity 0.2s;
}

.feedback-buttons:hover {
  opacity: 1;
}

.feedback-text {
  font-size: 12px;
  color: var(--color-text-secondary, #909399);
  margin-left: 4px;
}
</style>
