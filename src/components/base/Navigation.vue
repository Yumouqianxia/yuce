<template>
  <div class="navigation">
    <a href="/" class="nav-item" :class="{ active: currentPath === '/' }">
      <nav-icon icon="home" />
      <span>首页</span>
    </a>
    <a href="/matches" class="nav-item" :class="{ active: currentPath === '/matches' }">
      <nav-icon icon="trophy" />
      <span>比赛列表</span>
    </a>
    <a href="/leaderboard" class="nav-item" :class="{ active: currentPath === '/leaderboard' }">
      <nav-icon icon="list" />
      <span>排行榜</span>
    </a>
  </div>
</template>

<script lang="ts" setup>
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import NavIcon from './NavIcon.vue'

const route = useRoute()
const currentPath = computed(() => route.path)
</script>

<style scoped>
.navigation {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 30px;
  padding: 10px 0;
  width: 100%;
  max-width: 600px;
  margin: 0 auto;
}

.nav-item {
  display: flex;
  flex-direction: row;
  align-items: center;
  justify-content: center;
  text-decoration: none;
  color: var(--text-secondary);
  padding: 8px 16px;
  border-radius: var(--border-radius);
  transition: all 0.2s ease;
  background-color: #f5f7fa;
  border: 1px solid var(--border-color);
  min-width: 100px;
  box-shadow: var(--shadow-sm);
}

.nav-item:hover {
  background-color: #ebeef5;
  color: var(--primary-color);
}

.nav-item.active {
  color: var(--primary-color);
  background-color: #e6f1ff;
  font-weight: 500;
  border-color: var(--primary-light);
}

.nav-item span {
  margin-left: 6px;
  font-size: 14px;
  white-space: nowrap;
}
</style> 