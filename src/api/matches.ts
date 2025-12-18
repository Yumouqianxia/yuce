import { get, post, put, del } from './http'
import { Match, Prediction, CreatePredictionData } from '@/types/match'

/**
 * 获取所有比赛
 */
export const getMatches = (): Promise<Match[]> => {
  return get<Match[]>('/matches')
}

/**
 * 获取单个比赛详情
 */
export const getMatch = (id: number): Promise<Match> => {
  return get<Match>(`/matches/${id}`)
}

/**
 * 创建比赛
 */
export const createMatch = (matchData: Partial<Match>): Promise<Match> => {
  return post<Match>('/matches', matchData)
}

/**
 * 更新比赛
 */
export const updateMatch = (id: number, matchData: Partial<Match>): Promise<Match> => {
  return put<Match>(`/matches/${id}`, matchData)
}

/**
 * 删除比赛
 */
export const deleteMatch = (id: number): Promise<void> => {
  return del(`/matches/${id}`)
}

/**
 * 设置比赛结果
 */
export const setMatchResult = (id: number, resultData: { winner: string; scoreA?: number; scoreB?: number }): Promise<Match> => {
  return post<Match>(`/matches/${id}/result`, resultData)
}

/**
 * 提交比赛预测
 */
export const createPrediction = (matchId: number, prediction: CreatePredictionData): Promise<Prediction> => {
  // 使用前端命名约定的字段
  const payload = {
    matchId: Number(matchId),
    predictedWinner: prediction.predictedWinner,
    predictedScoreA: prediction.predictedScoreA,
    predictedScoreB: prediction.predictedScoreB
  }
  console.log('发送预测数据:', payload)
  return post<Prediction>('/predictions', payload)
}

/**
 * 获取已完成的比赛
 */
export const getFinishedMatches = (limit: number = 20): Promise<Match[]> => {
  return get<Match[]>('/matches/finished', { limit })
}

/**
 * 获取用户的预测
 */
export const getUserPredictions = (): Promise<Prediction[]> => {
  return get<Prediction[]>('/predictions/my')
}

/**
 * 获取当前用户在某场比赛的预测（若无则返回 null）
 */
export const getUserPredictionForMatch = (matchId: number): Promise<Prediction | null> => {
  // 使用前端命名约定的查询参数
  return get<Prediction[]>('/predictions/my', { matchId: matchId })
    .then(list => {
      if (Array.isArray(list) && list.length > 0) return list[0]
      return null
    })
    .catch(() => get<Prediction[]>('/predictions/my')
      .then(list => list.find(p => (p as any).matchId === Number(matchId)) ?? null))
}

/**
 * 获取指定比赛的预测列表
 */
export const getPredictionsByMatch = (matchId: number): Promise<Prediction[]> => {
  return get<Prediction[]>('/predictions', { matchId: matchId })
}

/**
 * 获取即将开始的比赛
 */
export const getUpcomingMatches = (limit: number = 10): Promise<Match[]> => {
  return get<Match[]>('/matches/upcoming', { limit })
}

/**
 * 获取正在进行的比赛
 */
export const getLiveMatches = (): Promise<Match[]> => {
  return get<Match[]>('/matches/live')
}