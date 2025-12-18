<template>
  <div class="navbar">
    <div class="navbar-container">
      <div class="logo">
        <a href="/">
          <span class="logo-text">预测系统</span>
        </a>
      </div>
      <div class="user-actions-mobile">
        <template v-if="isAuthenticated">
          <div class="user-select-container-mobile">
            <div class="user-avatar-mobile" @click="toggleMobileUserMenu">
              <img v-if="userAvatar" :src="userAvatar" alt="用户头像">
              <span v-else>{{ userInitial }}</span>
            </div>
            <div v-if="showMobileUserMenu" class="user-dropdown-menu-mobile">
              <div class="dropdown-item" @click="navigateToExternal('/profile'); closeMobileMenu();">
                <div class="dropdown-icon">•</div>
                <span>个人资料</span>
              </div>
              <div class="dropdown-item" @click="navigateToExternal('/prediction-history'); closeMobileMenu();">
                <div class="dropdown-icon">•</div>
                <span>我的预测</span>
              </div>
              <div v-if="isAdmin" class="dropdown-item" @click="navigateToExternal('/admin'); closeMobileMenu();">
                <div class="dropdown-icon">•</div>
                <span>管理中心</span>
              </div>
              <div class="dropdown-divider"></div>
              <div class="dropdown-item" @click="logout(); closeMobileMenu();">
                <div class="dropdown-icon">•</div>
                <span>退出登录</span>
              </div>
            </div>
          </div>
        </template>
        <template v-else>
          <a href="/login" class="btn-mobile btn-primary-mobile">登录</a>
        </template>
        <div class="mobile-nav-toggle">
          <hamburger-icon :is-active="showMobileMenu" @toggle="toggleMobileMenu" />
        </div>
      </div>
      <div class="menu" :class="{ 'mobile-menu-active': showMobileMenu }">
        <div class="menu-list">
          <a href="/" class="menu-item" :class="{ active: activeMenu === '/' }" @click="closeMobileMenu">
            <nav-icon icon="home" />首页
          </a>
          <a href="/matches" class="menu-item" :class="{ active: activeMenu.includes('/matches') }" @click="closeMobileMenu">
            <nav-icon icon="trophy" />比赛列表
          </a>

          <a v-if="isAuthenticated" href="/upcoming-matches" class="menu-item" :class="{ active: activeMenu.includes('/upcoming-matches') || activeMenu.includes('/prediction-history') || activeMenu.includes('/prediction-rules') }" @click="closeMobileMenu">
            <nav-icon icon="trophy" />比赛预测
          </a>

          <a href="/leaderboard" class="menu-item" :class="{ active: activeMenu.includes('/leaderboard') }" @click="closeMobileMenu">
            <nav-icon icon="chart-bar" />积分排行
          </a>
          
          <!-- 管理中心下拉菜单 -->
          <div v-if="isAdmin" class="dropdown-trigger menu-item" :class="{ active: activeMenu.includes('/admin') }" @click="toggleAdminMenu">
            <nav-icon icon="cog" />管理中心
            <div v-if="showAdminMenu" class="dropdown-menu admin-dropdown">
              <PermissionWrapper permission="sport_type.manage">
                <a href="/admin/sport-types" class="dropdown-item" @click="closeMobileMenu">
                  <div class="dropdown-icon">🏃</div>
                  <span>运动类型管理</span>
                </a>
              </PermissionWrapper>
              
              <PermissionWrapper permission="scoring_rule.manage">
                <a href="/admin/scoring-rules" class="dropdown-item" @click="closeMobileMenu">
                  <div class="dropdown-icon">📊</div>
                  <span>积分规则配置</span>
                </a>
              </PermissionWrapper>
              
              <PermissionWrapper permission="match.manage">
                <a href="/admin/matches" class="dropdown-item" @click="closeMobileMenu">
                  <div class="dropdown-icon">🏆</div>
                  <span>比赛管理</span>
                </a>
              </PermissionWrapper>
              
              <PermissionWrapper permission="user.manage">
                <a href="/admin/users" class="dropdown-item" @click="closeMobileMenu">
                  <div class="dropdown-icon">👥</div>
                  <span>用户管理</span>
                </a>
              </PermissionWrapper>
              
              <PermissionWrapper permission="admin.manage">
                <a href="/admin/admins" class="dropdown-item" @click="closeMobileMenu">
                  <div class="dropdown-icon">👨‍💼</div>
                  <span>管理员管理</span>
                </a>
              </PermissionWrapper>
              
              <div class="dropdown-divider"></div>
              
              <PermissionWrapper permission="audit_log.view">
                <a href="/admin/audit-logs" class="dropdown-item" @click="closeMobileMenu">
                  <div class="dropdown-icon">📋</div>
                  <span>审计日志</span>
                </a>
              </PermissionWrapper>
              
              <PermissionWrapper permission="system.config" :show-fallback="false">
                <a href="/admin/system-config" class="dropdown-item" @click="closeMobileMenu">
                  <div class="dropdown-icon">⚙️</div>
                  <span>系统配置</span>
                </a>
              </PermissionWrapper>
              
              <!-- 保留原有的站点管理入口 -->
              <div class="dropdown-divider"></div>
              <a href="/admin" class="dropdown-item" @click="closeMobileMenu">
                <div class="dropdown-icon">🏠</div>
                <span>站点管理</span>
              </a>
            </div>
          </div>
        </div>

        <div class="user-actions">
          <template v-if="isAuthenticated">
            <div class="select-container user-select-container">
              <div class="select-field user-select-field" @click="toggleUserMenu">
                <div class="user-select-content">
                  <div class="user-avatar" v-if="userAvatar">
                    <img :src="userAvatar" alt="用户头像">
                  </div>
                  <div class="user-avatar" v-else>{{ userInitial }}</div>
                  <span class="username-text">{{ username }}</span>
                </div>
              </div>
              <div v-if="showUserMenu" class="user-dropdown-menu">
                <div class="dropdown-item" @click="navigateToExternal('/profile'); closeMobileMenu();">
                  <div class="dropdown-icon">•</div>
                  <span>个人资料</span>
                </div>
                <div class="dropdown-item" @click="navigateToExternal('/prediction-history'); closeMobileMenu();">
                  <div class="dropdown-icon">•</div>
                  <span>我的预测</span>
                </div>
                <div v-if="isAdmin" class="dropdown-item" @click="navigateToExternal('/admin'); closeMobileMenu();">
                  <div class="dropdown-icon">•</div>
                  <span>管理中心</span>
                </div>
                <div class="dropdown-divider"></div>
                <div class="dropdown-item" @click="logout(); closeMobileMenu();">
                  <div class="dropdown-icon">•</div>
                  <span>退出登录</span>
                </div>
              </div>
            </div>
          </template>
          <template v-else>
            <a href="/login" class="btn btn-primary" @click="closeMobileMenu">登录</a>
            <a href="/register" class="btn btn-outline" @click="closeMobileMenu">注册</a>
          </template>
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts" setup>
import { computed, ref, onMounted, onBeforeUnmount, inject, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { usePermissionStore } from '@/stores/permissions'
import NavIcon from '@/components/base/NavIcon.vue'
import HamburgerIcon from '@/components/base/HamburgerIcon.vue'
import PermissionWrapper from '@/components/base/PermissionWrapper.vue'


// 不需要使用router，因为我们使用普通链接导航
const route = useRoute()
const userStore = useUserStore()
const permissionStore = usePermissionStore()

// 菜单状态
const showUserMenu = ref(false) // 桌面端用户菜单
const showMobileUserMenu = ref(false) // 移动端用户菜单
const showMobileMenu = ref(false) // 移动端主菜单
const showAdminMenu = ref(false) // 管理员下拉菜单

// 从 App.vue 注入移动端菜单状态
const appMobileMenuOpen = inject('mobileMenuOpen', ref(false))

// 切换桌面端用户菜单
const toggleUserMenu = () => {
  showUserMenu.value = !showUserMenu.value
  // 关闭移动端用户菜单
  showMobileUserMenu.value = false
}

// 切换移动端用户菜单
const toggleMobileUserMenu = () => {
  showMobileUserMenu.value = !showMobileUserMenu.value
  // 关闭桌面端用户菜单
  showUserMenu.value = false
}

// 切换移动端菜单
const toggleMobileMenu = () => {
  showMobileMenu.value = !showMobileMenu.value
  // 同步到App级别的状态
  appMobileMenuOpen.value = showMobileMenu.value
}

// 切换管理员菜单
const toggleAdminMenu = () => {
  showAdminMenu.value = !showAdminMenu.value
  // 关闭其他菜单
  showUserMenu.value = false
  showMobileUserMenu.value = false
}

// 关闭移动端菜单
const closeMobileMenu = () => {
  showMobileMenu.value = false
  showAdminMenu.value = false
  // 同步到App级别的状态
  appMobileMenuOpen.value = false
}

// 点击外部关闭菜单
const handleClickOutside = (event: MouseEvent) => {
  // 处理用户菜单
  const userSelectContainer = document.querySelector('.user-select-container')
  const userSelectContainerMobile = document.querySelector('.user-select-container-mobile')
  const adminDropdownTrigger = document.querySelector('.dropdown-trigger')

  // 处理桌面端用户菜单
  if (userSelectContainer && event.target instanceof Node && !userSelectContainer.contains(event.target)) {
    showUserMenu.value = false
  }

  // 处理移动端用户菜单
  if (userSelectContainerMobile && event.target instanceof Node && !userSelectContainerMobile.contains(event.target)) {
    showMobileUserMenu.value = false
  }

  // 处理管理员下拉菜单
  if (adminDropdownTrigger && event.target instanceof Node && !adminDropdownTrigger.contains(event.target)) {
    showAdminMenu.value = false
  }

  // 处理移动端菜单 - 只在点击非菜单区域和非汉堡按钮时关闭
  const mobileMenuContainer = document.querySelector('.menu')
  const hamburgerIcon = document.querySelector('.hamburger-icon')
  if (showMobileMenu.value && mobileMenuContainer && hamburgerIcon &&
      event.target instanceof Node &&
      !mobileMenuContainer.contains(event.target) &&
      !hamburgerIcon.contains(event.target)) {
    showMobileMenu.value = false
    appMobileMenuOpen.value = false
  }
}

// 监听 App 级别的菜单状态变化
const syncWithAppMenuState = () => {
  if (showMobileMenu.value !== appMobileMenuOpen.value) {
    showMobileMenu.value = appMobileMenuOpen.value
  }
}

// 添加和移除点击事件监听器
onMounted(() => {
  document.addEventListener('click', handleClickOutside)

  // 监听 App 级别的菜单状态变化
  watch(appMobileMenuOpen, syncWithAppMenuState)
})

onBeforeUnmount(() => {
  document.removeEventListener('click', handleClickOutside)
})

// 当前活动菜单
const activeMenu = computed(() => route.path)

// 用户认证状态和用户名
const isAuthenticated = computed(() => userStore.isAuthenticated)
const username = computed(() => userStore.displayName)
const userAvatar = computed(() => {
  const avatar = userStore.user?.avatar || ''
  if (!avatar) return ''

  // 已是完整 URL
  if (avatar.startsWith('http')) return avatar

  // 统一转成 /api/uploads/avatar/{filename}[?cacheBust]
  const [pathPart, query] = avatar.split('?')
  const filename = pathPart.split('/').pop() || ''
  if (!filename) return ''
  return `/api/uploads/avatar/${filename}${query ? `?${query}` : ''}`
})
const userInitial = computed(() => {
  const name = userStore.displayName
  return name ? name.charAt(0).toUpperCase() : 'U'
})

// 判断是否是管理员
const isAdmin = computed(() => userStore.user?.role === 'admin')

// 导航方法（使用普通链接导航而不是router）
const navigateToExternal = (path: string) => {
  window.location.href = path
}

// 登出方法
const logout = () => {
  try {
    userStore.logout()
    window.location.href = '/login'
  } catch (error) {
    console.error('登出失败', error)
  }
}
</script>

<style scoped>
.navbar {
  height: 64px;
  width: 100%;
  background-color: var(--bg-white);
}

.navbar-container {
  max-width: 1200px;
  margin: 0 auto;
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 100%;
  padding: 0 24px;
  position: relative;
}

.logo {
  font-size: 22px;
  font-weight: bold;
}

.logo a {
  text-decoration: none;
  display: flex;
  align-items: center;
}

.logo-text {
  color: var(--primary-color);
  background: linear-gradient(to right, var(--primary-color), var(--primary-dark));
  -webkit-background-clip: text;
  background-clip: text;
  -webkit-text-fill-color: transparent;
  font-weight: 700;
}

.menu {
  display: flex;
  align-items: center;
  flex: 1;
  justify-content: space-between;
  margin-left: 40px;
}

.mobile-nav-toggle {
  display: none;
  z-index: 1001;
}

.user-actions-mobile {
  display: none;
  margin-left: auto;
}

.menu-list {
  display: flex;
  flex-direction: row;
  height: 64px;
  align-items: center;
}

.menu-item {
  height: 64px;
  font-size: 15px;
  display: flex;
  align-items: center;
  padding: 0 20px;
  color: var(--text-secondary);
  text-decoration: none;
  position: relative;
  transition: color 0.3s;
}

.menu-item:hover {
  color: var(--primary-color);
}

.menu-item.active {
  color: var(--primary-color);
}

.menu-item.active::after {
  content: '';
  position: absolute;
  bottom: 0;
  left: 10%;
  right: 10%;
  height: 2px;
  background-color: var(--primary-color);
}

.dropdown-trigger {
  position: relative;
  cursor: pointer;
}

.dropdown-menu {
  position: absolute;
  top: 64px;
  left: 0;
  min-width: 180px;
  background-color: #fff;
  border-radius: 12px;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.15);
  z-index: 1000;
  padding: 8px 0;
  border: 1px solid #dcdfe6;
}

