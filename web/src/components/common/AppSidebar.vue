<template>
  <div class="bento-sidebar" :class="{ collapsed }">
    <!-- Logo 区域 -->
    <div class="sidebar-logo">
      <div class="logo-wrapper" v-if="!collapsed">
        <div class="logo-icon">
          <el-icon :size="22"><Monitor /></el-icon>
        </div>
        <div class="logo-text">
          <span class="logo-title">AI-Ops</span>
          <span class="logo-subtitle">智能运维</span>
        </div>
      </div>
      <div class="logo-icon-only" v-else>
        <el-icon :size="24"><Monitor /></el-icon>
      </div>
    </div>

    <!-- 导航菜单 -->
    <nav class="sidebar-nav">
      <div
        v-for="item in menuItems"
        :key="item.path"
        class="nav-item"
        :class="{ active: activeMenu === item.path }"
        @click="navigateTo(item.path)"
      >
        <div class="nav-icon" :style="{ background: item.gradient }">
          <el-icon><component :is="item.icon" /></el-icon>
        </div>
        <span class="nav-text" v-if="!collapsed">{{ item.title }}</span>
        <div class="nav-indicator" v-if="activeMenu === item.path"></div>
      </div>
    </nav>

    <!-- 底部区域 -->
    <div class="sidebar-footer">
      <div class="footer-content" v-if="!collapsed">
        <div class="version-badge">
          <span class="version-dot"></span>
          <span>v1.0.0</span>
        </div>
      </div>
      <el-tooltip v-else content="AI-Ops v1.0.0" placement="right">
        <div class="version-icon">
          <el-icon :size="14"><InfoFilled /></el-icon>
        </div>
      </el-tooltip>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  ChatDotRound, Monitor, SetUp, Document,
  List, Setting, InfoFilled, Tickets, Connection
} from '@element-plus/icons-vue'

defineProps<{
  collapsed?: boolean
}>()

const route = useRoute()
const router = useRouter()

const activeMenu = computed(() => route.path)

const menuItems = [
  {
    path: '/chat',
    title: '智能对话',
    icon: ChatDotRound,
    gradient: 'linear-gradient(135deg, #0071e3 0%, #64d2ff 100%)'
  },
  {
    path: '/hosts',
    title: '主机管理',
    icon: Monitor,
    gradient: 'linear-gradient(135deg, #5e5ce6 0%, #bf5af2 100%)'
  },
  {
    path: '/tools',
    title: '工具列表',
    icon: SetUp,
    gradient: 'linear-gradient(135deg, #30d158 0%, #34c759 100%)'
  },
  {
    path: '/scripts',
    title: '脚本管理',
    icon: Document,
    gradient: 'linear-gradient(135deg, #ff9500 0%, #ffcc00 100%)'
  },
  {
    path: '/tasks',
    title: '任务历史',
    icon: List,
    gradient: 'linear-gradient(135deg, #ff375f 0%, #ff6482 100%)'
  },
  {
    path: '/ops',
    title: '运维控制台',
    icon: Tickets,
    gradient: 'linear-gradient(135deg, #64d2ff 0%, #5ac8fa 100%)'
  },
  {
    path: '/mcp',
    title: 'MCP 配置',
    icon: Connection,
    gradient: 'linear-gradient(135deg, #bf5af2 0%, #da8fff 100%)'
  },
  {
    path: '/settings',
    title: '安全设置',
    icon: Setting,
    gradient: 'linear-gradient(135deg, #8e8e93 0%, #aeaeb2 100%)'
  },
]

const navigateTo = (path: string) => {
  router.push(path)
}
</script>

