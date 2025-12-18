-- 开发环境数据库初始化脚本

-- 创建开发数据库
CREATE DATABASE IF NOT EXISTS prediction_system_dev CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 创建测试数据库
CREATE DATABASE IF NOT EXISTS prediction_system_test CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 使用开发数据库
USE prediction_system_dev;

-- 创建用户表
CREATE TABLE IF NOT EXISTS users (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(50) NOT NULL UNIQUE,
    email VARCHAR(100) NOT NULL UNIQUE,
    nickname VARCHAR(50),
    password VARCHAR(255) NOT NULL,
    avatar VARCHAR(255),
    points INT DEFAULT 0,
    role VARCHAR(20) DEFAULT 'user',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_username (username),
    INDEX idx_email (email),
    INDEX idx_points (points DESC)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 创建比赛表
CREATE TABLE IF NOT EXISTS matches (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    team_a VARCHAR(100) NOT NULL,
    team_b VARCHAR(100) NOT NULL,
    tournament VARCHAR(50) DEFAULT 'SPRING',
    status VARCHAR(20) DEFAULT 'UPCOMING',
    start_time TIMESTAMP NOT NULL,
    winner VARCHAR(10),
    score_a INT DEFAULT 0,
    score_b INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_status (status),
    INDEX idx_start_time (start_time),
    INDEX idx_tournament (tournament)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 创建预测表
CREATE TABLE IF NOT EXISTS predictions (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    match_id BIGINT UNSIGNED NOT NULL,
    predicted_winner VARCHAR(10) NOT NULL,
    predicted_score_a INT NOT NULL,
    predicted_score_b INT NOT NULL,
    is_correct BOOLEAN DEFAULT FALSE,
    earned_points INT DEFAULT 0,
    modification_count INT DEFAULT 0,
    vote_count INT DEFAULT 0,
    is_featured BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    UNIQUE KEY idx_user_match (user_id, match_id),
    INDEX idx_match_votes (match_id, vote_count DESC),
    INDEX idx_featured (is_featured, created_at DESC),
    
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (match_id) REFERENCES matches(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 创建投票表
CREATE TABLE IF NOT EXISTS votes (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    prediction_id BIGINT UNSIGNED NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE KEY idx_user_prediction (user_id, prediction_id),
    INDEX idx_prediction (prediction_id),
    
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (prediction_id) REFERENCES predictions(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 插入测试数据
INSERT INTO users (username, email, nickname, password, points, role) VALUES
('admin', 'admin@example.com', '管理员', '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewdBPj/hL/.JklSS', 1000, 'admin'),
('testuser1', 'user1@example.com', '测试用户1', '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewdBPj/hL/.JklSS', 500, 'user'),
('testuser2', 'user2@example.com', '测试用户2', '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewdBPj/hL/.JklSS', 300, 'user'),
('testuser3', 'user3@example.com', '测试用户3', '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewdBPj/hL/.JklSS', 200, 'user');

INSERT INTO matches (team_a, team_b, tournament, status, start_time, winner, score_a, score_b) VALUES
('T1', 'DK', 'SPRING', 'FINISHED', '2024-01-15 14:00:00', 'A', 2, 1),
('GEN', 'KT', 'SPRING', 'FINISHED', '2024-01-16 15:00:00', 'B', 1, 2),
('DRX', 'LSB', 'SPRING', 'UPCOMING', '2024-02-01 16:00:00', NULL, 0, 0),
('HLE', 'NS', 'SPRING', 'UPCOMING', '2024-02-02 17:00:00', NULL, 0, 0);

INSERT INTO predictions (user_id, match_id, predicted_winner, predicted_score_a, predicted_score_b, is_correct, earned_points, vote_count) VALUES
(2, 1, 'A', 2, 1, TRUE, 30, 5),
(3, 1, 'B', 1, 2, FALSE, 0, 2),
(4, 1, 'A', 3, 0, FALSE, 10, 8),
(2, 2, 'A', 2, 0, FALSE, 0, 1),
(3, 2, 'B', 1, 2, TRUE, 30, 3),
(4, 2, 'B', 0, 2, FALSE, 10, 1);

INSERT INTO votes (user_id, prediction_id) VALUES
(3, 1), (4, 1), (2, 3), (3, 3), (4, 3),
(2, 5), (4, 5), (3, 6);

-- 更新用户积分
UPDATE users SET points = (
    SELECT COALESCE(SUM(earned_points), 0) 
    FROM predictions 
    WHERE predictions.user_id = users.id
) WHERE id IN (2, 3, 4);