// 路由守卫工具函数
import type { RouteLocationNormalized, NavigationGuardNext } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useUserStore } from '@/stores/user'
import { usePermissionStore } from '@/stores/permissions'

/**
 * 认证检查守卫
 */
export async function authGuard(
  to: RouteLocationNormalized,
  _from: RouteLocationNormalized,
  next: NavigationGuardNext
): Promise<void> {
  const userStore = useUserStore()

  // 如果用户已登录但访问的是仅限游客的页面
  if (to.meta.guest && userStore.isAuthenticated) {
    next({ name: 'Home' })
    return
  }

  // 验证是否需要登录
  if (to.meta.requiresAuth && !userStore.isAuthenticated) {
    next({ name: 'Login', query: { redirect: to.fullPath } })
    return
  }

  next()
}

/**
 * 管理员权限检查守卫
 */
export async function adminGuard(
  to: RouteLocationNormalized,
  _from: RouteLocationNormalized,
  next: NavigationGuardNext
): Promise<void> {
  const userStore = useUserStore()

  // 如果不需要管理员权限，直接通过
  if (!to.meta.requiresAdmin) {
    next()
    return
  }

  // 验证管理员身份
  if (!userStore.isAuthenticated || !userStore.isAdmin) {
    ElMessage.error('您没有管理员权限')
    next({ name: 'Home' })
    return
  }

  // 简化版权限检查：只基于用户role，不调用后端API
  // 管理员身份已在上面验证，直接允许访问
  // 如果需要更细粒度的权限控制，可以在具体页面中实现
  next()
}

/**
 * 权限初始化守卫
 */
export async function permissionInitGuard(
  to: RouteLocationNormalized,
  _from: RouteLocationNormalized,
  next: NavigationGuardNext
): Promise<void> {
  const permissionStore = usePermissionStore()
  const userStore = useUserStore()

  if (userStore.isAuthenticated && userStore.isAdmin) {
    await permissionStore.initializePermissions()
  }

  next()
}

/**
 * 页面标题守卫
 */
export function titleGuard(
  to: RouteLocationNormalized,
  _from: RouteLocationNormalized,
  next: NavigationGuardNext
): void {
  // 更改页面标题
  document.title = `${to.meta.title || '首页'} | 预测系统`
  next()
}

/**
 * 组合所有守卫
 */
export async function combinedGuard(
  to: RouteLocationNormalized,
  from: RouteLocationNormalized,
  next: NavigationGuardNext
): Promise<void> {
  try {
    const userStore = useUserStore()
    const permissionStore = usePermissionStore()

    // 页面标题
    document.title = `${to.meta.title || '首页'} | 预测系统`

    // 访客页重定向
    if (to.meta.guest && userStore.isAuthenticated) {
      next({ name: 'Home' })
      return
    }

    // 登录校验
    if (to.meta.requiresAuth && !userStore.isAuthenticated) {
      next({ name: 'Login', query: { redirect: to.fullPath } })
      return
    }

    // 管理员校验
    if (to.meta.requiresAdmin && !userStore.isAdmin) {
      ElMessage.error('您没有管理员权限')
      next({ name: 'Home' })
      return
    }

    // 超级管理员校验（如果需要）
    if (to.meta.requiresSuperAdmin && !permissionStore.isSuperAdmin) {
      ElMessage.error('需要超级管理员权限')
      next({ name: 'AdminDashboardHome' })
      return
    }

    // 权限数组校验（需要全部）
    if (to.meta.permissions && Array.isArray(to.meta.permissions)) {
      const ok = permissionStore.hasAllPermissions(to.meta.permissions)
      if (!ok) {
        ElMessage.error('权限不足')
        next({ name: 'AdminDashboardHome' })
        return
      }
    }

    // 任一权限校验
    if (to.meta.anyPermissions && Array.isArray(to.meta.anyPermissions)) {
      const okAny = permissionStore.hasAnyPermission(to.meta.anyPermissions)
      if (!okAny) {
        ElMessage.error('权限不足')
        next({ name: 'AdminDashboardHome' })
        return
      }
    }

    next()
  } catch (error) {
    console.error('路由守卫执行失败:', error)
    ElMessage.error('页面访问验证失败')
    next({ name: 'Home' })
  }
}

/**
 * 权限错误处理
 */
export function handlePermissionError(error: any, to: RouteLocationNormalized): void {
  console.error('权限检查错误:', error)
  
  if (error.code === 'PERMISSION_DENIED') {
    ElMessage.error(error.message || '权限不足')
  } else if (error.code === 'AUTH_REQUIRED') {
    ElMessage.error('请先登录')
  } else if (error.code === 'ADMIN_REQUIRED') {
    ElMessage.error('需要管理员权限')
  } else {
    ElMessage.error('访问验证失败')
  }
}

/**
 * 权限检查结果类型
 */
export interface PermissionCheckResult {
  hasPermission: boolean
  errorCode?: string
  errorMessage?: string
  redirectTo?: string
}

/**
 * 检查路由权限
 */
export function checkRoutePermissions(
  to: RouteLocationNormalized,
  userStore: any,
  permissionStore: any
): PermissionCheckResult {
  // 基础认证检查
  if (to.meta.requiresAuth && !userStore.isAuthenticated) {
    return {
      hasPermission: false,
      errorCode: 'AUTH_REQUIRED',
      errorMessage: '请先登录',
      redirectTo: '/login'
    }
  }

  // 管理员权限检查
  if (to.meta.requiresAdmin && !userStore.isAdmin) {
    return {
      hasPermission: false,
      errorCode: 'ADMIN_REQUIRED',
      errorMessage: '需要管理员权限',
      redirectTo: '/'
    }
  }

  // 超级管理员权限检查
  if (to.meta.requiresSuperAdmin && !permissionStore.isSuperAdmin) {
    return {
      hasPermission: false,
      errorCode: 'SUPER_ADMIN_REQUIRED',
      errorMessage: '需要超级管理员权限',
      redirectTo: '/admin'
    }
  }

  // 具体权限检查
  if (to.meta.permissions && Array.isArray(to.meta.permissions)) {
    const hasPermissions = permissionStore.hasAllPermissions(to.meta.permissions)
    if (!hasPermissions) {
      return {
        hasPermission: false,
        errorCode: 'PERMISSION_DENIED',
        errorMessage: '权限不足',
        redirectTo: '/admin'
      }
    }
  }

  return { hasPermission: true }
}