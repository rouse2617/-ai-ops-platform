<template>
  <ErrorBoundary>
    <GlobalTaskProgress />

    <div class="app-bento-layout">
      <!-- 侧边栏 -->
      <aside
        class="app-sidebar-wrapper"
        :class="{ collapsed: sidebarCollapsed }"
      >
        <AppSidebar :collapsed="sidebarCollapsed" />
      </aside>

      <!-- 主内容区 -->
      <div class="app-content-wrapper">
        <!-- 顶部栏 -->
        <header class="app-topbar">
          <button class="topbar-toggle" @click="toggleSidebar">
            <el-icon><component :is="sidebarCollapsed ? Expand : Fold" /></el-icon>
          </button>
          <AppHeader />
        </header>

        <!-- 进度条 -->
        <div class="topbar-progress" v-if="isLoading">
          <div class="progress-line" :style="{ width: progress + '%' }"></div>
        </div>

        <!-- 页面内容 -->
        <main class="app-page-content">
          <ErrorBoundary>
            <router-view v-slot="{ Component }">
              <transition name="bento-fade" mode="out-in">
                <component :is="Component" />
              </transition>
            </router-view>
          </ErrorBoundary>
        </main>
      </div>
    </div>
  </ErrorBoundary>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { Expand, Fold } from '@element-plus/icons-vue'
import AppHeader from '@/components/common/AppHeader.vue'
import AppSidebar from '@/components/common/AppSidebar.vue'
import ErrorBoundary from '@/components/common/ErrorBoundary.vue'
import GlobalTaskProgress from '@/components/common/GlobalTaskProgress.vue'
import { useAppStore } from '@/stores/app'

const route = useRoute()
const appStore = useAppStore()

const { sidebarCollapsed, toggleSidebar } = appStore

const isLoading = ref(false)
const progress = ref(0)

watch(() => route.path, () => {
  isLoading.value = true
  progress.value = 0

  const timer = setInterval(() => {
    progress.value += 15
    if (progress.value >= 90) {
      clearInterval(timer)
    }
  }, 40)

  setTimeout(() => {
    progress.value = 100
    setTimeout(() => {
      isLoading.value = false
    }, 150)
  }, 400)
})
</script>

<style>
/* 全局 Bento 样式覆盖 */
body {
  background: var(--bento-bg, #f5f5f7);
}

[data-theme="dark"] body {
  background: #000000;
}

/* 页面切换动画 */
.bento-fade-enter-active,
.bento-fade-leave-active {
  transition: opacity 0.2s ease, transform 0.2s ease;
}

.bento-fade-enter-from {
  opacity: 0;
  transform: translateY(8px);
}

.bento-fade-leave-to {
  opacity: 0;
  transform: translateY(-8px);
}

/* Element Plus 全局覆盖 - Bento 风格 */
.el-button {
  border-radius: 10px !important;
  font-weight: 500 !important;
}

.el-input__wrapper {
  border-radius: 10px !important;
  box-shadow: none !important;
  border: 1px solid var(--bento-border, rgba(0, 0, 0, 0.06)) !important;
}

.el-input__wrapper:hover {
  border-color: var(--bento-accent-blue, #0071e3) !important;
}

.el-input__wrapper.is-focus {
  border-color: var(--bento-accent-blue, #0071e3) !important;
  box-shadow: 0 0 0 3px rgba(0, 113, 227, 0.1) !important;
}

.el-select__wrapper {
  border-radius: 10px !important;
}

.el-dialog {
  border-radius: 20px !important;
  overflow: hidden;
}

.el-dialog__header {
  padding: 20px 24px 16px !important;
}

.el-dialog__body {
  padding: 0 24px 24px !important;
}

.el-message-box {
  border-radius: 16px !important;
  padding: 20px !important;
}

.el-notification {
  border-radius: 16px !important;
}

.el-dropdown-menu {
  border-radius: 12px !important;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.1) !important;
  border: 1px solid var(--bento-border) !important;
  padding: 6px !important;
}

.el-dropdown-menu__item {
  border-radius: 8px !important;
  padding: 10px 14px !important;
  margin: 2px 0 !important;
}

.el-table {
  border-radius: 12px !important;
  overflow: hidden;
}

.el-card {
  border-radius: var(--bento-radius, 20px) !important;
  border: 1px solid var(--bento-border) !important;
}

/* 滚动条美化 */
::-webkit-scrollbar {
  width: 8px;
  height: 8px;
}

::-webkit-scrollbar-track {
  background: transparent;
}

::-webkit-scrollbar-thumb {
  background: rgba(0, 0, 0, 0.15);
  border-radius: 4px;
}

::-webkit-scrollbar-thumb:hover {
  background: rgba(0, 0, 0, 0.25);
}

[data-theme="dark"] ::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.15);
}

[data-theme="dark"] ::-webkit-scrollbar-thumb:hover {
  background: rgba(255, 255, 255, 0.25);
}
</style>

<style scoped>
.app-bento-layout {
  display: flex;
  height: 100vh;
  width: 100vw;
  overflow: hidden;
  background: var(--bento-bg);
}

/* 侧边栏容器 */
.app-sidebar-wrapper {
  width: 220px;
  flex-shrink: 0;
  transition: width 0.3s cubic-bezier(0.25, 0.46, 0.45, 0.94);
  z-index: 100;
}

.app-sidebar-wrapper.collapsed {
  width: 72px;
}

/* 主内容区 */
.app-content-wrapper {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
  overflow: hidden;
}

/* 顶部栏 */
.app-topbar {
  height: 64px;
  display: flex;
  align-items: center;
  padding: 0 24px;
  background: var(--bento-card-bg);
  border-bottom: 1px solid var(--bento-border);
  flex-shrink: 0;
}

.topbar-toggle {
  width: 40px;
  height: 40px;
  border: none;
  background: var(--color-bg-secondary);
  border-radius: 10px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--color-text-secondary);
  margin-right: 16px;
  transition: all 0.2s;
}

.topbar-toggle:hover {
  background: var(--color-bg-tertiary);
  color: var(--color-text-primary);
}

/* 进度条 */
.topbar-progress {
  height: 2px;
  background: transparent;
  position: relative;
}

.progress-line {
  height: 100%;
  background: linear-gradient(90deg, var(--bento-accent-blue), var(--bento-accent-purple));
  border-radius: 1px;
  transition: width 0.15s ease;
}

/* 页面内容 */
.app-page-content {
  flex: 1;
  overflow: hidden;
  background: var(--bento-bg);
}

/* 响应式 */
@media (max-width: 768px) {
  .app-sidebar-wrapper {
    position: fixed;
    left: 0;
    top: 0;
    bottom: 0;
    z-index: 1000;
    transform: translateX(0);
  }

  .app-sidebar-wrapper.collapsed {
    transform: translateX(-100%);
    width: 220px;
  }

  .app-topbar {
    padding: 0 16px;
  }
}
</style>
