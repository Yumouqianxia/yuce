<template>
  <div class="match-card" @click="navigateToMatch">
    <div class="match-status" :class="statusClass">
      {{ statusText }}
    </div>

    <div class="match-info">
      <div class="match-time">
        <el-icon><Calendar /></el-icon>
        {{ formatDateTime((match as any).matchTime ?? (match as any).start_time) }}
      </div>
      <div class="match-metadata">
        <span class="match-type">{{ matchTypeText }}</span>
        <span class="match-series" v-if="match.series">{{ match.series }}</span>
      </div>
    </div>

    <div class="teams-container">
      <div class="team team-a" :class="{ 'team-winner': isTeamAWinner }">
        <TeamLogo :teamName="match.optionA" size="medium" />
        <div class="team-name">{{ match.optionA }}</div>
      </div>

      <div class="score-container">
        <template v-if="match.status === 'completed'">
          <div class="match-score" :class="{ 'team-a-win': isTeamAWinner, 'team-b-win': isTeamBWinner }">
            {{ displayScore }}
          </div>
        </template>
        <template v-else-if="match.status === 'in_progress'">
          <div class="match-live">
            <span class="live-indicator"></span>
            直播中
          </div>
        </template>
        <template v-else>
          <div class="match-vs">VS</div>
        </template>
      </div>

      <div class="team team-b" :class="{ 'team-winner': isTeamBWinner }">
        <TeamLogo :teamName="match.optionB" size="medium" />
        <div class="team-name">{{ match.optionB }}</div>
      </div>
    </div>

    <div class="match-actions" v-if="showPredictButton">
      <el-button
        size="small"
        :type="hasPrediction ? 'success' : 'primary'"
        :icon="hasPrediction ? Check : Edit"
        round
        @click.stop="onPredictClick"
      >
        {{ hasPrediction ? '修改预测' : '进行预测' }}
      </el-button>
    </div>
  </div>
</template>

<script lang="ts" setup>
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { ElButton } from 'element-plus'
import { Calendar, Check, Edit } from '@element-plus/icons-vue'
import { Match } from '@/types/match'
import { formatDateTime } from '@/utils/date'
import { useUserStore } from '@/stores/user'
import TeamLogo from '@/components/teams/TeamLogo.vue'

// 定义属性
const props = withDefaults(defineProps<{
  match: Match
  showPredictButton?: boolean
  hasPrediction?: boolean
}>(), {
  showPredictButton: true,
  hasPrediction: false
})

// 定义事件
const emit = defineEmits<{
  (e: 'predict', matchId: number): void
}>()

const router = useRouter()
const userStore = useUserStore()

// 计算状态CSS类
const statusClass = computed(() => {
  switch ((props.match.status as any)) {
    case 'not_started':
      return 'status-not-started'
    case 'in_progress':
      return 'status-in-progress'
    case 'completed':
      return 'status-finished'
    case 'cancelled':
      return 'status-cancelled'
    default:
      return ''
  }
})

