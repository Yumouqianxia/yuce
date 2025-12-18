-- 创建事件记录表
CREATE TABLE IF NOT EXISTS event_records (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    event_type VARCHAR(100) NOT NULL,
    event_id VARCHAR(100) NOT NULL UNIQUE,
    user_id BIGINT UNSIGNED,
    payload TEXT NOT NULL,
    metadata TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    processed_at TIMESTAMP NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    retry_count INT NOT NULL DEFAULT 0,
    error_message TEXT,
    
    INDEX idx_event_records_event_type (event_type),
    INDEX idx_event_records_user_id (user_id),
    INDEX idx_event_records_created_at (created_at),
    INDEX idx_event_records_status (status),
    INDEX idx_event_records_event_type_created_at (event_type, created_at),
    INDEX idx_event_records_user_id_created_at (user_id, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 添加外键约束（如果用户表存在）
-- ALTER TABLE event_records 
-- ADD CONSTRAINT fk_event_records_user_id 
-- FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL;