<template>
  <div class="leaderboard-container">
    <h1 class="page-title">积分排行榜</h1>
    <div class="leaderboard-content">
      <el-table
        v-loading="loading"
        :data="leaderboardData"
        style="width: 100%"
      >
        <el-table-column
          prop="rank"
          label="排名"
          width="100"
        >
          <template #default="{ $index }">
            <div class="rank-cell" :class="{ 'top-three': $index < 3 }">
              {{ $index + 1 + (currentPage - 1) * pageSize }}
            </div>
          </template>
        </el-table-column>

        <el-table-column
          prop="username"
          label="用户"
          min-width="200"
        >
          <template #default="{ row }">
            <div class="user-info">
              <div class="avatar" v-if="row.avatar">
                <img :src="row.avatar" alt="头像" />
              </div>
              <div class="username">{{ row.nickname || row.username }}</div>
            </div>
          </template>
        </el-table-column>

        <el-table-column
          prop="points"
          label="积分"
          sortable
          width="120"
        >
          <template #default="{ row }">
            <span class="points">{{ row.points }}</span>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        v-if="totalUsers > 0"
        layout="total, prev, pager, next"
        :total="totalUsers"
        :page-size="pageSize"
        :current-page="currentPage"
        @current-change="handlePageChange"
        class="pagination"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { ElTable, ElTableColumn, ElPagination } from 'element-plus'
import { getLeaderboard, LeaderboardUser } from '@/api/leaderboard'

// 状态
const loading = ref(true)
const allUsers = ref<LeaderboardUser[]>([])
const leaderboardData = computed(() => {
  const startIndex = (currentPage.value - 1) * pageSize.value
  const endIndex = startIndex + pageSize.value
  return allUsers.value.slice(startIndex, endIndex)
})
const totalUsers = computed(() => allUsers.value.length)
const currentPage = ref(1)
const pageSize = ref(10)

// 获取排行榜数据
const fetchLeaderboard = async () => {
  loading.value = true
  try {
    const users = await getLeaderboard()
    // 按积分排序
    allUsers.value = users.sort((a, b) => b.points - a.points)
  } catch (error) {
    console.error('获取排行榜失败', error)
  } finally {
    loading.value = false
  }
}

// 处理页码变化
const handlePageChange = (page: number) => {
  currentPage.value = page
}

// 初始化
onMounted(() => {
  fetchLeaderboard()
})
</script>

<style scoped>
.leaderboard-container {
  max-width: 800px;
  margin: 0 auto;
}

.page-title {
  font-size: 24px;
  font-weight: 600;
  margin-bottom: 20px;
  color: #303133;
}

.leaderboard-content {
  background-color: #fff;
  border-radius: 8px;
  padding: 20px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
}

.rank-cell {
  width: 36px;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: #f5f7fa;
  border-radius: 50%;
  font-weight: bold;
  color: #606266;
}

.rank-cell.top-three {
  background-color: #409eff;
  color: #fff;
}

.user-info {
  display: flex;
  align-items: center;
}

.avatar {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  overflow: hidden;
  margin-right: 12px;
}

.avatar img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.username {
  font-weight: 500;
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