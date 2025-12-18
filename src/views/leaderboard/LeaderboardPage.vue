<template>
  <div class="leaderboard-page-container">
    <h1 class="page-title">{{ pageTitle }}</h1>

    <div class="content-with-sidebar">
      <!-- 可折叠侧边栏 -->
      <CollapsibleSidebar currentTab="leaderboard" />

      <div class="main-content">
        <div v-if="loading" class="loading-indicator">
          <div class="spinner"></div>
          <span>加载中...</span>
        </div>

        <div v-else-if="leaderboard.length === 0" class="empty-state">
          暂无排行榜数据
        </div>

        <div v-else class="leaderboard-content">
          <div class="leaderboard-header">
            <div class="rank-column">排名</div>
            <div class="user-column">用户</div>
            <div class="predictions-column">预测数</div>
            <div class="correct-column">正确数</div>
            <div class="accuracy-column">准确率</div>
            <div class="points-column">积分</div>
          </div>

          <div v-for="(user, index) in leaderboard" :key="user.userId || user.id" class="leaderboard-row">
            <div class="rank-column" :class="{ 'top-rank': index < 3 }">{{ index + 1 }}</div>
            <div class="user-column">
              <div class="user-info">
                <div class="user-avatar" v-if="user.avatar">
                  <img :src="user.avatar" alt="用户头像" @error="handleAvatarError">
                </div>
                <div class="user-avatar placeholder" v-else>
                  {{ getInitial(user.nickname || user.username) }}
                </div>
                <span class="user-name">{{ user.nickname || user.username }}</span>
              </div>
            </div>
            <div class="predictions-column">{{ user.total_predictions || user.totalPredictions || 0 }}</div>
            <div class="correct-column">{{ user.accurate_predictions || user.correctPredictions || 0 }}</div>
            <div class="accuracy-column">{{ calculateAccuracy(user) }}%</div>
            <div class="points-column">{{ user.total_points || user.points || user.totalPoints || 0 }}</div>
          </div>
        </div>

        <!-- 我的排名信息 -->
        <div v-if="userStore.isAuthenticated && myRankInfo" class="my-rank-info">
          <div class="my-rank-label">我的排名:</div>
          <div class="my-rank-value">{{ myRankInfo.rank }}</div>
          <div class="my-points-label">我的积分:</div>
          <div class="my-points-value">{{ myRankInfo.points }}</div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed, watch } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useUserStore } from '@/stores/user'
import { get } from '@/api/http'
import CollapsibleSidebar from '@/components/matches/CollapsibleSidebar.vue'
import { getFullAvatarUrl } from '@/utils/url'

// 定义用户类型
interface LeaderboardUser {
  id?: number
  userId?: number
  username: string
  nickname?: string
  avatar?: string
  total_predictions?: number
  totalPredictions?: number
  accurate_predictions?: number
  correctPredictions?: number
  total_points?: number
  points?: number
  totalPoints?: number
}

// 我的排名信息
interface MyRankInfo {
  rank: number
  points: number
}

const route = useRoute()
const userStore = useUserStore()

// 状态
const loading = ref(true)
const leaderboard = ref<LeaderboardUser[]>([])
const myRankInfo = ref<MyRankInfo | null>(null)

// 计算属性
const tournamentType = computed(() => route.meta.tournamentType as string || '')
const pageTitle = computed(() => {
  if (tournamentType.value === 'spring') return '2025KPL春季赛积分榜'
  if (tournamentType.value === 'summer') return '2025KPL夏季赛积分榜'
  if (tournamentType.value === 'annual') return '2025KPL年度总决赛积分榜'
  if (tournamentType.value === 'challenger') return '2025KPL挑战者杯积分榜'
  return '积分排行榜'
})

// 计算准确率
const calculateAccuracy = (user: LeaderboardUser): number => {
  const totalPredictions = user.total_predictions || user.totalPredictions || 0
  const accuratePredictions = user.accurate_predictions || user.correctPredictions || 0

  if (totalPredictions === 0) return 0
  return Math.round((accuratePredictions / totalPredictions) * 100)
}

// 获取用户名或昵称的首字母
const getInitial = (name: string): string => {
  if (!name) return '?'

  // 如果是中文，返回第一个字
  if (/[\u4e00-\u9fa5]/.test(name.charAt(0))) {
    return name.charAt(0)
  }

  // 如果是英文，返回首字母大写
  return name.charAt(0).toUpperCase()
}

// 处理头像加载错误
const handleAvatarError = (event: Event) => {
  const img = event.target as HTMLImageElement
  img.style.display = 'none'

  // 查找父元素并添加首字母占位符
  const parent = img.parentElement
  if (parent) {
    parent.classList.add('placeholder')
    const user = leaderboard.value.find(u =>
      u.avatar === img.src ||
      u.avatar === img.src.split('?')[0]
    )
    if (user) {
      parent.textContent = getInitial(user.nickname || user.username)
    } else {
      parent.textContent = '?'
    }
  }
}

