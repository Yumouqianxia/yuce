# 管理员权限系统

## 概述

管理员权限系统为预测平台提供了细粒度的权限控制和操作审计功能。支持三级管理员权限，每个管理员可以拥有不同的权限组合和运动类型访问权限。

## 数据模型

### AdminUser (管理员用户)

管理员用户是对现有用户系统的扩展，为具有管理权限的用户提供额外的权限信息。

**字段说明:**
- `UserID`: 关联的用户ID (主键)
- `AdminLevel`: 管理员级别 (1-运动管理员, 2-系统管理员, 3-超级管理员)
- `IsActive`: 是否启用
- `Permissions`: 权限列表 (多对多关系)
- `SportTypes`: 可访问的运动类型 (多对多关系)

**管理员级别:**
- `AdminLevelSport` (1): 运动管理员 - 只能管理分配的运动类型
- `AdminLevelSystem` (2): 系统管理员 - 可以管理所有运动类型和用户
- `AdminLevelSuper` (3): 超级管理员 - 拥有所有权限

**业务方法:**
- `IsSuperAdmin()`: 检查是否为超级管理员
- `IsSystemAdmin()`: 检查是否为系统管理员或以上
- `HasPermission(permission)`: 检查是否有指定权限
- `HasSportAccess(sportTypeID)`: 检查是否有运动类型访问权限

### AdminPermission (管理员权限)

定义系统中所有可用的管理员权限。

**字段说明:**
- `ID`: 主键
- `Code`: 权限代码 (如: sport_type.manage)
- `Name`: 权限名称
- `Description`: 权限描述
- `Category`: 权限分类
- `IsActive`: 是否启用

**权限分类:**
- `sport_management`: 运动管理
- `match_management`: 比赛管理
- `user_management`: 用户管理
- `prediction_management`: 预测管理
- `system_management`: 系统管理
- `analytics`: 数据分析

### AdminAuditLog (管理员审计日志)

记录所有管理员操作，用于审计和问题追踪。

**字段说明:**
- `ID`: 主键
- `AdminUserID`: 操作的管理员ID
- `Action`: 操作动作
- `Resource`: 操作资源
- `ResourceID`: 资源ID
- `Method`: HTTP方法
- `Path`: 请求路径
- `IPAddress`: IP地址
- `UserAgent`: 用户代理
- `OldValues`: 变更前数据 (JSON)
- `NewValues`: 变更后数据 (JSON)
- `Changes`: 变更内容 (JSON)
- `Status`: 操作状态
- `ErrorMsg`: 错误信息
- `Duration`: 执行时间(毫秒)

**审计状态:**
- `AuditStatusSuccess` (1): 成功
- `AuditStatusFailed` (2): 失败
- `AuditStatusPartial` (3): 部分成功

**业务方法:**
- `IsSuccess()`: 检查操作是否成功
- `GetDurationMs()`: 获取执行时间

## 预定义权限

### 运动管理权限
- `sport_type.manage`: 运动类型管理
- `sport_config.manage`: 运动配置管理
- `scoring_rule.manage`: 积分规则管理

### 比赛管理权限
- `match.create`: 创建比赛
- `match.edit`: 编辑比赛
- `match.delete`: 删除比赛
- `match.result`: 设置比赛结果

### 用户管理权限
- `user.view`: 查看用户
- `user.edit`: 编辑用户
- `user.ban`: 封禁用户
- `user.points`: 管理用户积分

### 预测管理权限
- `prediction.view`: 查看预测
- `prediction.feature`: 精选预测
- `prediction.delete`: 删除预测

### 系统管理权限
- `admin.manage`: 管理员管理
- `audit_log.view`: 查看审计日志
- `system.config`: 系统配置
- `cache.manage`: 缓存管理

### 数据分析权限
- `analytics.view`: 数据分析
- `report.export`: 导出报告

## 数据库表结构

### admin_users 表
```sql
CREATE TABLE admin_users (
    user_id BIGINT UNSIGNED PRIMARY KEY,
    admin_level TINYINT DEFAULT 1,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
```

### admin_permissions 表
```sql
CREATE TABLE admin_permissions (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    code VARCHAR(50) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    category VARCHAR(50) DEFAULT '',
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);
```

### admin_user_permissions 表 (多对多关系)
```sql
CREATE TABLE admin_user_permissions (
    admin_user_user_id BIGINT UNSIGNED NOT NULL,
    admin_permission_id BIGINT UNSIGNED NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (admin_user_user_id, admin_permission_id),
    FOREIGN KEY (admin_user_user_id) REFERENCES admin_users(user_id) ON DELETE CASCADE,
    FOREIGN KEY (admin_permission_id) REFERENCES admin_permissions(id) ON DELETE CASCADE
);
```

