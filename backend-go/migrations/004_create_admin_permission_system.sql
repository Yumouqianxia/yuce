-- 004_create_admin_permission_system.sql
-- 创建管理员权限系统相关表

-- 管理员权限表
CREATE TABLE IF NOT EXISTS admin_permissions (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    code VARCHAR(50) NOT NULL UNIQUE COMMENT '权限代码',
    name VARCHAR(100) NOT NULL COMMENT '权限名称',
    description TEXT COMMENT '权限描述',
    category VARCHAR(50) DEFAULT '' COMMENT '权限分类',
    is_active BOOLEAN DEFAULT TRUE COMMENT '是否启用',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    
    INDEX idx_code (code),
    INDEX idx_category (category),
    INDEX idx_is_active (is_active)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='管理员权限表';

-- 管理员用户扩展表
CREATE TABLE IF NOT EXISTS admin_users (
    user_id INT UNSIGNED PRIMARY KEY COMMENT '用户ID',
    admin_level TINYINT UNSIGNED DEFAULT 1 COMMENT '管理员级别: 1-运动管理员, 2-系统管理员, 3-超级管理员',
    is_active BOOLEAN DEFAULT TRUE COMMENT '是否启用',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_admin_level (admin_level),
    INDEX idx_is_active (is_active),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='管理员用户扩展表';

-- 管理员权限关联表
CREATE TABLE IF NOT EXISTS admin_user_permissions (
    admin_user_id INT UNSIGNED NOT NULL COMMENT '管理员用户ID',
    admin_permission_id INT UNSIGNED NOT NULL COMMENT '权限ID',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    
    PRIMARY KEY (admin_user_id, admin_permission_id),
    FOREIGN KEY (admin_user_id) REFERENCES admin_users(user_id) ON DELETE CASCADE,
    FOREIGN KEY (admin_permission_id) REFERENCES admin_permissions(id) ON DELETE CASCADE,
    INDEX idx_admin_user_id (admin_user_id),
    INDEX idx_admin_permission_id (admin_permission_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='管理员权限关联表';

-- 管理员运动类型访问权限关联表
CREATE TABLE IF NOT EXISTS admin_sport_access (
    admin_user_id INT UNSIGNED NOT NULL COMMENT '管理员用户ID',
    sport_type_id INT UNSIGNED NOT NULL COMMENT '运动类型ID',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    
    PRIMARY KEY (admin_user_id, sport_type_id),
    FOREIGN KEY (admin_user_id) REFERENCES admin_users(user_id) ON DELETE CASCADE,
    FOREIGN KEY (sport_type_id) REFERENCES sport_types(id) ON DELETE CASCADE,
    INDEX idx_admin_user_id (admin_user_id),
    INDEX idx_sport_type_id (sport_type_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='管理员运动类型访问权限关联表';

-- 管理员操作审计日志表
CREATE TABLE IF NOT EXISTS admin_audit_logs (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    admin_user_id INT UNSIGNED NOT NULL COMMENT '管理员用户ID',
    action VARCHAR(100) NOT NULL COMMENT '操作类型',
    resource VARCHAR(100) NOT NULL COMMENT '资源类型',
    resource_id VARCHAR(50) DEFAULT '' COMMENT '资源ID',
    method VARCHAR(10) NOT NULL COMMENT 'HTTP方法',
    path VARCHAR(255) NOT NULL COMMENT '请求路径',
    ip_address VARCHAR(45) DEFAULT '' COMMENT 'IP地址',
    user_agent TEXT COMMENT '用户代理',
    
    -- 操作详情 (JSON格式)
    old_values JSON COMMENT '修改前的值',
    new_values JSON COMMENT '修改后的值',
    changes JSON COMMENT '变更内容',
    
    -- 操作结果
    status TINYINT UNSIGNED DEFAULT 1 COMMENT '操作状态: 1-成功, 2-失败, 3-部分成功',
    error_msg TEXT COMMENT '错误信息',
    duration BIGINT UNSIGNED DEFAULT 0 COMMENT '执行时间(毫秒)',
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    
    FOREIGN KEY (admin_user_id) REFERENCES admin_users(user_id) ON DELETE CASCADE,
    INDEX idx_admin_user_id (admin_user_id),
    INDEX idx_action (action),
    INDEX idx_resource (resource),
    INDEX idx_status (status),
    INDEX idx_created_at (created_at),
    INDEX idx_method (method),
    INDEX idx_path (path(100))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='管理员操作审计日志表';

-- 插入默认权限数据
INSERT INTO admin_permissions (code, name, description, category) VALUES
-- 运动类型管理权限
('sport_type.manage', '运动类型管理', '创建、编辑、删除运动类型', '运动管理'),
('sport_config.manage', '运动配置管理', '管理运动类型的功能配置', '运动管理'),

-- 积分规则管理权限
('scoring_rule.manage', '积分规则管理', '创建、编辑、删除积分规则', '积分管理'),

-- 比赛管理权限
('match.manage', '比赛管理', '创建、编辑、删除比赛', '比赛管理'),

-- 用户管理权限
('user.manage', '用户管理', '管理普通用户账户', '用户管理'),

-- 管理员管理权限
('admin.manage', '管理员管理', '管理管理员账户和权限', '管理员管理'),

-- 审计日志权限
('audit_log.view', '审计日志查看', '查看管理员操作审计日志', '审计管理'),

-- 系统配置权限
('system.config', '系统配置', '管理系统级别配置', '系统管理')

ON DUPLICATE KEY UPDATE
    name = VALUES(name),
    description = VALUES(description),
    category = VALUES(category),
    updated_at = CURRENT_TIMESTAMP;

-- 为现有的管理员用户创建管理员记录
INSERT INTO admin_users (user_id, admin_level, is_active)
SELECT id, 3, TRUE  -- 设置为超级管理员
FROM users 
WHERE role = 'admin'
ON DUPLICATE KEY UPDATE
    admin_level = VALUES(admin_level),
    is_active = VALUES(is_active),
    updated_at = CURRENT_TIMESTAMP;