import { get } from './http'
import { User } from '@/types/user'

// 排行榜用户接口
export interface LeaderboardUser extends User {
  rank?: number
  points: number
}

/**
 * 获取积分排行榜
 */
export const getLeaderboard = (tournament: string = 'GLOBAL', limit: number = 50): Promise<LeaderboardUser[]> => {
  return get<LeaderboardUser[]>('/api/leaderboard', { tournament, limit })
}

/**
 * 获取排行榜统计信息
 */
export const getLeaderboardStats = (tournament: string = 'GLOBAL'): Promise<any> => {
  return get<any>('/api/leaderboard/stats', { tournament })
}

/**
 * 获取用户排名信息
 */
export const getUserRank = (userId: number, tournament: string = 'GLOBAL'): Promise<any> => {
  return get<any>(`/api/leaderboard/users/${userId}/rank`, { tournament })
}