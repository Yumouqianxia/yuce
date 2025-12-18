package ports

import (
	"context"

	"backend-go/internal/core/domain/admin"
)

// AdminService 管理员服务接口
type AdminService interface {
	// 管理员管理
	CreateAdmin(ctx context.Context, req *CreateAdminRequest) (*admin.AdminUser, error)
	UpdateAdmin(ctx context.Context, userID uint, req *UpdateAdminRequest) (*admin.AdminUser, error)
	DeleteAdmin(ctx context.Context, userID uint) error
	GetAdmin(ctx context.Context, userID uint) (*admin.AdminUser, error)
	ListAdmins(ctx context.Context, req *ListAdminsRequest) (*ListAdminsResponse, error)

	// 权限检查
	HasPermission(ctx context.Context, userID uint, permission string) (bool, error)
	HasSportAccess(ctx context.Context, userID uint, sportTypeID uint) (bool, error)
	GetUserPermissions(ctx context.Context, userID uint) ([]string, error)

	// 权限管理
	GrantPermissions(ctx context.Context, userID uint, permissions []string) error
	RevokePermissions(ctx context.Context, userID uint, permissions []string) error
	GrantSportAccess(ctx context.Context, userID uint, sportTypeIDs []uint) error
	RevokeSportAccess(ctx context.Context, userID uint, sportTypeIDs []uint) error

	// 权限列表
	ListPermissions(ctx context.Context) ([]*admin.AdminPermission, error)
	GetPermission(ctx context.Context, code string) (*admin.AdminPermission, error)
}

// AdminAuditService 管理员审计服务接口
type AdminAuditService interface {
	// 审计日志记录
	LogAction(ctx context.Context, req *LogActionRequest) error
	
	// 审计日志查询
	GetAuditLog(ctx context.Context, id uint) (*admin.AdminAuditLog, error)
	ListAuditLogs(ctx context.Context, req *ListAuditLogsRequest) (*ListAuditLogsResponse, error)
	
	// 审计统计
	GetAuditStats(ctx context.Context, req *AuditStatsRequest) (*AuditStatsResponse, error)
}

// CreateAdminRequest 创建管理员请求
type CreateAdminRequest struct {
	UserID      uint                `json:"user_id" binding:"required"`
	AdminLevel  admin.AdminLevel    `json:"admin_level" binding:"required,min=1,max=3"`
	Permissions []string            `json:"permissions,omitempty"`
	SportTypes  []uint              `json:"sport_types,omitempty"`
}

// UpdateAdminRequest 更新管理员请求
type UpdateAdminRequest struct {
	AdminLevel  *admin.AdminLevel   `json:"admin_level,omitempty"`
	IsActive    *bool               `json:"is_active,omitempty"`
	Permissions []string            `json:"permissions,omitempty"`
	SportTypes  []uint              `json:"sport_types,omitempty"`
}

// ListAdminsRequest 管理员列表请求
type ListAdminsRequest struct {
	Page       int                 `json:"page" form:"page"`
	PageSize   int                 `json:"page_size" form:"page_size"`
	AdminLevel *admin.AdminLevel   `json:"admin_level,omitempty" form:"admin_level"`
	IsActive   *bool               `json:"is_active,omitempty" form:"is_active"`
	Search     string              `json:"search,omitempty" form:"search"`
}

// ListAdminsResponse 管理员列表响应
type ListAdminsResponse struct {
	Admins     []*AdminUserWithUser `json:"admins"`
	Total      int64                `json:"total"`
	Page       int                  `json:"page"`
	PageSize   int                  `json:"page_size"`
	TotalPages int                  `json:"total_pages"`
}

// AdminUserWithUser 包含用户信息的管理员
type AdminUserWithUser struct {
	*admin.AdminUser
	Username string `json:"username"`
	Email    string `json:"email"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}

// LogActionRequest 记录操作请求
type LogActionRequest struct {
	AdminUserID uint                `json:"admin_user_id"`
	Action      string              `json:"action"`
	Resource    string              `json:"resource"`
	ResourceID  string              `json:"resource_id,omitempty"`
	Method      string              `json:"method"`
	Path        string              `json:"path"`
	IPAddress   string              `json:"ip_address"`
	UserAgent   string              `json:"user_agent"`
	OldValues   interface{}         `json:"old_values,omitempty"`
	NewValues   interface{}         `json:"new_values,omitempty"`
	Changes     interface{}         `json:"changes,omitempty"`
	Status      admin.AuditStatus   `json:"status"`
	ErrorMsg    string              `json:"error_msg,omitempty"`
	Duration    int64               `json:"duration"`
}

// ListAuditLogsRequest 审计日志列表请求
type ListAuditLogsRequest struct {
	Page        int                 `json:"page" form:"page"`
	PageSize    int                 `json:"page_size" form:"page_size"`
	AdminUserID *uint               `json:"admin_user_id,omitempty" form:"admin_user_id"`
	Action      string              `json:"action,omitempty" form:"action"`
	Resource    string              `json:"resource,omitempty" form:"resource"`
	Status      *admin.AuditStatus  `json:"status,omitempty" form:"status"`
	StartTime   *string             `json:"start_time,omitempty" form:"start_time"`
	EndTime     *string             `json:"end_time,omitempty" form:"end_time"`
}

// ListAuditLogsResponse 审计日志列表响应
type ListAuditLogsResponse struct {
	Logs       []*admin.AdminAuditLog `json:"logs"`
	Total      int64                  `json:"total"`
	Page       int                    `json:"page"`
	PageSize   int                    `json:"page_size"`
	TotalPages int                    `json:"total_pages"`
}

// AuditStatsRequest 审计统计请求
type AuditStatsRequest struct {
	StartTime   *string `json:"start_time,omitempty" form:"start_time"`
	EndTime     *string `json:"end_time,omitempty" form:"end_time"`
	AdminUserID *uint   `json:"admin_user_id,omitempty" form:"admin_user_id"`
}

// AuditStatsResponse 审计统计响应
type AuditStatsResponse struct {
	TotalActions    int64                    `json:"total_actions"`
	SuccessActions  int64                    `json:"success_actions"`
	FailedActions   int64                    `json:"failed_actions"`
	ActionsByType   map[string]int64         `json:"actions_by_type"`
	ActionsByAdmin  map[uint]int64           `json:"actions_by_admin"`
	AvgDuration     float64                  `json:"avg_duration"`
}