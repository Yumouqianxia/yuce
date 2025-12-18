<template>
  <div class="predictions-container">
    <h1>比赛预测</h1>

    <div class="tabs">
      <div
        class="tab"
        :class="{ active: activeTab === 'upcoming' }"
        @click="activeTab = 'upcoming'; handleTabChange()"
      >
        最近比赛
      </div>
      <div
        class="tab"
        :class="{ active: activeTab === 'history' }"
        @click="activeTab = 'history'; handleTabChange()"
      >
        我的预测历史
      </div>
      <div
        class="tab"
        :class="{ active: activeTab === 'leaderboard' }"
        @click="activeTab = 'leaderboard'; handleTabChange()"
      >
        积分排行榜
      </div>
      <div
        class="tab"
        :class="{ active: activeTab === 'rules' }"
        @click="activeTab = 'rules'; handleTabChange()"
      >
        积分规则
      </div>
    </div>

    <!-- 最近比赛 -->
    <div v-if="activeTab === 'upcoming'" class="tab-content">
      <div v-if="loading" class="loading">加载中...</div>

      <div v-else-if="upcomingMatches.length === 0" class="empty-state">
        暂无最近比赛
      </div>

      <div v-else class="matches-list">
        <div v-for="match in upcomingMatches" :key="match.id" class="match-card">
          <div class="match-header">
            <span class="match-title">{{ match.title || `${match.optionA} vs ${match.optionB}` }}</span>
            <span class="match-time">{{ formatDate(match.matchTime) }}</span>
          </div>

          <!-- 如果用户已经预测过这场比赛，显示已有预测 -->
          <div v-if="match.userPrediction && !isEditingPrediction(match.id)" class="user-prediction">
            <div class="prediction-label">您的预测：</div>
            <div class="prediction-teams">
              <div class="team" :class="{ 'winner': match.userPrediction.predictedWinner === 'A' }">
                <TeamLogo :teamName="match.optionA" size="medium" />
                <span class="team-name">{{ match.optionA }}</span>
                <span class="team-score">{{ match.userPrediction.predictedScoreA }}</span>
              </div>

              <div class="vs">VS</div>

              <div class="team" :class="{ 'winner': match.userPrediction.predictedWinner === 'B' }">
                <TeamLogo :teamName="match.optionB" size="medium" />
                <span class="team-name">{{ match.optionB }}</span>
                <span class="team-score">{{ match.userPrediction.predictedScoreB }}</span>
              </div>
            </div>
            <button class="edit-prediction-btn" @click="editPrediction(match.userPrediction)">
              修改预测
            </button>
          </div>

          <!-- 如果用户没有预测过这场比赛或正在编辑预测，显示预测表单 -->
          <div v-else>
            <div class="match-teams">
              <div class="team team-a" :class="{ selected: currentPrediction.predictedWinner === 'A' || (currentPrediction.matchId === match.id && currentPrediction.predictedScoreA > currentPrediction.predictedScoreB) }" @click="selectWinner('A')">
                <TeamLogo :teamName="match.optionA" size="medium" />
                <span class="team-name">{{ match.optionA }}</span>
                <input
                  type="number"
                  v-model.number="currentPrediction.matchId === match.id ? currentPrediction.predictedScoreA : tempScoreA"
                  min="0"
                  class="score-input"
                  @input="updatePrediction(match, 'A')"
                >
              </div>

              <div class="vs">VS</div>

              <div class="team team-b" :class="{ selected: currentPrediction.predictedWinner === 'B' || (currentPrediction.matchId === match.id && currentPrediction.predictedScoreB > currentPrediction.predictedScoreA) }" @click="selectWinner('B')">
                <TeamLogo :teamName="match.optionB" size="medium" />
                <span class="team-name">{{ match.optionB }}</span>
                <input
                  type="number"
                  v-model.number="currentPrediction.matchId === match.id ? currentPrediction.predictedScoreB : tempScoreB"
                  min="0"
                  class="score-input"
                  @input="updatePrediction(match, 'B')"
                >
              </div>
            </div>

            <div class="match-actions">
              <button
                class="predict-button"
                @click="submitPrediction(match.id)"
                :disabled="!isPredictionValid"
              >
                {{ isEditingPrediction(match.id) && match.userPrediction ? '更新预测' : '提交预测' }}
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 预测历史 -->
    <div v-if="activeTab === 'history'" class="tab-content">
      <div v-if="loading" class="loading">加载中...</div>

      <div v-else-if="myPredictions.length === 0" class="empty-state">
        您还没有进行过预测
      </div>

      <div v-else class="predictions-list">
        <div v-for="prediction in myPredictions" :key="prediction.id" class="prediction-card">
          <div v-if="!prediction.match" class="prediction-loading">
            <p>正在加载比赛信息...</p>
            <p>请刷新页面查看完整信息</p>
          </div>

          <div v-else>
            <div class="prediction-header">
              <span class="match-title">{{ prediction.match.title || '比赛信息加载中' }}</span>
              <span class="match-time">{{ formatDate(prediction.match.matchTime) }}</span>
            </div>

            <div class="prediction-teams">
              <div class="team" :class="{ 'winner': prediction.predictedWinner === 'A' }">
                <TeamLogo :teamName="prediction.match.optionA" size="medium" />
                <span class="team-name">{{ prediction.match.optionA || '队伍A' }}</span>
                <span class="team-score">{{ prediction.predictedScoreA }}</span>
              </div>

              <div class="vs">VS</div>

              <div class="team" :class="{ 'winner': prediction.predictedWinner === 'B' }">
                <TeamLogo :teamName="prediction.match.optionB" size="medium" />
                <span class="team-name">{{ prediction.match.optionB || '队伍B' }}</span>
                <span class="team-score">{{ prediction.predictedScoreB }}</span>
              </div>
            </div>

            <div class="prediction-result" v-if="prediction.match.status === 'completed'">
              <div class="actual-result">
                实际结果:
                <span class="winner">{{ prediction.match.winner === 'A' ? prediction.match.optionA : prediction.match.optionB }}</span>
                {{ prediction.match.scoreA }} - {{ prediction.match.scoreB }}
              </div>

              <div class="result-info-container">
                <div class="points" v-if="prediction.isVerified">
                  获得积分: <span>{{ prediction.earnedPoints }}</span>
                </div>

                <div class="status" v-if="prediction.isVerified" :class="{ correct: prediction.isCorrect }">
                  {{ prediction.isCorrect ? '预测正确' : '预测错误' }}
                </div>
              </div>
            </div>

            <div class="prediction-status" v-else>
              <div class="status-text">
                比赛状态: {{ getMatchStatusText(prediction.match.status) }}
              </div>

              <!-- 如果比赛未开始，显示修改预测按钮 -->
              <button
                v-if="prediction.match.status === 'not_started'"
                class="edit-prediction-btn"
                @click="editPrediction(prediction)"
              >
                修改预测
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 积分排行榜 -->
    <div v-if="activeTab === 'leaderboard'" class="tab-content">
      <div v-if="loading" class="loading">加载中...</div>

      <div v-else-if="leaderboard.length === 0" class="empty-state">
        暂无排行数据
      </div>

      <div v-else class="leaderboard">
        <div class="leaderboard-header">
          <div class="rank">排名</div>
          <div class="user">用户</div>
          <div class="points">积分</div>
        </div>

        <div v-for="(user, index) in leaderboard" :key="user.userId" class="leaderboard-item">
          <div class="rank">{{ index + 1 }}</div>
          <div class="user">{{ user.nickname || user.username }}</div>
          <div class="points">{{ user.totalPoints }}</div>
        </div>
      </div>

      <div class="my-points">
        我的总积分: <span>{{ myTotalPoints }}</span>
      </div>
    </div>

    <!-- 积分规则 -->
    <div v-if="activeTab === 'rules'" class="tab-content">
      <div class="rules-container">
        <h2>预测积分规则</h2>
        <div class="rule-item">
          <div class="rule-title">预测胜利队伍和比分全对</div>
          <div class="rule-points">+5分</div>
          <div class="rule-description">如果您预测的获胜队伍和比分都正确，将获得5分。</div>
        </div>

        <div class="rule-item">
          <div class="rule-title">预测胜利队伍但比分错误</div>
          <div class="rule-points">+3分</div>
          <div class="rule-description">如果您预测的获胜队伍正确，但比分预测错误，将获得3分。</div>
        </div>

        <div class="rule-item">
          <div class="rule-title">预测错误队伍但比分正确</div>
          <div class="rule-points">+1分</div>
          <div class="rule-description">如果您预测的获胜队伍错误，但比分预测正确，将获得1分。</div>
        </div>

        <div class="rule-item">
          <div class="rule-title">全部预测错误</div>
          <div class="rule-points">0分</div>
          <div class="rule-description">如果您预测的获胜队伍和比分都错误，将不获得积分。</div>
        </div>

        <div class="rule-note">
          <p>注意：比赛结束后，系统将自动验证预测并计算积分。您可以在“我的预测历史”中查看您的预测结果和获得的积分。</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { ref, reactive, computed, onMounted, onUnmounted, watch } from 'vue'