.matches-dropdown,
.leaderboard-dropdown {
  width: 200px;
}

.admin-dropdown {
  width: 220px;
}

.dropdown-menu .dropdown-item {
  display: flex;
  align-items: center;
  padding: 10px 16px;
  cursor: pointer;
  font-size: 14px;
  color: #606266;
  transition: all 0.3s;
  margin: 0 6px;
  border-radius: 8px;
  text-decoration: none;
}

.dropdown-menu .dropdown-item:hover {
  background-color: #f0f7ff;
  color: #409eff;
}

.dropdown-menu .dropdown-item.active {
  background-color: #ecf5ff;
  color: #409eff;
  font-weight: 500;
}

.user-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.user-select-container {
  position: relative;
  width: auto;
  min-width: 120px;
}

.user-select-field {
  display: flex;
  align-items: center;
  justify-content: space-between;
  cursor: pointer;
  padding: 4px 16px;
  height: 38px;
  box-sizing: border-box;
  border: 1px solid #dcdfe6;
  border-radius: 19px;
  background-color: #fff;
  color: #606266;
  transition: all 0.3s;
}

.user-select-field:hover {
  border-color: var(--primary-color);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.user-select-content {
  display: flex;
  align-items: center;
  gap: 8px;
}

.user-avatar {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  background-color: var(--primary-color);
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 14px;
  overflow: hidden;
  margin-right: 2px;
}

.user-avatar img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.username-text {
  font-size: 14px;
  color: var(--text-primary);
}

.user-dropdown-menu {
  position: absolute;
  top: calc(100% + 5px);
  left: 0;
  right: 0;
  min-width: 150px;
  background-color: #fff;
  border-radius: 12px;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.15);
  z-index: 1000;
  padding: 8px 0;
  border: 1px solid #dcdfe6;
}

