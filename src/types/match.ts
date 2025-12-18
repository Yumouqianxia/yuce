// 比赛状态类型
export type MatchStatus = 'not_started' | 'in_progress' | 'completed' | 'cancelled' | 'finished'

// 比赛系列类型
export type MatchSeries = 'BO1' | 'BO2' | 'BO3' | 'BO5' | 'BO7' | 'BO9'

// 比赛类型
export type MatchType = 'regular' | 'playoff' | 'final'

// 赛事类型
export type TournamentType = 'spring' | 'summer' | 'annual' | 'challenger'

// 赛事阶段
export type TournamentStage = 'regular' | 'playoff' | 'group' | 'knockout'

// 比赛接口
export interface Match {
  id: number
  title: string
  description?: string
  optionA: string
  optionB: string
  matchTime: string
  status: MatchStatus
  matchType?: MatchType
  series?: MatchSeries
  winner?: string
  scoreA?: number
  scoreB?: number
  result_winner?: string // 兼容后端返回格式
  result_score?: string // 兼容后端返回格式
  isActive: boolean
  tournamentType?: TournamentType
  tournamentStage?: TournamentStage
  year?: number
  createdAt: string
  updatedAt: string

  // 兼容后端返回格式
  team_a?: string
  team_b?: string
  match_type?: MatchType
  match_series?: MatchSeries
  is_predictable?: boolean
  start_time?: string
  points_earned?: number
}

// 预测接口
export interface Prediction {
  id: number
  userId: number
  matchId: number
  match: Match
  predictedWinner: string
  predicted_winner?: string // 兼容后端返回格式
  predictedScoreA?: number
  predictedScoreB?: number
  predicted_score?: string // 兼容后端返回格式
  pointsEarned?: number
  points_earned?: number // 兼容后端返回格式
  isCorrect?: boolean
  createdAt: string
}

// 创建预测数据类型
export interface CreatePredictionData {
  matchId: number
  predictedWinner: string
  predictedScoreA: number
  predictedScoreB: number
}

// 比赛过滤条件
export interface MatchFilters {
  status?: MatchStatus
  match_type?: MatchType
  search?: string
  tournamentType?: TournamentType
  tournamentStage?: TournamentStage
  year?: number
}

// 分页响应
export interface PaginatedResponse<T> {
  count: number
  next: string | null
  previous: string | null
  results: T[]
}