// 管理员相关类型定义

// 管理员级别枚举
export enum AdminLevel {
  SPORT = 1,   // 运动管理员
  SYSTEM = 2,  // 系统管理员
  SUPER = 3    // 超级管理员
}

// 管理员用户
export interface AdminUser {
  user_id: number
  admin_level: AdminLevel
  is_active: boolean
  created_at: string
  updated_at: string
  permissions?: AdminPermission[]
  sport_types?: SportType[]
}

// 管理员权限
export interface AdminPermission {
  id: number
  code: string
  name: string
  description: string
  category: string
  is_active: boolean
  created_at: string
  updated_at: string
}

// 管理员审计日志
export interface AdminAuditLog {
  id: number
  admin_user_id: number
  action: string
  resource: string
  resource_id?: string
  method: string
  path: string
  ip_address?: string
  user_agent?: string
  old_values?: any
  new_values?: any
  changes?: any
  status: AuditStatus
  error_msg?: string
  duration: number
  created_at: string
  admin_user?: AdminUser
}

// 审计状态枚举
export enum AuditStatus {
  SUCCESS = 1,
  FAILED = 2,
  PARTIAL = 3
}

// 运动类别枚举
export enum SportCategory {
  ESPORTS = 'esports',
  TRADITIONAL = 'traditional'
}

// 运动类型
export interface SportType {
  id: number
  name: string
  code: string
  category: SportCategory
  icon?: string
  banner?: string
  description?: string
  is_active: boolean
  sort_order: number
  created_at: string
  updated_at: string
  configuration?: SportConfiguration
}

// 运动配置
export interface SportConfiguration {
  id: number
  sport_type_id: number
  
  // 功能开关
  enable_realtime: boolean
  enable_chat: boolean
  enable_voting: boolean
  enable_prediction: boolean
  enable_leaderboard: boolean
  
  // 预测设置
  allow_modification: boolean
  max_modifications: number
  modification_deadline: number
  
  // 投票设置
  enable_self_voting: boolean
  max_votes_per_user: number
  voting_deadline: number
  
  created_at: string
  updated_at: string
  sport_type?: SportType
}

// 积分规则
export interface ScoringRule {
  id: number
  sport_type_id: number
  name: string
  description?: string
  is_active: boolean
  
  // 基础积分设置
  base_points: number
  enable_difficulty: boolean
  difficulty_multiplier: number
  
  // 奖励组件开关
  enable_vote_reward: boolean
  vote_reward_points: number
  max_vote_reward: number
  
  enable_time_reward: boolean
  time_reward_points: number
  time_reward_hours: number
  
  // 惩罚组件开关
  enable_modify_penalty: boolean
  modify_penalty_points: number
  max_modify_penalty: number
  
  created_at: string
  updated_at: string
  sport_type?: SportType
}

// API请求类型

// 创建管理员请求
export interface CreateAdminRequest {
  user_id: number
  admin_level: AdminLevel
  permissions?: string[]
  sport_type_ids?: number[]
}

// 更新管理员请求
export interface UpdateAdminRequest {
  admin_level?: AdminLevel
  is_active?: boolean
  permissions?: string[]
  sport_type_ids?: number[]
}

// 管理员列表请求
export interface ListAdminsRequest {
  page?: number
  page_size?: number
  admin_level?: AdminLevel
  is_active?: boolean
  search?: string
}

// 管理员列表响应
export interface ListAdminsResponse {
  admins: AdminUser[]
  total: number
  page: number
  page_size: number
}

// 权限操作请求
export interface PermissionRequest {
  permissions: string[]
}

// 运动类型访问权限请求
export interface SportAccessRequest {
  sport_type_ids: number[]
}

// 创建运动类型请求
export interface CreateSportTypeRequest {
  name: string
  code: string
  category: SportCategory
  icon?: string
  banner?: string
  description?: string
  sort_order?: number
  configuration?: Partial<SportConfiguration>
}

// 更新运动类型请求
export interface UpdateSportTypeRequest {
  name?: string
  code?: string
  category?: SportCategory
  icon?: string
  banner?: string
  description?: string
  is_active?: boolean
  sort_order?: number
}

// 运动类型列表请求
export interface ListSportTypesRequest {
  page?: number
  page_size?: number
  category?: SportCategory
  is_active?: boolean
  search?: string
}

