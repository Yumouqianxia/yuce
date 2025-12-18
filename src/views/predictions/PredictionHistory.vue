<template>
  <div class="prediction-list-container">
    <h1 class="page-title">我的预测历史</h1>
    <button v-if="userStore.isAdmin" @click="showDebugInfo" class="debug-button">显示调试信息</button>

    <!-- 导航菜单 -->
    <div class="prediction-nav">
      <router-link to="/upcoming-matches" class="nav-item">最近比赛</router-link>
      <div class="nav-item active">我的预测历史</div>
      <router-link to="/prediction-rules" class="nav-item">积分规则</router-link>
    </div>

    <div class="content-with-sidebar">
      <!-- 侧边栏 -->
      <TournamentSidebar currentTab="history" />

      <div class="main-content">
        <div v-if="!userStore.isAuthenticated" class="login-prompt">
          <p>请登录后查看您的预测历史</p>
          <router-link to="/login" class="login-button">立即登录</router-link>
        </div>

        <template v-else>
      <div v-if="loading" class="loading-indicator">
        <div class="spinner"></div>
        <span>加载中...</span>
      </div>

      <div v-else-if="error" class="error-state">
        <div class="error-icon">⚠️</div>
        <p>获取预测历史失败</p>
        <button class="retry-button" @click="fetchPredictions">重试</button>
      </div>

      <div v-if="showDebug" class="debug-info">
        <h3>调试信息</h3>
        <pre>{{ JSON.stringify(predictions, null, 2) }}</pre>
      </div>

      <div v-else-if="predictions.length === 0" class="empty-state">
        您还没有进行过预测
      </div>

      <div v-else class="predictions-list">
        <div class="predictions-grid">
          <div v-for="prediction in predictions" :key="prediction.id" class="prediction-card">
            <div class="match-info">
              <div class="match-teams">
                <div class="team">
                  <TeamLogo :teamName="prediction.match.optionA || prediction.match.team_a || '未知队伍'" size="small" />
                  <span class="team-name">{{ prediction.match.optionA || prediction.match.team_a || '未知队伍' }}</span>
                </div>

                <div class="vs">VS</div>

                <div class="team">
                  <TeamLogo :teamName="prediction.match.optionB || prediction.match.team_b || '未知队伍'" size="small" />
                  <span class="team-name">{{ prediction.match.optionB || prediction.match.team_b || '未知队伍' }}</span>
                </div>
              </div>

              <div class="match-time">
                {{ formatDate(prediction.match.matchTime || prediction.match.start_time) }}
              </div>
            </div>

            <div class="prediction-details">
              <div class="prediction-header">您的预测</div>
              <div class="prediction-result">
                <div class="prediction-team-container">
                  <div class="prediction-team" :class="{ 'predicted-winner': getPredictedWinner(prediction) === 'A' }">
                    <TeamLogo :teamName="prediction.match.optionA || prediction.match.team_a || '未知队伍'" size="small" />
                    <span class="team-name">{{ prediction.match.optionA || prediction.match.team_a || '未知队伍' }}</span>
                    <span class="team-score">{{ prediction.predictedScoreA || 0 }}</span>
                  </div>

                  <div class="vs-container">
                    <div class="vs">VS</div>
                  </div>

                  <div class="prediction-team" :class="{ 'predicted-winner': getPredictedWinner(prediction) === 'B' }">
                    <TeamLogo :teamName="prediction.match.optionB || prediction.match.team_b || '未知队伍'" size="small" />
                    <span class="team-name">{{ prediction.match.optionB || prediction.match.team_b || '未知队伍' }}</span>
                    <span class="team-score">{{ prediction.predictedScoreB || 0 }}</span>
                  </div>
                </div>
              </div>
            </div>

            <div class="prediction-result-row" v-if="prediction.match.status === 'completed'">
              <div class="result-item actual-result">
                <div class="result-label">实际结果:</div>
                <div class="result-content">
                  <span class="winner">{{ prediction.match.winner === 'A' ? (prediction.match.optionA || prediction.match.team_a || '未知队伍') : (prediction.match.optionB || prediction.match.team_b || '未知队伍') }}</span>
                  {{ prediction.match.scoreA || 0 }} - {{ prediction.match.scoreB || 0 }}
                </div>
              </div>

              <div class="result-item points" v-if="prediction.isVerified">
                <div class="result-label">获得积分:</div>
                <div class="result-content">
                  <span>{{ prediction.earnedPoints || 0 }}</span>
                </div>
              </div>

              <div class="result-item">
                <div class="result-label">预测结果:</div>
                <div class="result-content">
                  <div class="status-badge" v-if="prediction.isVerified" :class="{ correct: prediction.isCorrect }">
                    {{ prediction.isCorrect ? '预测正确' : '预测错误' }}
                  </div>
                </div>
              </div>
            </div>

            <div class="prediction-result-row" v-else>
              <div class="result-item">
                <div class="result-label">比赛状态:</div>
                <div class="result-content">
                  <span class="status-badge pending">等待比赛结果</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
        </template>
      </div>
    </div>
  </div>
