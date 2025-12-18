package admin

import (
	"time"

	"gorm.io/datatypes"
)

// AdminLevel 管理员级别
type AdminLevel int

const (
	AdminLevelSport  AdminLevel = iota + 1 // 运动管理员
	AdminLevelSystem                       // 系统管理员
	AdminLevelSuper                        // 超级管理员
)

// AdminUser 管理员用户扩展
type AdminUser struct {
	UserID    uint       `json:"user_id" gorm:"primaryKey"`
	AdminLevel AdminLevel `json:"admin_level" gorm:"default:1"`
	IsActive  bool       `json:"is_active" gorm:"default:true"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`

	// 多对多关系
	Permissions []AdminPermission `json:"permissions,omitempty" gorm:"many2many:admin_user_permissions;"`
	SportTypes  []SportType       `json:"sport_types,omitempty" gorm:"many2many:admin_sport_access;"`
}

// AdminPermission 管理员权限
type AdminPermission struct {
	ID          uint   `json:"id" gorm:"primaryKey"`
	Code        string `json:"code" gorm:"uniqueIndex;size:50;not null"`
	Name        string `json:"name" gorm:"size:100;not null"`
	Description string `json:"description" gorm:"type:text"`
	Category    string `json:"category" gorm:"size:50"`
	IsActive    bool   `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// AdminAuditLog 管理员操作审计日志
type AdminAuditLog struct {
	ID          uint   `json:"id" gorm:"primaryKey"`
	AdminUserID uint   `json:"admin_user_id" gorm:"index;not null"`
	Action      string `json:"action" gorm:"size:100;not null"`
	Resource    string `json:"resource" gorm:"size:100;not null"`
	ResourceID  string `json:"resource_id" gorm:"size:50"`
	Method      string `json:"method" gorm:"size:10;not null"`
	Path        string `json:"path" gorm:"size:255;not null"`
	IPAddress   string `json:"ip_address" gorm:"size:45"`
	UserAgent   string `json:"user_agent" gorm:"type:text"`

	// 操作详情 (JSON格式存储)
	OldValues datatypes.JSON `json:"old_values,omitempty"`
	NewValues datatypes.JSON `json:"new_values,omitempty"`
	Changes   datatypes.JSON `json:"changes,omitempty"`

	// 操作结果
	Status   AuditStatus `json:"status" gorm:"default:1"`
	ErrorMsg string      `json:"error_msg,omitempty" gorm:"type:text"`
	Duration int64       `json:"duration"` // 毫秒

	CreatedAt time.Time `json:"created_at"`

	// 关联管理员信息
	AdminUser *AdminUser `json:"admin_user,omitempty" gorm:"foreignKey:AdminUserID"`
}

// AuditStatus 审计状态
type AuditStatus int

const (
	AuditStatusSuccess AuditStatus = iota + 1
	AuditStatusFailed
	AuditStatusPartial
)

// SportType 运动类型 (为了避免循环导入，这里只定义ID)
type SportType struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
}

// 预定义权限常量
const (
	PermissionSportTypeManage   = "sport_type.manage"
	PermissionSportConfigManage = "sport_config.manage"
	PermissionScoringRuleManage = "scoring_rule.manage"
	PermissionMatchManage       = "match.manage"
	PermissionUserManage        = "user.manage"
	PermissionAdminManage       = "admin.manage"
	PermissionAuditLogView      = "audit_log.view"
	PermissionSystemConfig      = "system.config"
)

// GetLevelName 获取管理员级别名称
func (level AdminLevel) GetLevelName() string {
	switch level {
	case AdminLevelSport:
		return "运动管理员"
	case AdminLevelSystem:
		return "系统管理员"
	case AdminLevelSuper:
		return "超级管理员"
	default:
		return "未知级别"
	}
}

// IsSuperAdmin 检查是否为超级管理员
func (au *AdminUser) IsSuperAdmin() bool {
	return au.AdminLevel == AdminLevelSuper
}

// IsSystemAdmin 检查是否为系统管理员或以上
func (au *AdminUser) IsSystemAdmin() bool {
	return au.AdminLevel >= AdminLevelSystem
}

// HasPermission 检查是否有指定权限
func (au *AdminUser) HasPermission(permission string) bool {
	// 超级管理员拥有所有权限
	if au.IsSuperAdmin() {
		return true
	}

	for _, perm := range au.Permissions {
		if perm.Code == permission && perm.IsActive {
			return true
		}
	}
	return false
}

// HasSportAccess 检查是否有运动类型访问权限
func (au *AdminUser) HasSportAccess(sportTypeID uint) bool {
	// 超级管理员和系统管理员拥有所有运动类型访问权限
	if au.IsSystemAdmin() {
		return true
	}

	for _, sport := range au.SportTypes {
		if sport.ID == sportTypeID {
			return true
		}
	}
	return false
}

// GetStatusName 获取审计状态名称
func (status AuditStatus) GetStatusName() string {
	switch status {
	case AuditStatusSuccess:
		return "成功"
	case AuditStatusFailed:
		return "失败"
	case AuditStatusPartial:
		return "部分成功"
	default:
		return "未知状态"
	}
}

// IsSuccess 检查操作是否成功
func (log *AdminAuditLog) IsSuccess() bool {
	return log.Status == AuditStatusSuccess
}

// GetDurationMs 获取执行时间（毫秒）
func (log *AdminAuditLog) GetDurationMs() int64 {
	return log.Duration
}