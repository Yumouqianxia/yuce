package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/datatypes"

	"backend-go/internal/core/domain/admin"
	"backend-go/internal/core/domain/user"
	"backend-go/internal/core/ports"
	"backend-go/pkg/database"
)

// adminService 管理员服务实现
type adminService struct {
	db *database.DB
}

// NewAdminService 创建管理员服务实例
func NewAdminService(db *database.DB) ports.AdminService {
	return &adminService{
		db: db,
	}
}

// CreateAdmin 创建管理员
func (s *adminService) CreateAdmin(ctx context.Context, req *ports.CreateAdminRequest) (*admin.AdminUser, error) {
	// 检查用户是否存在
	var existingUser user.User
	if err := s.db.WithContext(ctx).First(&existingUser, req.UserID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to check user existence: %w", err)
	}

	// 检查是否已经是管理员
	var existingAdmin admin.AdminUser
	if err := s.db.WithContext(ctx).First(&existingAdmin, req.UserID).Error; err == nil {
		return nil, fmt.Errorf("user is already an admin")
	} else if err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to check admin existence: %w", err)
	}

	// 创建管理员记录
	adminUser := &admin.AdminUser{
		UserID:     req.UserID,
		AdminLevel: req.AdminLevel,
		IsActive:   true,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// 开始事务
	return adminUser, s.db.Transaction(func(tx *gorm.DB) error {
		// 创建管理员记录
		if err := tx.WithContext(ctx).Create(adminUser).Error; err != nil {
			return fmt.Errorf("failed to create admin user: %w", err)
		}

		// 授予权限
		if len(req.Permissions) > 0 {
			if err := s.grantPermissionsInTx(ctx, tx, req.UserID, req.Permissions); err != nil {
				return fmt.Errorf("failed to grant permissions: %w", err)
			}
		}

		// 授予运动类型访问权限
		if len(req.SportTypes) > 0 {
			if err := s.grantSportAccessInTx(ctx, tx, req.UserID, req.SportTypes); err != nil {
				return fmt.Errorf("failed to grant sport access: %w", err)
			}
		}

		// 重新加载完整的管理员信息
		return tx.WithContext(ctx).
			Preload("Permissions").
			Preload("SportTypes").
			First(adminUser, req.UserID).Error
	})
}

// UpdateAdmin 更新管理员
func (s *adminService) UpdateAdmin(ctx context.Context, userID uint, req *ports.UpdateAdminRequest) (*admin.AdminUser, error) {
	var adminUser admin.AdminUser
	if err := s.db.WithContext(ctx).First(&adminUser, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("admin not found")
		}
		return nil, fmt.Errorf("failed to get admin: %w", err)
	}

	return &adminUser, s.db.Transaction(func(tx *gorm.DB) error {
		// 更新基本信息
		updates := make(map[string]interface{})
		if req.AdminLevel != nil {
			updates["admin_level"] = *req.AdminLevel
		}
		if req.IsActive != nil {
			updates["is_active"] = *req.IsActive
		}
		updates["updated_at"] = time.Now()

		if len(updates) > 0 {
			if err := tx.WithContext(ctx).Model(&adminUser).Updates(updates).Error; err != nil {
				return fmt.Errorf("failed to update admin: %w", err)
			}
		}

		// 更新权限
		if req.Permissions != nil {
			// 清除现有权限
			if err := tx.WithContext(ctx).Model(&adminUser).Association("Permissions").Clear(); err != nil {
				return fmt.Errorf("failed to clear permissions: %w", err)
			}
			// 授予新权限
			if len(req.Permissions) > 0 {
				if err := s.grantPermissionsInTx(ctx, tx, userID, req.Permissions); err != nil {
					return fmt.Errorf("failed to grant permissions: %w", err)
				}
			}
		}

		// 更新运动类型访问权限
		if req.SportTypes != nil {
			// 清除现有运动类型访问权限
			if err := tx.WithContext(ctx).Model(&adminUser).Association("SportTypes").Clear(); err != nil {
				return fmt.Errorf("failed to clear sport access: %w", err)
			}
			// 授予新的运动类型访问权限
			if len(req.SportTypes) > 0 {
				if err := s.grantSportAccessInTx(ctx, tx, userID, req.SportTypes); err != nil {
					return fmt.Errorf("failed to grant sport access: %w", err)
				}
			}
		}

		// 重新加载完整的管理员信息
		return tx.WithContext(ctx).
			Preload("Permissions").
			Preload("SportTypes").
			First(&adminUser, userID).Error
	})
}

