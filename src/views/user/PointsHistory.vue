<template>
  <div class="points-history-container">
    <div v-if="loading" class="loading-container">
      <el-skeleton :rows="10" animated />
    </div>

    <div v-else class="points-history-content">
      <div class="page-header">
        <h2>{{ isAdminView ? '用户积分历史' : '积分历史' }}</h2>
        <el-button class="blue-btn back-btn" @click="goBack">
          {{ isAdminView ? '返回用户管理' : '返回个人资料' }}
        </el-button>
      </div>

      <div class="history-card">
        <div class="total-points">
          <div class="points-label">总积分</div>
          <div class="points-value">{{ totalPoints }}</div>
        </div>

        <div v-if="pointsHistory.length === 0" class="empty-history">
          <el-empty description="暂无积分历史记录" />
        </div>
        <div v-else class="points-history-list">
          <div class="history-bubble" v-for="(item, index) in pointsHistory" :key="index" :class="getBubbleClass(item.points_change, item.change_type)">
            <div class="history-content">
              <div class="history-header">
                <div class="history-match" v-if="item.related_match">
                  {{ item.related_match }}
                </div>
                <div class="history-points" :class="{ 'positive': item.points_change > 0, 'negative': item.points_change < 0 }">
                  {{ item.points_change > 0 ? '+' : '' }}{{ item.points_change }}
                </div>
              </div>
              <div class="history-description">{{ item.description || getDescription(item) }}</div>
              <div class="history-timestamp">{{ formatDate(item.created_at) }}</div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage, ElSkeleton, ElButton, ElEmpty } from 'element-plus'
import { useUserStore } from '@/stores/user'
import type { PointsHistory } from '@/types/user'
import axios from 'axios'

const userStore = useUserStore()
const router = useRouter()
const route = useRoute()

// 状态
const loading = ref(true)
const pointsHistory = ref<PointsHistory[]>([])
const totalPoints = ref(0)
const userId = ref<number | null>(null)
const isAdminView = ref(false)

// 获取积分历史
const fetchPointsHistory = async () => {
  loading.value = true
  // 重置数据
  pointsHistory.value = []
  totalPoints.value = 0
  
  try {
    // 获取总积分
    await fetchTotalPoints()

    const urlUserId = route.query.userId ? Number(route.query.userId) : null

    // 仅使用排行榜积分历史接口；预测积分已包含在此表，不再合并额外数据
    const targetId = (urlUserId && userStore.user?.role === 'admin')
      ? urlUserId
      : userStore.user?.id

    if (!targetId) {
      throw new Error('无法确定用户ID')
    }

    const endpoint = `/api/leaderboard/users/${targetId}/points-history`

    const resp = await axios.get(endpoint, {
      headers: { Authorization: `Bearer ${userStore.token}` }
    })

    const historyData = Array.isArray(resp.data?.data) ? resp.data.data : (Array.isArray(resp.data) ? resp.data : [])
    pointsHistory.value = historyData
  } catch (error) {
    console.error('获取积分历史失败:', error)
    ElMessage.error('获取积分历史失败')
  } finally {
    loading.value = false
  }
}

// 合并两种积分历史记录
const mergePointsHistory = async (adminHistoryData: any[], predictionsData: any): Promise<PointsHistory[]> => {
  console.log('开始合并积分历史数据:')
  console.log('管理员API数据:', adminHistoryData)
  console.log('预测API原始数据:', predictionsData)
  
  // 处理预测数据
  let predictions = []
  if (Array.isArray(predictionsData)) {
    predictions = predictionsData
  } else if (predictionsData && typeof predictionsData === 'object' && Array.isArray(predictionsData.data)) {
    predictions = predictionsData.data
  }
  
  // 打印预测数据
  console.log('处理后的预测数据:', predictions.map((p: any) => ({ 
    id: p.id, 
    matchId: p.matchId, 
    points: p.earnedPoints, 
    isVerified: p.isVerified 
  })))
  
  // 创建一个Set来存储已处理的预测ID，用于去重
  const processedMatchIds = new Set()
  
  // 格式化管理员调整积分历史
  const adminHistory = adminHistoryData.map(record => {
    // 对于与比赛关联的记录，记录其matchId以避免重复
    if (record.matchId) {
      processedMatchIds.add(record.matchId)
      console.log(`从管理员API添加匹配ID: ${record.matchId}, 标题: ${record.matchTitle || '无标题'}`);
    }
    
    return {
      id: record.id,
      points_change: record.pointsChange,
      points_after: 0, // 计算累计积分
      change_type: record.type === 'admin_adjustment' ? 'admin' : 'prediction',
      description: record.reason || '',
      related_match: record.matchTitle || (record.matchId ? `比赛 #${record.matchId}` : undefined),
      created_at: record.createdAt,
      source: 'admin_api' // 标记来源
    } as PointsHistory;
  })
  
  // 直接返回空；前端不再合并积分历史，完全依赖后端返回
  return []
}

// 获取总积分
const fetchTotalPoints = async () => {
  try {
    console.log('开始获取用户总积分...')
    
    const profileResp = await axios.get('/api/auth/profile', {
      headers: { Authorization: `Bearer ${userStore.token}` }
    })

    const val = profileResp.data?.points ?? profileResp.data?.data?.points ?? 0
    totalPoints.value = Number(val) || 0
  } catch (error) {
    console.error('获取总积分失败:', error)
    totalPoints.value = 0
  }
}

