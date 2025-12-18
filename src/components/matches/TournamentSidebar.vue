<template>
  <div class="tournament-sidebar">
    <h3 class="sidebar-title">赛事分类</h3>

    <div class="year-section" v-for="year in years" :key="year">
      <div class="year-header">{{ year }}</div>
      <ul class="tournament-list">
        <li v-for="tournament in getTournamentsByYear(year)" :key="`${year}-${tournament.type}`"
            :class="{ active: isActive(year, tournament.type) }">
          <router-link :to="getRouteLink(year, tournament.type, currentTab)">
            {{ tournament.name }}
          </router-link>
        </li>
      </ul>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useRoute } from 'vue-router'

defineProps({
  currentTab: {
    type: String,
    default: 'today'
  }
})

const route = useRoute()

// 定义年份列表
const years = [2025, 2026]

// 定义每个年份的赛事
const tournaments: Record<number, Array<{type: string, name: string}>> = {
  2025: [
    { type: 'summer', name: '夏季赛' },
    { type: 'spring', name: '春季赛' },
    { type: 'annual', name: '年度总决赛' },
    { type: 'challenger', name: '挑战者杯' }
  ],
  2026: [
    { type: 'summer', name: '夏季赛' },
    { type: 'spring', name: '春季赛' },
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
  return `/${type}/${tab}?year=${year}`
}
</script>

<style scoped>
.tournament-sidebar {
  background-color: #f8f9fa;
  border-radius: 8px;
  padding: 15px;
  width: 200px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
}

.sidebar-title {
  font-size: 16px;
  font-weight: 600;
  margin-bottom: 15px;
  color: #333;
  text-align: center;
}

.year-section {
  margin-bottom: 15px;
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
  padding: 8px 15px;
  color: #495057;
  text-decoration: none;
  border-radius: 6px;
  transition: all 0.2s;
}

.tournament-list li a:hover {
  background-color: #e9ecef;
  color: #007bff;
}

.tournament-list li.active a {
  background-color: #007bff;
  color: white;
}
</style>
