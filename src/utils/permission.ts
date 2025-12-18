// 权限相关工具函数
import type { RouteLocationNormalized } from 'vue-router'
import { ADMIN_PERMISSIONS } from '@/types/admin'

/**
 * 权限检查配置接口
 */
export interface PermissionConfig {
  permission?: string
  permissions?: string[]
  anyPermissions?: string[]
  requireAdmin?: boolean
  requireSuperAdmin?: boolean
  sportTypeId?: number
}

/**
 * 路由权限配置接口
 */
export interface RoutePermissionConfig extends PermissionConfig {
  // 路由路径或路径模式
  path?: string | RegExp
  // 路由名称
  name?: string
  // 是否精确匹配
  exact?: boolean
}

/**
 * 权限检查结果接口
 */
export interface PermissionCheckResult {
  hasPermission: boolean
  errorMessage?: string
  errorCode?: string
}

/**
 * 预定义的路由权限配置
 */
export const ROUTE_PERMISSIONS: Record<string, RoutePermissionConfig> = {
  // 管理员相关路由
  admin: {
    path: /^\/admin/,
    requireAdmin: true
  },
  
  // 超级管理员路由
  superAdmin: {
    path: /^\/admin\/system/,
    requireSuperAdmin: true
  },
  
  // 运动类型管理
  sportTypeManagement: {
    path: '/admin/sport-types',
    permission: ADMIN_PERMISSIONS.SPORT_TYPE_MANAGE
  },
  
  // 积分规则管理
  scoringRuleManagement: {
    path: '/admin/scoring-rules',
    permission: ADMIN_PERMISSIONS.SCORING_RULE_MANAGE
  },
  
  // 比赛管理
  matchManagement: {
    path: '/admin/matches',
    permission: ADMIN_PERMISSIONS.MATCH_MANAGE
  },
  
  // 用户管理
  userManagement: {
    path: '/admin/users',
    permission: ADMIN_PERMISSIONS.USER_MANAGE
  },
  
  // 管理员管理
  adminManagement: {
    path: '/admin/admins',
    permission: ADMIN_PERMISSIONS.ADMIN_MANAGE
  },
  
  // 审计日志
  auditLogs: {
    path: '/admin/audit-logs',
    permission: ADMIN_PERMISSIONS.AUDIT_LOG_VIEW
  },
  
  // 系统配置
  systemConfig: {
    path: '/admin/system-config',
    permission: ADMIN_PERMISSIONS.SYSTEM_CONFIG
  }
}

/**
 * 检查路由是否匹配权限配置
 */
export function matchRoutePermission(
  route: RouteLocationNormalized,
  config: RoutePermissionConfig
): boolean {
  // 检查路径匹配
  if (config.path) {
    if (typeof config.path === 'string') {
      if (config.exact) {
        if (route.path !== config.path) {
          return false
        }
      } else {
        if (!route.path.startsWith(config.path)) {
          return false
        }
      }
    } else if (config.path instanceof RegExp) {
      if (!config.path.test(route.path)) {
        return false
      }
    }
  }
  
  // 检查路由名称匹配
  if (config.name && route.name !== config.name) {
    return false
  }
  
  return true
}

/**
 * 获取路由对应的权限配置
 */
export function getRoutePermissionConfig(
  route: RouteLocationNormalized
): RoutePermissionConfig | null {
  // 检查路由meta中的权限配置
  if (route.meta?.permission) {
    return route.meta.permission as RoutePermissionConfig
  }
  
  // 检查预定义的权限配置
  for (const config of Object.values(ROUTE_PERMISSIONS)) {
    if (matchRoutePermission(route, config)) {
      return config
    }
  }
  
  return null
}

/**
 * 格式化权限错误消息
 */
export function formatPermissionError(
  config: PermissionConfig,
  errorCode?: string
): string {
  if (config.requireSuperAdmin) {
    return '此功能需要超级管理员权限'
  }
  
  if (config.requireAdmin) {
    return '此功能需要管理员权限'
  }
  
  if (config.permission) {
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
    
    const permissionName = permissionNames[config.permission] || config.permission
    return `您没有${permissionName}权限`
  }
  
  if (config.sportTypeId !== undefined) {
    return '您没有访问此运动类型的权限'
  }
  
  return '您没有权限访问此功能'
}

