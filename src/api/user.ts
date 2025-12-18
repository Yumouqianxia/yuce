import { get, post, patch } from './http'
import { User, ProfileUpdateData, PointsHistory } from '@/types/user'

/**
 * 获取用户个人资料
 */
export const getUserProfile = (): Promise<any> => {
  return get<any>('/api/auth/profile')
}

/**
 * 更新用户个人资料
 */
export const updateProfile = (data: ProfileUpdateData): Promise<User> => {
  // Go后端使用PUT方法更新用户资料，允许只更新头像等任意字段
  const updateData: any = {}

  if (data.nickname !== undefined) {
    updateData.nickname = data.nickname
  }

  if (data.email !== undefined) {
    updateData.email = data.email
  }

  if (data.avatar !== undefined) {
    updateData.avatar = data.avatar
  }

  if (Object.keys(updateData).length === 0) {
    return Promise.reject(new Error('没有提供要更新的数据'))
  }

  return patch<User>('/api/auth/profile', updateData)
}

/**
 * 更新用户密码
 */
export const updateUserPassword = (data: {
  currentPassword: string
  newPassword: string
  newPasswordConfirm?: string
}): Promise<{ success: boolean; message: string }> => {
  return post<{ success: boolean; message: string }>('/api/auth/change-password', data)
}

/**
 * 上传用户头像
 */
export const uploadUserAvatar = (formData: FormData): Promise<any> => {
  return post<any>('/api/uploads/avatar', formData, {
    headers: {
      'Content-Type': 'multipart/form-data'
    }
  })
}

/**
 * 获取用户积分历史
 */
export const getPointsHistory = (userId?: number): Promise<PointsHistory[]> => {
  // Go后端的积分历史端点在leaderboard路由下
  if (userId) {
    return get<PointsHistory[]>(`/api/leaderboard/users/${userId}/points-history`)
  }
  // 如果没有指定用户ID，获取当前用户的积分历史
  return get<PointsHistory[]>('/api/leaderboard/users/me/points-history')
}