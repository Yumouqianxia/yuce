package services

import (
	"context"
	"time"

	"backend-go/internal/core/domain/admin"
	"backend-go/pkg/database"
)

// InitializeAdminPermissions 初始化管理员权限系统
func InitializeAdminPermissions(db *database.DB) error {
	ctx := context.Background()

	// 检查权限是否已经初始化
	var count int64
	if err := db.WithContext(ctx).Model(&admin.AdminPermission{}).Count(&count).Error; err != nil {
		return err
	}

	// 如果已有权限数据，跳过初始化
	if count > 0 {
		return nil
	}

	// 定义默认权限
	permissions := []*admin.AdminPermission{
		{
			Code:        admin.PermissionSportTypeManage,
			Name:        "运动类型管理",
			Description: "创建、编辑、删除运动类型",
			Category:    "运动管理",
			IsActive:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			Code:        admin.PermissionSportConfigManage,
			Name:        "运动配置管理",
			Description: "管理运动类型的功能配置",
			Category:    "运动管理",
			IsActive:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			Code:        admin.PermissionScoringRuleManage,
			Name:        "积分规则管理",
			Description: "创建、编辑、删除积分规则",
			Category:    "积分管理",
			IsActive:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			Code:        admin.PermissionMatchManage,
			Name:        "比赛管理",
			Description: "创建、编辑、删除比赛",
			Category:    "比赛管理",
			IsActive:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			Code:        admin.PermissionUserManage,
			Name:        "用户管理",
			Description: "管理普通用户账户",
			Category:    "用户管理",
			IsActive:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			Code:        admin.PermissionAdminManage,
			Name:        "管理员管理",
			Description: "管理管理员账户和权限",
			Category:    "管理员管理",
			IsActive:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			Code:        admin.PermissionAuditLogView,
			Name:        "审计日志查看",
			Description: "查看管理员操作审计日志",
			Category:    "审计管理",
			IsActive:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			Code:        admin.PermissionSystemConfig,
			Name:        "系统配置",
			Description: "管理系统级别配置",
			Category:    "系统管理",
			IsActive:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	// 批量创建权限
	if err := db.WithContext(ctx).Create(&permissions).Error; err != nil {
		return err
	}

	return nil
}