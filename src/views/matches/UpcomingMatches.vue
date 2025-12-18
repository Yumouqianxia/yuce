<template>
  <div class="matches-container">
    <h1 class="page-title">{{ pageTitle }}</h1>

    <div v-if="loading" class="loading-container">
      <el-skeleton :rows="3" animated />
    </div>
    <div v-else-if="matches.length === 0" class="empty-data">
      暂无未来赛程
    </div>
    <div v-else class="matches-grid">
      <MatchCard
        v-for="match in matches"
        :key="match.id"
        :match="match"
        class="match-card"
        :showPredictButton="false"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch, computed } from 'vue'
import { ElSkeleton, ElMessage } from 'element-plus'
import { useRoute } from 'vue-router'
import MatchCard from '@/components/matches/MatchCard.vue'
import { Match } from '@/types/match'
import { useUserStore } from '@/stores/user'
import axios from 'axios'
import { getBeijingStartOfDay, convertToBeijingTime } from '@/utils/date'

const userStore = useUserStore()
const route = useRoute()

// 状态
const loading = ref(true)
const matches = ref<Match[]>([])
const allMatches = ref<Match[]>([])

// 获取路由参数
const getTournamentType = () => route.meta.tournamentType as string || route.query.tournament as string || ''
const getYear = () => route.meta.year as number || Number(route.query.year) || 2025

// 页面标题
const pageTitle = computed(() => {
  const tournamentType = getTournamentType()
  const year = getYear()
  if (!tournamentType) return '未来赛程'

  const tournamentNames: Record<string, string> = {
    'spring': '春季赛',
    'summer': '夏季赛',
    'annual': '年度总决赛',
    'challenger': '挑战者杯'
  }

  return `${year}KPL${tournamentNames[tournamentType] || ''}未来赛程`
})

// 获取未来赛程
const fetchMatches = async () => {
  loading.value = true
  try {
    // 获取所有比赛
    const response = await axios.get('/api/matches', {
      headers: { Authorization: `Bearer ${userStore.token}` }
    })

    // 确保数据是数组
    let fetchedMatches = [];
    if (Array.isArray(response.data)) {
      fetchedMatches = response.data;
    } else if (response.data && Array.isArray(response.data.data)) {
      fetchedMatches = response.data.data;
    } else if (response.data && typeof response.data === 'object') {
      // 尝试其他可能的数据结构
      if (response.data.results && Array.isArray(response.data.results)) {
        fetchedMatches = response.data.results;
      } else if (response.data.matches && Array.isArray(response.data.matches)) {
        fetchedMatches = response.data.matches;
      }
    }

    // 保存所有比赛
    allMatches.value = fetchedMatches;

    // 筛选比赛
    filterMatches();

    console.log('未来比赛:', matches.value)
  } catch (error) {
    console.error('获取未来赛程失败:', error)
    ElMessage.error('获取未来赛程失败')
  } finally {
    loading.value = false
  }
}

// 筛选比赛
const filterMatches = () => {
  const today = getBeijingStartOfDay(new Date())
  const now = convertToBeijingTime(new Date())
  const fiveDaysMs = 5 * 24 * 60 * 60 * 1000

  let filteredMatches = allMatches.value.filter((match: Match) => {
    const timeStr = (match as any).matchTime ?? (match as any).start_time
    if (!timeStr) return false
    const matchDate = convertToBeijingTime(timeStr)
    const status = (match as any).status
    const diff = matchDate.getTime() - now.getTime()
    return diff >= -fiveDaysMs && diff <= fiveDaysMs && (status === 'not_started' || status === 'upcoming')
  })

  // 获取赛事类型和年份
  const tournamentType = getTournamentType()
  const year = getYear()

  // 如果指定了赛事类型，进一步筛选
  if (tournamentType) {
    filteredMatches = filteredMatches.filter(match => match.tournamentType === tournamentType)
  }

  // 如果指定了年份，进一步筛选
  if (year) {
    filteredMatches = filteredMatches.filter(match => match.year === year)
  }

  matches.value = filteredMatches
}

// 监听路由变化
watch(() => [route.meta.tournamentType, route.query.tournament, route.meta.year, route.query.year], () => {
  if (allMatches.value.length > 0) {
    filterMatches()
  }
}, { immediate: true })

// 初始化
onMounted(() => {
  fetchMatches()
})
</script>

<style scoped>
.matches-container {
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 20px;
}

.page-title {
  font-size: 28px;
  font-weight: 600;
  margin-bottom: 30px;
  color: #303133;
  text-align: center;
}

.loading-container {
  padding: 20px 0;
}

.empty-data {
  text-align: center;
  color: #909399;
  padding: 40px 0;
  font-size: 16px;
}

.matches-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 20px;
}

.match-card {
  transition: all 0.3s ease;
}

.match-card:hover {
  transform: translateY(-5px);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.12);
}

/* 响应式调整 */
@media (max-width: 768px) {
  .matches-grid {
    grid-template-columns: 1fr;
  }

  .page-title {
    font-size: 24px;
    margin-bottom: 20px;
  }
}
</style>
