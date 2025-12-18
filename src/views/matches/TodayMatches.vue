<template>
  <div class="matches-container">
    <h1 class="page-title">{{ pageTitle }}</h1>

    <div v-if="loading" class="loading-container">
      <el-skeleton :rows="3" animated />
    </div>
    <div v-else-if="matches.length === 0" class="empty-data">
      今日暂无赛程
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
import { getBeijingStartOfDay, isSameBeijingDay } from '@/utils/date'

const userStore = useUserStore()
const route = useRoute()

// 状态
const loading = ref(true)
const matches = ref<Match[]>([])
const allMatches = ref<Match[]>([])

// 获取路由参数
const getTournamentType = () => (route.meta.tournamentType as string) || (route.query.tournament as string) || ''
// 仅当路由明确提供年份时才返回；否则不按年份筛选
const getYearParam = (): number | undefined => {
  const metaYear = route.meta.year as number | undefined
  const queryYear = route.query.year !== undefined ? Number(route.query.year) : undefined
  return Number.isFinite(metaYear as number) ? (metaYear as number) : (Number.isFinite(queryYear as number) ? (queryYear as number) : undefined)
}

// 页面标题
const pageTitle = computed(() => {
  const tournamentType = getTournamentType()
  const year = getYearParam()
  if (!tournamentType) return '今日赛程'

  const tournamentNames: Record<string, string> = {
    'spring': '春季赛',
    'summer': '夏季赛',
    'annual': '年度总决赛',
    'challenger': '挑战者杯'
  }

  return `${year ?? ''}${year ? '' : ''}KPL${tournamentNames[tournamentType] || ''}今日赛程`
})

// 获取今日赛程
const fetchMatches = async () => {
  loading.value = true
  try {
    // 获取所有比赛
    const response = await axios.get('/api/matches', {
      headers: { Authorization: `Bearer ${userStore.token}` }
    })

    console.log('获取比赛列表成功:', response.data)
    console.log('响应数据类型:', typeof response.data)
    console.log('响应数据键:', response.data ? Object.keys(response.data) : 'null')

    // 确保数据是数组
    let fetchedMatches = [];
    if (Array.isArray(response.data)) {
      fetchedMatches = response.data;
      console.log('数据是直接数组')
    } else if (response.data && Array.isArray(response.data.data)) {
      fetchedMatches = response.data.data;
      console.log('数据在 data 字段中')
    } else if (response.data && typeof response.data === 'object') {
      // 尝试其他可能的数据结构
      console.log('尝试解析数据结构:', Object.keys(response.data));
      if (response.data.results && Array.isArray(response.data.results)) {
        fetchedMatches = response.data.results;
        console.log('数据在 results 字段中')
      } else if (response.data.matches && Array.isArray(response.data.matches)) {
        fetchedMatches = response.data.matches;
        console.log('数据在 matches 字段中')
      }
    }

    console.log('解析后的比赛数据数量:', fetchedMatches.length)
    console.log('前3场比赛示例(原始):', fetchedMatches.slice(0, 3))

    // 统一字段与状态（兼容后端不同命名/大小写）
    const normalizeStatus = (s: any) => {
      const v = (s || '').toString().toLowerCase()
      if (v === 'upcoming' || v === 'not_started') return 'not_started'
      if (v === 'live' || v === 'in_progress') return 'in_progress'
      if (v === 'finished' || v === 'completed') return 'completed'
      if (v === 'cancelled' || v === 'canceled') return 'cancelled'
      return v as any
    }

    const normalizedMatches = fetchedMatches.map((m: any) => {
      return {
        ...m,
        optionA: m.optionA ?? m.team_a ?? m.teamA,
        optionB: m.optionB ?? m.team_b ?? m.teamB,
        matchTime: m.matchTime ?? m.start_time ?? m.startTime,
        status: normalizeStatus(m.status),
        tournamentType: (m.tournamentType ?? m.tournament ?? '').toString().toLowerCase(),
        year: typeof m.year === 'string' ? Number(m.year) : m.year,
      }
    }) as any[]

    console.log('前3场比赛示例(规范化后):', normalizedMatches.slice(0, 3))
    
    // 详细检查第一场比赛的数据结构
    if (fetchedMatches.length > 0) {
      const firstMatch = fetchedMatches[0]
      console.log('第一场比赛详细信息:')
      console.log('- ID:', firstMatch.id)
      console.log('- optionA:', firstMatch.optionA)
      console.log('- optionB:', firstMatch.optionB)
      console.log('- matchTime:', firstMatch.matchTime)
      console.log('- matchTime类型:', typeof firstMatch.matchTime)
      console.log('- status:', firstMatch.status)
      console.log('- tournamentType:', firstMatch.tournamentType)
      console.log('- year:', firstMatch.year)
      console.log('- 完整对象:', JSON.stringify(firstMatch, null, 2))
    }
    
    // 查找今天的比赛
    const today = getBeijingStartOfDay(new Date())
    console.log('今天日期 (北京时间，用于筛选):', today.toISOString())
    
    const todayMatches = normalizedMatches.filter((match: any) => {
      if (!match || !match.matchTime) return false
      const same = isSameBeijingDay(match.matchTime as any, today)
      console.log('比赛', match.id, '时间:', match.matchTime, '是否今天(北京时间):', same)
      return same
    })
    
    console.log('手动筛选的今日比赛:', todayMatches.length, '场')
    if (todayMatches.length > 0) {
      console.log('今日比赛详情:', todayMatches)
    }

    // 保存所有比赛
    allMatches.value = normalizedMatches as any;

    // 筛选比赛
    filterMatches();

    console.log('今日比赛:', matches.value)
  } catch (error) {
    console.error('获取今日赛程失败:', error)
    ElMessage.error('获取今日赛程失败')
  } finally {
    loading.value = false
  }
}

// 筛选比赛
const filterMatches = () => {
  console.log('开始筛选比赛，总比赛数:', allMatches.value.length)
  
  // 获取今天的日期（只保留年月日）
  const today = getBeijingStartOfDay(new Date())
  console.log('今天的日期(北京时间):', today.toISOString())

  // 首先筛选今日比赛
  let filteredMatches = allMatches.value.filter((match: Match) => {
    if (!match || !match.matchTime) {
      console.log('比赛数据无效:', match)
      return false;
    }
    
    const isToday = isSameBeijingDay(match.matchTime as any, today)
    
    if (isToday) {
      console.log('找到今日比赛:', match.id, match.optionA, 'vs', match.optionB, match.matchTime)
    }
    
    return isToday
  })

  console.log('今日比赛筛选结果:', filteredMatches.length, '场比赛')

  // 获取赛事类型和年份
  const tournamentType = getTournamentType()
  const year = getYearParam()

  console.log('赛事类型筛选:', tournamentType, '年份筛选:', year)

  // 如果指定了赛事类型，进一步筛选
  if (tournamentType) {
    const beforeFilter = filteredMatches.length
    filteredMatches = filteredMatches.filter(match => match.tournamentType === tournamentType)
    console.log('赛事类型筛选后:', filteredMatches.length, '场比赛 (从', beforeFilter, '场)')
  }

  // 如果指定了年份，进一步筛选
  if (typeof year === 'number' && !Number.isNaN(year)) {
    const beforeFilter = filteredMatches.length
    filteredMatches = filteredMatches.filter(match => match.year === year)
    console.log('年份筛选后:', filteredMatches.length, '场比赛 (从', beforeFilter, '场)')
  }

  matches.value = filteredMatches
  console.log('最终显示比赛数:', matches.value.length)
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
