<template>
  <div class="profile-container">
    <div v-if="loading" class="loading-container">
      <el-skeleton :rows="10" animated />
    </div>

    <div v-else class="profile-content">
      <!-- 用户信息卡片 -->
      <div class="user-profile-card">
        <div class="profile-header">
          <div class="avatar-container">
            <div class="avatar">
              <img v-if="userInfo.avatar" :src="userInfo.avatar" alt="用户头像">
              <div v-else class="avatar-placeholder">
                {{ getInitial(userInfo.nickname || userInfo.username) }}
              </div>
            </div>
          </div>

          <div class="user-info">
            <h2 class="user-nickname">{{ userInfo.nickname || '未设置昵称' }}</h2>
            <div class="user-username">@{{ userInfo.username }}</div>
            <div class="user-register-time">注册时间: {{ formatDate(userInfo.createdAt) }}</div>

            <el-button class="blue-btn edit-profile-btn" @click="navigateTo('/profile/edit')">
              编辑资料
            </el-button>
          </div>
        </div>

        <!-- 用户统计信息 -->
        <div class="user-stats">
          <div class="stat-box">
            <div class="stat-number">{{ userStats.total_points || 0 }}</div>
            <div class="stat-label">总积分</div>
            <el-button class="blue-btn view-history-btn" @click="navigateTo('/profile/points-history')">
              查看积分历史
            </el-button>
          </div>

          <div class="stat-box">
            <div class="stat-number">{{ userStats.total_predictions || 0 }}</div>
            <div class="stat-label">预测次数</div>
          </div>

          <div class="stat-box">
            <div class="stat-number">{{ userStats.accurate_predictions || 0 }}</div>
            <div class="stat-label">正确预测</div>
          </div>

          <div class="stat-box">
            <div class="stat-number">{{ calculateAccuracy }}%</div>
            <div class="stat-label">准确率</div>
          </div>
        </div>

        <!-- 密码管理按钮已移至编辑资料页面 -->
      </div>

      <!-- 对话框已移除，改为单独页面 -->
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElSkeleton, ElButton } from 'element-plus'
import { useUserStore } from '@/stores/user'
import { getUserProfile } from '@/api/user'
import { get } from '@/api/http'
import type { UserProfile, UserStats } from '@/types/user'
import { getFullAvatarUrl } from '@/utils/url'

const userStore = useUserStore()
const router = useRouter()

// 导航到指定页面
const navigateTo = (path: string) => {
  router.push(path)
}

// 头像相关配置
// 已移除默认头像设置，使用首字母占位符

// 状态
const loading = ref(true)

// 用户信息
const userInfo = ref<UserProfile>({
  username: '',
  nickname: '',
  email: '',
  avatar: '',
})

// 用户统计信息
const userStats = ref<UserStats>({
  total_points: 0,
  rank: 0,
  total_predictions: 0,
  accurate_predictions: 0,
})

// 移除不再需要的表单数据和验证规则

// 计算准确率
const calculateAccuracy = computed(() => {
  if (!userStats.value.total_predictions || userStats.value.total_predictions === 0) {
    return 0
  }

  return Math.round((userStats.value.accurate_predictions / userStats.value.total_predictions) * 100)
})

// 定义用户资料接口
interface UserProfileResponse {
  id?: number;
  username?: string;
  nickname?: string;
  email?: string;
  avatar?: string;
  points?: number;
  createdAt?: string;
  [key: string]: any;
}

// 定义积分响应接口
interface PointsResponse {
  success?: boolean;
  message?: string;
  data?: number;
  timestamp?: string;
  statusCode?: number;
  points?: number;
  [key: string]: any;
}

// 获取用户信息
const fetchUserProfile = async () => {
  loading.value = true
  try {
    // 获取用户信息
    const userData = await getUserProfile()
    console.log('获取到的用户数据:', userData)

    // 处理头像 URL
    if (userData.avatar) {
      // 使用工具函数获取完整的头像 URL
      userData.avatar = getFullAvatarUrl(userData.avatar)
      console.log('处理后的头像 URL:', userData.avatar)

      // 更新全局状态中的头像 URL
      userStore.updateUserInfo({
        avatar: userData.avatar
      })
    }

    // 处理返回的用户数据
    // 确保类型兼容性
    userInfo.value = {
      username: userData.username || '',
      nickname: userData.nickname || '',
      email: userData.email || '',
      avatar: userData.avatar || '',
      createdAt: userData.createdAt || userData.created_at || '',
      id: userData.id || 0
    }

    // 调试日志
    console.log('原始用户数据:', userData)

    // 调试日志
    console.log('处理后的用户数据:', userInfo.value)

    // 获取用户统计信息（传递 profile 数据以避免重复请求）
    await fetchUserStats(userData)
  } catch (error) {
    console.error('获取用户信息失败:', error)
    ElMessage.error('获取用户信息失败')
  } finally {
    loading.value = false
  }
}

// 获取用户统计信息（赛季积分直接用 profile.points，统计用预测数据）
const fetchUserStats = async (profileData?: any) => {
  try {
    // 积分直接用 profile 中的 points 作为赛季积分
    const totalPoints = typeof profileData?.points === 'number'
      ? profileData.points
      : (parseInt(String(profileData?.points)) || 0)

    // 预测统计
    const predictions = await get('/api/predictions/my')
    const totalPredictions = Array.isArray(predictions) ? predictions.length : 0
    const accuratePredictions = Array.isArray(predictions)
      ? predictions.filter(p => p.isCorrect).length
      : 0

    // 排名（可选，从排行榜接口获取，如果失败则置 0）
    let rank = 0
    if (profileData?.id) {
      try {
        const rankResp = await get<{ rank?: number }>(`/api/leaderboard/users/${profileData.id}/rank`)
        if (rankResp && typeof rankResp.rank === 'number') {
          rank = rankResp.rank
        }
      } catch (e) {
        console.warn('获取排行榜排名失败，忽略', e)
      }
    }

    userStats.value = {
      total_points: totalPoints,
      rank,
      total_predictions: totalPredictions,
      accurate_predictions: accuratePredictions
    }
  } catch (error) {
    console.error('获取用户统计信息失败:', error)
    // 出错时设置默认值
    userStats.value = {
      total_points: 0,
      rank: 0,
      total_predictions: 0,
      accurate_predictions: 0
    }
  }
}