// DeleteAdmin 删除管理员
func (s *adminService) DeleteAdmin(ctx context.Context, userID uint) error {
	var adminUser admin.AdminUser
	if err := s.db.WithContext(ctx).First(&adminUser, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("admin not found")
		}
		return fmt.Errorf("failed to get admin: %w", err)
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		// 清除权限关联
		if err := tx.WithContext(ctx).Model(&adminUser).Association("Permissions").Clear(); err != nil {
			return fmt.Errorf("failed to clear permissions: %w", err)
		}

		// 清除运动类型访问权限关联
		if err := tx.WithContext(ctx).Model(&adminUser).Association("SportTypes").Clear(); err != nil {
			return fmt.Errorf("failed to clear sport access: %w", err)
		}

		// 删除管理员记录
		if err := tx.WithContext(ctx).Delete(&adminUser).Error; err != nil {
			return fmt.Errorf("failed to delete admin: %w", err)
		}

		return nil
	})
}

// GetAdmin 获取管理员信息
func (s *adminService) GetAdmin(ctx context.Context, userID uint) (*admin.AdminUser, error) {
	var adminUser admin.AdminUser
	if err := s.db.WithContext(ctx).
		Preload("Permissions").
		Preload("SportTypes").
		First(&adminUser, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("admin not found")
		}
		return nil, fmt.Errorf("failed to get admin: %w", err)
	}

	return &adminUser, nil
}

// ListAdmins 获取管理员列表
func (s *adminService) ListAdmins(ctx context.Context, req *ports.ListAdminsRequest) (*ports.ListAdminsResponse, error) {
	// 设置默认分页参数
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	query := s.db.WithContext(ctx).
		Table("admin_users").
		Select("admin_users.*, users.username, users.email, users.nickname, users.avatar").
		Joins("LEFT JOIN users ON admin_users.user_id = users.id")

	// 添加过滤条件
	if req.AdminLevel != nil {
		query = query.Where("admin_users.admin_level = ?", *req.AdminLevel)
	}
	if req.IsActive != nil {
		query = query.Where("admin_users.is_active = ?", *req.IsActive)
	}
	if req.Search != "" {
		query = query.Where("users.username LIKE ? OR users.email LIKE ? OR users.nickname LIKE ?",
			"%"+req.Search+"%", "%"+req.Search+"%", "%"+req.Search+"%")
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count admins: %w", err)
	}

	// 获取分页数据
	var results []struct {
		admin.AdminUser
		Username string `json:"username"`
		Email    string `json:"email"`
		Nickname string `json:"nickname"`
		Avatar   string `json:"avatar"`
	}

	offset := (req.Page - 1) * req.PageSize
	if err := query.
		Order("admin_users.created_at DESC").
		Limit(req.PageSize).
		Offset(offset).
		Find(&results).Error; err != nil {
		return nil, fmt.Errorf("failed to get admins: %w", err)
	}

	// 转换结果
	admins := make([]*ports.AdminUserWithUser, len(results))
	for i, result := range results {
		admins[i] = &ports.AdminUserWithUser{
			AdminUser: &result.AdminUser,
			Username:  result.Username,
			Email:     result.Email,
			Nickname:  result.Nickname,
			Avatar:    result.Avatar,
		}
	}

	totalPages := int((total + int64(req.PageSize) - 1) / int64(req.PageSize))

	return &ports.ListAdminsResponse{
		Admins:     admins,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
	}, nil
}

// HasPermission 检查权限
func (s *adminService) HasPermission(ctx context.Context, userID uint, permission string) (bool, error) {
	var adminUser admin.AdminUser
	if err := s.db.WithContext(ctx).
		Preload("Permissions").
		First(&adminUser, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil // 不是管理员，没有权限
		}
		return false, fmt.Errorf("failed to get admin: %w", err)
	}

	if !adminUser.IsActive {
		return false, nil // 管理员已禁用
	}

	return adminUser.HasPermission(permission), nil
}

// HasSportAccess 检查运动类型访问权限
func (s *adminService) HasSportAccess(ctx context.Context, userID uint, sportTypeID uint) (bool, error) {
	var adminUser admin.AdminUser
	if err := s.db.WithContext(ctx).
		Preload("SportTypes").
		First(&adminUser, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil // 不是管理员，没有权限
		}
		return false, fmt.Errorf("failed to get admin: %w", err)
	}

	if !adminUser.IsActive {
		return false, nil // 管理员已禁用
	}

	return adminUser.HasSportAccess(sportTypeID), nil
}

