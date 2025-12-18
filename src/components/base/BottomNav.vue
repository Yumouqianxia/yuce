<template>
  <div class="bottom-nav">
    <a href="/" class="nav-item" :class="{ active: currentPath === '/' }">
      <nav-icon icon="home" />
      <span>首页</span>
    </a>
    <a href="/matches" class="nav-item" :class="{ active: currentPath.includes('/matches') }">
      <nav-icon icon="trophy" />
      <span>比赛</span>
    </a>

    <a v-if="isAuthenticated" href="/upcoming-matches" class="nav-item" :class="{ active: currentPath === '/upcoming-matches' || currentPath.includes('/prediction-history') || currentPath.includes('/prediction-rules') }">
      <nav-icon icon="trophy" />
      <span>预测</span>
    </a>
    <a href="/leaderboard" class="nav-item" :class="{ active: currentPath.includes('/leaderboard') }">
      <nav-icon icon="chart-bar" />
      <span>排行</span>
    </a>
    <a v-if="isAdmin" href="/admin" class="nav-item" :class="{ active: currentPath === '/admin' }">
      <nav-icon icon="cog" />
      <span>管理</span>
    </a>
  </div>
</template>

<script lang="ts" setup>
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import NavIcon from './NavIcon.vue'
import { useUserStore } from '@/stores/user'

const route = useRoute()
const userStore = useUserStore()
const currentPath = computed(() => route.path)
const isAdmin = computed(() => userStore.user?.role === 'admin')
const isAuthenticated = computed(() => userStore.isAuthenticated)
</script>

<style scoped>
.bottom-nav {
  display: flex;
  justify-content: space-around;
  align-items: center;
  background-color: var(--bg-white);
  box-shadow: 0 -2px 10px rgba(0, 0, 0, 0.05);
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  height: 56px;
  z-index: 100;
  padding: 0 10px;
}

.nav-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  text-decoration: none;
  color: var(--text-secondary);
  flex: 1;
  height: 100%;
  transition: all 0.2s ease;
  position: relative;
  margin: 0 5px;
}

.nav-item:hover {
  background-color: var(--bg-light);
}

.nav-item.active {
  color: var(--primary-color);
}

.nav-item.active::after {
  content: '';
  position: absolute;
  bottom: 0;
  left: 25%;
  right: 25%;
  height: 3px;
  background-color: var(--primary-color);
  border-radius: 3px 3px 0 0;
}

.nav-item span {
  font-size: 12px;
  margin-top: 4px;
}

@media (min-width: 768px) {
  .bottom-nav {
    display: none;
  }
}
</style>