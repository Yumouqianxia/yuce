<template>
  <div class="prediction-list-container">
    <h1 class="page-title">最近比赛</h1>

    <!-- 导航菜单 -->
    <div class="prediction-nav">
      <div class="nav-item active">最近比赛</div>
      <router-link to="/prediction-history" class="nav-item">我的预测历史</router-link>
      <router-link to="/prediction-rules" class="nav-item">积分规则</router-link>
    </div>

    <div v-if="loading" class="loading-indicator">
      <div class="spinner"></div>
      <span>加载中...</span>
    </div>

    <div v-else-if="upcomingMatches.length === 0" class="empty-state">
      暂无预测比赛
    </div>

    <div v-else class="matches-grid">
      <div v-for="match in upcomingMatches" :key="match.id" class="match-card">
        <div class="match-header">
          <span class="match-title">{{ `${match.optionA || match.team_a} vs ${match.optionB || match.team_b}` }}</span>
          <span class="match-time">{{ formatDate(match.matchTime || match.start_time) }}</span>
        </div>

        <div class="match-teams">
          <div class="team">
            <TeamLogo :teamName="match.optionA || match.team_a" size="medium" />
            <span class="team-name">{{ match.optionA || match.team_a }}</span>
          </div>

          <div class="vs">VS</div>

          <div class="team">
            <TeamLogo :teamName="match.optionB || match.team_b" size="medium" />
            <span class="team-name">{{ match.optionB || match.team_b }}</span>
          </div>
        </div>

        <div class="prediction-form" v-if="!match.userPrediction || match.isEditing">
          <div class="prediction-options">
            <div
              class="prediction-option"
              :class="{ selected: match.tempPrediction && match.tempPrediction.value && getWinner(match) === 'A' }"
            >
              <span class="team-name">{{ match.optionA || match.team_a }}</span>
              <input
                type="number"
                :value="match.tempPrediction && match.tempPrediction.value ? match.tempPrediction.value.scoreA : 0"
                @input="updateScoreA(match, $event)"
                min="0"
                class="score-input"
                @click.stop
              >
            </div>

            <div
              class="prediction-option"
              :class="{ selected: match.tempPrediction && match.tempPrediction.value && getWinner(match) === 'B' }"
            >
              <span class="team-name">{{ match.optionB || match.team_b }}</span>
              <input
                type="number"
                :value="match.tempPrediction && match.tempPrediction.value ? match.tempPrediction.value.scoreB : 0"
                @input="updateScoreB(match, $event)"
                min="0"
                class="score-input"
                @click.stop
              >
            </div>
          </div>

          <button
            class="predict-button"
            @click="submitPrediction(match)"
            :disabled="!match.tempPrediction || !isPredictionValid(match)"
          >
            {{ match.userPrediction ? '更新预测' : '提交预测' }}
          </button>
        </div>

        <div class="prediction-info" v-else>
          <div class="prediction-header">您的预测</div>
          <div class="prediction-details">
            <div class="team" :class="{ 'predicted-winner': match.userPrediction && match.userPrediction.predictedWinner === 'A' }">
              <span class="team-name">{{ match.optionA || match.team_a }}</span>
              <span class="team-score">{{ match.userPrediction ? match.userPrediction.predictedScoreA : '' }}</span>
            </div>

            <div class="vs">VS</div>

            <div class="team" :class="{ 'predicted-winner': match.userPrediction && match.userPrediction.predictedWinner === 'B' }">
              <span class="team-name">{{ match.optionB || match.team_b }}</span>
              <span class="team-score">{{ match.userPrediction ? match.userPrediction.predictedScoreB : '' }}</span>
            </div>
          </div>
          <button class="edit-prediction-btn" @click="editPrediction(match)">
            修改预测
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { ref, onMounted, reactive } from 'vue'
import axios from 'axios'
import { useUserStore } from '@/stores/user'
import { ElMessage } from 'element-plus'
import TeamLogo from '@/components/teams/TeamLogo.vue'

