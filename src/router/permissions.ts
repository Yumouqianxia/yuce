// 路由权限配置
import { ADMIN_PERMISSIONS } from '@/types/admin'

/**
 * 路由权限配置接口
 */
export interface RoutePermissionMeta {
  // 基础权限
  requiresAuth?: boolean
  requiresAdmin?: boolean
  requiresSuperAdmin?: boolean
  guest?: boolean
  
  // 具体权限要求
  permissions?: string[]
  anyPermissions?: string[]
  
  // 运动类型访问权限
  sportTypeId?: number
  
  // 页面标题
  title?: string
  
  // 其他元数据
  [key: string]: any
}

/**
 * 管理员路由权限配置
 */
export const ADMIN_ROUTE_PERMISSIONS = {
  // 管理中心首页
  AdminDashboard: {
    requiresAuth: true,
    requiresAdmin: true,
    title: '管理中心'
  },
  
  // 公告管理
  AdminSite: {
    requiresAuth: true,
    requiresAdmin: true,
    permissions: [ADMIN_PERMISSIONS.SYSTEM_CONFIG],
    title: '公告管理'
  },
  
  // 用户管理
  AdminUsers: {
    requiresAuth: true,
    requiresAdmin: true,
    permissions: [ADMIN_PERMISSIONS.USER_MANAGE],
    title: '用户管理'
  },
  
  // 比赛管理
  AdminMatches: {
    requiresAuth: true,
    requiresAdmin: true,
    permissions: [ADMIN_PERMISSIONS.MATCH_MANAGE],
    title: '比赛管理'
  },
  
  // 运动类型管理
  AdminSportTypes: {
    requiresAuth: true,
    requiresAdmin: true,
    permissions: [ADMIN_PERMISSIONS.SPORT_TYPE_MANAGE],
    title: '运动类型管理'
  },
  
  // 积分规则管理
  AdminScoringRules: {
    requiresAuth: true,
    requiresAdmin: true,
    permissions: [ADMIN_PERMISSIONS.SCORING_RULE_MANAGE],
    title: '积分规则管理'
  },
  
  // 管理员管理
  AdminManagement: {
    requiresAuth: true,
    requiresAdmin: true,
    requiresSuperAdmin: true,
    permissions: [ADMIN_PERMISSIONS.ADMIN_MANAGE],
    title: '管理员管理'
  },
  
  // 审计日志
  AdminAuditLogs: {
    requiresAuth: true,
    requiresAdmin: true,
    permissions: [ADMIN_PERMISSIONS.AUDIT_LOG_VIEW],
    title: '审计日志'
  },
  
  // 系统设置
  AdminSettings: {
    requiresAuth: true,
    requiresAdmin: true,
    requiresSuperAdmin: true,
    permissions: [ADMIN_PERMISSIONS.SYSTEM_CONFIG],
    title: '系统设置'
  }
} as const

/**
 * 权限检查函数
 */
export function checkRoutePermission(
  routeName: string,
  userRole: string | undefined,
  userPermissions: string[],
  isSuperAdmin: boolean
): { hasPermission: boolean; errorMessage?: string } {
  const config = ADMIN_ROUTE_PERMISSIONS[routeName as keyof typeof ADMIN_ROUTE_PERMISSIONS]
  
  if (!config) {
    return { hasPermission: true }
  }
  
  // 检查是否需要管理员权限
  if (config.requiresAdmin && userRole !== 'admin') {
    return {
      hasPermission: false,
      errorMessage: '此功能需要管理员权限'
    }
  }
  
  // 检查是否需要超级管理员权限
  if ('requiresSuperAdmin' in config && config.requiresSuperAdmin && !isSuperAdmin) {
    return {
      hasPermission: false,
      errorMessage: '此功能需要超级管理员权限'
    }
  }
  
  // 检查具体权限
  if ('permissions' in config && config.permissions && config.permissions.length > 0) {
    const hasAllPermissions = config.permissions.every((permission: string) => 
      userPermissions.includes(permission)
    )
    
    if (!hasAllPermissions) {
      return {
        hasPermission: false,
        errorMessage: `您没有访问此页面的权限`
      }
    }
  }
  
  // 检查任一权限
  const anyPerms = (config as RoutePermissionMeta).anyPermissions
  if (anyPerms && anyPerms.length > 0) {
    const hasAnyPermission = anyPerms.some((permission: string) => 
      userPermissions.includes(permission)
    )
    
    if (!hasAnyPermission) {
      return {
        hasPermission: false,
        errorMessage: `您没有访问此页面的权限`
      }
    }
  }
  
  return { hasPermission: true }
}

