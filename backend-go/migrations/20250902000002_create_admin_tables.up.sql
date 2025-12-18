-- 创建管理员权限表
CREATE TABLE admin_permissions (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    code VARCHAR(50) NOT NULL UNIQUE COMMENT '权限代码',
    name VARCHAR(100) NOT NULL COMMENT '权限名称',
    description TEXT COMMENT '权限描述',
    category VARCHAR(50) DEFAULT '' COMMENT '权限分类',
    is_active BOOLEAN DEFAULT TRUE COMMENT '是否启用',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_admin_permissions_category (category),
    INDEX idx_admin_permissions_is_active (is_active)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='管理员权限表';

-- 创建管理员用户表
CREATE TABLE admin_users (
    user_id BIGINT UNSIGNED PRIMARY KEY COMMENT '用户ID',
    admin_level TINYINT DEFAULT 1 COMMENT '管理员级别: 1-运动管理员, 2-系统管理员, 3-超级管理员',
    is_active BOOLEAN DEFAULT TRUE COMMENT '是否启用',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_admin_users_level (admin_level),
    INDEX idx_admin_users_is_active (is_active),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='管理员用户表';

-- 创建管理员权限关联表
CREATE TABLE admin_user_permissions (
    admin_user_user_id BIGINT UNSIGNED NOT NULL COMMENT '管理员用户ID',
    admin_permission_id BIGINT UNSIGNED NOT NULL COMMENT '权限ID',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    PRIMARY KEY (admin_user_user_id, admin_permission_id),
    FOREIGN KEY (admin_user_user_id) REFERENCES admin_users(user_id) ON DELETE CASCADE,
    FOREIGN KEY (admin_permission_id) REFERENCES admin_permissions(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='管理员权限关联表';

-- 创建管理员运动类型访问权限表
CREATE TABLE admin_sport_access (
    admin_user_user_id BIGINT UNSIGNED NOT NULL COMMENT '管理员用户ID',
    sport_type_id BIGINT UNSIGNED NOT NULL COMMENT '运动类型ID',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    PRIMARY KEY (admin_user_user_id, sport_type_id),
    FOREIGN KEY (admin_user_user_id) REFERENCES admin_users(user_id) ON DELETE CASCADE,
    FOREIGN KEY (sport_type_id) REFERENCES sport_types(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='管理员运动类型访问权限表';

-- 创建管理员审计日志表
CREATE TABLE admin_audit_logs (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    admin_user_id BIGINT UNSIGNED NOT NULL COMMENT '管理员用户ID',
    action VARCHAR(100) NOT NULL COMMENT '操作动作',
    resource VARCHAR(100) NOT NULL COMMENT '操作资源',
    resource_id VARCHAR(50) DEFAULT '' COMMENT '资源ID',
    method VARCHAR(10) NOT NULL COMMENT 'HTTP方法',
    path VARCHAR(255) NOT NULL COMMENT '请求路径',
    ip_address VARCHAR(45) DEFAULT '' COMMENT 'IP地址',
    user_agent TEXT COMMENT '用户代理',
    
    -- 操作详情 (JSON格式)
    old_values JSON COMMENT '变更前数据',
    new_values JSON COMMENT '变更后数据',
    changes JSON COMMENT '变更内容',
    
    -- 操作结果
    status TINYINT DEFAULT 1 COMMENT '状态: 1-成功, 2-失败, 3-部分成功',
    error_msg TEXT COMMENT '错误信息',
    duration BIGINT DEFAULT 0 COMMENT '执行时间(毫秒)',
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_admin_audit_logs_admin_user (admin_user_id),
    INDEX idx_admin_audit_logs_action (action),
    INDEX idx_admin_audit_logs_resource (resource),
    INDEX idx_admin_audit_logs_status (status),
    INDEX idx_admin_audit_logs_created_at (created_at),
    FOREIGN KEY (admin_user_id) REFERENCES admin_users(user_id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='管理员审计日志表';