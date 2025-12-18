// 管理员操作API
import { get, post, put, del } from './http'
import type {
  AdminUser,
  CreateAdminRequest,
  UpdateAdminRequest,
  ListAdminsRequest,
  ListAdminsResponse,
  PermissionRequest,
  SportAccessRequest,
  AdminPermission,
  AdminAuditLog,
  ListAuditLogsRequest,
  ListAuditLogsResponse,
  AuditStatsRequest,
  AuditStatsResponse
} from '@/types/admin'

// ==================== 管理员管理 ====================

/**
 * 创建管理员
 */
export const createAdmin = (data: CreateAdminRequest): Promise<AdminUser> => {
  return post<AdminUser>('/admin/admins', data)
}

/**
 * 获取管理员列表
 */
export const getAdminList = (params?: ListAdminsRequest): Promise<ListAdminsResponse> => {
  return get<ListAdminsResponse>('/admin/admins', params)
}

/**
 * 获取管理员详情
 */
export const getAdmin = (id: number): Promise<AdminUser> => {
  return get<AdminUser>(`/admin/admins/${id}`)
}

/**
 * 更新管理员
 */
export const updateAdmin = (id: number, data: UpdateAdminRequest): Promise<AdminUser> => {
  return put<AdminUser>(`/admin/admins/${id}`, data)
}

/**
 * 删除管理员
 */
export const deleteAdmin = (id: number): Promise<void> => {
  return del<void>(`/admin/admins/${id}`)
}

// ==================== 权限管理 ====================

/**
 * 获取所有权限列表
 */
export const getAllPermissions = (): Promise<AdminPermission[]> => {
  return get<AdminPermission[]>('/admin/permissions')
}

/**
 * 获取用户权限
 */
export const getUserPermissions = (userId: number): Promise<AdminPermission[]> => {
  return get<AdminPermission[]>(`/admin/admins/${userId}/permissions`)
}

/**
 * 授予权限
 */
export const grantPermissions = (userId: number, data: PermissionRequest): Promise<void> => {
  return post<void>(`/admin/admins/${userId}/permissions`, data)
}

/**
 * 撤销权限
 */
export const revokePermissions = (userId: number, data: PermissionRequest): Promise<void> => {
  return del<void>(`/admin/admins/${userId}/permissions`, { data })
}

// ==================== 运动类型访问权限管理 ====================

/**
 * 获取用户运动类型访问权限
 */
export const getUserSportAccess = (userId: number): Promise<number[]> => {
  return get<number[]>(`/admin/admins/${userId}/sport-access`)
}

/**
 * 授予运动类型访问权限
 */
export const grantSportAccess = (userId: number, data: SportAccessRequest): Promise<void> => {
  return post<void>(`/admin/admins/${userId}/sport-access`, data)
}

/**
 * 撤销运动类型访问权限
 */
export const revokeSportAccess = (userId: number, data: SportAccessRequest): Promise<void> => {
  return del<void>(`/admin/admins/${userId}/sport-access`, { data })
}

// ==================== 审计日志 ====================

/**
 * 获取审计日志列表
 */
export const getAuditLogs = (params?: ListAuditLogsRequest): Promise<ListAuditLogsResponse> => {
  return get<ListAuditLogsResponse>('/admin/audit-logs', params)
}

/**
 * 获取审计日志详情
 */
export const getAuditLog = (id: number): Promise<AdminAuditLog> => {
  return get<AdminAuditLog>(`/admin/audit-logs/${id}`)
}

/**
 * 获取审计统计
 */
export const getAuditStats = (params?: AuditStatsRequest): Promise<AuditStatsResponse> => {
  return get<AuditStatsResponse>('/admin/audit-logs/stats', params)
}

// ==================== 系统管理 ====================

/**
 * 获取系统信息
 */
export const getSystemInfo = (): Promise<any> => {
  return get<any>('/admin/system/info')
}

/**
 * 获取系统统计
 */
export const getSystemStats = (): Promise<any> => {
  return get<any>('/admin/system/stats')
}

/**
 * 清理系统缓存
 */
export const clearSystemCache = (): Promise<void> => {
  return post<void>('/admin/system/cache/clear')
}

/**
 * 获取缓存统计
 */
export const getCacheStats = (): Promise<any> => {
  return get<any>('/admin/cache/stats')
}

/**
 * 刷新排行榜缓存
 */
export const refreshLeaderboardCache = (tournament?: string): Promise<void> => {
  return post<void>('/admin/cache/leaderboard/refresh', { tournament })
}

/**
 * 失效排行榜缓存
 */
export const invalidateLeaderboardCache = (tournament?: string): Promise<void> => {
  return post<void>('/admin/cache/leaderboard/invalidate', { tournament })
}

// ==================== 用户管理 ====================

