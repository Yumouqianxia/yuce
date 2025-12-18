// 权限相关的组合式函数
import { computed, ref, watch } from 'vue'
import { usePermissionStore } from '@/stores/permissions'
import { useAdminStore } from '@/stores/admin'
import { ADMIN_PERMISSIONS } from '@/types/admin'

export interface PermissionOptions {
  // 单个权限
  permission?: string
  // 多个权限（需要全部拥有）
  permissions?: string[]
  // 多个权限（拥有任一即可）
  anyPermissions?: string[]
  // 是否需要管理员身份
  requireAdmin?: boolean
  // 是否需要超级管理员身份
  requireSuperAdmin?: boolean
  // 运动类型ID
  sportTypeId?: number
  // 是否自动初始化权限数据
  autoInit?: boolean
}

/**
 * 权限检查组合式函数
 */
export function usePermission(options: PermissionOptions = {}) {
  const permissionStore = usePermissionStore()
  const adminStore = useAdminStore()
  
  // 权限检查结果
  const hasPermission = computed(() => {
    // 超级管理员检查
    if (options.requireSuperAdmin) {
      return permissionStore.isSuperAdmin
    }
    
    // 管理员身份检查
    if (options.requireAdmin) {
      if (!permissionStore.isAdmin) {
        return false
      }
    }
    
    // 运动类型访问权限检查
    if (options.sportTypeId !== undefined) {
      if (!permissionStore.canAccessSportType(options.sportTypeId)) {
        return false
      }
    }
    
    // 单个权限检查
    if (options.permission) {
      return permissionStore.hasPermission(options.permission)
    }
    
    // 多个权限检查（需要全部拥有）
    if (options.permissions && options.permissions.length > 0) {
      return permissionStore.hasAllPermissions(options.permissions)
    }
    
    // 多个权限检查（拥有任一即可）
    if (options.anyPermissions && options.anyPermissions.length > 0) {
      return permissionStore.hasAnyPermission(options.anyPermissions)
    }
    
    // 如果只要求管理员身份且已通过检查，则允许访问
    if (options.requireAdmin && permissionStore.isAdmin) {
      return true
    }
    
    // 默认允许访问（如果没有指定任何权限要求）
    return !options.requireAdmin && !options.requireSuperAdmin
  })
  
  // 加载状态
  const loading = computed(() => {
    return permissionStore.permissionsLoading || adminStore.adminLoading
  })
  
  // 错误信息
  const errorMessage = computed(() => {
    if (hasPermission.value) {
      return null
    }
    
    if (options.requireSuperAdmin) {
      return '此功能需要超级管理员权限'
    }
    
    if (options.requireAdmin && !permissionStore.isAdmin) {
      return '此功能需要管理员权限'
    }
    
    if (options.permission) {
      return permissionStore.getPermissionDeniedMessage(options.permission)
    }
    
    if (options.sportTypeId !== undefined) {
      return '您没有访问此运动类型的权限'
    }
    
    return '您没有权限访问此功能'
  })
  
  // 初始化权限数据
  const initializePermissions = async () => {
    if (options.autoInit !== false && permissionStore.isAdmin) {
      try {
        await permissionStore.initializePermissions()
      } catch (error) {
        console.error('初始化权限数据失败:', error)
      }
    }
  }
  
  // 自动初始化
  if (options.autoInit !== false) {
    initializePermissions()
  }
  
  return {
    hasPermission,
    loading,
    errorMessage,
    initializePermissions,
    // 权限store的引用
    permissionStore,
    adminStore
  }
}

/**
 * 管理员权限检查组合式函数
 */
export function useAdminPermission(permission?: string, sportTypeId?: number) {
  return usePermission({
    permission,
    requireAdmin: true,
    sportTypeId,
    autoInit: true
  })
}

/**
 * 超级管理员权限检查组合式函数
 */
export function useSuperAdminPermission() {
  return usePermission({
    requireSuperAdmin: true,
    autoInit: true
  })
}

/**
 * 运动类型访问权限检查组合式函数
 */
export function useSportTypePermission(sportTypeId: number) {
  return usePermission({
    requireAdmin: true,
    sportTypeId,
    autoInit: true
  })
}

/**
 * 多权限检查组合式函数
 */
export function useMultiplePermissions(permissions: string[], requireAll = true) {
  return usePermission({
    requireAdmin: true,
    permissions: requireAll ? permissions : undefined,
    anyPermissions: requireAll ? undefined : permissions,
    autoInit: true
  })
}

/**
 * 权限守卫组合式函数
 * 用于路由守卫或其他需要权限检查的场景
 */
export function usePermissionGuard() {
  const permissionStore = usePermissionStore()
  
  // 检查路由权限
  const checkRoutePermission = (permission: string): boolean => {
    return permissionStore.requirePermission(permission)
  }
  
  // 检查多个路由权限
  const checkRoutePermissions = (permissions: string[], requireAll = true): boolean => {
    if (requireAll) {
      return permissionStore.requireAllPermissions(permissions)
    } else {
      return permissionStore.requireAnyPermission(permissions)
    }
  }
  
  // 获取权限拒绝消息
  const getPermissionDeniedMessage = (permission: string): string => {
    return permissionStore.getPermissionDeniedMessage(permission)
  }
  
  return {
    checkRoutePermission,
    checkRoutePermissions,
    getPermissionDeniedMessage,
    permissionStore
  }
}

/**
 * 权限相关的常用功能
 */
export function usePermissionHelpers() {
  const permissionStore = usePermissionStore()
  const adminStore = useAdminStore()
  
  // 常用权限检查
  const canManageSportTypes = computed(() => permissionStore.canManageSportTypes)
  const canManageScoringRules = computed(() => permissionStore.canManageScoringRules)
  const canManageMatches = computed(() => permissionStore.canManageMatches)
  const canManageUsers = computed(() => permissionStore.canManageUsers)
  const canManageAdmins = computed(() => permissionStore.canManageAdmins)
  const canViewAuditLogs = computed(() => permissionStore.canViewAuditLogs)
  const canManageSystemConfig = computed(() => permissionStore.canManageSystemConfig)
  
  // 身份检查
  const isAdmin = computed(() => permissionStore.isAdmin)
  const isSuperAdmin = computed(() => permissionStore.isSuperAdmin)
  const isSystemAdmin = computed(() => adminStore.isSystemAdmin)
  
  // 可访问的运动类型
  const accessibleSportTypes = computed(() => {
    return permissionStore.filterAccessibleSportTypes(adminStore.allSportTypes)
  })
  
  return {
    // 权限检查
    canManageSportTypes,
    canManageScoringRules,
    canManageMatches,
    canManageUsers,
    canManageAdmins,
    canViewAuditLogs,
    canManageSystemConfig,
    
    // 身份检查
    isAdmin,
    isSuperAdmin,
    isSystemAdmin,
    
    // 数据
    accessibleSportTypes,
    
    // Store引用
    permissionStore,
    adminStore
  }
}

// 权限常量
export { ADMIN_PERMISSIONS } from '@/types/admin'