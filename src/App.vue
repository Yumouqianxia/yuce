<template>
  <el-config-provider namespace="ep">
    <div class="app-container">
      <el-container class="layout-container">
        <el-header height="64px">
          <Navbar />
        </el-header>
        <el-main @click="closeMobileMenuIfOpen">
          <CollapsibleAdminSidebar v-if="isAdminRoute" />
          <div :class="{'admin-main-content': isAdminRoute}">
            <router-view v-slot="{ Component }">
              <transition name="fade" mode="out-in">
                <component :is="Component" />
              </transition>
            </router-view>
          </div>
        </el-main>
        <el-footer height="50px">
          <Footer />
        </el-footer>
        <BottomNav />
      </el-container>
    </div>
  </el-config-provider>
</template>

<script lang="ts" setup>
import { computed, ref, provide } from 'vue'
import { useRoute } from 'vue-router'
import { ElConfigProvider, ElContainer, ElHeader, ElMain, ElFooter } from 'element-plus'
import Navbar from '@/components/layout/Navbar.vue'
import Footer from '@/components/layout/Footer.vue'
import BottomNav from '@/components/base/BottomNav.vue'
import CollapsibleAdminSidebar from '@/components/layout/CollapsibleAdminSidebar.vue'

const route = useRoute()
const isAdminRoute = computed(() => route.path.startsWith('/admin'))

// 移动端菜单状态
const mobileMenuOpen = ref(false)

// 提供给子组件使用
provide('mobileMenuOpen', mobileMenuOpen)

// 如果菜单打开状态，点击内容区域时关闭菜单
const closeMobileMenuIfOpen = () => {
  if (mobileMenuOpen.value) {
    mobileMenuOpen.value = false
  }
}
</script>

<style>
:root {
  --primary-color: #3b82f6;
  --primary-light: #93c5fd;
  --primary-dark: #1d4ed8;
  --success-color: #10b981;
  --warning-color: #f59e0b;
  --danger-color: #ef4444;
  --text-primary: #1f2937;
  --text-secondary: #4b5563;
  --text-light: #9ca3af;
  --bg-light: #f9fafb;
  --bg-white: #ffffff;
  --border-color: #e5e7eb;
  --border-radius: 8px;
  --shadow-sm: 0 1px 2px 0 rgba(0, 0, 0, 0.05);
  --shadow-md: 0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06);
  --shadow-lg: 0 10px 15px -3px rgba(0, 0, 0, 0.1), 0 4px 6px -2px rgba(0, 0, 0, 0.05);
  --sidebar-width: 180px; /* 侧边栏宽度 */
}

html, body {
  margin: 0;
  padding: 0;
  font-family: 'Inter', 'PingFang SC', 'Microsoft YaHei', sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  color: var(--text-primary);
  font-size: 14px;
  background-color: var(--bg-light);
}

.layout-container {
  min-height: 100vh;
}

.app-container {
  width: 100%;
  height: 100%;
}

.el-header {
  padding: 0;
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  z-index: 999;
  box-shadow: var(--shadow-sm);
}

.el-main {
  margin-top: 64px;
  padding: 0;
  position: relative;
}

.admin-main-content {
  margin-left: var(--sidebar-width);
  padding: 20px;
  min-height: calc(100vh - 64px - 50px);
  transition: margin-left 0.3s ease;
}

.el-footer {
  padding: 0;
  position: relative;
  z-index: 990;
}

.fade-enter-active, .fade-leave-active {
  transition: opacity 0.2s;
}

.fade-enter-from, .fade-leave-to {
  opacity: 0;
}

@media (max-width: 768px) {
  .el-footer {
    padding-bottom: 56px;
  }

  .admin-main-content {
    margin-left: 0;
    padding: 15px;
    margin-top: 15px;
  }
}

/* 全局消息样式 */
.custom-message {
  min-width: 240px !important;
  padding: 15px 20px !important;
  display: flex !important;
  align-items: center !important;
  border-radius: 6px !important;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.12) !important;
  position: fixed !important;
  top: 20px !important;
  left: 50% !important;
  transform: translateX(-50%) !important;
  z-index: 9999 !important;
}

/* 覆盖Element Plus的默认消息容器样式 */
.el-message {
  top: 20px !important;
  left: 50% !important;
  transform: translateX(-50%) !important;
  margin: 0 !important;
}
</style>

<style scoped>
.app-container {
  height: 100vh;
  width: 100%;
}

.layout-container {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.el-header {
  padding: 0;
  box-shadow: var(--shadow-sm);
  z-index: 10;
  background-color: var(--bg-white);
}

.el-main {
  flex: 1;
  padding: 0;
  background-color: var(--bg-light);
  overflow-x: hidden;
  position: relative;
}

.admin-main-content {
  margin-left: 200px;
  min-height: calc(100vh - 64px - 50px); /* 减去header和footer的高度 */
}

.el-footer {
  padding: 0;
  text-align: center;
  background-color: var(--bg-white);
  border-top: 1px solid var(--border-color);
}

@media (max-width: 768px) {
  .admin-main-content {
    margin-left: 0;
  }
}
</style>