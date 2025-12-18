-- 创建比赛积分计算记录表
CREATE TABLE IF NOT EXISTS match_points_calculations (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    match_id INT UNSIGNED NOT NULL,
    results TEXT NOT NULL COMMENT '积分计算结果JSON',
    total_points INT NOT NULL DEFAULT 0 COMMENT '总积分',
    processed_at DATETIME NOT NULL COMMENT '处理时间',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_match_id (match_id),
    UNIQUE KEY uk_match_id (match_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='比赛积分计算记录表';

-- 创建积分更新事件表
CREATE TABLE IF NOT EXISTS points_update_events (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id INT UNSIGNED NOT NULL,
    match_id INT UNSIGNED NOT NULL,
    prediction_id INT UNSIGNED NOT NULL,
    old_points INT NOT NULL DEFAULT 0 COMMENT '原积分',
    new_points INT NOT NULL DEFAULT 0 COMMENT '新积分',
    points_change INT NOT NULL DEFAULT 0 COMMENT '积分变化',
    tournament VARCHAR(50) NOT NULL DEFAULT 'GLOBAL' COMMENT '锦标赛类型',
    timestamp DATETIME NOT NULL COMMENT '事件时间',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_user_id (user_id),
    INDEX idx_match_id (match_id),
    INDEX idx_prediction_id (prediction_id),
    INDEX idx_tournament (tournament),
    INDEX idx_timestamp (timestamp)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='积分更新事件表';

-- 为用户表的积分字段添加索引（如果不存在）
ALTER TABLE users ADD INDEX IF NOT EXISTS idx_points (points DESC);

-- 为比赛表添加锦标赛字段索引（如果不存在）
ALTER TABLE matches ADD INDEX IF NOT EXISTS idx_tournament (tournament);