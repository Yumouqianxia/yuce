-- 创建比赛表
CREATE TABLE IF NOT EXISTS matches (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    team_a VARCHAR(100) NOT NULL COMMENT '队伍A名称',
    team_b VARCHAR(100) NOT NULL COMMENT '队伍B名称',
    tournament VARCHAR(50) NOT NULL DEFAULT 'SPRING' COMMENT '赛事类型',
    status VARCHAR(20) NOT NULL DEFAULT 'UPCOMING' COMMENT '比赛状态',
    start_time DATETIME NOT NULL COMMENT '开始时间',
    winner VARCHAR(10) DEFAULT NULL COMMENT '获胜者 (A/B)',
    score_a INT NOT NULL DEFAULT 0 COMMENT '队伍A得分',
    score_b INT NOT NULL DEFAULT 0 COMMENT '队伍B得分',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    
    INDEX idx_matches_status (status),
    INDEX idx_matches_start_time (start_time),
    INDEX idx_matches_tournament (tournament),
    INDEX idx_matches_status_start_time (status, start_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='比赛表';

-- 添加约束检查
ALTER TABLE matches 
ADD CONSTRAINT chk_matches_status 
CHECK (status IN ('UPCOMING', 'LIVE', 'FINISHED', 'CANCELLED'));

ALTER TABLE matches 
ADD CONSTRAINT chk_matches_tournament 
CHECK (tournament IN ('SPRING', 'SUMMER', 'WORLDS'));

ALTER TABLE matches 
ADD CONSTRAINT chk_matches_winner 
CHECK (winner IS NULL OR winner IN ('A', 'B'));

ALTER TABLE matches 
ADD CONSTRAINT chk_matches_scores 
CHECK (score_a >= 0 AND score_b >= 0);

-- 插入一些测试数据
INSERT INTO matches (team_a, team_b, tournament, status, start_time) VALUES
('T1', 'GenG', 'WORLDS', 'UPCOMING', '2025-09-07 20:00:00'),
('DK', 'KT', 'SPRING', 'FINISHED', '2025-09-06 18:00:00'),
('BRO', 'HLE', 'SUMMER', 'LIVE', '2025-09-06 22:00:00');

-- 更新已完成比赛的结果
UPDATE matches SET winner = 'A', score_a = 2, score_b = 1 WHERE team_a = 'DK' AND team_b = 'KT';