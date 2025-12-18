-- 创建积分规则表
CREATE TABLE IF NOT EXISTS scoring_rules (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL COMMENT '规则名称',
    description VARCHAR(500) DEFAULT '' COMMENT '规则描述',
    correct_team_correct_score INT DEFAULT 0 COMMENT '预测正确队伍和比分的积分',
    correct_team_wrong_score INT DEFAULT 0 COMMENT '预测正确队伍错误比分的积分',
    wrong_team_correct_score INT DEFAULT 0 COMMENT '预测错误队伍正确比分的积分',
    wrong_team_wrong_score INT DEFAULT 0 COMMENT '预测错误队伍错误比分的积分',
    is_active BOOLEAN DEFAULT FALSE COMMENT '是否激活',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    
    INDEX idx_scoring_rules_active (is_active),
    INDEX idx_scoring_rules_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='积分规则表';

-- 插入默认积分规则
INSERT INTO scoring_rules (
    name, 
    description, 
    correct_team_correct_score, 
    correct_team_wrong_score, 
    wrong_team_correct_score, 
    wrong_team_wrong_score, 
    is_active
) VALUES (
    '默认积分规则',
    '系统默认的积分计算规则：预测正确队伍和比分30分，预测正确队伍错误比分10分，其他情况0分',
    30,  -- 预测正确队伍和比分
    10,  -- 预测正确队伍错误比分
    0,   -- 预测错误队伍正确比分
    0,   -- 预测错误队伍错误比分
    TRUE -- 默认激活
);

-- 插入示例积分规则
INSERT INTO scoring_rules (
    name, 
    description, 
    correct_team_correct_score, 
    correct_team_wrong_score, 
    wrong_team_correct_score, 
    wrong_team_wrong_score, 
    is_active
) VALUES (
    '平衡积分规则',
    '更平衡的积分规则：各种情况都有相应积分奖励',
    50,  -- 预测正确队伍和比分
    20,  -- 预测正确队伍错误比分
    10,  -- 预测错误队伍正确比分
    5,   -- 预测错误队伍错误比分
    FALSE
);

INSERT INTO scoring_rules (
    name, 
    description, 
    correct_team_correct_score, 
    correct_team_wrong_score, 
    wrong_team_correct_score, 
    wrong_team_wrong_score, 
    is_active
) VALUES (
    '高风险高回报规则',
    '高风险高回报：完全正确获得高分，其他情况分数较低',
    100, -- 预测正确队伍和比分
    15,  -- 预测正确队伍错误比分
    5,   -- 预测错误队伍正确比分
    1,   -- 预测错误队伍错误比分
    FALSE
);