// 计算状态文本
const statusText = computed(() => {
  switch ((props.match.status as any)) {
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
})

// 计算比赛类型文本
const matchTypeText = computed(() => {
  switch (props.match.matchType) {
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

// 显示比分
const displayScore = computed(() => {
  if (props.match.scoreA !== undefined && props.match.scoreB !== undefined) {
    return `${props.match.scoreA}-${props.match.scoreB}`
  }
  return '0-0'
})

// 判断获胜队伍
const isTeamAWinner = computed(() => {
  return props.match.status === 'completed' && props.match.winner === 'A'
})

const isTeamBWinner = computed(() => {
  return props.match.status === 'completed' && props.match.winner === 'B'
})



// 导航到比赛详情
const navigateToMatch = () => {
  router.push(`/matches/${props.match.id}`)
}

// 进行预测
const onPredictClick = () => {
  if (userStore.isAuthenticated) {
    emit('predict', props.match.id)
  } else {
    router.push({
      path: '/login',
      query: { redirect: `/matches/${props.match.id}` }
    })
  }
}
</script>

<style scoped>
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

.match-status {
  position: absolute;
  top: 12px;
  right: 12px;
  font-size: 12px;
  padding: 3px 10px;
  border-radius: 12px;
  font-weight: 500;
}

.status-not-started {
  background-color: rgba(59, 130, 246, 0.1);
  color: var(--primary-color);
}

.status-in-progress {
  background-color: rgba(245, 158, 11, 0.1);
  color: var(--warning-color);
}

.status-finished {
  background-color: rgba(16, 185, 129, 0.1);
  color: var(--success-color);
}

.status-cancelled {
  background-color: rgba(239, 68, 68, 0.1);
  color: var(--danger-color);
}

.match-info {
  display: flex;
  flex-direction: column;
  margin-bottom: 20px;
}

.match-time {
  font-size: 14px;
  color: var(--text-secondary);
  display: flex;
  align-items: center;
  margin-bottom: 6px;
}

.match-time .el-icon {
  margin-right: 4px;
}

.match-metadata {
  display: flex;
  gap: 8px;
}

.match-type, .match-series {
  font-size: 12px;
  color: var(--text-secondary);
  background-color: var(--bg-light);
  padding: 3px 8px;
  border-radius: 4px;
}

.teams-container {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}

.team {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  transition: transform 0.2s;
}

.team-logo {
  width: 48px;
  height: 48px;
  background: linear-gradient(135deg, var(--primary-light), var(--primary-color));
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 8px;
  color: white;
  font-weight: bold;
  font-size: 22px;
  overflow: hidden;
}

.team-logo-img {
  width: 100%;
  height: 100%;
  object-fit: contain;
}

.team-a .team-logo {
  background: linear-gradient(135deg, #60a5fa, #3b82f6);
}

.team-b .team-logo {
  background: linear-gradient(135deg, #f87171, #ef4444);
}

.team-name {
  font-size: 15px;
  font-weight: 600;
  text-align: center;
  color: var(--text-primary);
}

.team-winner {
  transform: scale(1.05);
}

.team-winner .team-name {
  color: var(--primary-dark);
}

.team-winner .team-logo {
  box-shadow: 0 0 15px rgba(59, 130, 246, 0.4);
}

.score-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 0 16px;
  min-width: 80px;
}

.match-score {
  font-size: 24px;
  font-weight: bold;
  color: var(--text-primary);
  background-color: var(--bg-light);
  padding: 4px 12px;
  border-radius: 12px;
}

.team-a-win {
  background: linear-gradient(to right, rgba(96, 165, 250, 0.2), rgba(59, 130, 246, 0.1));
  color: #3b82f6;
}

.team-b-win {
  background: linear-gradient(to right, rgba(248, 113, 113, 0.2), rgba(239, 68, 68, 0.1));
  color: #ef4444;
}

.match-vs {
  font-size: 18px;
  font-weight: bold;
  color: var(--text-light);
  background-color: var(--bg-light);
  padding: 4px 12px;
  border-radius: 12px;
}

.match-live {
  display: flex;
  align-items: center;
  font-size: 14px;
  font-weight: 600;
  color: var(--warning-color);
}

.live-indicator {
  width: 8px;
  height: 8px;
  background-color: var(--warning-color);
  border-radius: 50%;
  display: inline-block;
  margin-right: 6px;
  animation: pulse 1.5s infinite;
}

@keyframes pulse {
  0% {
    transform: scale(0.95);
    box-shadow: 0 0 0 0 rgba(245, 158, 11, 0.5);
  }
  70% {
    transform: scale(1);
    box-shadow: 0 0 0 6px rgba(245, 158, 11, 0);
  }
  100% {
    transform: scale(0.95);
    box-shadow: 0 0 0 0 rgba(245, 158, 11, 0);
  }
}

.match-actions {
  display: flex;
  justify-content: center;
}
</style>