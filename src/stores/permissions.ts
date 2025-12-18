// 权限管理状态
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { useUserStore } from './user'
import { useAdminStore } from './admin'
import { ADMIN_PERMISSIONS } from '@/types/admin'
import type { AdminPermission } from '@/types/admin'

export const usePermissionStore = defineStore('permissions', () => {
  const userStore = useUserStore()
  const adminStore = useAdminStore()
  
  // ==================== 状态定义 ====================
  
  // 当前用户的权限列表
  const currentUserPermissions = ref<AdminPermission[]>([])
  
  // 权限检查缓存
  const permissionCache = ref<Record<string, boolean>>({})
  
  // 权限加载状态
  const permissionsLoading = ref(false)
  
  // ==================== 计算属性 ====================
  
  // 当前用户是否为管理员
  const isAdmin = computed(() => {
    return userStore.user?.role === 'admin'
  })
  
  // 当前用户是否为超级管理员
  const isSuperAdmin = computed(() => {
    // 后端当前未区分超级管理员，暂时将管理员视为通过
    return adminStore.isSuperAdmin || isAdmin.value
  })
  
  // 权限代码列表
  const permissionCodes = computed(() => {
    return currentUserPermissions.value.map(p => p.code)
  })
  
  // 按分类分组的权限
  const permissionsByCategory = computed(() => {
    const grouped: Record<string, AdminPermission[]> = {}
    currentUserPermissions.value.forEach(permission => {
      const category = permission.category || 'other'
      if (!grouped[category]) {
        grouped[category] = []
      }
      grouped[category].push(permission)
    })
    return grouped
  })
  
  // ==================== 权限检查方法 ====================
  
  // 检查单个权限
  const hasPermission = (permission: string): boolean => {
    if (!isAdmin.value) return false

    // 如果拉取到权限列表，则按列表判断；否则管理员默认放行
    if (permissionCodes.value.length > 0) {
      return permissionCodes.value.includes(permission)
    }

    return true
  }
  
  // 批量检查权限
  const hasPermissions = (permissions: string[]): Record<string, boolean> => {
    const result: Record<string, boolean> = {}
    permissions.forEach(permission => {
      result[permission] = hasPermission(permission)
    })
    return result
  }
  
  // 检查是否拥有任一权限
  const hasAnyPermission = (permissions: string[]): boolean => {
    return permissions.some(permission => hasPermission(permission))
  }
  
  // 检查是否拥有所有权限
  const hasAllPermissions = (permissions: string[]): boolean => {
    return permissions.every(permission => hasPermission(permission))
  }
  
  // ==================== 特定权限检查 ====================
  
  // 运动类型管理权限
  const canManageSportTypes = computed(() => {
    return hasPermission(ADMIN_PERMISSIONS.SPORT_TYPE_MANAGE)
  })
  
  // 运动配置管理权限
  const canManageSportConfig = computed(() => {
    return hasPermission(ADMIN_PERMISSIONS.SPORT_CONFIG_MANAGE)
  })
  
  // 积分规则管理权限
  const canManageScoringRules = computed(() => {
    return hasPermission(ADMIN_PERMISSIONS.SCORING_RULE_MANAGE)
  })
  
  // 比赛管理权限
  const canManageMatches = computed(() => {
    return hasPermission(ADMIN_PERMISSIONS.MATCH_MANAGE)
  })
  
  // 用户管理权限
  const canManageUsers = computed(() => {
    return hasPermission(ADMIN_PERMISSIONS.USER_MANAGE)
  })
  
  // 管理员管理权限
  const canManageAdmins = computed(() => {
    return hasPermission(ADMIN_PERMISSIONS.ADMIN_MANAGE)
  })
  
  // 审计日志查看权限
  const canViewAuditLogs = computed(() => {
    return hasPermission(ADMIN_PERMISSIONS.AUDIT_LOG_VIEW)
  })
  
  // 系统配置权限
  const canManageSystemConfig = computed(() => {
    return hasPermission(ADMIN_PERMISSIONS.SYSTEM_CONFIG)
  })
  
  // ==================== 运动类型访问权限 ====================
  
  // 检查运动类型访问权限
  const canAccessSportType = (sportTypeId: number): boolean => {
    // 超级管理员和系统管理员可以访问所有运动类型
    if (isSuperAdmin.value || adminStore.isSystemAdmin) {
      return true
    }
    
    // 检查是否有该运动类型的访问权限
    const adminUser = adminStore.currentAdmin
    if (!adminUser || !adminUser.sport_types) {
      return false
    }
    
    return adminUser.sport_types.some(sport => sport.id === sportTypeId)
  }
  
  // 获取可访问的运动类型ID列表
  const accessibleSportTypeIds = computed(() => {
    // 超级管理员和系统管理员可以访问所有运动类型
    if (isSuperAdmin.value || adminStore.isSystemAdmin) {
      return adminStore.allSportTypes.map(sport => sport.id)
    }
    
    const adminUser = adminStore.currentAdmin
    if (!adminUser || !adminUser.sport_types) {
      return []
    }
    
    return adminUser.sport_types.map(sport => sport.id)
  })
  
  // 过滤可访问的运动类型
  const filterAccessibleSportTypes = <T extends { sport_type_id?: number; id?: number }>(
    items: T[]
  ): T[] => {
    if (isSuperAdmin.value || adminStore.isSystemAdmin) {
      return items
    }
    
    const accessibleIds = accessibleSportTypeIds.value
    return items.filter(item => {
      const sportTypeId = item.sport_type_id || item.id
      return sportTypeId && accessibleIds.includes(sportTypeId)
    })
  }
  
  // ==================== 权限加载和管理 ====================
  
  // 加载当前用户权限
  const loadCurrentUserPermissions = async () => {
    if (!isAdmin.value || !userStore.user) {
      currentUserPermissions.value = []
      permissionCache.value = {}
      return
    }

    // TODO: 可在此处调用后端权限接口后填充 currentUserPermissions
    currentUserPermissions.value = []
    permissionCache.value = {}
  }
  
  // 刷新权限缓存
  const refreshPermissionCache = () => {
    permissionCache.value = {}
  }
  
  // 清除权限数据
  const clearPermissions = () => {
    currentUserPermissions.value = []
    permissionCache.value = {}
  }
  
  // ==================== 权限守卫 ====================
  
  // 权限守卫函数，用于路由守卫
  const requirePermission = (permission: string) => {
    return hasPermission(permission)
  }
  
  // 权限守卫函数，用于组件显示控制
  const requireAnyPermission = (permissions: string[]) => {
    return hasAnyPermission(permissions)
  }
  
  // 权限守卫函数，用于严格权限控制
  const requireAllPermissions = (permissions: string[]) => {
    return hasAllPermissions(permissions)
  }
  
  // ==================== 权限提示 ====================
  
  // 获取权限不足的提示信息
  const getPermissionDeniedMessage = (permission: string): string => {
    const permissionNames: Record<string, string> = {
      [ADMIN_PERMISSIONS.SPORT_TYPE_MANAGE]: '运动类型管理',
      [ADMIN_PERMISSIONS.SPORT_CONFIG_MANAGE]: '运动配置管理',
      [ADMIN_PERMISSIONS.SCORING_RULE_MANAGE]: '积分规则管理',
      [ADMIN_PERMISSIONS.MATCH_MANAGE]: '比赛管理',
      [ADMIN_PERMISSIONS.USER_MANAGE]: '用户管理',
      [ADMIN_PERMISSIONS.ADMIN_MANAGE]: '管理员管理',
      [ADMIN_PERMISSIONS.AUDIT_LOG_VIEW]: '审计日志查看',
      [ADMIN_PERMISSIONS.SYSTEM_CONFIG]: '系统配置管理'
    }
    
    const permissionName = permissionNames[permission] || permission
    return `您没有${permissionName}权限，请联系超级管理员`
  }
  
  // ==================== 初始化 ====================
  
  // 初始化权限系统
  const initializePermissions = async () => {
    if (isAdmin.value) {
      await loadCurrentUserPermissions()
    }
  }
  
  return {
    // 状态
    currentUserPermissions,
    permissionCache,
    permissionsLoading,
    
    // 计算属性
    isAdmin,
    isSuperAdmin,
    permissionCodes,
    permissionsByCategory,
    accessibleSportTypeIds,
    
    // 基础权限检查
    hasPermission,
    hasPermissions,
    hasAnyPermission,
    hasAllPermissions,
    
    // 特定权限检查
    canManageSportTypes,
    canManageSportConfig,
    canManageScoringRules,
    canManageMatches,
    canManageUsers,
    canManageAdmins,
    canViewAuditLogs,
    canManageSystemConfig,
    
    // 运动类型访问权限
    canAccessSportType,
    filterAccessibleSportTypes,
    
    // 权限管理
    loadCurrentUserPermissions,
    refreshPermissionCache,
    clearPermissions,
    
    // 权限守卫
    requirePermission,
    requireAnyPermission,
    requireAllPermissions,
    
    // 工具方法
    getPermissionDeniedMessage,
    initializePermissions
  }
})