// GetUserPermissions 获取用户权限列表
func (s *adminService) GetUserPermissions(ctx context.Context, userID uint) ([]string, error) {
	var adminUser admin.AdminUser
	if err := s.db.WithContext(ctx).
		Preload("Permissions").
		First(&adminUser, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return []string{}, nil // 不是管理员，返回空权限列表
		}
		return nil, fmt.Errorf("failed to get admin: %w", err)
	}

	if !adminUser.IsActive {
		return []string{}, nil // 管理员已禁用
	}

	// 超级管理员拥有所有权限
	if adminUser.IsSuperAdmin() {
		var allPermissions []admin.AdminPermission
		if err := s.db.WithContext(ctx).Where("is_active = ?", true).Find(&allPermissions).Error; err != nil {
			return nil, fmt.Errorf("failed to get all permissions: %w", err)
		}

		permissions := make([]string, len(allPermissions))
		for i, perm := range allPermissions {
			permissions[i] = perm.Code
		}
		return permissions, nil
	}

	// 普通管理员返回已授予的权限
	permissions := make([]string, 0, len(adminUser.Permissions))
	for _, perm := range adminUser.Permissions {
		if perm.IsActive {
			permissions = append(permissions, perm.Code)
		}
	}

	return permissions, nil
}

// GrantPermissions 授予权限
func (s *adminService) GrantPermissions(ctx context.Context, userID uint, permissions []string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		return s.grantPermissionsInTx(ctx, tx, userID, permissions)
	})
}

// RevokePermissions 撤销权限
func (s *adminService) RevokePermissions(ctx context.Context, userID uint, permissions []string) error {
	if len(permissions) == 0 {
		return nil
	}

	var adminUser admin.AdminUser
	if err := s.db.WithContext(ctx).First(&adminUser, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("admin not found")
		}
		return fmt.Errorf("failed to get admin: %w", err)
	}

	// 获取要撤销的权限
	var permsToRevoke []admin.AdminPermission
	if err := s.db.WithContext(ctx).
		Where("code IN ? AND is_active = ?", permissions, true).
		Find(&permsToRevoke).Error; err != nil {
		return fmt.Errorf("failed to get permissions: %w", err)
	}

	if len(permsToRevoke) == 0 {
		return nil // 没有有效权限需要撤销
	}

	// 撤销权限关联
	if err := s.db.WithContext(ctx).Model(&adminUser).Association("Permissions").Delete(permsToRevoke); err != nil {
		return fmt.Errorf("failed to revoke permissions: %w", err)
	}

	return nil
}

// GrantSportAccess 授予运动类型访问权限
func (s *adminService) GrantSportAccess(ctx context.Context, userID uint, sportTypeIDs []uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		return s.grantSportAccessInTx(ctx, tx, userID, sportTypeIDs)
	})
}

// RevokeSportAccess 撤销运动类型访问权限
func (s *adminService) RevokeSportAccess(ctx context.Context, userID uint, sportTypeIDs []uint) error {
	if len(sportTypeIDs) == 0 {
		return nil
	}

	var adminUser admin.AdminUser
	if err := s.db.WithContext(ctx).First(&adminUser, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("admin not found")
		}
		return fmt.Errorf("failed to get admin: %w", err)
	}

	// 获取要撤销的运动类型
	var sportsToRevoke []admin.SportType
	if err := s.db.WithContext(ctx).
		Table("sport_types").
		Where("id IN ?", sportTypeIDs).
		Find(&sportsToRevoke).Error; err != nil {
		return fmt.Errorf("failed to get sport types: %w", err)
	}

	if len(sportsToRevoke) == 0 {
		return nil // 没有有效运动类型需要撤销
	}

	// 撤销运动类型访问权限关联
	if err := s.db.WithContext(ctx).Model(&adminUser).Association("SportTypes").Delete(sportsToRevoke); err != nil {
		return fmt.Errorf("failed to revoke sport access: %w", err)
	}

	return nil
}

// ListPermissions 获取权限列表
func (s *adminService) ListPermissions(ctx context.Context) ([]*admin.AdminPermission, error) {
	var permissions []*admin.AdminPermission
	if err := s.db.WithContext(ctx).
		Where("is_active = ?", true).
		Order("category, name").
		Find(&permissions).Error; err != nil {
		return nil, fmt.Errorf("failed to get permissions: %w", err)
	}

	return permissions, nil
}

