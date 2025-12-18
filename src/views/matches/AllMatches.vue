<template>
  <div class="matches-container">
    <h1 class="page-title">所有比赛 (调试用)</h1>

    <div v-if="loading" class="loading-container">
      <el-skeleton :rows="3" animated />
    </div>
    <div v-else-if="matches.length === 0" class="empty-data">
      没有找到任何比赛
    </div>
    <div v-else>
      <div class="stats">
        <p>总比赛数: {{ matches.length }}</p>
        <p>今日比赛数: {{ todayMatches.length }}</p>
        <p>即将开始的比赛数: {{ upcomingMatches.length }}</p>
      </div>
      
      <div class="matches-grid">
        <div v-for="match in matches.slice(0, 10)" :key="match.id" class="match-debug">
          <h3>{{ match.optionA }} vs {{ match.optionB }}</h3>
          <p>ID: {{ match.id }}</p>
          <p>时间: {{ match.matchTime }}</p>
          <p>状态: {{ match.status }}</p>
          <p>赛事类型: {{ match.tournamentType }}</p>
          <p>年份: {{ match.year }}</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { ElSkeleton, ElMessage } from 'element-plus'
import { Match } from '@/types/match'
import { useUserStore } from '@/stores/user'
import axios from 'axios'
import { getBeijingStartOfDay, convertToBeijingTime } from '@/utils/date'

const userStore = useUserStore()

// 状态
const loading = ref(true)
const matches = ref<Match[]>([])

// 计算属性
const todayMatches = computed(() => {
  const today = getBeijingStartOfDay(new Date())
  
  return matches.value.filter((match: any) => {
    const t = match.matchTime ?? match.start_time
    if (!t) return false
    const matchDate = getBeijingStartOfDay(t)
    return matchDate.getTime() === today.getTime()
  })
})

const upcomingMatches = computed(() => {
  const now = convertToBeijingTime(new Date())
  const fiveDaysMs = 5 * 24 * 60 * 60 * 1000
  return matches.value.filter((match: Match) => {
    const timeStr = (match as any).matchTime ?? (match as any).start_time
    if (!timeStr) return false
    const matchDate = convertToBeijingTime(timeStr)
    const status = (match as any).status
    const diff = matchDate.getTime() - now.getTime()
    return diff >= -fiveDaysMs && diff <= fiveDaysMs && (status === 'not_started' || status === 'upcoming')
  })
})

// 获取所有比赛
const fetchMatches = async () => {
  loading.value = true
  try {
    const response = await axios.get('/api/matches', {
      headers: { Authorization: `Bearer ${userStore.token}` }
    })

    console.log('调试页面 - 获取比赛列表成功:', response.data)

    // 确保数据是数组
    let fetchedMatches = [];
    if (Array.isArray(response.data)) {
      fetchedMatches = response.data;
    } else if (response.data && Array.isArray(response.data.data)) {
      fetchedMatches = response.data.data;
    } else if (response.data && typeof response.data === 'object') {
      console.log('调试页面 - 尝试解析数据结构:', Object.keys(response.data));
      if (response.data.results && Array.isArray(response.data.results)) {
        fetchedMatches = response.data.results;
      } else if (response.data.matches && Array.isArray(response.data.matches)) {
        fetchedMatches = response.data.matches;
      }
    }

    matches.value = fetchedMatches
    console.log('调试页面 - 解析后的比赛数据:', fetchedMatches.length, '场比赛')
    console.log('调试页面 - 前5场比赛:', fetchedMatches.slice(0, 5))
  } catch (error) {
    console.error('调试页面 - 获取比赛失败:', error)
    ElMessage.error('获取比赛失败')
  } finally {
    loading.value = false
  }
}

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

.stats {
  background: #f5f7fa;
  padding: 20px;
  border-radius: 8px;
  margin-bottom: 20px;
}

.stats p {
  margin: 5px 0;
  font-weight: 600;
}

.matches-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 20px;
}

.match-debug {
  background: white;
  border: 1px solid #e4e7ed;
  border-radius: 8px;
  padding: 15px;
}

.match-debug h3 {
  margin: 0 0 10px 0;
  color: #303133;
}

.match-debug p {
  margin: 5px 0;
  font-size: 14px;
  color: #606266;
}
</style>