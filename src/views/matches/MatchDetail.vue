<template>
  <div class="match-detail-container">
    <div v-if="loading" class="loading-container">
      <el-skeleton :rows="10" animated />
    </div>

    <div v-else-if="!match" class="not-found">
      <h2>比赛未找到</h2>
      <p>该比赛可能已被删除或不存在</p>
      <el-button type="primary" @click="$router.push('/matches')">
        返回比赛列表
      </el-button>
    </div>

    <div v-else class="match-detail">
      <!-- 比赛基本信息 -->
      <div class="match-header">
        <el-tag :type="statusTagType" class="status-tag">{{ statusText }}</el-tag>
        <div class="teams-container">
          <div class="team team-a" :class="{ 'team-winner': isTeamAWinner }">
            <div class="team-name">{{ match.team_a }}</div>
          </div>

          <div class="match-score-container">
            <div class="match-series">{{ match.match_series }}</div>
            <div v-if="match.status === 'finished'" class="match-score">
              {{ match.result_score }}
            </div>
            <div v-else class="match-vs">VS</div>
          </div>

          <div class="team team-b" :class="{ 'team-winner': isTeamBWinner }">
            <div class="team-name">{{ match.team_b }}</div>
          </div>
        </div>

        <div class="match-info">
          <div class="info-item">
            <span class="label">比赛时间：</span>
            <span class="value">{{ formatDateTime(match.start_time || match.matchTime) }}</span>
          </div>
          <div class="info-item">
            <span class="label">比赛类型：</span>
            <span class="value">{{ matchTypeText }}</span>
          </div>
        </div>
      </div>

      <!-- 预测区域 -->
      <div v-if="isPredictable" class="prediction-section">
        <div class="section-title">
          <h3>进行预测</h3>
          <p class="countdown" v-if="countdown">
            距离比赛开始还有：{{ countdownText }}
          </p>
        </div>

        <el-form
          ref="predictionForm"
          :model="predictionData"
          :rules="predictionRules"
          label-position="top"
          class="prediction-form"
        >
          <el-form-item label="预测获胜队伍" prop="predicted_winner">
            <el-radio-group v-model="predictionData.predicted_winner">
              <el-radio :label="match.team_a">{{ match.team_a }}</el-radio>
              <el-radio :label="match.team_b">{{ match.team_b }}</el-radio>
            </el-radio-group>
          </el-form-item>

          <el-form-item label="预测比分" prop="predicted_score">
            <el-radio-group v-model="predictionData.predicted_score">
              <template v-if="match.match_series === 'BO1'">
                <el-radio label="1-0">1-0</el-radio>
              </template>

              <template v-else-if="match.match_series === 'BO2'">
                <el-radio label="2-0">2-0</el-radio>
                <el-radio label="1-1">1-1</el-radio>
                <el-radio label="0-2">0-2</el-radio>
              </template>

              <template v-else-if="match.match_series === 'BO3'">
                <el-radio label="2-0">2-0</el-radio>
                <el-radio label="2-1">2-1</el-radio>
                <el-radio label="1-2">1-2</el-radio>
                <el-radio label="0-2">0-2</el-radio>
              </template>

              <template v-else-if="match.match_series === 'BO5'">
                <el-radio label="3-0">3-0</el-radio>
                <el-radio label="3-1">3-1</el-radio>
                <el-radio label="3-2">3-2</el-radio>
                <el-radio label="2-3">2-3</el-radio>
                <el-radio label="1-3">1-3</el-radio>
                <el-radio label="0-3">0-3</el-radio>
              </template>

              <template v-else>
                <el-radio label="2-0">2-0</el-radio>
                <el-radio label="2-1">2-1</el-radio>
                <el-radio label="1-2">1-2</el-radio>
                <el-radio label="0-2">0-2</el-radio>
              </template>
            </el-radio-group>
          </el-form-item>

          <el-form-item>
            <el-button
              type="primary"
              @click="submitPrediction"
              :loading="submitting"
              :disabled="!userStore.isAuthenticated"
            >
              {{ userPrediction ? '修改预测' : '提交预测' }}
            </el-button>
            <p v-if="!userStore.isAuthenticated" class="login-tip">
              请<router-link to="/login">登录</router-link>后进行预测
            </p>
          </el-form-item>
        </el-form>
      </div>

      <!-- 比赛结果区域 -->
      <div v-if="match.status === 'finished'" class="result-section">
        <div class="section-title">
          <h3>比赛结果</h3>
        </div>

        <div class="result-content">
          <div class="winner">
            <span class="label">获胜队伍：</span>
            <span class="value">{{ match.result_winner }}</span>
          </div>
          <div class="final-score">
            <span class="label">最终比分：</span>
            <span class="value">{{ match.result_score }}</span>
          </div>
        </div>
      </div>

      <!-- 用户预测区域 -->
      <div v-if="userPrediction" class="user-prediction-section">
        <div class="section-title">
          <h3>您的预测</h3>
        </div>

        <div class="user-prediction">
          <div class="prediction-item">
            <span class="label">预测队伍：</span>
            <span class="value" :class="{ 'correct': isPredictionWinnerCorrect }">
              {{ userPrediction.predicted_winner }}
              <el-icon v-if="match.status === 'finished' && isPredictionWinnerCorrect" color="#67C23A">
                <Check />
              </el-icon>
              <el-icon v-else-if="match.status === 'finished' && !isPredictionWinnerCorrect" color="#F56C6C">
                <Close />
              </el-icon>
            </span>
          </div>

          <div class="prediction-item">
            <span class="label">预测比分：</span>
            <span class="value" :class="{ 'correct': isPredictionScoreCorrect }">
              {{ userPrediction.predicted_score }}
              <el-icon v-if="match.status === 'finished' && isPredictionScoreCorrect" color="#67C23A">
                <Check />
              </el-icon>
              <el-icon v-else-if="match.status === 'finished' && !isPredictionScoreCorrect" color="#F56C6C">
                <Close />
              </el-icon>
            </span>
          </div>

          <div v-if="match.status === 'finished'" class="prediction-item">
            <span class="label">获得积分：</span>
            <span class="value points">{{ userPrediction.points_earned }}</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, reactive, onMounted, onBeforeUnmount } from 'vue'