// GetPermission 获取权限信息
func (s *adminService) GetPermission(ctx context.Context, code string) (*admin.AdminPermission, error) {
	var permission admin.AdminPermission
	if err := s.db.WithContext(ctx).
		Where("code = ? AND is_active = ?", code, true).
		First(&permission).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("permission not found")
		}
		return nil, fmt.Errorf("failed to get permission: %w", err)
	}

	return &permission, nil
}

// grantPermissionsInTx 在事务中授予权限
func (s *adminService) grantPermissionsInTx(ctx context.Context, tx *gorm.DB, userID uint, permissions []string) error {
	if len(permissions) == 0 {
		return nil
	}

	var adminUser admin.AdminUser
	if err := tx.WithContext(ctx).First(&adminUser, userID).Error; err != nil {
		return fmt.Errorf("admin not found: %w", err)
	}

	// 获取有效的权限
	var validPermissions []admin.AdminPermission
	if err := tx.WithContext(ctx).
		Where("code IN ? AND is_active = ?", permissions, true).
		Find(&validPermissions).Error; err != nil {
		return fmt.Errorf("failed to get permissions: %w", err)
	}

	if len(validPermissions) == 0 {
		return nil // 没有有效权限需要授予
	}

	// 授予权限
	if err := tx.WithContext(ctx).Model(&adminUser).Association("Permissions").Append(validPermissions); err != nil {
		return fmt.Errorf("failed to grant permissions: %w", err)
	}

	return nil
}

// grantSportAccessInTx 在事务中授予运动类型访问权限
func (s *adminService) grantSportAccessInTx(ctx context.Context, tx *gorm.DB, userID uint, sportTypeIDs []uint) error {
	if len(sportTypeIDs) == 0 {
		return nil
	}

	var adminUser admin.AdminUser
	if err := tx.WithContext(ctx).First(&adminUser, userID).Error; err != nil {
		return fmt.Errorf("admin not found: %w", err)
	}

	// 获取有效的运动类型
	var validSportTypes []admin.SportType
	if err := tx.WithContext(ctx).
		Table("sport_types").
		Where("id IN ?", sportTypeIDs).
		Find(&validSportTypes).Error; err != nil {
		return fmt.Errorf("failed to get sport types: %w", err)
	}

	if len(validSportTypes) == 0 {
		return nil // 没有有效运动类型需要授予
	}

	// 授予运动类型访问权限
	if err := tx.WithContext(ctx).Model(&adminUser).Association("SportTypes").Append(validSportTypes); err != nil {
		return fmt.Errorf("failed to grant sport access: %w", err)
	}

	return nil
}

// adminAuditService 管理员审计服务实现
type adminAuditService struct {
	db *database.DB
}

// NewAdminAuditService 创建管理员审计服务实例
func NewAdminAuditService(db *database.DB) ports.AdminAuditService {
	return &adminAuditService{
		db: db,
	}
}

// LogAction 记录操作
func (s *adminAuditService) LogAction(ctx context.Context, req *ports.LogActionRequest) error {
	auditLog := &admin.AdminAuditLog{
		AdminUserID: req.AdminUserID,
		Action:      req.Action,
		Resource:    req.Resource,
		ResourceID:  req.ResourceID,
		Method:      req.Method,
		Path:        req.Path,
		IPAddress:   req.IPAddress,
		UserAgent:   req.UserAgent,
		Status:      req.Status,
		ErrorMsg:    req.ErrorMsg,
		Duration:    req.Duration,
		CreatedAt:   time.Now(),
	}

	// 序列化JSON数据
	if req.OldValues != nil {
		if data, err := json.Marshal(req.OldValues); err == nil {
			auditLog.OldValues = datatypes.JSON(data)
		}
	}
	if req.NewValues != nil {
		if data, err := json.Marshal(req.NewValues); err == nil {
			auditLog.NewValues = datatypes.JSON(data)
		}
	}
	if req.Changes != nil {
		if data, err := json.Marshal(req.Changes); err == nil {
			auditLog.Changes = datatypes.JSON(data)
		}
	}

	if err := s.db.WithContext(ctx).Create(auditLog).Error; err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}

	return nil
}

// GetAuditLog 获取审计日志
func (s *adminAuditService) GetAuditLog(ctx context.Context, id uint) (*admin.AdminAuditLog, error) {
	var auditLog admin.AdminAuditLog
	if err := s.db.WithContext(ctx).
		Preload("AdminUser").
		First(&auditLog, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("audit log not found")
		}
		return nil, fmt.Errorf("failed to get audit log: %w", err)
	}

	return &auditLog, nil
}