.dropdown-item {
  display: flex;
  align-items: center;
  padding: 10px 16px;
  cursor: pointer;
  font-size: 14px;
  color: #606266;
  transition: all 0.3s;
  margin: 0 6px;
  border-radius: 8px;
}

.dropdown-item:hover {
  background-color: #f0f7ff;
  color: #409eff;
}

.dropdown-icon {
  margin-right: 8px;
  font-size: 16px;
  color: #909399;
}

.dropdown-divider {
  height: 1px;
  background-color: #ebeef5;
  margin: 5px 0;
}

.select-container::after {
  content: "\25BC";
  font-size: 10px;
  color: #C0C4CC;
  position: absolute;
  right: 18px;
  top: 50%;
  transform: translateY(-50%);
  pointer-events: none;
  transition: all 0.3s;
}

.user-select-container:hover::after {
  color: var(--primary-color);
}

.btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  height: 36px;
  padding: 0 16px;
  border-radius: 18px;
  font-size: 14px;
  font-weight: 500;
  text-decoration: none;
  transition: all 0.3s;
}

.btn-primary {
  background-color: var(--primary-color);
  color: white;
}

.btn-primary:hover {
  background-color: var(--primary-dark);
}

.btn-outline {
  border: 1px solid var(--border-color);
  color: var(--text-secondary);
  background-color: transparent;
}

