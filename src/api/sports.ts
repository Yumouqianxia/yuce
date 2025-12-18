// 运动类型管理API
import { get, post, put, del } from './http'
import type {
  SportType,
  SportConfiguration,
  CreateSportTypeRequest,
  UpdateSportTypeRequest,
  ListSportTypesRequest,
  ListSportTypesResponse,
  UpdateSportConfigurationRequest,
  SportTypeStats
} from '@/types/admin'

// ==================== 运动类型管理 ====================

/**
 * 创建运动类型
 */
export const createSportType = (data: CreateSportTypeRequest): Promise<SportType> => {
  return post<SportType>('/admin/sport-types', data)
}

/**
 * 获取运动类型列表
 */
export const getSportTypes = (params?: ListSportTypesRequest): Promise<ListSportTypesResponse> => {
  return get<ListSportTypesResponse>('/admin/sport-types', params)
}

/**
 * 获取所有运动类型（不分页，用于下拉选择）
 */
export const getAllSportTypes = (): Promise<SportType[]> => {
  return get<SportType[]>('/admin/sport-types/all')
}

/**
 * 获取运动类型详情
 */
export const getSportType = (id: number): Promise<SportType> => {
  return get<SportType>(`/admin/sport-types/${id}`)
}

/**
 * 根据代码获取运动类型
 */
export const getSportTypeByCode = (code: string): Promise<SportType> => {
  return get<SportType>(`/admin/sport-types/by-code/${code}`)
}

/**
 * 更新运动类型
 */
export const updateSportType = (id: number, data: UpdateSportTypeRequest): Promise<SportType> => {
  return put<SportType>(`/admin/sport-types/${id}`, data)
}

/**
 * 删除运动类型
 */
export const deleteSportType = (id: number): Promise<void> => {
  return del<void>(`/admin/sport-types/${id}`)
}

/**
 * 启用/禁用运动类型
 */
export const toggleSportTypeStatus = (id: number, isActive: boolean): Promise<SportType> => {
  return post<SportType>(`/admin/sport-types/${id}/toggle-status`, { is_active: isActive })
}

/**
 * 批量更新运动类型排序
 */
export const updateSportTypesOrder = (orders: Array<{ id: number; sort_order: number }>): Promise<void> => {
  return post<void>('/admin/sport-types/batch-order', { orders })
}

// ==================== 运动配置管理 ====================

/**
 * 获取运动配置
 */
export const getSportConfiguration = (sportTypeId: number): Promise<SportConfiguration> => {
  return get<SportConfiguration>(`/admin/sport-types/${sportTypeId}/configuration`)
}

/**
 * 更新运动配置
 */
export const updateSportConfiguration = (
  sportTypeId: number, 
  data: UpdateSportConfigurationRequest
): Promise<SportConfiguration> => {
  return put<SportConfiguration>(`/admin/sport-types/${sportTypeId}/configuration`, data)
}

/**
 * 重置运动配置为默认值
 */
export const resetSportConfiguration = (sportTypeId: number): Promise<SportConfiguration> => {
  return post<SportConfiguration>(`/admin/sport-types/${sportTypeId}/configuration/reset`)
}

/**
 * 批量更新运动配置
 */
export const batchUpdateSportConfigurations = (
  sportTypeIds: number[], 
  data: UpdateSportConfigurationRequest
): Promise<void> => {
  return post<void>('/admin/sport-configurations/batch-update', {
    sport_type_ids: sportTypeIds,
    configuration: data
  })
}

// ==================== 运动类型统计 ====================

/**
 * 获取运动类型统计信息
 */
export const getSportTypeStats = (sportTypeId: number): Promise<SportTypeStats> => {
  return get<SportTypeStats>(`/admin/sport-types/${sportTypeId}/stats`)
}

/**
 * 获取所有运动类型统计概览
 */
export const getAllSportTypeStats = (): Promise<SportTypeStats[]> => {
  return get<SportTypeStats[]>('/admin/sport-types/stats/overview')
}

// ==================== 公开API（不需要管理员权限） ====================

/**
 * 获取活跃的运动类型列表（公开接口）
 */
export const getActiveSportTypes = (): Promise<SportType[]> => {
  return get<SportType[]>('/sport-types')
}

/**
 * 获取运动类型详情（公开接口）
 */
export const getPublicSportType = (id: number): Promise<SportType> => {
  return get<SportType>(`/sport-types/${id}`)
}

/**
 * 根据代码获取运动类型（公开接口）
 */
export const getPublicSportTypeByCode = (code: string): Promise<SportType> => {
  return get<SportType>(`/sport-types/by-code/${code}`)
}

// ==================== 工具函数 ====================

/**
 * 验证运动类型代码是否可用
 */
export const checkSportTypeCodeAvailability = (code: string, excludeId?: number): Promise<{ available: boolean }> => {
  const params = excludeId ? { exclude_id: excludeId } : {}
  return get<{ available: boolean }>(`/admin/sport-types/check-code/${code}`, params)
}

/**
 * 获取运动类型配置模板
 */
export const getSportConfigurationTemplate = (category: string): Promise<Partial<SportConfiguration>> => {
  return get<Partial<SportConfiguration>>(`/admin/sport-configurations/template/${category}`)
}

/**
 * 导出运动类型配置
 */
export const exportSportTypeConfigurations = (sportTypeIds?: number[]): Promise<Blob> => {
  const params = sportTypeIds ? { sport_type_ids: sportTypeIds.join(',') } : {}
  return get<Blob>('/admin/sport-types/export', params, {
    responseType: 'blob'
  })
}

/**
 * 导入运动类型配置
 */
export const importSportTypeConfigurations = (file: File): Promise<{ success: number; failed: number; errors: string[] }> => {
  const formData = new FormData()
  formData.append('file', file)
  
  return post<{ success: number; failed: number; errors: string[] }>('/admin/sport-types/import', formData, {
    headers: {
      'Content-Type': 'multipart/form-data'
    }
  })
}

// ==================== 预设配置 ====================

/**
 * 获取预设的运动类型配置
 */
export const getPresetSportConfigurations = (): Promise<Record<string, Partial<SportConfiguration>>> => {
  return Promise.resolve({
    // 电子竞技默认配置
    esports: {
      enable_realtime: true,
      enable_chat: true,
      enable_voting: true,
      enable_prediction: true,
      enable_leaderboard: true,
      allow_modification: true,
      max_modifications: 3,
      modification_deadline: 30,
      enable_self_voting: false,
      max_votes_per_user: 10,
      voting_deadline: 0
    },
    // 传统体育默认配置
    traditional: {
      enable_realtime: true,
      enable_chat: false,
      enable_voting: true,
      enable_prediction: true,
      enable_leaderboard: true,
      allow_modification: true,
      max_modifications: 2,
      modification_deadline: 60,
      enable_self_voting: false,
      max_votes_per_user: 5,
      voting_deadline: 30
    }
  })
}

/**
 * 应用预设配置到运动类型
 */
export const applySportConfigurationPreset = (
  sportTypeId: number, 
  presetName: string
): Promise<SportConfiguration> => {
  return post<SportConfiguration>(`/admin/sport-types/${sportTypeId}/configuration/apply-preset`, {
    preset: presetName
  })
}