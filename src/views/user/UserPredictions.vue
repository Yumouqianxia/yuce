<template>
  <div class="user-predictions-container">
    <h1 class="page-title">我的预测</h1>

    <div class="predictions-content">
      <div v-if="loading" class="loading-container">
        <el-skeleton :rows="5" animated />
      </div>

      <div v-else-if="predictions.length === 0" class="empty-data">
        您还没有进行任何预测
        <div class="empty-action">
          <router-link to="/matches">
            <el-button type="primary">去预测比赛</el-button>
          </router-link>
        </div>
      </div>

      <div v-else>
        <el-table
          v-loading="loading"
          :data="predictions"
          style="width: 100%"
        >
          <el-table-column label="比赛" min-width="220">
            <template #default="{ row }">
              <div class="match-info">
                <div class="teams">
                  <span class="team" :class="{ 'winner': row.match.winner === 'A' }">
                    {{ row.match.optionA || row.match.team_a }}
                  </span>
                  <span class="vs">VS</span>
                  <span class="team" :class="{ 'winner': row.match.winner === 'B' }">
                    {{ row.match.optionB || row.match.team_b }}
                  </span>
                </div>
                <div class="match-time">{{ formatDateTime(row.match.matchTime || row.match.start_time) }}</div>
              </div>
            </template>
          </el-table-column>

          <el-table-column label="比赛状态" width="100">
            <template #default="{ row }">
              <el-tag :type="getStatusType(row.match.status)">
                {{ getStatusText(row.match.status) }}
              </el-tag>
            </template>
          </el-table-column>

          <el-table-column label="您的预测" width="180">
            <template #default="{ row }">
              <div class="prediction-info">
                <div class="predicted-winner">
                  获胜方:
                  <span :class="{ 'correct': isPredictionWinnerCorrect(row), 'incorrect': isMatchFinishedAndWrongPrediction(row) }">
                    {{ row.predicted_winner }}
                    <el-icon v-if="isPredictionWinnerCorrect(row)" color="#67C23A">
                      <Check />
                    </el-icon>
                    <el-icon v-else-if="isMatchFinishedAndWrongPrediction(row)" color="#F56C6C">
                      <Close />
                    </el-icon>
                  </span>
                </div>
                <div class="predicted-score">
                  比分:
                  <span :class="{ 'correct': isPredictionScoreCorrect(row), 'incorrect': isMatchFinishedAndWrongScore(row) }">
                    {{ row.predicted_score }}
                    <el-icon v-if="isPredictionScoreCorrect(row)" color="#67C23A">
                      <Check />
                    </el-icon>
                    <el-icon v-else-if="isMatchFinishedAndWrongScore(row)" color="#F56C6C">
                      <Close />
                    </el-icon>
                  </span>
                </div>
              </div>
            </template>
          </el-table-column>

          <el-table-column label="实际结果" width="160">
            <template #default="{ row }">
              <div v-if="row.match.status === 'completed'" class="result-info">
                <div class="result-winner">获胜方: {{ row.match.winner }}</div>
                <div class="result-score">比分: {{ row.match.scoreA }} - {{ row.match.scoreB }}</div>
              </div>
              <span v-else>-</span>
            </template>
          </el-table-column>

          <el-table-column label="获得积分" width="100">
            <template #default="{ row }">
              <span v-if="row.match.status === 'completed'" class="points">
                {{ row.points_earned || 0 }}
              </span>
              <span v-else>-</span>
            </template>
          </el-table-column>

          <el-table-column label="操作" width="100">
            <template #default="{ row }">
              <router-link :to="`/matches/${row.match.id}`">
                <el-button type="primary" size="small" text>查看详情</el-button>
              </router-link>
            </template>
          </el-table-column>
        </el-table>

        <el-pagination
          v-if="totalPredictions > 0"
          layout="total, prev, pager, next"
          :total="totalPredictions"
          :page-size="pageSize"
          :current-page="currentPage"
          @current-change="handlePageChange"
          class="pagination"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElSkeleton, ElTable, ElTableColumn, ElTag, ElButton, ElPagination, ElIcon } from 'element-plus'
import { Check, Close } from '@element-plus/icons-vue'
import { getUserPredictions } from '@/api/matches'
import { Prediction, MatchStatus } from '@/types/match'
import { formatDateTime } from '@/utils/date'