.btn-outline:hover {
  border-color: var(--primary-color);
  color: var(--primary-color);
}

/* 移动端样式 */
@media (max-width: 768px) {
  .mobile-nav-toggle {
    display: block;
  }

  .user-actions-mobile {
    display: flex;
    align-items: center;
    gap: 15px;
  }

  .user-select-container-mobile {
    position: relative;
  }

  .user-avatar-mobile {
    width: 32px;
    height: 32px;
    border-radius: 50%;
    background-color: var(--primary-color);
    color: white;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 14px;
    overflow: hidden;
    cursor: pointer;
  }

  .user-avatar-mobile img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }

  .user-dropdown-menu-mobile {
    position: absolute;
    top: calc(100% + 5px);
    right: 0;
    min-width: 150px;
    background-color: #fff;
    border-radius: 12px;
    box-shadow: 0 4px 16px rgba(0, 0, 0, 0.15);
    z-index: 1002;
    padding: 8px 0;
    border: 1px solid #dcdfe6;
  }

  .btn-mobile {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    height: 32px;
    padding: 0 12px;
    border-radius: 16px;
    font-size: 13px;
    font-weight: 500;
    text-decoration: none;
    transition: all 0.3s;
  }

  .btn-primary-mobile {
    background-color: var(--primary-color);
    color: white;
  }

  .menu {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background-color: var(--bg-white);
    flex-direction: column;
    justify-content: flex-start;
    padding-top: 64px;
    margin-left: 0;
    transform: translateX(100%);
    transition: transform 0.3s ease;
    z-index: 1000;
    overflow-y: auto;
  }

  .mobile-menu-active {
    transform: translateX(0);
  }

  .menu-list {
    flex-direction: column;
    width: 100%;
    height: auto;
    padding: 20px 0;
  }

  .menu-item {
    height: auto;
    padding: 15px 24px;
    width: 100%;
    box-sizing: border-box;
    border-bottom: 1px solid var(--border-color);
  }

  .menu-item.active::after {
    display: none;
  }

  .menu-item.active {
    background-color: var(--bg-light);
  }

  .user-actions {
    width: 100%;
    padding: 20px 24px;
    justify-content: center;
    border-top: 1px solid var(--border-color);
  }

  .user-select-container {
    width: 100%;
    max-width: 300px;
  }

  .user-dropdown-menu {
    width: 100%;
    max-width: 300px;
  }
}
</style>