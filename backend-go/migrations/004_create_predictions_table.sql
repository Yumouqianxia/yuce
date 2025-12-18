-- 创建预测表
CREATE TABLE IF NOT EXISTS predictions (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    match_id BIGINT UNSIGNED NOT NULL,
    predicted_winner VARCHAR(10) NOT NULL,
    predicted_score_a INT NOT NULL DEFAULT 0,
    predicted_score_b INT NOT NULL DEFAULT 0,
    is_correct BOOLEAN DEFAULT FALSE,
    earned_points INT DEFAULT 0,
    modification_count INT DEFAULT 0,
    vote_count INT DEFAULT 0,
    is_featured BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    -- 索引
    INDEX idx_predictions_user_id (user_id),
    INDEX idx_predictions_match_id (match_id),
    INDEX idx_predictions_user_match (user_id, match_id),
    INDEX idx_predictions_match_votes (match_id, vote_count DESC),
    INDEX idx_predictions_featured (is_featured, created_at DESC),
    
    -- 外键约束
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (match_id) REFERENCES matches(id) ON DELETE CASCADE,
    
    -- 唯一约束：每个用户对每场比赛只能有一个预测
    UNIQUE KEY uk_user_match (user_id, match_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 创建投票表
CREATE TABLE IF NOT EXISTS votes (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    prediction_id BIGINT UNSIGNED NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- 索引
    INDEX idx_votes_user_id (user_id),
    INDEX idx_votes_prediction_id (prediction_id),
    
    -- 外键约束
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (prediction_id) REFERENCES predictions(id) ON DELETE CASCADE,
    
    -- 唯一约束：每个用户对每个预测只能投一票
    UNIQUE KEY uk_user_prediction (user_id, prediction_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;