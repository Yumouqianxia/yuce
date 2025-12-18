<template>
  <div class="collapsible-sidebar-container">
    <!-- 侧边栏切换按钮 -->
    <div class="sidebar-toggle" @click="toggleSidebar">
      <div class="toggle-icon">
        <span class="line"></span>
        <span class="line"></span>
        <span class="line"></span>
      </div>
      <span class="toggle-text">赛事分类</span>
    </div>

    <!-- 侧边栏内容 -->
    <div class="sidebar-overlay" v-if="isOpen" @click="closeSidebar"></div>
    <div class="tournament-sidebar" :class="{ 'sidebar-open': isOpen }">
      <div class="sidebar-header">
        <h3 class="sidebar-title">赛事分类</h3>
        <button class="close-button" @click="closeSidebar">×</button>
      </div>

      <div class="year-section" v-for="year in years" :key="year">
        <div class="year-header">{{ year }}</div>
        <ul class="tournament-list">
          <li v-for="tournament in getTournamentsByYear(year)" :key="`${year}-${tournament.type}`"
              :class="{ active: isActive(year, tournament.type) }">
            <router-link :to="getRouteLink(year, tournament.type, currentTab)" @click="closeSidebar">
              {{ tournament.name }}
            </router-link>
          </li>
        </ul>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onUnmounted, onMounted } from 'vue'
import { useRoute } from 'vue-router'

defineProps({
  currentTab: {
    type: String,
    default: 'today'
  }
})

const route = useRoute()
const isOpen = ref(false)

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
  if (isOpen.value) {
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
  if (event.key === 'Escape' && isOpen.value) {
    closeSidebar()
  }
}

// 添加和移除事件监听器
onMounted(() => {
  document.addEventListener('keydown', handleKeyDown)
})

// 组件卸载时恢复滚动并移除事件监听器
onUnmounted(() => {
  enableBodyScroll()
  document.removeEventListener('keydown', handleKeyDown)
})

// 定义年份列表
const years = [2025, 2026]

// 定义每个年份的赛事
const tournaments: Record<number, Array<{type: string, name: string}>> = {
  2025: [
    { type: 'spring', name: '春季赛' },
    { type: 'summer', name: '夏季赛' },
    { type: 'annual', name: '年度总决赛' },
    { type: 'challenger', name: '挑战者杯' }
  ],
  2026: [
    { type: 'spring', name: '春季赛' },
    { type: 'summer', name: '夏季赛' },
    { type: 'annual', name: '年度总决赛' },
    { type: 'challenger', name: '挑战者杯' }
  ]
}

// 根据年份获取赛事列表
const getTournamentsByYear = (year: number) => {
  return tournaments[year] || []
}

// 判断当前路由是否激活
const isActive = (year: number, type: string) => {
  const currentYear = route.meta.year || 2025
  const currentType = route.meta.tournamentType || ''

  return currentYear === year && currentType === type
}

// 获取路由链接
const getRouteLink = (year: number, type: string, tab: string) => {
  // 如果是积分排行榜页面
  if (tab === 'leaderboard') {
    return `/leaderboard/${type}?year=${year}`
  }
  // 其他页面
  return `/${type}/${tab}?year=${year}`
}
</script>

<style scoped>
.collapsible-sidebar-container {
  position: relative;
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
  z-index: 998;
}

.tournament-sidebar {
  position: fixed;
  top: 0;
  left: -250px;
  width: 250px;
  height: 100%;
  background-color: white;
  box-shadow: 2px 0 8px rgba(0, 0, 0, 0.1);
  z-index: 999;
  transition: left 0.3s ease;
  overflow-y: auto;
  padding: 20px;
}

.tournament-sidebar.sidebar-open {
  left: 0;
}

.sidebar-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.sidebar-title {
  font-size: 18px;
  font-weight: 600;
  margin: 0;
  color: var(--text-primary);
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

.year-section {
  margin-bottom: 20px;
}

.year-header {
  font-weight: 600;
  padding: 8px 10px;
  background-color: #e9ecef;
  border-radius: 6px;
  margin-bottom: 8px;
  font-size: 14px;
}

.tournament-list {
  list-style: none;
  padding: 0;
  margin: 0;
}

.tournament-list li {
  margin-bottom: 5px;
  padding: 0;
}

.tournament-list li a {
  display: block;
  padding: 10px 15px;
  color: var(--text-primary);
  text-decoration: none;
  border-radius: 6px;
  transition: all 0.2s;
}

.tournament-list li a:hover {
  background-color: #e9ecef;
  color: var(--primary-color);
}

.tournament-list li.active a {
  background-color: var(--primary-color);
  color: white;
}

@media (max-width: 768px) {
  .tournament-sidebar {
    width: 80%;
    max-width: 300px;
    left: -100%;
  }

  .sidebar-toggle {
    width: 100%;
    justify-content: center;
  }
}
</style>