/**
 * 权限相关的CSS类名
 */
export const PERMISSION_CLASSES = {
  DENIED: 'permission-denied',
  LOADING: 'permission-loading',
  GRANTED: 'permission-granted',
  DISABLED: 'permission-disabled',
  HIDDEN: 'permission-hidden'
} as const

/**
 * 权限检查的错误代码
 */
export const PERMISSION_ERROR_CODES = {
  NOT_ADMIN: 'NOT_ADMIN',
  NOT_SUPER_ADMIN: 'NOT_SUPER_ADMIN',
  INSUFFICIENT_PERMISSION: 'INSUFFICIENT_PERMISSION',
  SPORT_TYPE_ACCESS_DENIED: 'SPORT_TYPE_ACCESS_DENIED',
  PERMISSION_LOADING: 'PERMISSION_LOADING'
} as const

/**
 * 权限检查的操作类型
 */
export const PERMISSION_ACTIONS = {
  HIDE: 'hide',
  DISABLE: 'disable',
  CLASS: 'class',
  REDIRECT: 'redirect'
} as const

/**
 * 创建权限检查函数
 */
export function createPermissionChecker(permissionStore: any) {
  return function checkPermission(config: PermissionConfig): PermissionCheckResult {
    try {
      // 超级管理员检查
      if (config.requireSuperAdmin) {
        if (!permissionStore.isSuperAdmin) {
          return {
            hasPermission: false,
            errorMessage: formatPermissionError(config),
            errorCode: PERMISSION_ERROR_CODES.NOT_SUPER_ADMIN
          }
        }
      }
      
      // 管理员身份检查
      if (config.requireAdmin) {
        if (!permissionStore.isAdmin) {
          return {
            hasPermission: false,
            errorMessage: formatPermissionError(config),
            errorCode: PERMISSION_ERROR_CODES.NOT_ADMIN
          }
        }
      }
      
      // 运动类型访问权限检查
      if (config.sportTypeId !== undefined) {
        if (!permissionStore.canAccessSportType(config.sportTypeId)) {
          return {
            hasPermission: false,
            errorMessage: formatPermissionError(config),
            errorCode: PERMISSION_ERROR_CODES.SPORT_TYPE_ACCESS_DENIED
          }
        }
      }
      
      // 单个权限检查
      if (config.permission) {
        if (!permissionStore.hasPermission(config.permission)) {
          return {
            hasPermission: false,
            errorMessage: formatPermissionError(config),
            errorCode: PERMISSION_ERROR_CODES.INSUFFICIENT_PERMISSION
          }
        }
      }
      
      // 多个权限检查（需要全部拥有）
      if (config.permissions && config.permissions.length > 0) {
        if (!permissionStore.hasAllPermissions(config.permissions)) {
          return {
            hasPermission: false,
            errorMessage: formatPermissionError(config),
            errorCode: PERMISSION_ERROR_CODES.INSUFFICIENT_PERMISSION
          }
        }
      }
      
      // 多个权限检查（拥有任一即可）
      if (config.anyPermissions && config.anyPermissions.length > 0) {
        if (!permissionStore.hasAnyPermission(config.anyPermissions)) {
          return {
            hasPermission: false,
            errorMessage: formatPermissionError(config),
            errorCode: PERMISSION_ERROR_CODES.INSUFFICIENT_PERMISSION
          }
        }
      }
      
      return { hasPermission: true }
    } catch (error) {
      return {
        hasPermission: false,
        errorMessage: '权限检查失败',
        errorCode: 'PERMISSION_CHECK_ERROR'
      }
    }
  }
}

/**
 * 权限相关的默认配置
 */
export const DEFAULT_PERMISSION_CONFIG = {
  // 默认的权限检查失败处理方式
  defaultAction: PERMISSION_ACTIONS.HIDE,
  
  // 默认的权限检查失败CSS类名
  defaultFailedClass: PERMISSION_CLASSES.DENIED,
  
  // 是否自动初始化权限数据
  autoInit: true,
  
  // 权限检查失败时是否显示错误消息
  showErrorMessage: true,
  
  // 权限检查失败时的默认错误消息
  defaultErrorMessage: '您没有权限访问此功能'
} as const