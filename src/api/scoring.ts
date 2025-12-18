// 积分规则配置API
import { get, post, put, del } from './http'
import type {
  ScoringRule,
  CreateScoringRuleRequest,
  UpdateScoringRuleRequest,
  ListScoringRulesRequest,
  ListScoringRulesResponse
} from '@/types/admin'

// ==================== 积分规则管理 ====================

/**
 * 创建积分规则
 */
export const createScoringRule = (data: CreateScoringRuleRequest): Promise<ScoringRule> => {
  return post<ScoringRule>('/admin/scoring-rules', data)
}

/**
 * 获取积分规则列表
 */
export const getScoringRules = (params?: ListScoringRulesRequest): Promise<ListScoringRulesResponse> => {
  return get<ListScoringRulesResponse>('/admin/scoring-rules', params)
}

/**
 * 获取积分规则详情
 */
export const getScoringRule = (id: number): Promise<ScoringRule> => {
  return get<ScoringRule>(`/admin/scoring-rules/${id}`)
}

/**
 * 更新积分规则
 */
export const updateScoringRule = (id: number, data: UpdateScoringRuleRequest): Promise<ScoringRule> => {
  return put<ScoringRule>(`/admin/scoring-rules/${id}`, data)
}

/**
 * 删除积分规则
 */
export const deleteScoringRule = (id: number): Promise<void> => {
  return del<void>(`/admin/scoring-rules/${id}`)
}

/**
 * 激活积分规则
 */
export const activateScoringRule = (id: number): Promise<ScoringRule> => {
  return post<ScoringRule>(`/admin/scoring-rules/${id}/activate`)
}

/**
 * 停用积分规则
 */
export const deactivateScoringRule = (id: number): Promise<ScoringRule> => {
  return post<ScoringRule>(`/admin/scoring-rules/${id}/deactivate`)
}

// ==================== 运动类型相关积分规则 ====================

/**
 * 获取运动类型的积分规则
 */
export const getScoringRulesBySportType = (sportTypeId: number): Promise<ScoringRule[]> => {
  return get<ScoringRule[]>(`/admin/sport-types/${sportTypeId}/scoring-rules`)
}

/**
 * 获取运动类型的活跃积分规则
 */
export const getActiveScoringRule = (sportTypeId: number): Promise<ScoringRule> => {
  return get<ScoringRule>(`/admin/sport-types/${sportTypeId}/scoring-rules/active`)
}

/**
 * 为运动类型设置默认积分规则
 */
export const setDefaultScoringRule = (sportTypeId: number, ruleId: number): Promise<void> => {
  return post<void>(`/admin/sport-types/${sportTypeId}/scoring-rules/${ruleId}/set-default`)
}

// ==================== 积分计算和重算 ====================

/**
 * 使用指定规则重新计算比赛积分
 */
export const recalculateMatchPointsWithRule = (matchId: number, ruleId?: number): Promise<{
  affected_predictions: number
  total_points_changed: number
  calculation_time: number
}> => {
  const data = ruleId ? { rule_id: ruleId } : {}
  return post<{
    affected_predictions: number
    total_points_changed: number
    calculation_time: number
  }>(`/admin/matches/${matchId}/recalculate-points`, data)
}

/**
 * 批量重新计算运动类型的所有比赛积分
 */
export const recalculateSportTypePoints = (sportTypeId: number, ruleId?: number): Promise<{
  affected_matches: number
  affected_predictions: number
  total_points_changed: number
  calculation_time: number
}> => {
  const data = ruleId ? { rule_id: ruleId } : {}
  return post<{
    affected_matches: number
    affected_predictions: number
    total_points_changed: number
    calculation_time: number
  }>(`/admin/sport-types/${sportTypeId}/recalculate-points`, data)
}

/**
 * 预览积分规则计算结果
 */
export const previewScoringRuleCalculation = (ruleId: number, matchId: number): Promise<{
  predictions: Array<{
    prediction_id: number
    user_id: number
    current_points: number
    new_points: number
    point_difference: number
    calculation_details: any
  }>
  total_predictions: number
  average_points: number
  max_points: number
  min_points: number
}> => {
  return get<{
    predictions: Array<{
      prediction_id: number
      user_id: number
      current_points: number
      new_points: number
      point_difference: number
      calculation_details: any
    }>
    total_predictions: number
    average_points: number
    max_points: number
    min_points: number
  }>(`/admin/scoring-rules/${ruleId}/preview/${matchId}`)
}

// ==================== 积分规则模板 ====================

/**
 * 获取积分规则模板
 */
