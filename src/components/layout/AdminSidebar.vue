<template>
  <div class="admin-sidebar">
    <div class="sidebar-header">
      <h3>管理面板</h3>
    </div>
    <ul class="sidebar-menu">
      <li class="sidebar-item" :class="{ active: $route.name === 'AdminSite' }">
        <a href="/admin/site">公告管理</a>
      </li>
      <li class="sidebar-item" :class="{ active: $route.name === 'AdminUsers' }">
        <a href="/admin/users">用户管理</a>
      </li>
      <li class="sidebar-item" :class="{ active: $route.name === 'AdminMatches' }">
        <a href="/admin/matches">比赛管理</a>
      </li>
      <li class="sidebar-item" :class="{ active: $route.name === 'AdminSettings' }">
        <a href="/admin/settings">系统设置</a>
      </li>
    </ul>
  </div>
</template>

<script setup lang="ts">
// 直接使用$route，不需要导入useRoute
</script>

<style scoped>
.admin-sidebar {
  width: 180px;
  position: fixed;
  top: 64px; /* 导航栏高度 */
  left: 0;
  height: calc(100vh - 64px);
  background-color: #fff;
  border-right: 1px solid #e6e6e6;
  z-index: 1000;
}

.sidebar-header {
  padding: 16px;
  border-bottom: 1px solid #e6e6e6;
}

.sidebar-header h3 {
  margin: 0;
  font-size: 16px;
  color: #303133;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.sidebar-menu {
  list-style: none;
  padding: 0;
  margin: 0;
}

.sidebar-item {
  margin-bottom: 1px;
  width: 100%;
}

.sidebar-item a {
  display: block;
  width: 100%;
  text-align: left;
  padding: 12px 16px;
  text-decoration: none;
  border: none;
  background-color: transparent;
  font-size: 14px;
  color: #606266;
  cursor: pointer;
  transition: all 0.3s;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  box-sizing: border-box;
}

.sidebar-item a:hover {
  background-color: #f5f7fa;
  color: #409EFF;
}

.sidebar-item.active a {
  background-color: #f0f9ff;
  color: #409EFF;
  border-left: 3px solid #409EFF;
}
</style>