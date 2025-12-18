// 用户类型
export interface User {
  id: number
  username: string
  nickname?: string
  email: string
  avatar?: string
  points?: number
  role?: 'user' | 'admin'
  createdAt: string
  updatedAt?: string
}

// 用户资料类型
export interface UserProfile {
  username: string
  nickname: string
  email: string
  avatar?: string
  createdAt?: string
  id?: number
}

// 用户统计类型
export interface UserStats {
  total_points: number
  rank: number
  total_predictions: number
  accurate_predictions: number
}

// 登录数据类型
export interface LoginData {
  username: string
  password: string
}

// 注册数据类型
export interface RegisterData {
  username: string
  nickname: string
  email: string
  password: string
  password_confirm: string
}

// 用户资料更新数据类型
export interface ProfileUpdateData {
  nickname?: string
  email?: string
  avatar?: File | string
  current_password?: string
  new_password?: string
  new_password_confirm?: string
}

// 登录响应类型
export interface LoginResponse {
  access_token: string
  user: User
}

// 注册响应类型
export interface RegisterResponse {
  success: boolean
  message: string
}

// 积分历史记录类型
export interface PointsHistory {
  id: number
  points_change: number
  points_after: number
  change_type: 'prediction' | 'admin' | 'system' | 'match_deleted'
  description: string
  related_match?: string
  created_at: string
  source?: 'admin_api' | 'prediction_api'
}