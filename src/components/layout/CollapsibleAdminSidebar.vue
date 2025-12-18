<template>
  <div class="admin-sidebar-container">
    <!-- 移动端侧边栏切换按钮 -->
    <div class="sidebar-toggle" v-if="isMobile" @click="toggleSidebar">
      <div class="toggle-icon">
        <span class="line"></span>
        <span class="line"></span>
        <span class="line"></span>
      </div>
      <span class="toggle-text">管理菜单</span>
    </div>

    <!-- 侧边栏遮罩层 -->
    <div class="sidebar-overlay" v-if="isMobile && isOpen" @click="closeSidebar"></div>

    <!-- 侧边栏内容 -->
    <div class="admin-sidebar" :class="{ 'sidebar-open': isMobile && isOpen, 'mobile-sidebar': isMobile }">
      <div class="sidebar-header">
        <h3>管理面板</h3>
        <button v-if="isMobile" class="close-button" @click="closeSidebar">×</button>
      </div>
      <ul class="sidebar-menu">
        <li class="sidebar-item" :class="{ active: $route.name === 'AdminSite' }">
          <a href="/admin/site" @click="isMobile && closeSidebar()">公告管理</a>
        </li>
        <li class="sidebar-item" :class="{ active: $route.name === 'AdminUsers' }">
          <a href="/admin/users" @click="isMobile && closeSidebar()">用户管理</a>
        </li>
        <li class="sidebar-item" :class="{ active: $route.name === 'AdminMatches' }">
          <a href="/admin/matches" @click="isMobile && closeSidebar()">比赛管理</a>
        </li>
        <li class="sidebar-item" :class="{ active: $route.name === 'AdminSettings' }">
          <a href="/admin/settings" @click="isMobile && closeSidebar()">系统设置</a>
        </li>
      </ul>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch } from 'vue'

// 状态
const isOpen = ref(false)
const isMobile = ref(false)

// 检查是否为移动设备
const checkMobile = () => {
  isMobile.value = window.innerWidth <= 768
}

// 打开/关闭侧边栏
const toggleSidebar = () => {
  isOpen.value = !isOpen.value
  toggleBodyScroll()
}

// 关闭侧边栏
const closeSidebar = () => {
  isOpen.value = false
  enableBodyScroll()
}

// 切换身体滚动
const toggleBodyScroll = () => {
  if (isOpen.value && isMobile.value) {
    disableBodyScroll()
  } else {
    enableBodyScroll()
  }
}

// 禁用身体滚动
const disableBodyScroll = () => {
  document.body.style.overflow = 'hidden'
}

// 启用身体滚动
const enableBodyScroll = () => {
  document.body.style.overflow = ''
}

// 处理ESC键关闭侧边栏
const handleKeyDown = (event: KeyboardEvent) => {
  if (event.key === 'Escape' && isOpen.value && isMobile.value) {
    closeSidebar()
  }
}

// 监听窗口大小变化
const handleResize = () => {
  checkMobile()
  // 如果从移动端切换到桌面端，关闭侧边栏并恢复滚动
  if (!isMobile.value && isOpen.value) {
    isOpen.value = false
    enableBodyScroll()
  }
}

// 添加和移除事件监听器
onMounted(() => {
  checkMobile() // 初始检查
  window.addEventListener('resize', handleResize)
  document.addEventListener('keydown', handleKeyDown)
})

// 组件卸载时恢复滚动并移除事件监听器
onUnmounted(() => {
  enableBodyScroll()
  window.removeEventListener('resize', handleResize)
  document.removeEventListener('keydown', handleKeyDown)
})

// 监听移动状态变化
watch(isMobile, (newValue) => {
  if (!newValue) {
    // 如果不是移动设备，确保侧边栏是打开的
    isOpen.value = false
    enableBodyScroll()
  }
})
</script>

<style scoped>
.admin-sidebar-container {
  position: relative;
}

/* 桌面端侧边栏 */
.admin-sidebar {
  width: 180px;
  position: fixed;
  top: 64px; /* 导航栏高度 */
  left: 0;
  height: calc(100vh - 64px);
  background-color: #fff;
  border-right: 1px solid #e6e6e6;
  z-index: 1000;
  transition: left 0.3s ease;
}

/* 移动端侧边栏 */
.admin-sidebar.mobile-sidebar {
  left: -250px;
  width: 250px;
  z-index: 1001;
}

.admin-sidebar.mobile-sidebar.sidebar-open {
  left: 0;
}

.sidebar-toggle {
  display: flex;
  align-items: center;
  background-color: var(--primary-color);
  color: white;
  padding: 10px 15px;
  border-radius: 8px;
  cursor: pointer;
  margin-bottom: 15px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  transition: all 0.3s ease;
}

.sidebar-toggle:hover {
  background-color: var(--primary-dark);
}

.toggle-icon {
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  width: 18px;
  height: 14px;
  margin-right: 10px;
}

.toggle-icon .line {
  height: 2px;
  background-color: white;
  border-radius: 1px;
}

.toggle-text {
  font-weight: 500;
  font-size: 14px;
}

.sidebar-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.5);
  z-index: 1000;
}

.sidebar-header {
  padding: 16px;
  border-bottom: 1px solid #e6e6e6;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.sidebar-header h3 {
  margin: 0;
  font-size: 16px;
  color: #303133;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.close-button {
  background: none;
  border: none;
  font-size: 24px;
  color: var(--text-secondary);
  cursor: pointer;
  padding: 0;
  line-height: 1;
}

.sidebar-menu {
  list-style: none;
  padding: 0;
  margin: 0;
}

.sidebar-item {
  margin-bottom: 1px;
  width: 100%;
}

.sidebar-item a {
  display: block;
  width: 100%;
  text-align: left;
  padding: 12px 16px;
  text-decoration: none;
  border: none;
  background-color: transparent;
  font-size: 14px;
  color: #606266;
  cursor: pointer;
  transition: all 0.3s;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  box-sizing: border-box;
}

.sidebar-item a:hover {
  background-color: #f5f7fa;
  color: #409EFF;
}

.sidebar-item.active a {
  background-color: #f0f9ff;
  color: #409EFF;
  border-left: 3px solid #409EFF;
}

@media (max-width: 768px) {
  .sidebar-toggle {
    width: 100%;
    justify-content: center;
  }
}
</style>