### admin_sport_access 表 (多对多关系)
```sql
CREATE TABLE admin_sport_access (
    admin_user_user_id BIGINT UNSIGNED NOT NULL,
    sport_type_id BIGINT UNSIGNED NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (admin_user_user_id, sport_type_id),
    FOREIGN KEY (admin_user_user_id) REFERENCES admin_users(user_id) ON DELETE CASCADE,
    FOREIGN KEY (sport_type_id) REFERENCES sport_types(id) ON DELETE CASCADE
);
```

### admin_audit_logs 表
```sql
CREATE TABLE admin_audit_logs (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    admin_user_id BIGINT UNSIGNED NOT NULL,
    action VARCHAR(100) NOT NULL,
    resource VARCHAR(100) NOT NULL,
    resource_id VARCHAR(50) DEFAULT '',
    method VARCHAR(10) NOT NULL,
    path VARCHAR(255) NOT NULL,
    ip_address VARCHAR(45) DEFAULT '',
    user_agent TEXT,
    old_values JSON,
    new_values JSON,
    changes JSON,
    status TINYINT DEFAULT 1,
    error_msg TEXT,
    duration BIGINT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (admin_user_id) REFERENCES admin_users(user_id) ON DELETE CASCADE
);
```

## 权限控制逻辑

### 权限检查优先级

1. **超级管理员**: 拥有所有权限，跳过所有权限检查
2. **系统管理员**: 拥有所有运动类型访问权限，需要检查具体功能权限
3. **运动管理员**: 只能访问分配的运动类型，需要检查功能权限和运动类型权限

### 运动类型访问控制

```go
func (au *AdminUser) HasSportAccess(sportTypeID uint) bool {
    // 超级管理员和系统管理员拥有所有运动类型访问权限
    if au.IsSystemAdmin() {
        return true
    }

    // 运动管理员只能访问分配的运动类型
    for _, sport := range au.SportTypes {
        if sport.ID == sportTypeID {
            return true
        }
    }
    return false
}
```

## 使用示例

### 创建管理员

```go
// 创建运动管理员
adminUser := &admin.AdminUser{
    UserID:     123,
    AdminLevel: admin.AdminLevelSport,
    IsActive:   true,
}

// 分配权限
permissions := []admin.AdminPermission{
    {Code: admin.PermissionSportTypeManage},
    {Code: admin.PermissionMatchManage},
}
adminUser.Permissions = permissions

// 分配运动类型访问权限
sportTypes := []admin.SportType{
    {ID: 1}, // LOL
    {ID: 2}, // 王者荣耀
}
adminUser.SportTypes = sportTypes
```

### 权限检查

```go
// 检查是否有权限
if adminUser.HasPermission(admin.PermissionSportTypeManage) {
    // 允许管理运动类型
}

// 检查运动类型访问权限
if adminUser.HasSportAccess(1) { // LOL
    // 允许访问LOL相关功能
}

// 检查管理员级别
if adminUser.IsSuperAdmin() {
    // 超级管理员特殊处理
}
```

### 审计日志记录

```go
// 记录操作日志
auditLog := &admin.AdminAuditLog{
    AdminUserID: adminUser.UserID,
    Action:      "CREATE",
    Resource:    "sport_type",
    ResourceID:  "1",
    Method:      "POST",
    Path:        "/api/v1/admin/sport-types",
    IPAddress:   "192.168.1.1",
    Status:      admin.AuditStatusSuccess,
    Duration:    150, // 毫秒
}

// 记录变更内容
auditLog.NewValues = datatypes.JSON(`{"name": "英雄联盟", "code": "lol"}`)
```

## 安全考虑

1. **权限最小化原则**: 管理员只分配必要的权限
2. **运动类型隔离**: 运动管理员只能访问分配的运动类型
3. **操作审计**: 所有管理员操作都被记录
4. **会话管理**: 管理员会话有效期控制
5. **IP限制**: 可以限制管理员登录IP范围

## 扩展性

系统设计支持以下扩展：

1. **新权限添加**: 通过权限表添加新的权限类型
2. **权限分组**: 可以创建权限组简化权限分配
3. **时间限制**: 可以为权限添加时间限制
4. **地域限制**: 可以为管理员添加地域访问限制
5. **审批流程**: 可以为敏感操作添加审批流程