export const getScoringRuleTemplates = (): Promise<Record<string, Partial<ScoringRule>>> => {
  return Promise.resolve({
    // 基础模板
    basic: {
      name: '基础积分规则',
      description: '简单的积分计算规则',
      base_points: 10,
      enable_difficulty: false,
      difficulty_multiplier: 1.0,
      enable_vote_reward: false,
      enable_time_reward: false,
      enable_modify_penalty: false
    },
    // 竞技模板
    competitive: {
      name: '竞技积分规则',
      description: '适合竞技类比赛的积分规则',
      base_points: 15,
      enable_difficulty: true,
      difficulty_multiplier: 1.5,
      enable_vote_reward: true,
      vote_reward_points: 1,
      max_vote_reward: 5,
      enable_time_reward: true,
      time_reward_points: 3,
      time_reward_hours: 24,
      enable_modify_penalty: true,
      modify_penalty_points: 2,
      max_modify_penalty: 6
    },
    // 休闲模板
    casual: {
      name: '休闲积分规则',
      description: '适合休闲类比赛的积分规则',
      base_points: 8,
      enable_difficulty: false,
      difficulty_multiplier: 1.0,
      enable_vote_reward: true,
      vote_reward_points: 1,
      max_vote_reward: 3,
      enable_time_reward: false,
      enable_modify_penalty: false
    }
  })
}

/**
 * 应用积分规则模板
 */
export const applyScoringRuleTemplate = (
  sportTypeId: number, 
  templateName: string, 
  customName?: string
): Promise<ScoringRule> => {
  return post<ScoringRule>('/admin/scoring-rules/apply-template', {
    sport_type_id: sportTypeId,
    template: templateName,
    name: customName
  })
}

/**
 * 复制积分规则
 */
export const copyScoringRule = (ruleId: number, newName: string, sportTypeId?: number): Promise<ScoringRule> => {
  return post<ScoringRule>(`/admin/scoring-rules/${ruleId}/copy`, {
    name: newName,
    sport_type_id: sportTypeId
  })
}

// ==================== 积分规则验证 ====================

/**
 * 验证积分规则配置
 */
export const validateScoringRule = (ruleData: Partial<ScoringRule>): Promise<{
  valid: boolean
  errors: string[]
  warnings: string[]
  suggestions: string[]
}> => {
  return post<{
    valid: boolean
    errors: string[]
    warnings: string[]
    suggestions: string[]
  }>('/admin/scoring-rules/validate', ruleData)
}

/**
 * 测试积分规则计算
 */
export const testScoringRuleCalculation = (
  ruleData: Partial<ScoringRule>,
  testScenarios: Array<{
    predicted_winner: string
    actual_winner: string
    predicted_score_a: number
    predicted_score_b: number
    actual_score_a: number
    actual_score_b: number
    vote_count?: number
    modification_count?: number
    prediction_time?: string
    match_start_time?: string
  }>
): Promise<Array<{
  scenario_index: number
  calculated_points: number
  calculation_breakdown: {
    base_points: number
    difficulty_bonus: number
    vote_bonus: number
    time_bonus: number
    modification_penalty: number
    final_points: number
  }
}>> => {
  return post<Array<{
    scenario_index: number
    calculated_points: number
    calculation_breakdown: {
      base_points: number
      difficulty_bonus: number
      vote_bonus: number
      time_bonus: number
      modification_penalty: number
      final_points: number
    }
  }>>('/admin/scoring-rules/test-calculation', {
    rule: ruleData,
    scenarios: testScenarios
  })
}

// ==================== 积分规则统计 ====================

/**
 * 获取积分规则使用统计
 */
export const getScoringRuleStats = (ruleId: number): Promise<{
  rule_id: number
  total_matches: number
  total_predictions: number
  average_points: number
  max_points: number
  min_points: number
  points_distribution: Array<{
    points_range: string
    count: number
    percentage: number
  }>
  usage_over_time: Array<{
    date: string
    match_count: number
    prediction_count: number
    average_points: number
  }>
}> => {
  return get<{
    rule_id: number
    total_matches: number
    total_predictions: number
    average_points: number
    max_points: number
    min_points: number
    points_distribution: Array<{
      points_range: string
      count: number
      percentage: number
    }>
    usage_over_time: Array<{
      date: string
      match_count: number
      prediction_count: number
      average_points: number
    }>
  }>(`/admin/scoring-rules/${ruleId}/stats`)
}

/**
 * 比较多个积分规则的效果
 */
export const compareScoringRules = (ruleIds: number[]): Promise<{
  rules: Array<{
    rule_id: number
    rule_name: string
    total_matches: number
    total_predictions: number
    average_points: number
    points_std_deviation: number
  }>
  comparison_metrics: {
    fairness_score: number
    engagement_score: number
    difficulty_balance: number
  }
}> => {
  return post<{
    rules: Array<{
      rule_id: number
      rule_name: string
      total_matches: number
      total_predictions: number
      average_points: number
      points_std_deviation: number
    }>
    comparison_metrics: {
      fairness_score: number
      engagement_score: number
      difficulty_balance: number
    }
  }>('/admin/scoring-rules/compare', { rule_ids: ruleIds })
}