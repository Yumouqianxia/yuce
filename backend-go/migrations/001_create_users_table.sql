-- 创建用户表
CREATE TABLE IF NOT EXISTS users (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(50) NOT NULL,
    email VARCHAR(100) NOT NULL,
    nickname VARCHAR(50) DEFAULT '',
    password VARCHAR(255) NOT NULL,
    avatar VARCHAR(255) DEFAULT '',
    points INT NOT NULL DEFAULT 0,
    role VARCHAR(20) NOT NULL DEFAULT 'user',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    last_password_change DATETIME NULL,
    
    -- 唯一性约束
    UNIQUE KEY idx_username (username),
    UNIQUE KEY idx_email (email),
    
    -- 性能索引
    INDEX idx_points (points DESC),
    INDEX idx_created_at (created_at),
    INDEX idx_role (role),
    
    -- 复合索引用于排行榜查询
    INDEX idx_points_created (points DESC, created_at DESC)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 添加约束检查
ALTER TABLE users 
ADD CONSTRAINT chk_role CHECK (role IN ('user', 'admin')),
ADD CONSTRAINT chk_points CHECK (points >= 0),
ADD CONSTRAINT chk_username_length CHECK (CHAR_LENGTH(username) >= 3),
ADD CONSTRAINT chk_email_format CHECK (email REGEXP '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\\.[A-Za-z]{2,}$');