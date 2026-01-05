<template>
  <div class="data-description-container" :class="{ 'side-by-side': layout === 'horizontal' }">
    <div
      v-for="(section, index) in sections"
      :key="index"
      class="data-section"
      :class="getSectionClass(section)"
    >
      <div v-if="section.title" class="section-title">{{ section.title }}</div>
      <dl class="description-list">
        <div v-for="item in section.items" :key="item.label" class="desc-item">
          <dt>{{ item.label }}</dt>
          <dd :class="getValueClass(item)">
            <!-- 进度条类型 -->
            <template v-if="item.type === 'progress'">
              <div class="progress-wrapper">
                <el-progress
                  :percentage="item.value"
                  :stroke-width="5"
                  :color="getProgressColor(item.value)"
                  :show-text="false"
                />
                <span class="progress-text">{{ item.value }}%</span>
              </div>
            </template>
            <!-- 状态标签类型 -->
            <template v-else-if="item.type === 'status'">
              <el-tag :type="item.statusType || 'info'" size="small" effect="plain">
                {{ item.value }}
              </el-tag>
            </template>
            <!-- 普通文本 -->
            <template v-else>
              {{ item.value }}
            </template>
          </dd>
        </div>
      </dl>
    </div>
  </div>
</template>

<script setup lang="ts">
export interface DescItem {
  label: string
  value: string | number
  type?: 'text' | 'progress' | 'status'
  statusType?: 'success' | 'warning' | 'danger' | 'info'
  highlight?: boolean
}

export interface DataSection {
  title?: string
  status?: 'normal' | 'warning' | 'error'
  items: DescItem[]
}

defineProps<{
  sections: DataSection[]
  layout?: 'vertical' | 'horizontal'
}>()

const getSectionClass = (section: DataSection) => ({
  'status-warning': section.status === 'warning',
  'status-error': section.status === 'error'
})

const getValueClass = (item: DescItem) => ({
  'highlight': item.highlight
})

const getProgressColor = (value: number) => {
  if (value >= 90) return '#f56c6c'
  if (value >= 70) return '#e6a23c'
  return '#67c23a'
}
</script>

<style scoped>
.data-description-container {
  display: flex;
  flex-direction: column;
  gap: 12px;
  margin: 8px 0;
}

.data-description-container.side-by-side {
  flex-direction: row;
  flex-wrap: wrap;
}

.data-description-container.side-by-side .data-section {
  flex: 1;
  min-width: 200px;
}

.data-section {
  background: #fafafa;
  border-radius: 6px;
  padding: 10px 14px;
  border-left: 3px solid #67c23a;
}

.data-section.status-warning {
  background: linear-gradient(90deg, rgba(230, 162, 60, 0.08) 0%, #fafafa 30%);
  border-left-color: #e6a23c;
}

.data-section.status-error {
  background: linear-gradient(90deg, rgba(245, 108, 108, 0.08) 0%, #fafafa 30%);
  border-left-color: #f56c6c;
}

.section-title {
  font-size: 13px;
  font-weight: 600;
  color: #303133;
  margin-bottom: 8px;
  padding-bottom: 6px;
  border-bottom: 1px solid #ebeef5;
}

.description-list {
  margin: 0;
  padding: 0;
}

.desc-item {
  display: flex;
  align-items: center;
  padding: 4px 0;
  font-size: 13px;
}

.desc-item dt {
  color: #909399;
  min-width: 70px;
  flex-shrink: 0;
}

.desc-item dd {
  margin: 0;
  color: #606266;
  flex: 1;
}

.desc-item dd.highlight {
  color: #f56c6c;
  font-weight: 500;
}

.progress-wrapper {
  display: flex;
  align-items: center;
  gap: 8px;
}

.progress-wrapper .el-progress {
  flex: 1;
  max-width: 100px;
}

.progress-text {
  font-size: 12px;
  min-width: 36px;
}

/* 响应式 */
@media (max-width: 500px) {
  .data-description-container.side-by-side {
    flex-direction: column;
  }
}
</style>