// 状态
const loading = ref(true)
const predictions = ref<Prediction[]>([])
const totalPredictions = ref(0)
const currentPage = ref(1)
const pageSize = ref(10)

// 获取用户预测列表
const fetchUserPredictions = async () => {
  loading.value = true
  try {
    const list = await getUserPredictions()
    // 兼容返回格式：标准成功时 http.ts 已返回 data
    // 这里假设为数组
    predictions.value = Array.isArray(list) ? list as any : (list as any)?.results || []
    totalPredictions.value = Array.isArray(list) ? (list as any).length : ((list as any)?.count || (predictions.value as any[]).length)
  } catch (error) {
    console.error('获取用户预测失败', error)
  } finally {
    loading.value = false
  }
}

// 处理页码变化
const handlePageChange = (page: number) => {
  currentPage.value = page
  fetchUserPredictions()
}

// 获取状态类型
const getStatusType = (status: MatchStatus) => {
  switch (status) {
    case 'not_started':
      return 'info'
    case 'in_progress':
      return 'warning'
    case 'completed':
      return 'success'
    case 'cancelled':
      return 'danger'
    default:
      return 'info'
  }
}

// 获取状态文本
const getStatusText = (status: MatchStatus) => {
  switch (status) {
    case 'not_started':
      return '未开始'
    case 'in_progress':
      return '进行中'
    case 'completed':
      return '已结束'
    case 'cancelled':
      return '已取消'
    default:
      return ''
  }
}

// 预测结果是否正确
const isPredictionWinnerCorrect = (prediction: Prediction) => {
  if (prediction.match.status !== 'completed') return false
  return prediction.predicted_winner === prediction.match.winner
}

const isPredictionScoreCorrect = (prediction: Prediction) => {
  if (prediction.match.status !== 'completed') return false
  // 适配后端: 使用 predicted_score_a / predicted_score_b 与 match.scoreA / scoreB 比较
  const pa = (prediction as any).predicted_score_a ?? prediction.predictedScoreA
  const pb = (prediction as any).predicted_score_b ?? prediction.predictedScoreB
  return pa === (prediction.match as any).scoreA && pb === (prediction.match as any).scoreB
}

// 比赛已结束且预测错误
const isMatchFinishedAndWrongPrediction = (prediction: Prediction) => {
  if (prediction.match.status !== 'completed') return false
  return prediction.predicted_winner !== prediction.match.winner
}

const isMatchFinishedAndWrongScore = (prediction: Prediction) => {
  if (prediction.match.status !== 'completed') return false
  const pa = (prediction as any).predicted_score_a ?? prediction.predictedScoreA
  const pb = (prediction as any).predicted_score_b ?? prediction.predictedScoreB
  return !(pa === (prediction.match as any).scoreA && pb === (prediction.match as any).scoreB)
}

// 初始化
onMounted(() => {
  fetchUserPredictions()
})
</script>

<style scoped>
.user-predictions-container {
  max-width: 1000px;
  margin: 0 auto;
}

.page-title {
  font-size: 24px;
  font-weight: 600;
  margin-bottom: 20px;
  color: #303133;
}

.predictions-content {
  background-color: #fff;
  border-radius: 8px;
  padding: 20px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
}

.loading-container {
  padding: 20px 0;
}

.empty-data {
  text-align: center;
  color: #909399;
  padding: 40px 0;
}

.empty-action {
  margin-top: 20px;
}

.match-info {
  display: flex;
  flex-direction: column;
}

.teams {
  display: flex;
  align-items: center;
  margin-bottom: 5px;
}

.team {
  font-weight: 500;
}

.team.winner {
  color: #f56c6c;
  font-weight: bold;
}

.vs {
  margin: 0 8px;
  color: #909399;
}

.match-time {
  font-size: 12px;
  color: #909399;
}

.prediction-info, .result-info {
  display: flex;
  flex-direction: column;
  gap: 5px;
}

.predicted-winner, .predicted-score, .result-winner, .result-score {
  font-size: 14px;
}

.correct {
  color: #67C23A;
  font-weight: bold;
}

.incorrect {
  color: #F56C6C;
}

.points {
  font-weight: bold;
  color: #f56c6c;
}

.pagination {
  margin-top: 20px;
  display: flex;
  justify-content: center;
}
</style>