import { useRoute } from 'vue-router'
import { ElSkeleton, ElTag, ElForm, ElFormItem, ElRadio, ElRadioGroup, ElButton, ElMessage, ElIcon } from 'element-plus'
import { Check, Close } from '@element-plus/icons-vue'
import { getMatch, getUserPredictionForMatch, createPrediction } from '@/api/matches'
import { Match, Prediction, CreatePredictionData } from '@/types/match'
import { formatDateTime, calculateCountdown } from '@/utils/date'
import { useUserStore } from '@/stores/user'

const route = useRoute()
// 不需要使用router
const userStore = useUserStore()
const predictionForm = ref<typeof ElForm>()

// 状态
const loading = ref(true)
const match = ref<Match | null>(null)
const userPrediction = ref<Prediction | null>(null)
const submitting = ref(false)
const countdown = ref<{ days: number; hours: number; minutes: number; seconds: number } | null>(null)
const countdownInterval = ref<number | null>(null)

// 预测表单数据
const predictionData = reactive<{
  predicted_winner: string;
  predicted_score: string;
}>({
  predicted_winner: '',
  predicted_score: ''
})

// 表单验证规则
const predictionRules = {
  predicted_winner: [
    { required: true, message: '请选择预测获胜队伍', trigger: 'change' }
  ],
  predicted_score: [
    { required: true, message: '请选择预测比分', trigger: 'change' }
  ]
}

// 计算属性
const matchId = computed(() => {
  const id = route.params.id
  return typeof id === 'string' ? parseInt(id, 10) : 0
})

// 状态标签类型
const statusTagType = computed(() => {
  if (!match.value) return 'info'

  switch (match.value.status) {
    case 'not_started':
      return 'info'
    case 'in_progress':
      return 'warning'
    case 'finished':
      return 'success'
    case 'cancelled':
      return 'danger'
    default:
      return 'info'
  }
})

// 状态文本
const statusText = computed(() => {
  if (!match.value) return ''

  switch (match.value.status) {
    case 'not_started':
      return '未开始'
    case 'in_progress':
      return '进行中'
    case 'finished':
      return '已结束'
    case 'cancelled':
      return '已取消'
    default:
      return ''
  }
})

// 比赛类型文本
const matchTypeText = computed(() => {
  if (!match.value) return ''

  switch (match.value.match_type) {
    case 'regular':
      return '常规赛'
    case 'playoff':
      return '季后赛'
    case 'final':
      return '决赛'
    default:
      return ''
  }
})

// 是否可预测
const isPredictable = computed(() => {
  return match.value?.is_predictable || false
})

// 判断获胜队伍
const isTeamAWinner = computed(() => {
  return match.value?.status === 'finished' && match.value?.result_winner === match.value?.team_a
})

const isTeamBWinner = computed(() => {
  return match.value?.status === 'finished' && match.value?.result_winner === match.value?.team_b
})

// 预测结果是否正确
const isPredictionWinnerCorrect = computed(() => {
  if (!match.value || !userPrediction.value || match.value.status !== 'finished') return false
  return userPrediction.value.predicted_winner === match.value.result_winner
})

const isPredictionScoreCorrect = computed(() => {
  if (!match.value || !userPrediction.value || match.value.status !== 'finished') return false
  return userPrediction.value.predicted_score === match.value.result_score
})

// 倒计时文本
const countdownText = computed(() => {
  if (!countdown.value) return ''

  const { days, hours, minutes, seconds } = countdown.value

  if (days > 0) {
    return `${days}天 ${hours}小时 ${minutes}分钟 ${seconds}秒`
  } else if (hours > 0) {
    return `${hours}小时 ${minutes}分钟 ${seconds}秒`
  } else {
    return `${minutes}分钟 ${seconds}秒`
  }
})