/**
 * 获取路由显示名称
 */
export function getRouteDisplayName(routeName: string): string {
  const config = ADMIN_ROUTE_PERMISSIONS[routeName as keyof typeof ADMIN_ROUTE_PERMISSIONS]
  return config?.title || routeName
}

/**
 * 获取用户可访问的管理员路由
 */
export function getAccessibleAdminRoutes(
  userRole: string | undefined,
  userPermissions: string[],
  isSuperAdmin: boolean
): string[] {
  const accessibleRoutes: string[] = []
  
  for (const [routeName, config] of Object.entries(ADMIN_ROUTE_PERMISSIONS)) {
    const { hasPermission } = checkRoutePermission(routeName, userRole, userPermissions, isSuperAdmin)
    
    if (hasPermission) {
      accessibleRoutes.push(routeName)
    }
  }
  
  return accessibleRoutes
}

/**
 * 管理员菜单配置
 */
export interface AdminMenuItem {
  name: string
  routeName: string
  icon?: string
  permissions?: string[]
  requiresSuperAdmin?: boolean
  children?: AdminMenuItem[]
}

export const ADMIN_MENU_ITEMS: AdminMenuItem[] = [
  {
    name: '公告管理',
    routeName: 'AdminSite',
    icon: 'Notification',
    permissions: [ADMIN_PERMISSIONS.SYSTEM_CONFIG]
  },
  {
    name: '用户管理',
    routeName: 'AdminUsers',
    icon: 'User',
    permissions: [ADMIN_PERMISSIONS.USER_MANAGE]
  },
  {
    name: '比赛管理',
    routeName: 'AdminMatches',
    icon: 'Trophy',
    permissions: [ADMIN_PERMISSIONS.MATCH_MANAGE]
  },
  {
    name: '运动类型管理',
    routeName: 'AdminSportTypes',
    icon: 'Basketball',
    permissions: [ADMIN_PERMISSIONS.SPORT_TYPE_MANAGE]
  },
  {
    name: '积分规则管理',
    routeName: 'AdminScoringRules',
    icon: 'Medal',
    permissions: [ADMIN_PERMISSIONS.SCORING_RULE_MANAGE]
  },
  {
    name: '管理员管理',
    routeName: 'AdminManagement',
    icon: 'UserFilled',
    permissions: [ADMIN_PERMISSIONS.ADMIN_MANAGE],
    requiresSuperAdmin: true
  },
  {
    name: '审计日志',
    routeName: 'AdminAuditLogs',
    icon: 'Document',
    permissions: [ADMIN_PERMISSIONS.AUDIT_LOG_VIEW]
  },
  {
    name: '系统设置',
    routeName: 'AdminSettings',
    icon: 'Setting',
    permissions: [ADMIN_PERMISSIONS.SYSTEM_CONFIG],
    requiresSuperAdmin: true
  }
]

/**
 * 过滤用户可访问的菜单项
 */
export function filterAccessibleMenuItems(
  menuItems: AdminMenuItem[],
  userRole: string | undefined,
  userPermissions: string[],
  isSuperAdmin: boolean
): AdminMenuItem[] {
  return menuItems.filter(item => {
    // 检查超级管理员权限
    if (item.requiresSuperAdmin && !isSuperAdmin) {
      return false
    }
    
    // 检查具体权限
    if (item.permissions && item.permissions.length > 0) {
      const hasPermissions = item.permissions.every(permission => 
        userPermissions.includes(permission)
      )
      
      if (!hasPermissions) {
        return false
      }
    }
    
    // 递归过滤子菜单
    if (item.children) {
      item.children = filterAccessibleMenuItems(item.children, userRole, userPermissions, isSuperAdmin)
    }
    
    return true
  })
}