// 获取积分排行榜
const fetchLeaderboard = async () => {
  loading.value = true
  try {
    // 构建API URL，根据赛事类型添加参数（后端 /api/leaderboard）
    let url = '/api/leaderboard'
    if (tournamentType.value) {
      url += `?tournament=${tournamentType.value.toUpperCase()}`
    }

    // 获取排行榜数据
    const response = await get(url)
    console.log('获取到的排行榜数据:', response)

    // 处理响应数据
    let leaderboardData: LeaderboardUser[] = []
    if (Array.isArray(response)) {
      leaderboardData = response
    } else if (response && typeof response === 'object' && 'data' in response && Array.isArray((response as any).data)) {
      leaderboardData = (response as any).data
    }

    // 处理头像URL
    leaderboardData.forEach(user => {
      if (user.avatar) {
        user.avatar = getFullAvatarUrl(user.avatar)
      }
    })

    // 更新排行榜数据
    leaderboard.value = leaderboardData

    // 查找当前用户的排名
    if (userStore.isAuthenticated && userStore.user) {
      const currentUserId = userStore.user.id
      const myRank = leaderboardData.findIndex(user =>
        (user.id === currentUserId) || (user.userId === currentUserId)
      )

      if (myRank !== -1) {
        const user = leaderboardData[myRank]
        myRankInfo.value = {
          rank: myRank + 1,
          points: user.total_points || user.points || user.totalPoints || 0
        }
      } else {
        // 如果在排行榜中找不到当前用户，获取用户积分
        await fetchMyPoints()
      }
    }
  } catch (error) {
    console.error('获取积分排行榜失败:', error)
    ElMessage({
      message: '获取积分排行榜失败，请稍后再试',
      type: 'warning',
      duration: 3000
    })
  } finally {
    loading.value = false
  }
}

// 获取当前用户积分
const fetchMyPoints = async () => {
  if (!userStore.isAuthenticated) return

  try {
    // 如果后端未提供该端点，可忽略; 这里保持调用占位符
    const points = await get('/api/leaderboard/stats', { self: true })
    console.log('获取到的用户积分:', points)

    // 更新我的排名信息
    myRankInfo.value = {
      rank: leaderboard.value.length + 1, // 排在最后
      points: typeof points === 'number' ? points : 0
    }
  } catch (error) {
    console.error('获取用户积分失败:', error)
  }
}

// 监听路由变化，重新获取数据
watch(() => route.path, () => {
  fetchLeaderboard()
})

// 初始化
onMounted(() => {
  fetchLeaderboard()
})
</script>

<style scoped>
.leaderboard-page-container {
  max-width: 1200px;
  margin: 0 auto;
  padding: 20px;
}

.page-title {
  font-size: 24px;
  font-weight: 600;
  margin-bottom: 20px;
  color: var(--text-primary);
}

.content-with-sidebar {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.main-content {
  flex: 1;
  background-color: var(--bg-white);
  border-radius: var(--border-radius);
  box-shadow: var(--shadow-sm);
  padding: 20px;
  min-height: 500px;
  position: relative;
}

.loading-indicator {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 300px;
}

.spinner {
  width: 40px;
  height: 40px;
  border: 4px solid rgba(0, 0, 0, 0.1);
  border-left-color: var(--primary-color);
  border-radius: 50%;
  animation: spin 1s linear infinite;
  margin-bottom: 10px;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.empty-state {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 300px;
  color: var(--text-light);
  font-size: 16px;
}

.leaderboard-content {
  width: 100%;
}

.leaderboard-header {
  display: flex;
  background-color: #f5f7fa;
  padding: 12px 16px;
  border-radius: var(--border-radius) var(--border-radius) 0 0;
  font-weight: 600;
  color: var(--text-secondary);
}

.leaderboard-row {
  display: flex;
  padding: 16px;
  border-bottom: 1px solid var(--border-color);
  transition: background-color 0.2s;
}

.leaderboard-row:hover {
  background-color: #f9fafb;
}

.leaderboard-row:last-child {
  border-bottom: none;
  border-radius: 0 0 var(--border-radius) var(--border-radius);
}

.rank-column {
  width: 60px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 600;
}

.top-rank {
  color: var(--primary-color);
  font-size: 18px;
}

.user-column {
  flex: 1;
  min-width: 200px;
}

.user-info {
  display: flex;
  align-items: center;
}

.user-avatar {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  overflow: hidden;
  margin-right: 12px;
  background-color: #e5e7eb;
}

.user-avatar img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.user-avatar.placeholder {
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: var(--primary-color);
  color: white;
  font-weight: 600;
  font-size: 16px;
}

.user-name {
  font-weight: 500;
  color: var(--text-primary);
}

.predictions-column,
.correct-column,
.accuracy-column,
.points-column {
  width: 100px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.points-column {
  font-weight: 600;
  color: var(--primary-color);
}

.my-rank-info {
  margin-top: 30px;
  padding: 16px;
  background-color: #f0f7ff;
  border-radius: var(--border-radius);
  border: 1px solid #d0e3ff;
  display: flex;
  align-items: center;
}

.my-rank-label,
.my-points-label {
  font-weight: 500;
  color: var(--text-secondary);
  margin-right: 8px;
}

.my-rank-value,
.my-points-value {
  font-weight: 600;
  color: var(--primary-color);
  margin-right: 24px;
}

@media (max-width: 768px) {
  .content-with-sidebar {
    flex-direction: column;
  }

  .leaderboard-header,
  .leaderboard-row {
    font-size: 12px;
    padding: 10px 8px;
  }

  .rank-column {
    width: 40px;
  }

  .user-column {
    min-width: 120px;
  }

  .predictions-column,
  .correct-column,
  .accuracy-column,
  .points-column {
    width: 60px;
  }

  .user-avatar {
    width: 28px;
    height: 28px;
    margin-right: 8px;
  }

  .my-rank-info {
    flex-direction: column;
    align-items: flex-start;
  }

  .my-rank-value,
  .my-points-value {
    margin-bottom: 8px;
  }
}
</style>
