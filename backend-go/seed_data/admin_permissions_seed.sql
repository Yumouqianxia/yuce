-- 管理员权限种子数据
INSERT INTO admin_permissions (code, name, description, category, is_active) VALUES
-- 运动类型管理权限
('sport_type.manage', '运动类型管理', '创建、编辑、删除运动类型', 'sport_management', TRUE),
('sport_config.manage', '运动配置管理', '配置运动类型的功能开关和参数', 'sport_management', TRUE),
('scoring_rule.manage', '积分规则管理', '配置和管理积分计算规则', 'sport_management', TRUE),

-- 比赛管理权限
('match.create', '创建比赛', '创建新的比赛', 'match_management', TRUE),
('match.edit', '编辑比赛', '编辑比赛信息', 'match_management', TRUE),
('match.delete', '删除比赛', '删除比赛', 'match_management', TRUE),
('match.result', '设置比赛结果', '设置比赛结果和比分', 'match_management', TRUE),

-- 用户管理权限
('user.view', '查看用户', '查看用户信息和统计', 'user_management', TRUE),
('user.edit', '编辑用户', '编辑用户信息', 'user_management', TRUE),
('user.ban', '封禁用户', '封禁和解封用户', 'user_management', TRUE),
('user.points', '管理用户积分', '调整用户积分', 'user_management', TRUE),

-- 预测管理权限
('prediction.view', '查看预测', '查看所有用户预测', 'prediction_management', TRUE),
('prediction.feature', '精选预测', '设置精选预测', 'prediction_management', TRUE),
('prediction.delete', '删除预测', '删除不当预测', 'prediction_management', TRUE),

-- 系统管理权限
('admin.manage', '管理员管理', '创建、编辑、删除管理员账户', 'system_management', TRUE),
('audit_log.view', '查看审计日志', '查看管理员操作日志', 'system_management', TRUE),
('system.config', '系统配置', '修改系统全局配置', 'system_management', TRUE),
('cache.manage', '缓存管理', '清理和管理系统缓存', 'system_management', TRUE),

-- 数据分析权限
('analytics.view', '数据分析', '查看系统数据分析报告', 'analytics', TRUE),
('report.export', '导出报告', '导出各类数据报告', 'analytics', TRUE);