/**
 * 获取用户列表（管理员视图）
 */
export const getUsers = (params?: any): Promise<any> => {
  return get<any>('/api/users', params)
}

/**
 * 创建用户（管理员）
 */
export const createUser = (data: any): Promise<any> => {
  return post<any>('/api/users', data)
}

/**
 * 获取用户详情（管理员视图）
 */
export const getUserDetail = (id: number): Promise<any> => {
  return get<any>(`/api/users/${id}`)
}

/**
 * 更新用户信息（管理员操作）
 */
export const updateUser = (id: number, data: any): Promise<any> => {
  return put<any>(`/api/users/${id}`, data)
}

/**
 * 删除用户
 */
export const deleteUser = (id: number): Promise<void> => {
  return del<void>(`/api/users/${id}`)
}

/**
 * 禁用/启用用户
 */
export const toggleUserStatus = (id: number, isActive: boolean): Promise<void> => {
  return post<void>(`/api/users/${id}/toggle-status`, { is_active: isActive })
}

/**
 * 重置用户密码
 */
export const resetUserPassword = (id: number, data?: any): Promise<{ new_password: string }> => {
  return post<{ new_password: string }>(`/api/users/${id}/password`, data)
}

// ==================== 比赛管理 ====================

/**
 * 获取比赛列表（管理员视图）
 */
export const getMatches = (params?: any): Promise<any> => {
  return get<any>('/api/matches', params)
}

/**
 * 创建比赛
 */
export const createMatch = (data: any): Promise<any> => {
  return post<any>('/api/matches', data)
}

/**
 * 更新比赛
 */
export const updateMatch = (id: number, data: any): Promise<any> => {
  return put<any>(`/api/matches/${id}`, data)
}

/**
 * 删除比赛
 */
export const deleteMatch = (id: number): Promise<void> => {
  return del<void>(`/api/matches/${id}`)
}

/**
 * 设置比赛结果
 */
export const setMatchResult = (id: number, data: any): Promise<any> => {
  return post<any>(`/api/matches/${id}/result`, data)
}

/**
 * 开始比赛
 */
export const startMatch = (id: number): Promise<any> => {
  return post<any>(`/api/matches/${id}/start`)
}

/**
 * 取消比赛
 */
export const cancelMatch = (id: number, reason?: string): Promise<any> => {
  return post<any>(`/api/matches/${id}/cancel`, { reason })
}

// ==================== 预测管理 ====================

/**
 * 获取预测列表（管理员视图）
 */
export const getPredictions = (params?: any): Promise<any> => {
  return get<any>('/admin/predictions', params)
}

/**
 * 删除预测
 */
export const deletePrediction = (id: number): Promise<void> => {
  return del<void>(`/admin/predictions/${id}`)
}

// ==================== 公告管理 ====================

export const listAnnouncements = (params?: any): Promise<any> => {
  return get<any>('/api/announcements', params)
}

export const getAnnouncementDetail = (id: number): Promise<any> => {
  return get<any>(`/api/announcements/${id}`)
}

export const createAnnouncement = (data: any): Promise<any> => {
  return post<any>('/api/announcements', data)
}

export const updateAnnouncement = (id: number, data: any): Promise<any> => {
  return put<any>(`/api/announcements/${id}`, data)
}

export const deleteAnnouncement = (id: number): Promise<void> => {
  return del<void>(`/api/announcements/${id}`)
}

// ==================== 系统设置 ====================

export const getSystemSettings = (): Promise<any> => {
  return get<any>('/api/admin/settings')
}

export const updateSystemSettings = (data: any): Promise<any> => {
  return post<any>('/api/admin/settings', data)
}

/**
 * 重新计算比赛积分
 */
export const recalculateMatchPoints = (matchId: number): Promise<any> => {
  return post<any>(`/admin/matches/${matchId}/recalculate-points`)
}

// ==================== 工具函数 ====================

/**
 * 检查当前用户是否有指定权限
 */
export const checkPermission = async (permission: string): Promise<boolean> => {
  try {
    const permissions = await getUserPermissions(0) // 0表示当前用户
    return permissions.some(p => p.code === permission && p.is_active)
  } catch (error) {
    console.error('检查权限失败:', error)
    return false
  }
}

/**
 * 批量检查权限
 */
export const checkPermissions = async (permissions: string[]): Promise<Record<string, boolean>> => {
  try {
    const userPermissions = await getUserPermissions(0)
    const result: Record<string, boolean> = {}
    
    permissions.forEach(permission => {
      result[permission] = userPermissions.some(p => p.code === permission && p.is_active)
    })
    
    return result
  } catch (error) {
    console.error('批量检查权限失败:', error)
    return permissions.reduce((acc, permission) => {
      acc[permission] = false
      return acc
    }, {} as Record<string, boolean>)
  }
}