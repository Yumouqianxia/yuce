3+6--+3
<template>
  <div class="match-list-container">
    <div class="page-header">
      <h1 class="page-title">比赛列表</h1>
      
      <!-- 管理员操作按钮 -->
      <AdminPermissionCheck permission="match.manage" :show-fallback="false">
        <div class="admin-actions">
          <el-button type="primary" @click="navigateToMatchManagement">
            <el-icon><Plus /></el-icon>
            创建比赛
          </el-button>
          <el-button type="default" @click="navigateToMatchManagement">
            <el-icon><Setting /></el-icon>
            比赛管理
          </el-button>
        </div>
      </AdminPermissionCheck>
    </div>

    <!-- 导航菜单 -->
    <div class="match-nav">
      <router-link :to="{ name: 'TodayMatches' }" class="nav-item" active-class="active">
        今日赛程
      </router-link>
      <router-link :to="{ name: 'UpcomingMatches' }" class="nav-item" active-class="active">
        未来赛程
      </router-link>
      <router-link :to="{ name: 'HistoryMatches' }" class="nav-item" active-class="active">
        历史比赛
      </router-link>
    </div>

    <div class="content-with-sidebar">
      <!-- 可折叠侧边栏 -->
      <CollapsibleSidebar :currentTab="getCurrentTab()" />

      <!-- 子路由视图 -->
      <div class="main-content">
        <router-view></router-view>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useRoute, useRouter } from 'vue-router'
import { Plus, Setting } from '@element-plus/icons-vue'
import CollapsibleSidebar from '@/components/matches/CollapsibleSidebar.vue'
import AdminPermissionCheck from '@/components/admin/AdminPermissionCheck.vue'

const route = useRoute()
const router = useRouter()

// 获取当前标签页
// 从路由中提取当前标签页（today, upcoming, history）
const getCurrentTab = () => {
  const path = route.path
  if (path.includes('/today')) return 'today'
  if (path.includes('/upcoming')) return 'upcoming'
  if (path.includes('/history')) return 'history'
  return 'today' // 默认值
}

// 导航到比赛管理页面
const navigateToMatchManagement = () => {
  router.push('/admin/matches')
}
</script>

<style scoped>
.match-list-container {
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 20px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 30px;
}

.admin-actions {
  display: flex;
  gap: 12px;
}

.content-with-sidebar {
  display: flex;
  flex-direction: column;
  gap: 20px;
  margin-top: 20px;
}

.main-content {
  flex: 1;
}

.page-title {
  font-size: 28px;
  font-weight: 600;
  margin: 0;
  color: #303133;
}

.match-nav {
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

/* 响应式调整 */
@media (max-width: 768px) {
  .page-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 16px;
  }
  
  .admin-actions {
    width: 100%;
    justify-content: flex-end;
  }
  
  .match-nav {
    flex-direction: column;
    align-items: center;
  }

  .nav-item {
    margin: 5px 0;
    width: 90%;
    text-align: center;
    box-sizing: border-box;
  }

  .page-title {
    font-size: 24px;
  }
}
</style>