import axios from 'axios'
import { useUserStore } from '@/stores/user'
import { ElMessage } from 'element-plus'
import TeamLogo from '@/components/teams/TeamLogo.vue'

export default {
  components: {
    TeamLogo
  },
  setup() {
    const userStore = useUserStore()
    const activeTab = ref('upcoming')
    const loading = ref(false)
    const upcomingMatches = ref([])
    const myPredictions = ref([])
    const leaderboard = ref([])
    const myTotalPoints = ref(0)

    // 添加临时变量用于非当前选中比赛的输入
    const tempScoreA = ref(0)
    const tempScoreB = ref(0)

    const currentPrediction = reactive({
      matchId: null,
      predictedWinner: '',
      predictedScoreA: 0,
      predictedScoreB: 0
    })

    // 更新预测方法
    const updatePrediction = (match, team) => {
      // 如果是当前选中的比赛
      if (currentPrediction.matchId === match.id) {
        // 根据比分自动判断获胜方
        if (currentPrediction.predictedScoreA > currentPrediction.predictedScoreB) {
          currentPrediction.predictedWinner = 'A';
        } else if (currentPrediction.predictedScoreB > currentPrediction.predictedScoreA) {
          currentPrediction.predictedWinner = 'B';
        }
      } else {
        // 如果不是当前选中的比赛，则将其设置为当前选中的比赛
        currentPrediction.matchId = match.id;

        if (team === 'A') {
          currentPrediction.predictedScoreA = tempScoreA.value;
          currentPrediction.predictedScoreB = 0;
          if (tempScoreA.value > 0) {
            currentPrediction.predictedWinner = 'A';
          }
        } else {
          currentPrediction.predictedScoreB = tempScoreB.value;
          currentPrediction.predictedScoreA = 0;
          if (tempScoreB.value > 0) {
            currentPrediction.predictedWinner = 'B';
          }
        }
      }
    }

    const isPredictionValid = computed(() => {
      return (
        currentPrediction.predictedWinner === 'A' ||
        currentPrediction.predictedWinner === 'B'
      ) &&
      currentPrediction.predictedScoreA >= 0 &&
      currentPrediction.predictedScoreB >= 0
    })

    // 获取最近比赛和用户预测
    const fetchUpcomingMatches = async () => {
      // 如果组件已卸载，不执行请求
      if (!isComponentMounted.value) return

      // 创建新的AbortController
      abortController = new AbortController()
      const signal = abortController.signal

      loading.value = true
      try {
        console.log('开始获取最近比赛...')

        // 获取所有未开始的比赛
        const matchesResponse = await axios.get('/api/matches', {
          headers: { Authorization: `Bearer ${userStore.token}` },
          signal
        })

        console.log('获取比赛列表成功:', matchesResponse)

        // 如果组件已卸载，不处理数据
        if (!isComponentMounted.value) return

        // 处理返回的比赛数据
        let fetchedMatches = []

        if (Array.isArray(matchesResponse.data)) {
          fetchedMatches = matchesResponse.data
        } else if (matchesResponse.data && typeof matchesResponse.data === 'object') {
          if (matchesResponse.data.data && Array.isArray(matchesResponse.data.data)) {
            fetchedMatches = matchesResponse.data.data
          }
        }

        // 过滤出未开始的比赛
        const notStartedMatches = fetchedMatches
          .filter(match => match.status === 'not_started')
          .map(match => ({
            ...match,
            matchType: match.matchType || 'regular',
            series: match.series || 'BO3',
            userPrediction: null // 初始化用户预测字段
          }))

        console.log('未开始的比赛列表:', notStartedMatches)

        // 获取用户的预测历史
        const predictionsResponse = await axios.get('/api/predictions/my-predictions', {
          headers: { Authorization: `Bearer ${userStore.token}` },
          signal
        })

        console.log('获取预测历史成功:', predictionsResponse)

        // 如果组件已卸载，不处理数据
        if (!isComponentMounted.value) return

        // 处理返回的预测数据
        let userPredictions = []

        if (Array.isArray(predictionsResponse.data)) {
          userPredictions = predictionsResponse.data
        } else if (predictionsResponse.data && typeof predictionsResponse.data === 'object') {
          if (predictionsResponse.data.data && Array.isArray(predictionsResponse.data.data)) {
            userPredictions = predictionsResponse.data.data
          }
        }

        console.log('用户预测历史:', userPredictions)

        // 将用户预测与比赛关联
        notStartedMatches.forEach(match => {
          // 查找该比赛的用户预测
          const prediction = userPredictions.find(p => p.matchId === match.id)
          if (prediction) {
            match.userPrediction = prediction
          }
        })

        // 更新比赛列表
        upcomingMatches.value = notStartedMatches
        console.log('处理后的最近比赛列表:', upcomingMatches.value)
      } catch (error) {
        console.error('获取比赛或预测失败:', error)
        if (error.response) {
          console.error('错误状态:', error.response.status)
          console.error('错误数据:', JSON.stringify(error.response.data, null, 2))
        }
        ElMessage({
          message: '获取比赛列表失败',
          type: 'error',
          customClass: 'custom-message',
          offset: 80,
          duration: 3000
        })
        upcomingMatches.value = []
      } finally {
        loading.value = false
      }
    }

    // 获取我的预测历史
    const fetchMyPredictions = async () => {
      // 如果组件已卸载，不执行请求
      if (!isComponentMounted.value) return

      console.log('开始获取预测历史...')

      // 创建新的AbortController
      abortController = new AbortController()
      const signal = abortController.signal

      loading.value = true
      try {
        console.log('发送请求获取预测历史...')
        const response = await axios.get('/api/predictions/my', {
          headers: { Authorization: `Bearer ${userStore.token}` },
          signal // 添加信号以支持取消请求
        })

        console.log('获取预测历史成功:', response)
        console.log('API响应类型:', typeof response.data)
        console.log('API响应内容:', JSON.stringify(response.data, null, 2))

        // 如果组件已卸载，不更新状态
        if (!isComponentMounted.value) return

        // 处理响应数据
        let predictionsData = [];

        // 检查响应数据的格式
        if (Array.isArray(response.data)) {
          // 直接是数组格式
          predictionsData = response.data;
          console.log('响应数据是数组格式');
        } else if (response.data && typeof response.data === 'object') {
          // 是对象格式，检查是否有data属性
          if (Array.isArray(response.data.data)) {
            predictionsData = response.data.data;
            console.log('响应数据是对象格式，包含data数组');
          } else {
            console.error('响应数据中的data属性不是数组:', response.data);
          }
        } else {
          console.error('预测历史数据格式不正确:', typeof response.data);
        }

        // 更新预测数据
        myPredictions.value = predictionsData;
        console.log('预测历史数据已更新，数量:', myPredictions.value.length);

        // 检查每个预测的match属性
        myPredictions.value.forEach((prediction, index) => {
          console.log(`预测[${index}] - ID: ${prediction.id}, 有match属性: ${!!prediction.match}`);
          if (prediction.match) {
            console.log(`预测[${index}] - match属性内容:`, {
              id: prediction.match.id,
              title: prediction.match.title,
              optionA: prediction.match.optionA,
              optionB: prediction.match.optionB
            });
          }
        });
      } catch (error) {
        // 如果是取消请求导致的错误，不显示错误消息
        if (axios.isCancel(error)) {
          console.log('请求已取消')
          return
        }

        console.error('获取预测历史失败:', error)

        // 如果组件已卸载，不显示错误消息
        if (!isComponentMounted.value) return

        // 检查是否是认证错误
        if (error.response && error.response.status === 401) {
          ElMessage({
            message: '登录已过期，请重新登录',
            type: 'warning',
            customClass: 'custom-message',
            offset: 80,
            duration: 3000
          })

          // 将用户重定向到登录页面
          setTimeout(() => {
            window.location.href = '/login'
          }, 1500)
          return
        }

        ElMessage({
          message: '获取预测历史失败',
          type: 'error',
          customClass: 'custom-message',
          offset: 80,
          duration: 3000
        })
      } finally {
        if (isComponentMounted.value) {
          loading.value = false
        }
      }
    }

    // 获取积分排行榜
    const fetchLeaderboard = async (retryCount = 0) => {
      // 如果组件已卸载，不执行请求
      if (!isComponentMounted.value) return

      // 创建新的AbortController
      abortController = new AbortController()
      const signal = abortController.signal

      loading.value = true
      try {
        console.log('开始获取积分排行榜...')
        const response = await axios.get('/api/leaderboard', { signal })

        // 如果组件已卸载，不更新状态
        if (!isComponentMounted.value) return

        // 处理响应数据
        let leaderboardData = []

        console.log('排行榜原始响应数据:', response.data)

        if (Array.isArray(response.data)) {
          // 直接是数组格式
          leaderboardData = response.data
          console.log('排行榜数据是数组格式，数据条数:', response.data.length)
        } else if (response.data && typeof response.data === 'object') {
          if (response.data.data && Array.isArray(response.data.data)) {
            // 标准结构：{ success, message, data: [...] }
            console.log('排行榜数据是对象格式，包含data数组，数据条数:', response.data.data.length)
            leaderboardData = response.data.data
          } else {
            // 如果没有数据，可能是因为没有预测记录或没有已验证的预测
            console.log('排行榜数据为空或格式不符合预期，使用空数组')
            leaderboardData = []
          }
        } else {
          console.log('排行榜数据为空或格式异常:', typeof response.data)
          leaderboardData = []
        }

        leaderboard.value = leaderboardData
      } catch (error) {
        // 如果是取消请求导致的错误，不显示错误消息
        if (axios.isCancel(error)) {
          console.log('请求已取消')
          return
        }

        console.error('获取积分排行榜失败:', error)

        // 如果组件已卸载，不显示错误消息
        if (!isComponentMounted.value) return

        // 尝试重试，最多重试3次
        if (retryCount < 3) {
          console.log(`尝试重新获取排行榜，重试次数: ${retryCount + 1}`);
          setTimeout(() => {
            fetchLeaderboard(retryCount + 1);
          }, 1000); // 等待1秒后重试
          return;
        }

        // 重试失败，设置空数组并显示错误消息
        leaderboard.value = [];

        ElMessage({
          message: '获取积分排行榜失败，请稍后再试',
          type: 'warning',
          customClass: 'custom-message',
          offset: 80,
          duration: 3000
        })
      } finally {
        if (isComponentMounted.value) {
          loading.value = false
        }
      }
    }

    // 获取我的总积分
    const fetchMyTotalPoints = async () => {
      // 如果组件已卸载，不执行请求
      if (!isComponentMounted.value) return

      // 创建新的AbortController
      abortController = new AbortController()
      const signal = abortController.signal

      try {
        const response = await axios.get('/api/predictions/my-points', {
          headers: { Authorization: `Bearer ${userStore.token}` },
          signal
        })

        // 如果组件已卸载，不更新状态
        if (!isComponentMounted.value) return

        // 处理响应数据
        let pointsData = 0

        console.log('总积分原始响应数据:', response.data)

        if (typeof response.data === 'number') {
          // 直接是数字
          pointsData = response.data
          console.log('总积分数据是数字格式:', response.data)
        } else if (response.data && typeof response.data === 'object') {
          if (typeof response.data.data === 'number') {
            // 标准结构：{ success, message, data: number }
            console.log('总积分数据是对象格式，包含data数字:', response.data.data)
            pointsData = response.data.data
          } else if (response.data.data !== undefined) {
            // 尝试将data转换为数字
            const parsedPoints = Number(response.data.data)
            if (!isNaN(parsedPoints)) {
              console.log('将data转换为数字:', parsedPoints)
              pointsData = parsedPoints
            } else {
              console.log('无法将data转换为数字:', response.data.data)
              pointsData = 0
            }
          } else {
            console.log('总积分数据不包含data属性或格式不符合预期:', response.data)
            pointsData = 0
          }
        } else {
          console.log('总积分数据为空或格式异常:', typeof response.data)
          pointsData = 0
        }

        myTotalPoints.value = pointsData
      } catch (error) {
        // 如果是取消请求导致的错误，不显示错误消息
        if (axios.isCancel(error)) {
          console.log('请求已取消')
          return
        }

        console.error('获取总积分失败:', error)
      }
    }

    // 选择获胜方
    const selectWinner = (winner) => {
      currentPrediction.predictedWinner = winner
    }

    // 提交预测
    const submitPrediction = async (matchId) => {
      // 如果组件已卸载，不执行请求
      if (!isComponentMounted.value) return

      if (!isPredictionValid.value) {
        ElMessage({
          message: '请选择获胜方并填写有效的比分',
          type: 'warning',
          customClass: 'custom-message',
          offset: 80,
          duration: 3000
        })
        return
      }

      // 创建新的AbortController
      abortController = new AbortController()
      const signal = abortController.signal

      try {
        currentPrediction.matchId = matchId
        await axios.post('/api/predictions', currentPrediction, {
          headers: { Authorization: `Bearer ${userStore.token}` },
          signal
        })

        // 如果组件已卸载，不继续执行
        if (!isComponentMounted.value) return

        ElMessage({
          message: '预测提交成功',
          type: 'success',
          customClass: 'custom-message',
          offset: 80,
          duration: 3000
        })

        // 重置预测表单
        currentPrediction.matchId = null
        currentPrediction.predictedWinner = ''
        currentPrediction.predictedScoreA = 0
        currentPrediction.predictedScoreB = 0

        // 切换到预测历史标签
        console.log('提交预测成功，切换到预测历史标签...')
        activeTab.value = 'history'
        // 不需要手动调用fetchMyPredictions，因为我们已经添加了监听器
        // 在标签页变化时会自动调用handleTabChange函数
      } catch (error) {
        // 如果是取消请求导致的错误，不显示错误消息
        if (axios.isCancel(error)) {
          console.log('请求已取消')
          return
        }

        console.error('提交预测失败:', error)

        // 如果组件已卸载，不显示错误消息
        if (!isComponentMounted.value) return

        // 检查是否是认证错误
        if (error.response && error.response.status === 401) {
          ElMessage({
            message: '登录已过期，请重新登录',
            type: 'warning',
            customClass: 'custom-message',
            offset: 80,
            duration: 3000
          })

          // 将用户重定向到登录页面
          setTimeout(() => {
            window.location.href = '/login'
          }, 1500)
          return
        }

        let errorMessage = '提交预测失败'

        if (error.response && error.response.data && error.response.data.message) {
          errorMessage = error.response.data.message
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

    // 格式化日期
    const formatDate = (dateString) => {
      if (!dateString) return 'N/A'

      try {
        const date = new Date(dateString)

        // 检查日期是否有效
        if (isNaN(date.getTime())) {
          console.warn('无效的日期格式:', dateString)
          return dateString
        }

        const year = date.getFullYear()
        const month = String(date.getMonth() + 1).padStart(2, '0')
        const day = String(date.getDate()).padStart(2, '0')
        const hours = String(date.getHours()).padStart(2, '0')
        const minutes = String(date.getMinutes()).padStart(2, '0')

        return `${year}-${month}-${day} ${hours}:${minutes}`
      } catch (error) {
        console.error('格式化日期出错:', error)
        return dateString
      }
    }

    // 获取比赛状态文本
    const getMatchStatusText = (status) => {
      const statusMap = {
        'not_started': '未开始',
        'in_progress': '进行中',
        'completed': '已结束',
        'cancelled': '已取消'
      }
      return statusMap[status] || status
    }



    // 监听标签切换
    const handleTabChange = () => {
      if (activeTab.value === 'upcoming') {
        fetchUpcomingMatches()
      } else if (activeTab.value === 'history') {
        fetchMyPredictions()
      } else if (activeTab.value === 'leaderboard') {
        fetchLeaderboard()
        fetchMyTotalPoints()
      }
    }

    // 用于跟踪组件是否已卸载
    const isComponentMounted = ref(false)

    // 取消请求的控制器
    let abortController = new AbortController()

    onMounted(() => {
      console.log('预测页面加载完成，开始获取比赛数据...')
      isComponentMounted.value = true
      fetchUpcomingMatches()
    })

    // 组件卸载时清理资源
    onUnmounted(() => {
      console.log('预测页面卸载，清理资源...')
      isComponentMounted.value = false
      // 取消所有进行中的请求
      abortController.abort()
    })

    // 监听比赛数据变化
    watch(upcomingMatches, (newMatches) => {
      // 检查组件是否已卸载
      if (!isComponentMounted.value) return

      console.log('比赛数据已更新，数量:', newMatches.length)
      if (newMatches.length > 0) {
        console.log('第一场比赛数据示例:', newMatches[0])
      }
    })

    // 监听标签页变化
    watch(activeTab, (newTab) => {
      // 检查组件是否已卸载
      if (!isComponentMounted.value) return

      console.log('标签页切换为:', newTab)
      handleTabChange()
    })

    // 判断是否正在编辑某个比赛的预测
    const isEditingPrediction = (matchId) => {
      // 检查当前选中的比赛是否与传入的比赛 ID 相同
      return currentPrediction.matchId === matchId;
    };

    // 修改预测功能
    const editPrediction = (prediction) => {
      // 将预测数据复制到当前预测对象
      currentPrediction.matchId = prediction.matchId;
      currentPrediction.predictedWinner = prediction.predictedWinner;
      currentPrediction.predictedScoreA = prediction.predictedScoreA;
      currentPrediction.predictedScoreB = prediction.predictedScoreB;

      // 切换到预测比赛标签页
      activeTab.value = 'upcoming';

      // 显示成功消息
      ElMessage({
        message: '已加载您的预测，请进行修改并重新提交',
        type: 'info',
        customClass: 'custom-message',
        offset: 80,
        duration: 3000
      });
    };

    return {
      activeTab,
      loading,
      upcomingMatches,
      myPredictions,
      leaderboard,
      myTotalPoints,
      currentPrediction,
      tempScoreA,
      tempScoreB,
      isPredictionValid,
      selectWinner,
      submitPrediction,
      formatDate,
      getMatchStatusText,
      handleTabChange,
      editPrediction,
      updatePrediction,
      isEditingPrediction
    }
  }
}
</script>

<style scoped>
.predictions-container {
  max-width: 1200px;
  margin: 0 auto;
  padding: 20px;
}

h1 {
  text-align: center;
  margin-bottom: 30px;
  color: #333;
}

.tabs {
  display: flex;
  justify-content: center;
  margin-bottom: 30px;
  border-bottom: 1px solid #eee;
}

.tab {
  padding: 10px 20px;
  cursor: pointer;
  margin: 0 10px;
  font-weight: 500;
  color: #666;
  border-bottom: 2px solid transparent;
  transition: all 0.3s;
}

.tab.active {
  color: #409EFF;
  border-bottom: 2px solid #409EFF;
}

.tab-content {
  min-height: 400px;
}

.loading, .empty-state {
  text-align: center;
  padding: 50px;
  color: #999;
}

.matches-list, .predictions-list {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
  gap: 20px;
}

.match-card, .prediction-card {
  border: 1px solid #eee;
  border-radius: 8px;
  padding: 20px;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.05);
  transition: all 0.3s;
}

.match-card:hover {
  transform: translateY(-5px);
  box-shadow: 0 5px 15px rgba(0, 0, 0, 0.1);
}

.match-header, .prediction-header {
  display: flex;
  justify-content: space-between;
  margin-bottom: 15px;
  padding-bottom: 10px;
  border-bottom: 1px solid #eee;
}

.match-title {
  font-weight: bold;
  font-size: 16px;
}

.match-time {
  color: #999;
  font-size: 14px;
}

.match-teams, .prediction-teams {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.team {
  display: flex;
  flex-direction: column;
  align-items: center;
  width: 40%;
  padding: 10px;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.3s;
}

.team.selected {
  background-color: rgba(64, 158, 255, 0.1);
  border: 1px solid #409EFF;
}

.team.winner {
  background-color: rgba(103, 194, 58, 0.1);
  border: 1px solid #67C23A;
}

/* 使用TeamLogo组件的样式 */

.team-name {
  font-weight: bold;
  margin-bottom: 10px;
}

.team-score {
  font-size: 24px;
  font-weight: bold;
}

.vs {
  font-weight: bold;
  color: #999;
}

.score-input {
  width: 60px;
  height: 40px;
  text-align: center;
  font-size: 18px;
  border: 1px solid #ddd;
  border-radius: 4px;
}

.match-actions {
  display: flex;
  justify-content: center;
}

.predict-button {
  background-color: #409EFF;
  color: white;
  border: none;
  padding: 10px 20px;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.3s;
}

.predict-button:hover {
  background-color: #66b1ff;
}

.predict-button:disabled {
  background-color: #a0cfff;
  cursor: not-allowed;
}

.prediction-result {
  margin-top: 15px;
  padding-top: 15px;
  border-top: 1px dashed #eee;
}

.actual-result {
  margin-bottom: 15px;
  text-align: center;
  font-size: 16px;
}

.result-info-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  margin-top: 15px;
}

.winner {
  font-weight: bold;
  color: #67C23A;
}

.points {
  font-weight: bold;
  color: #E6A23C;
  font-size: 16px;
}

.status {
  display: inline-block;
  padding: 6px 12px;
  border-radius: 4px;
  background-color: #F56C6C;
  color: white;
  font-size: 14px;
}

.status.correct {
  background-color: #67C23A;
}

.prediction-status {
  color: #999;
  font-style: italic;
  margin-top: 15px;
  padding-top: 15px;
  border-top: 1px dashed #eee;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 10px;
}

.user-prediction {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 15px;
  padding: 15px 0;
}

.prediction-label {
  font-weight: bold;
  color: #409EFF;
  margin-bottom: 5px;
}

.status-text {
  margin-bottom: 5px;
}

.edit-prediction-btn {
  background-color: #409EFF;
  color: white;
  border: none;
  padding: 8px 16px;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.3s;
  font-size: 14px;
}

.edit-prediction-btn:hover {
  background-color: #66b1ff;
  transform: translateY(-2px);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.leaderboard {
  width: 100%;
  border: 1px solid #eee;
  border-radius: 8px;
  overflow: hidden;
}

.leaderboard-header {
  display: flex;
  background-color: #f5f7fa;
  padding: 15px;
  font-weight: bold;
}

.leaderboard-item {
  display: flex;
  padding: 15px;
  border-top: 1px solid #eee;
}

.leaderboard-item:nth-child(even) {
  background-color: #fafafa;
}

.rank {
  width: 15%;
  text-align: center;
}

.user {
  width: 60%;
}

.points {
  width: 25%;
  text-align: right;
  font-weight: bold;
}

.my-points {
  margin-top: 30px;
  text-align: center;
  font-size: 18px;
}

.my-points span {
  font-weight: bold;
  color: #E6A23C;
  font-size: 24px;
}

:global(.custom-message) {
  min-width: 240px;
  padding: 15px 20px;
  display: flex;
  align-items: center;
  border-radius: 6px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.12);
  position: fixed;
  top: 20px;
  left: 50%;
  transform: translateX(-50%);
  z-index: 9999;
}

/* 积分规则样式 */
.rules-container {
  max-width: 800px;
  margin: 0 auto;
  padding: 20px;
  background-color: #fff;
  border-radius: 8px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
}

.rules-container h2 {
  text-align: center;
  margin-bottom: 20px;
  color: #333;
}

.rule-item {
  margin-bottom: 20px;
  padding: 15px;
  border-radius: 8px;
  background-color: #f9fafb;
  border-left: 4px solid #3b82f6;
}

.rule-title {
  font-size: 18px;
  font-weight: bold;
  color: #333;
  margin-bottom: 5px;
}

.rule-points {
  font-size: 16px;
  font-weight: bold;
  color: #3b82f6;
  margin-bottom: 10px;
}

.rule-description {
  font-size: 14px;
  color: #666;
  line-height: 1.5;
}

.rule-note {
  margin-top: 30px;
  padding: 15px;
  background-color: #f0f9ff;
  border-radius: 8px;
  border-left: 4px solid #60a5fa;
}

.rule-note p {
  font-size: 14px;
  color: #555;
  line-height: 1.5;
  margin: 0;
}

.prediction-loading {
  padding: 20px;
  text-align: center;
  color: #999;
  background-color: #f9f9f9;
  border-radius: 8px;
  margin: 10px 0;
}

.prediction-loading p {
  margin: 5px 0;
}
</style>