<style scoped>
.bento-sidebar {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: linear-gradient(180deg, #1d1d1f 0%, #000000 100%);
  padding: 8px;
  transition: all 0.3s cubic-bezier(0.25, 0.46, 0.45, 0.94);
}

/* Logo 区域 */
.sidebar-logo {
  padding: 16px 12px;
  margin-bottom: 8px;
}

.logo-wrapper {
  display: flex;
  align-items: center;
  gap: 12px;
}

.logo-icon {
  width: 42px;
  height: 42px;
  background: linear-gradient(135deg, #0071e3 0%, #5e5ce6 100%);
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  box-shadow: 0 4px 12px rgba(0, 113, 227, 0.3);
}

.logo-icon-only {
  width: 42px;
  height: 42px;
  background: linear-gradient(135deg, #0071e3 0%, #5e5ce6 100%);
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  margin: 0 auto;
  box-shadow: 0 4px 12px rgba(0, 113, 227, 0.3);
}

.logo-text {
  display: flex;
  flex-direction: column;
}

.logo-title {
  font-size: 18px;
  font-weight: 700;
  color: #ffffff;
  letter-spacing: 0.5px;
}

.logo-subtitle {
  font-size: 11px;
  color: rgba(255, 255, 255, 0.5);
  letter-spacing: 1px;
}

/* 导航菜单 */
.sidebar-nav {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding: 0 4px;
  overflow-y: auto;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 12px;
  border-radius: 12px;
  cursor: pointer;
  transition: all 0.2s ease;
  position: relative;
}

.nav-item:hover {
  background: rgba(255, 255, 255, 0.06);
}

.nav-item.active {
  background: rgba(255, 255, 255, 0.1);
}

.nav-icon {
  width: 36px;
  height: 36px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  flex-shrink: 0;
  transition: all 0.2s ease;
  font-size: 16px;
}

.nav-item:not(.active) .nav-icon {
  background: rgba(255, 255, 255, 0.08) !important;
  color: rgba(255, 255, 255, 0.6);
}

.nav-item:hover:not(.active) .nav-icon {
  background: rgba(255, 255, 255, 0.12) !important;
  color: rgba(255, 255, 255, 0.8);
}

.nav-text {
  font-size: 14px;
  font-weight: 500;
  color: rgba(255, 255, 255, 0.7);
  white-space: nowrap;
  transition: all 0.2s ease;
}

.nav-item.active .nav-text {
  color: #ffffff;
}

.nav-item:hover .nav-text {
  color: rgba(255, 255, 255, 0.9);
}

.nav-indicator {
  position: absolute;
  left: 0;
  top: 50%;
  transform: translateY(-50%);
  width: 3px;
  height: 20px;
  background: linear-gradient(180deg, #0071e3 0%, #5e5ce6 100%);
  border-radius: 0 3px 3px 0;
}

/* 折叠状态 */
.collapsed .sidebar-logo {
  padding: 16px 8px;
}

.collapsed .nav-item {
  justify-content: center;
  padding: 10px;
}

.collapsed .nav-text {
  display: none;
}

.collapsed .nav-indicator {
  left: -4px;
}

/* 底部区域 */
.sidebar-footer {
  padding: 12px;
  border-top: 1px solid rgba(255, 255, 255, 0.06);
  margin-top: 8px;
}

.footer-content {
  display: flex;
  align-items: center;
  justify-content: center;
}

.version-badge {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 14px;
  background: rgba(255, 255, 255, 0.06);
  border-radius: 20px;
  font-size: 12px;
  color: rgba(255, 255, 255, 0.5);
}

.version-dot {
  width: 6px;
  height: 6px;
  background: #30d158;
  border-radius: 50%;
  animation: pulse 2s ease-in-out infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

.version-icon {
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(255, 255, 255, 0.06);
  border-radius: 8px;
  color: rgba(255, 255, 255, 0.4);
  margin: 0 auto;
}

/* 滚动条 */
.sidebar-nav::-webkit-scrollbar {
  width: 4px;
}

.sidebar-nav::-webkit-scrollbar-track {
  background: transparent;
}

.sidebar-nav::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.1);
  border-radius: 2px;
}

.sidebar-nav::-webkit-scrollbar-thumb:hover {
  background: rgba(255, 255, 255, 0.2);
}
</style>