export default {
  name: 'UpcomingMatches',
  components: {
    TeamLogo
  },
  setup() {
    const userStore = useUserStore()
    const upcomingMatches = ref([])
    const loading = ref(true)

    // 获取预测比赛
    const fetchUpcomingMatches = async () => {
      loading.value = true
      try {
        const response = await axios.get('/api/matches', {
          params: { status: 'not_started' },
          headers: userStore.isAuthenticated ? { Authorization: `Bearer ${userStore.token}` } : {}
        })

        let matches = []
        if (Array.isArray(response.data)) {
          matches = response.data
        } else if (response.data && Array.isArray(response.data.data)) {
          matches = response.data.data
        }

        console.log('获取到的比赛数据:', matches)

        // 按比赛时间排序
        matches = matches.sort((a, b) =>
          new Date(a.matchTime || a.start_time) - new Date(b.matchTime || b.start_time)
        )

        // 初始化每个比赛的临时预测状态
        matches.forEach(match => {
          // 确保 id 是数字类型
          if (match.id && typeof match.id === 'string') {
            match.id = parseInt(match.id)
          }

          // 始终确保tempPrediction已初始化
          match.tempPrediction = ref({
            winner: '',
            scoreA: 0,
            scoreB: 0
          })
        })

        // 获取用户已有的预测
        if (userStore.isAuthenticated) {
          try {
            const predictionsResponse = await axios.get('/api/predictions/my-predictions', {
              headers: { Authorization: `Bearer ${userStore.token}` }
            })

            let userPredictions = []
            if (Array.isArray(predictionsResponse.data)) {
              userPredictions = predictionsResponse.data
            } else if (predictionsResponse.data && Array.isArray(predictionsResponse.data.data)) {
              userPredictions = predictionsResponse.data.data
            }

            console.log('获取到的用户预测:', userPredictions)

            // 将用户预测与比赛关联
            matches.forEach(match => {
              const prediction = userPredictions.find(p => p.matchId === match.id)
              if (prediction) {
                match.userPrediction = prediction
              }
            })
          } catch (error) {
            console.error('获取用户预测失败:', error)
          }
        }

        upcomingMatches.value = matches
      } catch (error) {
        console.error('获取预测比赛失败:', error)
        upcomingMatches.value = []
        ElMessage({
          message: '获取比赛数据失败，请稍后再试',
          type: 'error',
          customClass: 'custom-message',
          offset: 80,
          duration: 3000
        })
      } finally {
        loading.value = false
      }
    }

    // 根据比分获取获胜方
    const getWinner = (match) => {
      if (!match.tempPrediction || !match.tempPrediction.value) return ''

      const scoreA = parseInt(match.tempPrediction.value.scoreA) || 0
      const scoreB = parseInt(match.tempPrediction.value.scoreB) || 0

      if (scoreA > scoreB) return 'A'
      if (scoreB > scoreA) return 'B'
      return '' // 平局或未设置比分
    }

    // 验证预测是否有效
    const isPredictionValid = (match) => {
      if (!match || !match.tempPrediction || !match.tempPrediction.value) return false

      const valueObj = match.tempPrediction.value;
      if (!valueObj.hasOwnProperty('scoreA') || !valueObj.hasOwnProperty('scoreB')) {
        return false;
      }

      // 确保比分是数字类型
      const scoreA = parseInt(valueObj.scoreA) || 0
      const scoreB = parseInt(valueObj.scoreB) || 0

      // 比分必须不相等，不允许平局
      return (
        scoreA >= 0 &&
        scoreB >= 0 &&
        scoreA !== scoreB
      )
    }

    // 提交预测
    const submitPrediction = async (match) => {
      if (!userStore.isAuthenticated) {
        ElMessage({
          message: '请先登录后再进行预测',
          type: 'warning',
          customClass: 'custom-message',
          offset: 80,
          duration: 3000
        })
        return
      }

      if (!match || !match.tempPrediction || !match.tempPrediction.value || !isPredictionValid(match)) {
        ElMessage({
          message: '请选择获胜方并填写有效的比分',
          type: 'warning',
          customClass: 'custom-message',
          offset: 80,
          duration: 3000
        })
        return
      }

      try {
        // 确保id存在
        if (!match.id) {
          ElMessage({
            message: '比赛ID无效',
            type: 'error',
            customClass: 'custom-message',
            offset: 80,
            duration: 3000
          })
          return
        }

        // 确保数据类型正确
        const scoreA = parseInt(match.tempPrediction.value.scoreA) || 0
        const scoreB = parseInt(match.tempPrediction.value.scoreB) || 0

        // 根据比分自动判断获胜方
        const winner = scoreA > scoreB ? 'A' : 'B'

        // 确保比分不相等
        if (scoreA === scoreB) {
          ElMessage({
            message: '比分不能相等，请设置有效的比分',
            type: 'warning',
            customClass: 'custom-message',
            offset: 80,
            duration: 3000
          })
          return
        }

        const predictionData = {
          matchId: parseInt(match.id),
          predictedWinner: winner,
          predictedScoreA: scoreA,
          predictedScoreB: scoreB
        }

        console.log('提交预测数据:', predictionData)

        let response

        // 区分创建新预测和更新已有预测
        if (match.isEditing && match.userPrediction && match.userPrediction.id) {
          // 更新已有预测
          console.log('更新已有预测:', match.userPrediction.id)
          response = await axios.put(`/api/predictions/${match.userPrediction.id}`, predictionData, {
            headers: { Authorization: `Bearer ${userStore.token}` }
          })
        } else {
          // 创建新预测
          console.log('创建新预测')
          response = await axios.post('/api/predictions', predictionData, {
            headers: { Authorization: `Bearer ${userStore.token}` }
          })
        }

        console.log('预测操作响应:', response.data)

        // 更新UI显示
        match.userPrediction = {
          ...predictionData,
          id: response.data.id || (match.userPrediction ? match.userPrediction.id : Date.now())
        }

        // 根据操作类型显示不同的成功消息
        const wasUpdate = match.isEditing && match.userPrediction && match.userPrediction.id

        // 重置编辑模式
        match.isEditing = false

        ElMessage({
          message: wasUpdate ? '预测更新成功' : '预测提交成功',
          type: 'success',
          customClass: 'custom-message',
          offset: 80,
          duration: 3000
        })
      } catch (error) {
        console.error('提交预测失败:', error)

        let errorMessage = '提交预测失败，请稍后再试'

        if (error.response) {
          console.error('错误状态:', error.response.status)
          console.error('错误数据:', error.response.data)

          // 根据错误状态码提供更具体的错误信息
          switch (error.response.status) {
            case 400:
              // 处理"已有预测"的情况
              if (error.response.data && error.response.data.message === "您已经对该比赛进行了预测") {
                console.log('检测到已有预测，尝试自动切换到编辑模式')

                // 尝试获取用户的所有预测
                try {
                  // 获取用户预测并找到当前比赛的预测
                  const predictionsResponse = await axios.get('/api/predictions/my-predictions', {
                    headers: { Authorization: `Bearer ${userStore.token}` }
                  })

                  let userPredictions = []
                  if (Array.isArray(predictionsResponse.data)) {
                    userPredictions = predictionsResponse.data
                  } else if (predictionsResponse.data && Array.isArray(predictionsResponse.data.data)) {
                    userPredictions = predictionsResponse.data.data
                  }

                  const existingPrediction = userPredictions.find(p => p.matchId === match.id)

                  if (existingPrediction) {
                    // 更新当前比赛的预测信息
                    match.userPrediction = existingPrediction

                    // 切换到编辑模式
                    match.isEditing = true

                    // 初始化临时预测数据
                    match.tempPrediction.value = {
                      winner: existingPrediction.predictedWinner,
                      scoreA: existingPrediction.predictedScoreA,
                      scoreB: existingPrediction.predictedScoreB
                    }

                    ElMessage({
                      message: '已切换到编辑模式，您可以修改预测',
                      type: 'info',
                      customClass: 'custom-message',
                      offset: 80,
                      duration: 3000
                    })

                    return // 防止显示错误消息
                  }
                } catch (retryError) {
                  console.error('获取预测失败:', retryError)
                  errorMessage = '无法获取已有预测，请刷新页面后重试'
                }
              } else {
                errorMessage = error.response.data?.message || '请求参数无效'
              }
              break
            case 401:
              errorMessage = '登录状态已失效，请重新登录'
              // 可以选择重定向到登录页面
              setTimeout(() => {
                window.location.href = '/login'
              }, 1500)
              break
            case 403:
              errorMessage = '您没有权限执行此操作'
              break
            case 404:
              errorMessage = '比赛不存在或已被删除'
              break
            case 409:
              errorMessage = '比赛已结束，无法修改预测'
              break
            case 422:
              errorMessage = '提交的预测数据无效'
              break
            default:
              // 如果后端返回了具体的错误消息，则使用后端的错误消息
              if (error.response.data && error.response.data.message) {
                errorMessage = error.response.data.message
              }
          }
        }

        ElMessage({
          message: errorMessage,
          type: 'error',
          customClass: 'custom-message',
          offset: 80,
          duration: 3000
        })
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

    // 更新scoreA
    const updateScoreA = (match, event) => {
      if (!userStore.isAuthenticated) {
        ElMessage({
          message: '请先登录后再进行预测',
          type: 'warning',
          customClass: 'custom-message',
          offset: 80,
          duration: 3000
        })
        return
      }

      // 确保tempPrediction已初始化
      if (!match.tempPrediction) {
        match.tempPrediction = ref({
          winner: '',
          scoreA: 0,
          scoreB: 0
        })
      }

      try {
        // 确保输入是数字
        const newValue = parseInt(event.target.value) || 0

        // 确保value对象存在
        if (!match.tempPrediction.value) {
          match.tempPrediction.value = {
            winner: '',
            scoreA: newValue,
            scoreB: 0
          }
        } else {
          // 更新响应式对象
          match.tempPrediction.value.scoreA = newValue
        }
      } catch (error) {
        console.error('更新scoreA失败:', error)
      }
    }

    // 更新scoreB
    const updateScoreB = (match, event) => {
      if (!userStore.isAuthenticated) {
        ElMessage({
          message: '请先登录后再进行预测',
          type: 'warning',
          customClass: 'custom-message',
          offset: 80,
          duration: 3000
        })
        return
      }

      // 确保tempPrediction已初始化
      if (!match.tempPrediction) {
        match.tempPrediction = ref({
          winner: '',
          scoreA: 0,
          scoreB: 0
        })
      }

      try {
        // 确保输入是数字
        const newValue = parseInt(event.target.value) || 0

        // 确保value对象存在
        if (!match.tempPrediction.value) {
          match.tempPrediction.value = {
            winner: '',
            scoreA: 0,
            scoreB: newValue
          }
        } else {
          // 更新响应式对象
          match.tempPrediction.value.scoreB = newValue
        }
      } catch (error) {
        console.error('更新scoreB失败:', error)
      }
    }

    onMounted(() => {
      fetchUpcomingMatches()
    })

    // 编辑预测
    const editPrediction = (match) => {
      if (!userStore.isAuthenticated) {
        ElMessage({
          message: '请先登录后再进行预测',
          type: 'warning',
          customClass: 'custom-message',
          offset: 80,
          duration: 3000
        })
        return
      }

      if (!match || !match.userPrediction) {
        ElMessage({
          message: '无法编辑预测',
          type: 'warning',
          customClass: 'custom-message',
          offset: 80,
          duration: 3000
        })
        return
      }

      // 初始化临时预测数据
      if (!match.tempPrediction) {
        match.tempPrediction = ref({
          winner: match.userPrediction.predictedWinner,
          scoreA: match.userPrediction.predictedScoreA,
          scoreB: match.userPrediction.predictedScoreB
        })
      } else {
        match.tempPrediction.value = {
          winner: match.userPrediction.predictedWinner,
          scoreA: match.userPrediction.predictedScoreA,
          scoreB: match.userPrediction.predictedScoreB
        }
      }

      // 切换到编辑模式
      match.isEditing = true
    }

    return {
      upcomingMatches,
      loading,
      getWinner,
      isPredictionValid,
      submitPrediction,
      formatDate,
      updateScoreA,
      updateScoreB,
      editPrediction
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
}

.matches-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 20px;
  margin-bottom: 40px;
}

.match-card {
  background-color: var(--bg-white);
  border-radius: var(--border-radius);
  padding: 20px;
  box-shadow: var(--shadow-sm);
  position: relative;
  cursor: pointer;
  transition: all 0.3s;
  border: 1px solid transparent;
}

.match-card:hover {
  box-shadow: var(--shadow-md);
  transform: translateY(-2px);
  border-color: var(--primary-light);
}

.match-header {
  display: flex;
  justify-content: space-between;
  margin-bottom: 20px;
}

.match-title {
  font-weight: 600;
  color: #333;
}

.match-time {
  color: #666;
  font-size: 14px;
}

.match-teams {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 25px;
}

.team {
  display: flex;
  flex-direction: column;
  align-items: center;
  width: 40%;
}

.team-name {
  margin-top: 10px;
  font-weight: 500;
  text-align: center;
}

.vs {
  font-weight: bold;
  color: #999;
}

.prediction-form {
  border-top: 1px solid #eee;
  padding-top: 20px;
}

.prediction-options {
  display: flex;
  justify-content: space-between;
  margin-bottom: 20px;
}

.prediction-option {
  display: flex;
  flex-direction: column;
  align-items: center;
  width: 45%;
  padding: 15px;
  border-radius: 8px;
  border: 1px solid #eee;
  cursor: pointer;
  transition: all 0.2s;
}

.prediction-option:hover {
  background-color: #f9f9f9;
}

.prediction-option.selected {
  background-color: rgba(58, 123, 213, 0.1);
  border-color: #3a7bd5;
}

.score-input {
  width: 60px;
  height: 40px;
  margin-top: 10px;
  text-align: center;
  font-size: 18px;
  border: 1px solid #ddd;
  border-radius: 4px;
}

.predict-button {
  display: block;
  width: 100%;
  padding: 12px;
  background-color: #3a7bd5;
  color: white;
  border: none;
  border-radius: 6px;
  font-weight: 500;
  cursor: pointer;
  transition: background-color 0.2s;
}

.predict-button:hover {
  background-color: #2c5aa0;
}

.predict-button:disabled {
  background-color: #a0cfff;
  cursor: not-allowed;
}

.prediction-info {
  border-top: 1px solid #eee;
  padding-top: 20px;
}

.prediction-header {
  text-align: center;
  font-weight: 500;
  color: #666;
  margin-bottom: 15px;
}

.prediction-details {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.predicted-winner {
  color: #3a7bd5;
  font-weight: 600;
}

.team-score {
  font-size: 20px;
  font-weight: 700;
  margin-top: 5px;
}

.edit-prediction-btn {
  display: block;
  width: 100%;
  margin-top: 15px;
  padding: 10px;
  background-color: #ecf5ff;
  color: #409EFF;
  border: 1px solid #d9ecff;
  border-radius: 6px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.edit-prediction-btn:hover {
  background-color: #409EFF;
  color: white;
  border-color: #409EFF;
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

  .matches-grid {
    grid-template-columns: 1fr;
  }
}
</style>