</template>

<script>
import { ref, onMounted } from 'vue'
import axios from 'axios'
import { useUserStore } from '@/stores/user'
import TeamLogo from '@/components/teams/TeamLogo.vue'
import TournamentSidebar from '@/components/matches/TournamentSidebar.vue'
import { ElMessage } from 'element-plus'

export default {
  name: 'PredictionHistory',
  components: {
    TeamLogo,
    TournamentSidebar
  },

  setup() {
    const userStore = useUserStore()
    const predictions = ref([])
    const loading = ref(true)
    const error = ref(false)
    const showDebug = ref(false)

    const showDebugInfo = () => {
      showDebug.value = !showDebug.value
    }

    // 获取用户预测历史
    const fetchPredictions = async () => {
      if (!userStore.isAuthenticated) {
        loading.value = false
        return
      }

      loading.value = true
      error.value = false
      predictions.value = [] // 清空现有数据

      try {
        console.log('开始获取预测历史...')
        const response = await axios.get('/api/predictions/my-predictions', {
          headers: { Authorization: `Bearer ${userStore.token}` }
        })

        console.log('获取到的原始数据:', response.data)

        let userPredictions = []
        if (Array.isArray(response.data)) {
          userPredictions = response.data
        } else if (response.data && Array.isArray(response.data.data)) {
          userPredictions = response.data.data
        } else if (response.data && typeof response.data === 'object') {
          // 尝试其他可能的数据结构
          console.log('尝试解析数据结构:', Object.keys(response.data))
          if (response.data.predictions && Array.isArray(response.data.predictions)) {
            userPredictions = response.data.predictions
          } else if (response.data.results && Array.isArray(response.data.results)) {
            userPredictions = response.data.results
          } else if (response.data.data && Array.isArray(response.data.data)) {
            userPredictions = response.data.data
          }
        }

        console.log('解析后的预测数据:', userPredictions)

        if (userPredictions.length === 0) {
          console.log('没有找到预测数据')
          loading.value = false
          return
        }

        // 获取每个预测对应的比赛信息
        const predictionsWithMatches = await Promise.all(
          userPredictions.map(async (prediction) => {
            try {
              console.log(`获取比赛信息 (ID: ${prediction.matchId})...`)
              const matchResponse = await axios.get(`/api/matches/${prediction.matchId}`, {
                headers: { Authorization: `Bearer ${userStore.token}` }
              })

              console.log(`比赛信息 (ID: ${prediction.matchId}):`, matchResponse.data)

              // 确保比赛数据中有队伍信息
              // 处理后端返回的嵌套数据结构
              const matchData = matchResponse.data.data || matchResponse.data
              console.log('处理后的比赛数据:', matchData)

              const teamA = matchData.optionA || matchData.team_a || '未知队伍A'
              const teamB = matchData.optionB || matchData.team_b || '未知队伍B'

              return {
                ...prediction,
                match: {
                  ...matchData,
                  optionA: teamA,
                  optionB: teamB
                }
              }
            } catch (error) {
              console.error(`获取比赛信息失败 (ID: ${prediction.matchId}):`, error)
              return {
                ...prediction,
                match: {
                  id: prediction.matchId,
                  optionA: '未知队伍A',
                  optionB: '未知队伍B',
                  status: 'unknown',
                  matchTime: new Date().toISOString()
                }
              }
            }
          })
        )

        console.log('处理后的预测数据:', predictionsWithMatches)

        // 按比赛时间倒序排序
        predictions.value = predictionsWithMatches.sort((a, b) => {
          const dateA = new Date(a.match.matchTime || a.match.start_time || 0)
          const dateB = new Date(b.match.matchTime || b.match.start_time || 0)
          return dateB - dateA
        })

        console.log('最终处理后的预测数据:', predictions.value)
      } catch (err) {
        console.error('获取预测历史失败:', err)
        predictions.value = []
        error.value = true
        // 显示错误提示
        ElMessage.error('获取预测历史失败，请稍后再试')
      } finally {
        loading.value = false
      }
    }

    // 格式化日期时间
    const formatDate = (dateString) => {
      if (!dateString) return ''

      const date = new Date(dateString)
      const year = date.getFullYear()
      const month = String(date.getMonth() + 1).padStart(2, '0')
      const day = String(date.getDate()).padStart(2, '0')
      const hours = String(date.getHours()).padStart(2, '0')
      const minutes = String(date.getMinutes()).padStart(2, '0')

      return `${year}-${month}-${day} ${hours}:${minutes}`
    }

    // 根据比分获取预测的获胜方
    const getPredictedWinner = (prediction) => {
      if (!prediction) return ''

      const scoreA = parseInt(prediction.predictedScoreA) || 0
      const scoreB = parseInt(prediction.predictedScoreB) || 0

      if (scoreA > scoreB) return 'A'
      if (scoreB > scoreA) return 'B'
      return '' // 平局或未设置比分
    }

    onMounted(() => {
      fetchPredictions()
    })


    return {
      userStore,
      predictions,
      loading,
      error,
      showDebug,
      formatDate,
      fetchPredictions,
      showDebugInfo,
      getPredictedWinner
    }
  }
}
</script>