// 这些方法将移至各自的页面组件中

// 格式化日期
const formatDate = (dateString?: string | Date) => {
  if (!dateString) return 'N/A'

  try {
    const date = new Date(dateString)
    if (isNaN(date.getTime())) return 'N/A'

    const year = date.getFullYear()
    const month = String(date.getMonth() + 1).padStart(2, '0')
    const day = String(date.getDate()).padStart(2, '0')

    return `${year}-${month}-${day}`
  } catch (error) {
    console.error('格式化日期出错:', error)
    return 'N/A'
  }
}

// 获取昵称或用户名的首字母
const getInitial = (name: string): string => {
  if (!name) return '?'

  // 如果是中文，返回第一个字
  if (/[\u4e00-\u9fa5]/.test(name.charAt(0))) {
    return name.charAt(0)
  }

  // 如果是英文，返回首字母大写
  return name.charAt(0).toUpperCase()
}

// 初始化
onMounted(() => {
  fetchUserProfile()
})
</script>

<style scoped>
.profile-container {
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

.user-profile-card {
  background-color: #fff;
  border-radius: 12px;
  padding: 40px;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.08);
  margin-bottom: 20px;
  transition: all 0.3s ease;
}

.profile-header {
  display: flex;
  align-items: center;
  margin-bottom: 40px;
}

.avatar-container {
  margin-right: 40px;
}

.avatar {
  width: 80px;
  height: 80px;
  border-radius: 50%;
  overflow: hidden;
}

.avatar img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.avatar-placeholder {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: #4DA1FF;
  color: white;
  font-size: 32px;
  font-weight: bold;
}

.user-info {
  flex: 1;
}

.user-nickname {
  font-size: 28px;
  font-weight: bold;
  margin: 0 0 8px 0;
  color: #303133;
}

.user-username {
  font-size: 16px;
  color: #606266;
  margin-bottom: 12px;
}

.user-register-time {
  font-size: 14px;
  color: #909399;
  margin-bottom: 20px;
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

.edit-profile-btn {
  margin-top: 10px;
}

.user-stats {
  display: flex;
  justify-content: space-between;
  padding: 30px 0;
  border-top: 1px solid #eee;
  margin-bottom: 20px;
}

.stat-box {
  flex: 1;
  text-align: center;
  padding: 25px 20px;
  border-radius: 12px;
  background-color: #f8f9fa;
  margin: 0 15px;
  display: flex;
  flex-direction: column;
  align-items: center;
  transition: all 0.3s ease;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
}

.stat-box:hover {
  transform: translateY(-5px);
  box-shadow: 0 6px 16px rgba(0, 0, 0, 0.1);
}

.stat-box:first-child {
  margin-left: 0;
}

.stat-box:last-child {
  margin-right: 0;
}

.stat-number {
  font-size: 42px;
  font-weight: bold;
  color: #409eff;
  margin-bottom: 15px;
}

.stat-label {
  font-size: 16px;
  color: #606266;
  margin-bottom: 20px;
}

.view-history-btn {
  margin-top: auto;
  transition: all 0.3s ease;
}

.password-management {
  display: flex;
  justify-content: center;
  padding-top: 10px;
}

/* 对话框样式 */
.custom-dialog :deep(.el-dialog__header) {
  padding: 20px 24px;
  border-bottom: 1px solid #eee;
}

.custom-dialog :deep(.el-dialog__body) {
  padding: 30px 24px;
}

.avatar-section {
  display: flex;
  flex-direction: column;
  align-items: center;
  margin-bottom: 30px;
}

.avatar-uploader {
  margin-top: 20px;
}

.form-buttons {
  margin-top: 30px;
  display: flex;
  justify-content: flex-end;
}

/* 积分历史样式 */
.loading-history {
  padding: 30px;
}

.points-history-list {
  max-height: 500px;
  overflow-y: auto;
  padding: 10px 0;
}

.history-content {
  display: flex;
  flex-direction: column;
  padding: 5px 0;
}

.history-match {
  font-weight: bold;
  margin-bottom: 8px;
  color: #303133;
}

.history-description {
  font-size: 15px;
  color: #606266;
  margin-bottom: 8px;
}

.history-points {
  font-size: 18px;
  font-weight: bold;
  align-self: flex-end;
}

.history-points.positive {
  color: #67c23a;
}

.history-points.negative {
  color: #f56c6c;
}

:deep(.el-timeline-item__node--success) {
  background-color: #67c23a;
}

:deep(.el-timeline-item__node--danger) {
  background-color: #f56c6c;
}

@media (max-width: 768px) {
  .profile-container {
    padding: 20px 15px;
  }

  .user-profile-card {
    padding: 25px 20px;
  }

  .profile-header {
    flex-direction: column;
    text-align: center;
  }

  .avatar-container {
    margin-right: 0;
    margin-bottom: 25px;
  }

  .user-stats {
    flex-direction: column;
    padding: 20px 0;
  }

  .stat-box {
    margin: 10px 0;
    padding: 20px;
  }

  .stat-number {
    font-size: 36px;
  }
}
</style>