// 格式化日期
const formatDate = (dateString: string | Date) => {
  if (!dateString) return 'N/A'

  try {
    const date = new Date(dateString)
    if (isNaN(date.getTime())) return 'N/A'

    const year = date.getFullYear()
    const month = String(date.getMonth() + 1).padStart(2, '0')
    const day = String(date.getDate()).padStart(2, '0')
    const hours = String(date.getHours()).padStart(2, '0')
    const minutes = String(date.getMinutes()).padStart(2, '0')

    return `${year}-${month}-${day} ${hours}:${minutes}`
  } catch (error) {
    console.error('格式化日期出错:', error)
    return 'N/A'
  }
}

// 根据记录类型和积分变化获取描述
const getDescription = (item: PointsHistory): string => {
  // 如果已有描述，直接返回
  if (item.description) return item.description
  
  // 根据类型和积分生成描述
  if (item.change_type === 'admin') {
    return item.points_change > 0 
      ? `管理员增加了 ${item.points_change} 积分` 
      : `管理员减少了 ${Math.abs(item.points_change)} 积分`
  }
  
  if (item.change_type === 'prediction') {
    if (item.points_change === 5) return '预测比赛结果和比分全部正确'
    if (item.points_change === 3) return '预测获胜队伍正确'
    if (item.points_change === 1) return '预测比分正确'
    if (item.points_change === 0) return '预测失败'
  }
  
  return '积分变动'
}

// 获取气泡框类型
const getBubbleClass = (pointsChange: number, changeType: string): string => {
  // 如果是管理员修改积分，无论积分正负都使用黄色背景
  if (changeType === 'admin') return 'bubble-admin'
  
  if (pointsChange > 3) return 'bubble-success'
  if (pointsChange > 0) return 'bubble-primary'
  if (pointsChange < 0) return 'bubble-danger'
  return 'bubble-info'
}

// 返回页面
const goBack = () => {
  if (isAdminView.value) {
    // 管理员视图返回到用户管理页面
    router.push('/admin/users')
  } else {
    // 普通用户返回到个人资料页
    router.push('/profile')
  }
}

// 初始化
onMounted(() => {
  fetchPointsHistory()
})
</script>

<style scoped>
.points-history-container {
  max-width: 900px;
  margin: 0 auto;
  padding: 30px 20px;
}

.loading-container {
  padding: 40px 20px;
  background-color: #fff;
  border-radius: 12px;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.08);
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 30px;
}

.page-header h2 {
  margin: 0;
  font-size: 24px;
  color: #303133;
}

.history-card {
  background-color: #fff;
  border-radius: 12px;
  padding: 40px;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.08);
}

.total-points {
  text-align: center;
  margin-bottom: 30px;
  padding-bottom: 20px;
  border-bottom: 1px solid #eee;
}

.points-label {
  font-size: 18px;
  color: #606266;
  margin-bottom: 10px;
}

.points-value {
  font-size: 48px;
  font-weight: bold;
  color: #409eff;
}

.empty-history {
  padding: 40px 0;
  text-align: center;
}

.points-history-list {
  max-height: 600px;
  overflow-y: auto;
  padding: 10px 0;
  display: flex;
  flex-direction: column;
  gap: 15px;
}

.history-bubble {
  border-radius: 12px;
  padding: 15px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
  margin-bottom: 15px;
  transition: all 0.3s ease;
}

.history-bubble:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.12);
}

.bubble-admin {
  background-color: #fdf6ec;
  border-left: 4px solid #e6a23c;
}

.bubble-success {
  background-color: #f0f9eb;
  border-left: 4px solid #67c23a;
}

.bubble-primary {
  background-color: #ecf5ff;
  border-left: 4px solid #409eff;
}

.bubble-danger {
  background-color: #fef0f0;
  border-left: 4px solid #f56c6c;
}

.bubble-info {
  background-color: #f4f4f5;
  border-left: 4px solid #909399;
}

.history-content {
  display: flex;
  flex-direction: column;
}

.history-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.history-match {
  font-weight: bold;
  color: #303133;
  font-size: 16px;
}

.history-description {
  font-size: 15px;
  color: #606266;
  margin-bottom: 8px;
}

.history-timestamp {
  font-size: 13px;
  color: #909399;
  margin-top: 5px;
  align-self: flex-end;
}

.history-points {
  font-size: 18px;
  font-weight: bold;
  padding: 2px 8px;
  border-radius: 4px;
}

.history-points.positive {
  color: #67c23a;
  background-color: rgba(103, 194, 58, 0.1);
}

.history-points.negative {
  color: #f56c6c;
  background-color: rgba(245, 108, 108, 0.1);
}

/* 蓝色按钮样式 */
.blue-btn {
  background-color: #4DA1FF !important;
  border-color: #4DA1FF !important;
  color: white !important;
  border-radius: 4px !important;
  padding: 10px 20px !important;
  font-size: 14px !important;
  transition: all 0.3s ease !important;
  border: none !important;
  font-weight: normal !important;
}

.blue-btn:hover {
  background-color: #3A90F8 !important;
  box-shadow: 0 2px 8px rgba(77, 161, 255, 0.3) !important;
}

.blue-btn:active {
  background-color: #2980F5 !important;
}

@media (max-width: 768px) {
  .points-history-container {
    padding: 20px 15px;
  }

  .history-card {
    padding: 25px 20px;
  }

  .page-header {
    flex-direction: column;
    gap: 15px;
    margin-bottom: 20px;
  }

  .points-value {
    font-size: 36px;
  }
}
</style>