<style scoped>
.prediction-list-container {
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 20px;
  position: relative;
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

.login-prompt {
  text-align: center;
  padding: 50px 0;
  background-color: #f9f9f9;
  border-radius: 8px;
  margin-bottom: 30px;
}

.login-prompt p {
  margin-bottom: 20px;
  color: #666;
}

.login-button {
  display: inline-block;
  padding: 12px 24px;
  background-color: #3a7bd5;
  color: white;
  border-radius: 6px;
  text-decoration: none;
  font-weight: 500;
  transition: background-color 0.2s;
}

.login-button:hover {
  background-color: #2c5aa0;
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

.empty-state, .error-state {
  text-align: center;
  padding: 50px 0;
  color: #666;
  background-color: #f9f9f9;
  border-radius: 8px;
  margin-bottom: 30px;
}

.error-state {
  background-color: #fff2f0;
}

.error-icon {
  font-size: 32px;
  margin-bottom: 15px;
}

.retry-button {
  margin-top: 15px;
  padding: 10px 20px;
  background-color: #3a7bd5;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-weight: 500;
  transition: background-color 0.2s;
}

.retry-button:hover {
  background-color: #2c5aa0;
}

.debug-button {
  position: absolute;
  top: 20px;
  right: 20px;
  padding: 5px 10px;
  background-color: #f56c6c;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 12px;
}

.debug-info {
  margin: 20px 0;
  padding: 15px;
  background-color: #f8f8f8;
  border: 1px solid #ddd;
  border-radius: 4px;
  overflow: auto;
}

.debug-info pre {
  white-space: pre-wrap;
  font-family: monospace;
  font-size: 12px;
}

.predictions-list {
  margin-bottom: 40px;
}

.predictions-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 20px;
  margin-bottom: 30px;
  max-width: 1400px;
  margin-left: auto;
  margin-right: auto;
}

.prediction-card {
  background-color: var(--bg-white);
  border-radius: var(--border-radius);
  padding: 20px;
  box-shadow: var(--shadow-sm);
  position: relative;
  cursor: pointer;
  transition: all 0.3s;
  border: 1px solid transparent;
  height: 100%;
  min-height: 350px;
  display: flex;
  flex-direction: column;
}

.prediction-card:hover {
  box-shadow: var(--shadow-md);
  transform: translateY(-2px);
  border-color: var(--primary-light);
}

.match-info {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  flex: 0 0 auto;
}

.match-teams {
  display: flex;
  align-items: center;
}

.team {
  display: flex;
  flex-direction: column;
  align-items: center;
  margin: 0 15px;
}



.team-name {
  margin-top: 5px;
  font-weight: 500;
  text-align: center;
}

.vs {
  font-weight: bold;
  color: #999;
  margin: 0 10px;
}

.match-time {
  color: #666;
  font-size: 14px;
}

.prediction-details {
  border-top: 1px solid #eee;
  padding-top: 15px;
  margin-bottom: 15px;
  flex: 1 0 auto;
}

.prediction-header {
  text-align: center;
  font-weight: 500;
  color: #666;
  margin-bottom: 15px;
}

.prediction-result {
  margin-top: 15px;
}

.prediction-team-container {
  display: flex;
  justify-content: space-between;
  align-items: center;
  width: 100%;
  max-width: 600px;
  margin: 0 auto;
}

.prediction-team {
  display: flex;
  flex-direction: column;
  align-items: center;
  width: 40%;
  padding: 10px;
  border-radius: 8px;
  transition: background-color 0.2s;
}

.vs-container {
  display: flex;
  justify-content: center;
  align-items: center;
  width: 20%;
}

.prediction-result-row {
  display: flex;
  justify-content: space-around;
  align-items: center;
  padding-top: 15px;
  border-top: 1px dashed #eee;
  flex-wrap: wrap;
  gap: 15px;
  margin-top: auto;
}

.result-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  min-width: 120px;
  padding: 10px;
}

