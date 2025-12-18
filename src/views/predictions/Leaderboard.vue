<template>
  <div class="prediction-list-container">
    <h1 class="page-title">{{ pageTitle }}</h1>

    <!-- 导航菜单 -->
    <div class="prediction-nav">
      <router-link to="/upcoming-matches" class="nav-item">预测比赛</router-link>
      <router-link to="/prediction-history" class="nav-item">我的预测历史</router-link>
      <div class="nav-item active">积分排行榜</div>
      <router-link to="/prediction-rules" class="nav-item">积分规则</router-link>
    </div>

    <div class="content-with-sidebar">
      <!-- 侧边栏 -->
      <TournamentSidebar currentTab="leaderboard" />

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

      <div class="leaderboard-list">
        <div
          v-for="(user, index) in leaderboard"
          :key="user.id"
          class="leaderboard-item"
          :class="{
            'top-rank': index < 3,
            'current-user': userStore.isLoggedIn && (user.id === userStore.user.id || user.username === userStore.user.username)
          }"
        >
          <div class="rank-column">
            <div class="rank" :class="{ 'rank-1': index === 0, 'rank-2': index === 1, 'rank-3': index === 2 }">
              {{ index + 1 }}
            </div>
          </div>

          <div class="user-column">
            <div class="user-avatar">
              <span v-if="!user.avatar">{{ getUserInitial(user) }}</span>
              <img v-else :src="user.avatar" alt="用户头像">
            </div>
            <div class="user-info">
              <div class="username">{{ user.nickname || user.username }}</div>
              <div class="user-id" v-if="user.username">@{{ user.username }}</div>
            </div>
          </div>

          <div class="predictions-column">{{ user.total_predictions || user.totalPredictions || 0 }}</div>
          <div class="correct-column">{{ user.accurate_predictions || user.correctPredictions || 0 }}</div>
          <div class="accuracy-column">{{ calculateAccuracy(user) }}%</div>
          <div class="points-column">{{ user.total_points || user.points || user.totalPoints || 0 }}</div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { ref, onMounted, computed } from 'vue'
import { useRoute } from 'vue-router'
import { get } from '@/api/http'
import { useUserStore } from '@/stores/user'
import TournamentSidebar from '@/components/matches/TournamentSidebar.vue'

export default {
  name: 'Leaderboard',
  components: {
    TournamentSidebar
  },
  setup() {
    const userStore = useUserStore()
    const route = useRoute()
    const leaderboard = ref([])
    const loading = ref(true)

    // 获取路由参数
    const getTournamentType = () => route.meta.tournamentType || route.query.tournament || ''

    // 页面标题
    const pageTitle = computed(() => {
      const tournamentType = getTournamentType()
      if (!tournamentType) return '积分排行榜'

      const tournamentNames = {
        'spring': '2025KPL春季赛',
        'summer': '2025KPL夏季赛',
        'annual': '2025KPL年度总决赛',
        'challenger': '2025KPL挑战者杯'
      }

      return `${tournamentNames[tournamentType] || ''}积分排行榜`
    })

    // 获取排行榜数据
    const fetchLeaderboard = async () => {
      loading.value = true
      try {
        console.log('开始获取排行榜数据')
        const tournamentType = getTournamentType()
        const endpoint = tournamentType
          ? `/api/leaderboard?tournament=${tournamentType.toUpperCase()}`
          : '/api/leaderboard'

        console.log('请求端点:', endpoint)
        const data = await get(endpoint)
        console.log('获取到的排行榜数据:', data)

        let leaderboardData = []
        if (Array.isArray(data)) {
          leaderboardData = data
        } else if (data && Array.isArray(data.data)) {
          leaderboardData = data.data
        }

        // 处理数据字段名称和头像 URL
        leaderboardData = leaderboardData.map(user => {
          // 添加预测次数和正确预测数
          if (!user.total_predictions && !user.totalPredictions) {
            // 从后端获取用户的预测数据
            // 这里我们暂时设置为0，因为后端没有提供这些数据
            user.total_predictions = 0
            user.accurate_predictions = 0
          }

          // 处理积分字段
          if (user.totalPoints && !user.total_points) {
            user.total_points = user.totalPoints
          }

          // 处理头像 URL
          if (user.avatar && !user.avatar.startsWith('http')) {
            // 如果头像路径不是完整URL，使用相对路径
            const avatarFilename = user.avatar.split('/').pop()
            user.avatar = `/api/uploads/avatar/${avatarFilename}`
          }

          return user
        })

        console.log('处理后的排行榜数据:', leaderboardData)

        leaderboard.value = leaderboardData
      } catch (error) {
        console.error('获取排行榜失败:', error)
        // 使用模拟数据
        leaderboard.value = [
          { id: 1, username: 'user1', nickname: '用户一', totalPredictions: 20, correctPredictions: 15, points: 120 },
          { id: 2, username: 'user2', nickname: '用户二', totalPredictions: 18, correctPredictions: 12, points: 105 },
          { id: 3, username: 'user3', nickname: '用户三', totalPredictions: 22, correctPredictions: 10, points: 98 },
          { id: 4, username: 'user4', nickname: '用户四', totalPredictions: 15, correctPredictions: 8, points: 86 },
          { id: 5, username: 'user5', nickname: '用户五', totalPredictions: 12, correctPredictions: 7, points: 72 },
          { id: 6, username: 'user6', nickname: '用户六', totalPredictions: 10, correctPredictions: 6, points: 65 },
          { id: 7, username: 'user7', nickname: '用户七', totalPredictions: 8, correctPredictions: 5, points: 58 },
          { id: 8, username: 'user8', nickname: '用户八', totalPredictions: 6, correctPredictions: 4, points: 45 },
          { id: 9, username: 'user9', nickname: '用户九', totalPredictions: 5, correctPredictions: 3, points: 38 },
          { id: 10, username: 'user10', nickname: '用户十', totalPredictions: 4, correctPredictions: 2, points: 25 }
        ]
      } finally {
        loading.value = false
      }
    }

    // 计算准确率
    const calculateAccuracy = (user) => {
      const totalPredictions = user.total_predictions || user.totalPredictions || 0
      const correctPredictions = user.accurate_predictions || user.correctPredictions || 0

      if (totalPredictions === 0) {
        return 0
      }

      return Math.round((correctPredictions / totalPredictions) * 100)
    }

    // 获取用户名首字母（用于头像）
    const getUserInitial = (user) => {
      const name = user.nickname || user.username || ''
      return name.charAt(0).toUpperCase()
    }

    // 监听路由变化
    const watchRoute = () => {
      const tournamentType = getTournamentType()
      console.log('路由变化, 赛事类型:', tournamentType)
      fetchLeaderboard()
    }

    onMounted(() => {
      fetchLeaderboard()

      // 监听路由变化
      if (route) {
        route.meta.tournamentType && watchRoute()
        route.query.tournament && watchRoute()
      }
    })

    return {
      userStore,
      leaderboard,
      loading,
      getUserInitial,
      calculateAccuracy,
      pageTitle
    }
  }
}
</script>

