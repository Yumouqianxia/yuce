-- 创建运动类型表
CREATE TABLE sport_types (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE COMMENT '运动名称',
    code VARCHAR(20) NOT NULL UNIQUE COMMENT '运动代码',
    category VARCHAR(20) NOT NULL COMMENT '运动类别: esports, traditional',
    icon VARCHAR(255) DEFAULT '' COMMENT '运动图标URL',
    banner VARCHAR(255) DEFAULT '' COMMENT '运动横幅图URL',
    description TEXT COMMENT '运动描述',
    is_active BOOLEAN DEFAULT TRUE COMMENT '是否启用',
    sort_order INT DEFAULT 0 COMMENT '显示顺序',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_sport_types_category (category),
    INDEX idx_sport_types_is_active (is_active),
    INDEX idx_sport_types_sort_order (sort_order)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='运动类型表';

-- 创建运动配置表
CREATE TABLE sport_configurations (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    sport_type_id BIGINT UNSIGNED NOT NULL UNIQUE COMMENT '运动类型ID',
    
    -- 功能开关
    enable_realtime BOOLEAN DEFAULT TRUE COMMENT '启用实时通信',
    enable_chat BOOLEAN DEFAULT FALSE COMMENT '启用聊天功能',
    enable_voting BOOLEAN DEFAULT TRUE COMMENT '启用投票功能',
    enable_prediction BOOLEAN DEFAULT TRUE COMMENT '启用预测功能',
    enable_leaderboard BOOLEAN DEFAULT TRUE COMMENT '启用排行榜',
    
    -- 预测设置
    allow_modification BOOLEAN DEFAULT TRUE COMMENT '允许修改预测',
    max_modifications INT DEFAULT 3 COMMENT '最大修改次数',
    modification_deadline INT DEFAULT 30 COMMENT '修改截止时间(分钟)',
    
    -- 投票设置
    enable_self_voting BOOLEAN DEFAULT FALSE COMMENT '允许给自己投票',
    max_votes_per_user INT DEFAULT 10 COMMENT '每用户最大投票数',
    voting_deadline INT DEFAULT 0 COMMENT '投票截止时间(分钟)',
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    FOREIGN KEY (sport_type_id) REFERENCES sport_types(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='运动配置表';

-- 创建积分规则表
CREATE TABLE scoring_rules (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    sport_type_id BIGINT UNSIGNED NOT NULL COMMENT '运动类型ID',
    name VARCHAR(100) NOT NULL COMMENT '规则名称',
    description TEXT COMMENT '规则描述',
    is_active BOOLEAN DEFAULT TRUE COMMENT '是否启用',
    
    -- 基础积分设置
    base_points INT DEFAULT 10 COMMENT '基础积分',
    enable_difficulty BOOLEAN DEFAULT FALSE COMMENT '启用难度系数',
    difficulty_multiplier DECIMAL(3,2) DEFAULT 1.00 COMMENT '难度系数',
    
    -- 奖励组件
    enable_vote_reward BOOLEAN DEFAULT FALSE COMMENT '启用投票奖励',
    vote_reward_points INT DEFAULT 1 COMMENT '每票奖励积分',
    max_vote_reward INT DEFAULT 10 COMMENT '最大投票奖励',
    
    enable_time_reward BOOLEAN DEFAULT FALSE COMMENT '启用时间奖励',
    time_reward_points INT DEFAULT 5 COMMENT '时间奖励积分',
    time_reward_hours INT DEFAULT 24 COMMENT '时间奖励小时数',
    
    -- 惩罚组件
    enable_modify_penalty BOOLEAN DEFAULT FALSE COMMENT '启用修改惩罚',
    modify_penalty_points INT DEFAULT 2 COMMENT '每次修改扣分',
    max_modify_penalty INT DEFAULT 6 COMMENT '最大修改惩罚',
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_scoring_rules_sport_type (sport_type_id),
    INDEX idx_scoring_rules_is_active (is_active),
    FOREIGN KEY (sport_type_id) REFERENCES sport_types(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='积分规则表';