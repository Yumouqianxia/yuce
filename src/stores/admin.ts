// 管理员状态管理
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type {
  AdminUser,
  AdminPermission,
  AdminAuditLog,
  SportType,
  SportConfiguration,
  ScoringRule,
  ListAdminsRequest,
  ListSportTypesRequest,
  ListScoringRulesRequest,
  ListAuditLogsRequest,
  AuditStatsResponse
} from '@/types/admin'
import { AdminLevel } from '@/types/admin'

// 导入API函数
import * as adminApi from '@/api/admin'
import * as sportsApi from '@/api/sports'
import * as scoringApi from '@/api/scoring'

export const useAdminStore = defineStore('admin', () => {
  // ==================== 状态定义 ====================
  
  // 管理员相关状态
  const admins = ref<AdminUser[]>([])
  const currentAdmin = ref<AdminUser | null>(null)
  const adminTotal = ref(0)
  const adminLoading = ref(false)
  
  // 权限相关状态
  const allPermissions = ref<AdminPermission[]>([])
  const userPermissions = ref<AdminPermission[]>([])
  const permissionLoading = ref(false)
  
  // 运动类型相关状态
  const sportTypes = ref<SportType[]>([])
  const allSportTypes = ref<SportType[]>([]) // 不分页的完整列表
  const currentSportType = ref<SportType | null>(null)
  const sportTypeTotal = ref(0)
  const sportTypeLoading = ref(false)
  
  // 积分规则相关状态
  const scoringRules = ref<ScoringRule[]>([])
  const currentScoringRule = ref<ScoringRule | null>(null)
  const scoringRuleTotal = ref(0)
  const scoringRuleLoading = ref(false)
  
  // 审计日志相关状态
  const auditLogs = ref<AdminAuditLog[]>([])
  const auditLogTotal = ref(0)
  const auditStats = ref<AuditStatsResponse | null>(null)
  const auditLoading = ref(false)
  
  // 系统状态
  const systemInfo = ref<any>(null)
  const systemStats = ref<any>(null)
  const cacheStats = ref<any>(null)
  
  // 错误状态
  const lastError = ref<string | null>(null)
  
  // ==================== 计算属性 ====================
  
  // 当前用户是否为管理员
  const isAdmin = computed(() => {
    return currentAdmin.value !== null
  })
  
  // 当前用户是否为超级管理员
  const isSuperAdmin = computed(() => {
    return currentAdmin.value?.admin_level === AdminLevel.SUPER
  })
  
  // 当前用户是否为系统管理员或以上
  const isSystemAdmin = computed(() => {
    return currentAdmin.value && currentAdmin.value.admin_level >= AdminLevel.SYSTEM
  })
  
  // 活跃的运动类型
  const activeSportTypes = computed(() => {
    return sportTypes.value.filter(sport => sport.is_active)
  })
  
  // 活跃的积分规则
  const activeScoringRules = computed(() => {
    return scoringRules.value.filter(rule => rule.is_active)
  })
  
  // 按运动类型分组的积分规则
  const scoringRulesBySportType = computed(() => {
    const grouped: Record<number, ScoringRule[]> = {}
    scoringRules.value.forEach(rule => {
      if (!grouped[rule.sport_type_id]) {
        grouped[rule.sport_type_id] = []
      }
      grouped[rule.sport_type_id].push(rule)
    })
    return grouped
  })
  
  // ==================== 管理员管理 ====================
  
  // 获取管理员列表
  const fetchAdmins = async (params?: ListAdminsRequest) => {
    try {
      adminLoading.value = true
      lastError.value = null
      
      const response = await adminApi.getAdminList(params)
      admins.value = response.admins
      adminTotal.value = response.total
      
      return response
    } catch (error) {
      lastError.value = error instanceof Error ? error.message : '获取管理员列表失败'
      throw error
    } finally {
      adminLoading.value = false
    }
  }
  
  // 获取管理员详情
  const fetchAdmin = async (id: number) => {
    try {
      adminLoading.value = true
      lastError.value = null
      
      const admin = await adminApi.getAdmin(id)
      currentAdmin.value = admin
      
      return admin
    } catch (error) {
      lastError.value = error instanceof Error ? error.message : '获取管理员详情失败'
      throw error
    } finally {
      adminLoading.value = false
    }
  }
  
  // 创建管理员
  const createAdmin = async (data: any) => {
    try {
      lastError.value = null
      
      const admin = await adminApi.createAdmin(data)
      admins.value.unshift(admin)
      adminTotal.value++
      
      return admin
    } catch (error) {
      lastError.value = error instanceof Error ? error.message : '创建管理员失败'
      throw error
    }
  }
  
  // 更新管理员
  const updateAdmin = async (id: number, data: any) => {
    try {
      lastError.value = null
      
      const admin = await adminApi.updateAdmin(id, data)
      const index = admins.value.findIndex(a => a.user_id === id)
      if (index > -1) {
        admins.value[index] = admin
      }
      
      if (currentAdmin.value?.user_id === id) {
        currentAdmin.value = admin
      }
      
      return admin
    } catch (error) {
      lastError.value = error instanceof Error ? error.message : '更新管理员失败'
      throw error
    }
  }
  
  // 删除管理员
  const deleteAdmin = async (id: number) => {
    try {
      lastError.value = null
      
      await adminApi.deleteAdmin(id)
      admins.value = admins.value.filter(a => a.user_id !== id)
      adminTotal.value--
      
      if (currentAdmin.value?.user_id === id) {
        currentAdmin.value = null
      }
    } catch (error) {
      lastError.value = error instanceof Error ? error.message : '删除管理员失败'
      throw error
    }
  }
  
  // ==================== 权限管理 ====================
  
  // 获取所有权限
  const fetchAllPermissions = async () => {
    try {
      permissionLoading.value = true
      lastError.value = null
      
      const permissions = await adminApi.getAllPermissions()
      allPermissions.value = permissions
      
      return permissions
    } catch (error) {
      lastError.value = error instanceof Error ? error.message : '获取权限列表失败'
      throw error
    } finally {
      permissionLoading.value = false
    }
  }
  
  // 获取用户权限
  const fetchUserPermissions = async (userId: number) => {
    try {
      permissionLoading.value = true
      lastError.value = null
      
      const permissions = await adminApi.getUserPermissions(userId)
      userPermissions.value = permissions
      
      return permissions
    } catch (error) {
      lastError.value = error instanceof Error ? error.message : '获取用户权限失败'
      throw error
    } finally {
      permissionLoading.value = false
    }
  }
  
  // 检查权限
  const hasPermission = (permission: string): boolean => {
    if (isSuperAdmin.value) return true
    return userPermissions.value.some(p => p.code === permission && p.is_active)
  }
  
  // 批量检查权限
  const hasPermissions = (permissions: string[]): Record<string, boolean> => {
    const result: Record<string, boolean> = {}
    permissions.forEach(permission => {
      result[permission] = hasPermission(permission)
    })
    return result
  }
  
  // 授予权限
  const grantPermissions = async (userId: number, permissions: string[]) => {
    try {
      lastError.value = null
      
      await adminApi.grantPermissions(userId, { permissions })
      
      // 如果是当前用户，刷新权限
      if (currentAdmin.value?.user_id === userId) {
        await fetchUserPermissions(userId)
      }
    } catch (error) {
      lastError.value = error instanceof Error ? error.message : '授予权限失败'
      throw error
    }
  }
  
  // 撤销权限
  const revokePermissions = async (userId: number, permissions: string[]) => {
    try {
      lastError.value = null
      
      await adminApi.revokePermissions(userId, { permissions })
      
      // 如果是当前用户，刷新权限
      if (currentAdmin.value?.user_id === userId) {
        await fetchUserPermissions(userId)
      }
    } catch (error) {
      lastError.value = error instanceof Error ? error.message : '撤销权限失败'
      throw error
    }
  }
  
  // ==================== 运动类型管理 ====================
  
  // 获取运动类型列表
  const fetchSportTypes = async (params?: ListSportTypesRequest) => {
    try {
      sportTypeLoading.value = true
      lastError.value = null
      
      const response = await sportsApi.getSportTypes(params)
      sportTypes.value = response.sport_types
      sportTypeTotal.value = response.total
      
      return response
    } catch (error) {
      lastError.value = error instanceof Error ? error.message : '获取运动类型列表失败'
      throw error
    } finally {
      sportTypeLoading.value = false
    }
  }
  
  // 获取所有运动类型（不分页）
  const fetchAllSportTypes = async () => {
    try {
      lastError.value = null
      
      const sportTypes = await sportsApi.getAllSportTypes()
      allSportTypes.value = sportTypes
      
      return sportTypes
    } catch (error) {
      lastError.value = error instanceof Error ? error.message : '获取运动类型失败'
      throw error
    }
  }
  
  // 获取运动类型详情
  const fetchSportType = async (id: number) => {
    try {
      sportTypeLoading.value = true
      lastError.value = null
      
      const sportType = await sportsApi.getSportType(id)
      currentSportType.value = sportType
      
      return sportType
    } catch (error) {
      lastError.value = error instanceof Error ? error.message : '获取运动类型详情失败'
      throw error
    } finally {
      sportTypeLoading.value = false
    }
  }
  
  // 创建运动类型
  const createSportType = async (data: any) => {
    try {
      lastError.value = null
      
      const sportType = await sportsApi.createSportType(data)
      sportTypes.value.unshift(sportType)
      allSportTypes.value.unshift(sportType)
      sportTypeTotal.value++
      
      return sportType
    } catch (error) {
      lastError.value = error instanceof Error ? error.message : '创建运动类型失败'
      throw error
    }
  }
  
  // 更新运动类型
  const updateSportType = async (id: number, data: any) => {
    try {
      lastError.value = null
      
      const sportType = await sportsApi.updateSportType(id, data)
      
      // 更新列表中的数据
      const index = sportTypes.value.findIndex(s => s.id === id)
      if (index > -1) {
        sportTypes.value[index] = sportType
      }
      
      const allIndex = allSportTypes.value.findIndex(s => s.id === id)
      if (allIndex > -1) {
        allSportTypes.value[allIndex] = sportType
      }
      
      if (currentSportType.value?.id === id) {
        currentSportType.value = sportType
      }
      
      return sportType
    } catch (error) {
      lastError.value = error instanceof Error ? error.message : '更新运动类型失败'
      throw error
    }
  }
  
  // 删除运动类型
  const deleteSportType = async (id: number) => {
    try {
      lastError.value = null
      
      await sportsApi.deleteSportType(id)
      
      sportTypes.value = sportTypes.value.filter(s => s.id !== id)
      allSportTypes.value = allSportTypes.value.filter(s => s.id !== id)
      sportTypeTotal.value--
      
      if (currentSportType.value?.id === id) {
        currentSportType.value = null
      }
    } catch (error) {
      lastError.value = error instanceof Error ? error.message : '删除运动类型失败'
      throw error
    }
  }
  
  // 更新运动配置
  const updateSportConfiguration = async (sportTypeId: number, data: any) => {
    try {
      lastError.value = null
      
      const configuration = await sportsApi.updateSportConfiguration(sportTypeId, data)
      
      // 更新运动类型中的配置
      const updateSportTypeConfig = (sportType: SportType) => {
        if (sportType.id === sportTypeId) {
          sportType.configuration = configuration
        }
      }
      
      sportTypes.value.forEach(updateSportTypeConfig)
      allSportTypes.value.forEach(updateSportTypeConfig)
      
      if (currentSportType.value?.id === sportTypeId) {
        currentSportType.value.configuration = configuration
      }
      
      return configuration
    } catch (error) {
      lastError.value = error instanceof Error ? error.message : '更新运动配置失败'
      throw error
    }
  }
  
  // ==================== 积分规则管理 ====================
  
  // 获取积分规则列表
  const fetchScoringRules = async (params?: ListScoringRulesRequest) => {
    try {
      scoringRuleLoading.value = true
      lastError.value = null
      
      const response = await scoringApi.getScoringRules(params)
      scoringRules.value = response.scoring_rules
      scoringRuleTotal.value = response.total
      
      return response
    } catch (error) {
      lastError.value = error instanceof Error ? error.message : '获取积分规则列表失败'
      throw error
    } finally {
      scoringRuleLoading.value = false
    }
  }
  
  // 获取积分规则详情
  const fetchScoringRule = async (id: number) => {
    try {
      scoringRuleLoading.value = true
      lastError.value = null
      
      const rule = await scoringApi.getScoringRule(id)
      currentScoringRule.value = rule
      
      return rule
    } catch (error) {
      lastError.value = error instanceof Error ? error.message : '获取积分规则详情失败'
      throw error
    } finally {
      scoringRuleLoading.value = false
    }
  }
  
  // 创建积分规则
  const createScoringRule = async (data: any) => {
    try {
      lastError.value = null
      
      const rule = await scoringApi.createScoringRule(data)
      scoringRules.value.unshift(rule)
      scoringRuleTotal.value++
      
      return rule
    } catch (error) {
      lastError.value = error instanceof Error ? error.message : '创建积分规则失败'
      throw error
    }
  }
  
  // 更新积分规则
  const updateScoringRule = async (id: number, data: any) => {
    try {
      lastError.value = null
      
      const rule = await scoringApi.updateScoringRule(id, data)
      const index = scoringRules.value.findIndex(r => r.id === id)
      if (index > -1) {
        scoringRules.value[index] = rule
      }
      
      if (currentScoringRule.value?.id === id) {
        currentScoringRule.value = rule
      }
      
      return rule
    } catch (error) {
      lastError.value = error instanceof Error ? error.message : '更新积分规则失败'
      throw error
    }
  }
  
  // 删除积分规则
  const deleteScoringRule = async (id: number) => {
    try {
      lastError.value = null
      
      await scoringApi.deleteScoringRule(id)
      scoringRules.value = scoringRules.value.filter(r => r.id !== id)
      scoringRuleTotal.value--
      
      if (currentScoringRule.value?.id === id) {
        currentScoringRule.value = null
      }
    } catch (error) {
      lastError.value = error instanceof Error ? error.message : '删除积分规则失败'
      throw error
    }
  }
  
  // ==================== 审计日志 ====================
  
  // 获取审计日志列表
  const fetchAuditLogs = async (params?: ListAuditLogsRequest) => {
    try {
      auditLoading.value = true
      lastError.value = null
      
      const response = await adminApi.getAuditLogs(params)
      auditLogs.value = response.logs
      auditLogTotal.value = response.total
      
      return response
    } catch (error) {
      lastError.value = error instanceof Error ? error.message : '获取审计日志失败'
      throw error
    } finally {
      auditLoading.value = false
    }
  }
  
  // 获取审计统计
  const fetchAuditStats = async (params?: any) => {
    try {
      auditLoading.value = true
      lastError.value = null
      
      const stats = await adminApi.getAuditStats(params)
      auditStats.value = stats
      
      return stats
    } catch (error) {
      lastError.value = error instanceof Error ? error.message : '获取审计统计失败'
      throw error
    } finally {
      auditLoading.value = false
    }
  }
  
  // ==================== 系统管理 ====================
  
  // 获取系统信息
  const fetchSystemInfo = async () => {
    try {
      lastError.value = null
      
      const info = await adminApi.getSystemInfo()
      systemInfo.value = info
      
      return info
    } catch (error) {
      lastError.value = error instanceof Error ? error.message : '获取系统信息失败'
      throw error
    }
  }
  
  // 获取系统统计
  const fetchSystemStats = async () => {
    try {
      lastError.value = null
      
      const stats = await adminApi.getSystemStats()
      systemStats.value = stats
      
      return stats
    } catch (error) {
      lastError.value = error instanceof Error ? error.message : '获取系统统计失败'
      throw error
    }
  }
  
  // 获取缓存统计
  const fetchCacheStats = async () => {
    try {
      lastError.value = null
      
      const stats = await adminApi.getCacheStats()
      cacheStats.value = stats
      
      return stats
    } catch (error) {
      lastError.value = error instanceof Error ? error.message : '获取缓存统计失败'
      throw error
    }
  }
  
  // ==================== 工具方法 ====================
  
  // 清除错误
  const clearError = () => {
    lastError.value = null
  }
  
  // 重置状态
  const resetState = () => {
    admins.value = []
    currentAdmin.value = null
    adminTotal.value = 0
    
    allPermissions.value = []
    userPermissions.value = []
    
    sportTypes.value = []
    allSportTypes.value = []
    currentSportType.value = null
    sportTypeTotal.value = 0
    
    scoringRules.value = []
    currentScoringRule.value = null
    scoringRuleTotal.value = 0
    
    auditLogs.value = []
    auditLogTotal.value = 0
    auditStats.value = null
    
    systemInfo.value = null
    systemStats.value = null
    cacheStats.value = null
    
    lastError.value = null
  }
  
  // 初始化管理员数据
  const initializeAdminData = async () => {
    // 简化版：不需要初始化管理员数据
    // 基于role的简单权限控制不需要从后端加载权限数据
    return Promise.resolve()
  }
  
  return {
    // 状态
    admins,
    currentAdmin,
    adminTotal,
    adminLoading,
    
    allPermissions,
    userPermissions,
    permissionLoading,
    
    sportTypes,
    allSportTypes,
    currentSportType,
    sportTypeTotal,
    sportTypeLoading,
    
    scoringRules,
    currentScoringRule,
    scoringRuleTotal,
    scoringRuleLoading,
    
    auditLogs,
    auditLogTotal,
    auditStats,
    auditLoading,
    
    systemInfo,
    systemStats,
    cacheStats,
    
    lastError,
    
    // 计算属性
    isAdmin,
    isSuperAdmin,
    isSystemAdmin,
    activeSportTypes,
    activeScoringRules,
    scoringRulesBySportType,
    
    // 管理员管理
    fetchAdmins,
    fetchAdmin,
    createAdmin,
    updateAdmin,
    deleteAdmin,
    
    // 权限管理
    fetchAllPermissions,
    fetchUserPermissions,
    hasPermission,
    hasPermissions,
    grantPermissions,
    revokePermissions,
    
    // 运动类型管理
    fetchSportTypes,
    fetchAllSportTypes,
    fetchSportType,
    createSportType,
    updateSportType,
    deleteSportType,
    updateSportConfiguration,
    
    // 积分规则管理
    fetchScoringRules,
    fetchScoringRule,
    createScoringRule,
    updateScoringRule,
    deleteScoringRule,
    
    // 审计日志
    fetchAuditLogs,
    fetchAuditStats,
    
    // 系统管理
    fetchSystemInfo,
    fetchSystemStats,
    fetchCacheStats,
    
    // 工具方法
    clearError,
    resetState,
    initializeAdminData
  }
})