// 获取比赛详情
const fetchMatchDetail = async () => {
  loading.value = true
  try {
    match.value = await getMatch(matchId.value)

    // 获取用户预测
    if (userStore.isAuthenticated) {
      userPrediction.value = await getUserPredictionForMatch(matchId.value)

      // 如果有预测，则填充表单
      if (userPrediction.value) {
        predictionData.predicted_winner = userPrediction.value.predicted_winner || ''
        predictionData.predicted_score = userPrediction.value.predicted_score || ''
      }
    }

    // 如果比赛未开始，开始倒计时
    if (match.value.status === 'not_started' && match.value.is_predictable) {
      updateCountdown()
      countdownInterval.value = window.setInterval(updateCountdown, 1000)
    }
  } catch (error) {
    console.error('获取比赛详情失败', error)
    match.value = null
  } finally {
    loading.value = false
  }
}

// 更新倒计时
const updateCountdown = () => {
  if (!match.value) return

  countdown.value = calculateCountdown(match.value.start_time || match.value.matchTime)

  // 如果倒计时结束，刷新比赛数据
  if (countdown.value.days === 0 && countdown.value.hours === 0 &&
      countdown.value.minutes === 0 && countdown.value.seconds === 0) {
    if (countdownInterval.value) {
      clearInterval(countdownInterval.value)
      countdownInterval.value = null

      // 刷新比赛数据
      setTimeout(() => {
        fetchMatchDetail()
      }, 1000)
    }
  }
}

// 提交预测
const submitPrediction = async () => {
  if (!match.value) return

  try {
    await predictionForm.value?.validate()

    submitting.value = true

    // 解析比分字符串
    const scoreParts = predictionData.predicted_score.split('-')
    const scoreA = parseInt(scoreParts[0], 10)
    const scoreB = parseInt(scoreParts[1], 10)

    const data: CreatePredictionData = {
      matchId: match.value.id,
      predictedWinner: predictionData.predicted_winner,
      predictedScoreA: scoreA,
      predictedScoreB: scoreB
    }

    console.log('提交预测数据:', data)

    const prediction = await createPrediction(match.value.id, data)
    userPrediction.value = prediction

    ElMessage.success(userPrediction.value ? '预测修改成功' : '预测提交成功')
  } catch (error) {
    console.error('提交预测失败:', error)
    if (error instanceof Error) {
      ElMessage.error(error.message || '预测提交失败')
    } else {
      ElMessage.error('预测提交失败')
    }
  } finally {
    submitting.value = false
  }
}

// 初始化
onMounted(() => {
  fetchMatchDetail()
})

// 组件销毁前清除倒计时
onBeforeUnmount(() => {
  if (countdownInterval.value) {
    clearInterval(countdownInterval.value)
  }
})
</script>

<style scoped>
.match-detail-container {
  max-width: 800px;
  margin: 0 auto;
  padding: 20px 0;
}

.loading-container {
  padding: 40px 20px;
  background-color: #fff;
  border-radius: 8px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
}

.not-found {
  padding: 40px 20px;
  text-align: center;
  background-color: #fff;
  border-radius: 8px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
}

.match-detail {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.match-header, .prediction-section, .result-section, .user-prediction-section {
  background-color: #fff;
  border-radius: 8px;
  padding: 20px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
}

.status-tag {
  display: inline-block;
  margin-bottom: 10px;
}

.teams-container {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin: 20px 0;
}

.team {
  flex: 1;
  text-align: center;
}

.team-name {
  font-size: 24px;
  font-weight: bold;
  color: #303133;
}

.team-winner .team-name {
  color: #f56c6c;
}

.match-score-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 0 20px;
}

.match-series {
  font-size: 14px;
  color: #909399;
  margin-bottom: 5px;
}

.match-score {
  font-size: 24px;
  font-weight: bold;
  color: #303133;
}

.match-vs {
  font-size: 20px;
  font-weight: bold;
  color: #909399;
}

.match-info {
  display: flex;
  flex-wrap: wrap;
  margin-top: 20px;
  border-top: 1px solid #eee;
  padding-top: 15px;
}

.info-item {
  width: 50%;
  margin-bottom: 10px;
}

.label {
  color: #909399;
  margin-right: 5px;
}

.value {
  color: #303133;
  font-weight: 500;
}

.section-title {
  margin-bottom: 20px;
  border-bottom: 1px solid #eee;
  padding-bottom: 10px;
}

.section-title h3 {
  font-size: 18px;
  margin: 0;
  color: #303133;
}

.countdown {
  margin-top: 5px;
  color: #f56c6c;
  font-size: 14px;
}

.prediction-form {
  max-width: 600px;
}

.login-tip {
  margin-top: 10px;
  font-size: 14px;
  color: #909399;
}

.login-tip a {
  color: #409eff;
  text-decoration: none;
}

.result-content, .user-prediction {
  padding: 10px 0;
}

.prediction-item {
  margin-bottom: 15px;
}

.correct {
  color: #67C23A !important;
}

.points {
  color: #f56c6c !important;
  font-weight: bold;
}

@media (max-width: 768px) {
  .info-item {
    width: 100%;
  }

  .team-name {
    font-size: 18px;
  }

  .match-score, .match-vs {
    font-size: 18px;
  }
}
</style>