<style scoped>
.prediction-list-container {
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 20px;
}

.content-with-sidebar {
  display: flex;
  gap: 20px;
  margin-top: 20px;
}

.main-content {
  flex: 1;
}

.page-title {
  font-size: 28px;
  font-weight: 600;
  margin-bottom: 30px;
  color: #303133;
  text-align: center;
}

.prediction-nav {
  display: flex;
  justify-content: center;
  margin-bottom: 30px;
  background-color: #fff;
  border-radius: 8px;
  padding: 15px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
}

.nav-item {
  padding: 10px 20px;
  margin: 0 10px;
  color: #606266;
  text-decoration: none;
  font-size: 16px;
  border-radius: 4px;
  transition: all 0.3s ease;
  cursor: pointer;
}

.nav-item:hover {
  color: #409EFF;
  background-color: #ecf5ff;
}

.nav-item.active {
  color: #409EFF;
  font-weight: 500;
  background-color: #ecf5ff;
}

.loading-indicator {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 50px 0;
  color: #666;
}

.spinner {
  width: 40px;
  height: 40px;
  border: 4px solid rgba(0, 0, 0, 0.1);
  border-radius: 50%;
  border-top-color: #3a7bd5;
  animation: spin 1s ease-in-out infinite;
  margin-bottom: 15px;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.empty-state {
  text-align: center;
  padding: 50px 0;
  color: #666;
  background-color: #f9f9f9;
  border-radius: 8px;
  margin-bottom: 30px;
}

.leaderboard-content {
  background-color: var(--bg-white);
  border-radius: var(--border-radius);
  box-shadow: var(--shadow-sm);
  margin-bottom: 40px;
  overflow: hidden;
  border: 1px solid transparent;
  transition: all 0.3s;
}

.leaderboard-content:hover {
  box-shadow: var(--shadow-md);
  border-color: var(--primary-light);
}

.leaderboard-header {
  display: flex;
  padding: 15px 20px;
  background-color: #f9f9f9;
  font-weight: 600;
  color: #333;
  border-bottom: 1px solid #eee;
}

.leaderboard-list {
  display: flex;
  flex-direction: column;
}

.leaderboard-item {
  display: flex;
  padding: 15px 20px;
  border-bottom: 1px solid #f5f5f5;
  transition: background-color 0.2s;
}

.leaderboard-item:last-child {
  border-bottom: none;
}

.leaderboard-item:hover {
  background-color: #f9f9f9;
}

.top-rank {
  background-color: rgba(58, 123, 213, 0.05);
}

.current-user {
  background-color: rgba(103, 194, 58, 0.05);
}

.rank-column {
  width: 80px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.user-column {
  flex: 1;
  display: flex;
  align-items: center;
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
  color: #3a7bd5;
}

.accuracy-column {
  font-weight: 600;
  color: #67c23a;
}

.rank {
  width: 36px;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  background-color: #f0f0f0;
  color: #333;
  font-weight: bold;
}

.rank-1 {
  background: linear-gradient(to bottom right, #FFD700, #FFA500);
  color: white;
}

.rank-2 {
  background: linear-gradient(to bottom right, #C0C0C0, #A0A0A0);
  color: white;
}

.rank-3 {
  background: linear-gradient(to bottom right, #CD7F32, #A0522D);
  color: white;
}

.user-avatar {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background-color: #3a7bd5;
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 600;
  margin-right: 15px;
  overflow: hidden;
}

.user-avatar img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.user-info {
  display: flex;
  flex-direction: column;
}

.username {
  font-weight: 500;
  color: #333;
}

.user-id {
  font-size: 12px;
  color: #999;
}

/* 响应式调整 */
@media (max-width: 768px) {
  .prediction-nav {
    flex-direction: column;
    align-items: center;
  }

  .nav-item {
    margin: 5px 0;
    width: 100%;
    text-align: center;
  }

  .page-title {
    font-size: 24px;
    margin-bottom: 20px;
  }

  .leaderboard-header {
    display: none;
  }

  .leaderboard-item {
    flex-wrap: wrap;
  }

  .rank-column {
    width: 60px;
  }

  .user-column {
    flex: 1;
  }

  .predictions-column,
  .correct-column,
  .accuracy-column,
  .points-column {
    width: 33.33%;
    padding: 10px 0 0 60px;
  }
}
</style>