// ListAuditLogs 获取审计日志列表
func (s *adminAuditService) ListAuditLogs(ctx context.Context, req *ports.ListAuditLogsRequest) (*ports.ListAuditLogsResponse, error) {
	// 设置默认分页参数
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	query := s.db.WithContext(ctx).Model(&admin.AdminAuditLog{})

	// 添加过滤条件
	if req.AdminUserID != nil {
		query = query.Where("admin_user_id = ?", *req.AdminUserID)
	}
	if req.Action != "" {
		query = query.Where("action LIKE ?", "%"+req.Action+"%")
	}
	if req.Resource != "" {
		query = query.Where("resource = ?", req.Resource)
	}
	if req.Status != nil {
		query = query.Where("status = ?", *req.Status)
	}
	if req.StartTime != nil && *req.StartTime != "" {
		query = query.Where("created_at >= ?", *req.StartTime)
	}
	if req.EndTime != nil && *req.EndTime != "" {
		query = query.Where("created_at <= ?", *req.EndTime)
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count audit logs: %w", err)
	}

	// 获取分页数据
	var logs []*admin.AdminAuditLog
	offset := (req.Page - 1) * req.PageSize
	if err := query.
		Preload("AdminUser").
		Order("created_at DESC").
		Limit(req.PageSize).
		Offset(offset).
		Find(&logs).Error; err != nil {
		return nil, fmt.Errorf("failed to get audit logs: %w", err)
	}

	totalPages := int((total + int64(req.PageSize) - 1) / int64(req.PageSize))

	return &ports.ListAuditLogsResponse{
		Logs:       logs,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
	}, nil
}

// GetAuditStats 获取审计统计
func (s *adminAuditService) GetAuditStats(ctx context.Context, req *ports.AuditStatsRequest) (*ports.AuditStatsResponse, error) {
	query := s.db.WithContext(ctx).Model(&admin.AdminAuditLog{})

	// 添加时间过滤
	if req.StartTime != nil && *req.StartTime != "" {
		query = query.Where("created_at >= ?", *req.StartTime)
	}
	if req.EndTime != nil && *req.EndTime != "" {
		query = query.Where("created_at <= ?", *req.EndTime)
	}
	if req.AdminUserID != nil {
		query = query.Where("admin_user_id = ?", *req.AdminUserID)
	}

	// 获取总操作数
	var totalActions int64
	if err := query.Count(&totalActions).Error; err != nil {
		return nil, fmt.Errorf("failed to count total actions: %w", err)
	}

	// 获取成功操作数
	var successActions int64
	if err := query.Where("status = ?", admin.AuditStatusSuccess).Count(&successActions).Error; err != nil {
		return nil, fmt.Errorf("failed to count success actions: %w", err)
	}

	// 获取失败操作数
	var failedActions int64
	if err := query.Where("status = ?", admin.AuditStatusFailed).Count(&failedActions).Error; err != nil {
		return nil, fmt.Errorf("failed to count failed actions: %w", err)
	}

	// 按操作类型统计
	var actionsByType []struct {
		Action string `json:"action"`
		Count  int64  `json:"count"`
	}
	if err := query.Select("action, COUNT(*) as count").Group("action").Find(&actionsByType).Error; err != nil {
		return nil, fmt.Errorf("failed to get actions by type: %w", err)
	}

	actionsByTypeMap := make(map[string]int64)
	for _, item := range actionsByType {
		actionsByTypeMap[item.Action] = item.Count
	}

	// 按管理员统计
	var actionsByAdmin []struct {
		AdminUserID uint  `json:"admin_user_id"`
		Count       int64 `json:"count"`
	}
	if err := query.Select("admin_user_id, COUNT(*) as count").Group("admin_user_id").Find(&actionsByAdmin).Error; err != nil {
		return nil, fmt.Errorf("failed to get actions by admin: %w", err)
	}

	actionsByAdminMap := make(map[uint]int64)
	for _, item := range actionsByAdmin {
		actionsByAdminMap[item.AdminUserID] = item.Count
	}

	// 计算平均执行时间
	var avgDuration float64
	if err := query.Select("AVG(duration) as avg_duration").Scan(&avgDuration).Error; err != nil {
		return nil, fmt.Errorf("failed to get average duration: %w", err)
	}

	return &ports.AuditStatsResponse{
		TotalActions:   totalActions,
		SuccessActions: successActions,
		FailedActions:  failedActions,
		ActionsByType:  actionsByTypeMap,
		ActionsByAdmin: actionsByAdminMap,
		AvgDuration:    avgDuration,
	}, nil
}