// 运动类型列表响应
export interface ListSportTypesResponse {
  sport_types: SportType[]
  total: number
  page: number
  page_size: number
}

// 更新运动配置请求
export interface UpdateSportConfigurationRequest {
  enable_realtime?: boolean
  enable_chat?: boolean
  enable_voting?: boolean
  enable_prediction?: boolean
  enable_leaderboard?: boolean
  allow_modification?: boolean
  max_modifications?: number
  modification_deadline?: number
  enable_self_voting?: boolean
  max_votes_per_user?: number
  voting_deadline?: number
}

// 创建积分规则请求
export interface CreateScoringRuleRequest {
  sport_type_id: number
  name: string
  description?: string
  base_points?: number
  enable_difficulty?: boolean
  difficulty_multiplier?: number
  enable_vote_reward?: boolean
  vote_reward_points?: number
  max_vote_reward?: number
  enable_time_reward?: boolean
  time_reward_points?: number
  time_reward_hours?: number
  enable_modify_penalty?: boolean
  modify_penalty_points?: number
  max_modify_penalty?: number
}

// 更新积分规则请求
export interface UpdateScoringRuleRequest {
  name?: string
  description?: string
  is_active?: boolean
  base_points?: number
  enable_difficulty?: boolean
  difficulty_multiplier?: number
  enable_vote_reward?: boolean
  vote_reward_points?: number
  max_vote_reward?: number
  enable_time_reward?: boolean
  time_reward_points?: number
  time_reward_hours?: number
  enable_modify_penalty?: boolean
  modify_penalty_points?: number
  max_modify_penalty?: number
}

// 积分规则列表请求
export interface ListScoringRulesRequest {
  page?: number
  page_size?: number
  sport_type_id?: number
  is_active?: boolean
  search?: string
}

// 积分规则列表响应
export interface ListScoringRulesResponse {
  scoring_rules: ScoringRule[]
  total: number
  page: number
  page_size: number
}

// 审计日志列表请求
export interface ListAuditLogsRequest {
  page?: number
  page_size?: number
  admin_user_id?: number
  action?: string
  resource?: string
  status?: AuditStatus
  start_date?: string
  end_date?: string
}

// 审计日志列表响应
export interface ListAuditLogsResponse {
  logs: AdminAuditLog[]
  total: number
  page: number
  page_size: number
}

// 审计统计请求
export interface AuditStatsRequest {
  start_date?: string
  end_date?: string
  admin_user_id?: number
}

// 审计统计响应
export interface AuditStatsResponse {
  total_actions: number
  success_actions: number
  failed_actions: number
  success_rate: number
  actions_by_day: Array<{
    date: string
    count: number
  }>
  actions_by_admin: Array<{
    admin_user_id: number
    admin_name: string
    count: number
  }>
  actions_by_resource: Array<{
    resource: string
    count: number
  }>
}

// 运动类型统计
export interface SportTypeStats {
  sport_type_id: number
  match_count: number
  prediction_count: number
  active_users: number
  last_activity: string
}

// 预定义权限常量
export const ADMIN_PERMISSIONS = {
  SPORT_TYPE_MANAGE: 'sport_type.manage',
  SPORT_CONFIG_MANAGE: 'sport_config.manage',
  SCORING_RULE_MANAGE: 'scoring_rule.manage',
  MATCH_MANAGE: 'match.manage',
  USER_MANAGE: 'user.manage',
  ADMIN_MANAGE: 'admin.manage',
  AUDIT_LOG_VIEW: 'audit_log.view',
  SYSTEM_CONFIG: 'system.config'
} as const

// 管理员级别显示名称
export const ADMIN_LEVEL_NAMES = {
  [AdminLevel.SPORT]: '运动管理员',
  [AdminLevel.SYSTEM]: '系统管理员',
  [AdminLevel.SUPER]: '超级管理员'
} as const

// 审计状态显示名称
export const AUDIT_STATUS_NAMES = {
  [AuditStatus.SUCCESS]: '成功',
  [AuditStatus.FAILED]: '失败',
  [AuditStatus.PARTIAL]: '部分成功'
} as const

// 运动类别显示名称
export const SPORT_CATEGORY_NAMES = {
  [SportCategory.ESPORTS]: '电子竞技',
  [SportCategory.TRADITIONAL]: '传统体育'
} as const