.result-label {
  font-size: 14px;
  color: #666;
  margin-bottom: 5px;
}

.result-content {
  font-weight: 500;
  text-align: center;
}

.predicted-winner {
  color: #3a7bd5;
  font-weight: 600;
  background-color: rgba(58, 123, 213, 0.1);
  border: 1px solid rgba(58, 123, 213, 0.3);
  border-radius: 8px;
}

.team-score {
  font-size: 20px;
  font-weight: 700;
  margin-top: 5px;
}

.actual-result .result-content {
  font-size: 16px;
}

.winner {
  font-weight: bold;
  color: #67C23A;
}

.points .result-content {
  font-weight: bold;
  color: #E6A23C;
  font-size: 16px;
}

.status-badge {
  display: inline-block;
  padding: 6px 12px;
  border-radius: 4px;
  background-color: #F56C6C;
  color: white;
  font-size: 14px;
}

.status-badge.correct {
  background-color: #67C23A;
}

.status-badge.pending {
  background-color: #E6A23C;
  color: white;
}

@media (min-width: 1600px) {
  .predictions-grid {
    grid-template-columns: repeat(3, 1fr);
  }
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

  .predictions-grid {
    grid-template-columns: 1fr;
  }

  .match-info {
    flex-direction: column;
  }

  .match-teams {
    margin-bottom: 10px;
  }

  .prediction-result-row {
    flex-direction: column;
    gap: 20px;
  }

  .result-item {
    width: 100%;
    padding: 5px 0;
  }

  .prediction-team-container {
    flex-direction: column;
    gap: 15px;
  }

  .prediction-team, .vs-container {
    width: 100%;
  }

  .vs-container {
    margin: 5px 0;
  }
}
</style>
