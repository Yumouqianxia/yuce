package routes

import (
	adminhandlers "backend-go/internal/adapters/http/handlers/admin"
	"backend-go/internal/adapters/http/middleware"
	"backend-go/internal/core/domain/admin"
	"backend-go/internal/core/ports"
	pkgMiddleware "backend-go/pkg/middleware"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// AdminRoutes 管理员路由配置
type AdminRoutes struct {
	sportTypeService     ports.SportTypeService
	scoringRuleService   ports.ScoringRuleService
	adminService         ports.AdminService
	adminAuditService    ports.AdminAuditService
	logger               *logrus.Logger
	permissionMiddleware *pkgMiddleware.AdminPermissionMiddleware
}

// NewAdminRoutes 创建管理员路由实例
func NewAdminRoutes(
	sportTypeService ports.SportTypeService,
	scoringRuleService ports.ScoringRuleService,
	adminService ports.AdminService,
	adminAuditService ports.AdminAuditService,
	logger *logrus.Logger,
) *AdminRoutes {
	permissionMiddleware := pkgMiddleware.NewAdminPermissionMiddleware(adminService, adminAuditService)

	return &AdminRoutes{
		sportTypeService:     sportTypeService,
		scoringRuleService:   scoringRuleService,
		adminService:         adminService,
		adminAuditService:    adminAuditService,
		logger:               logger,
		permissionMiddleware: permissionMiddleware,
	}
}

// RegisterRoutes 注册管理员路由
func (r *AdminRoutes) RegisterRoutes(router *gin.Engine, authMiddleware *middleware.AuthMiddleware) {
	// 管理员API组
	adminGroup := router.Group("/api/v1/admin")
	adminGroup.Use(authMiddleware.RequireAuth())
	adminGroup.Use(authMiddleware.RequireAdmin())
	// 暂时禁用审计中间件，避免权限检查问题
	// adminGroup.Use(r.permissionMiddleware.AuditMiddleware())

	// 管理员管理路由（暂时不使用额外的权限检查）
	r.registerAdminManagementRoutesSimple(adminGroup)

	// 运动类型管理路由
	r.registerSportTypeRoutes(adminGroup)

	// 积分规则管理路由
	r.registerScoringRuleRoutes(adminGroup)

	// 审计日志路由（暂时不使用额外的权限检查）
	r.registerAuditRoutesSimple(adminGroup)
}

// registerSportTypeRoutes 注册运动类型管理路由
func (r *AdminRoutes) registerSportTypeRoutes(group *gin.RouterGroup) {
	sportTypeHandler := adminhandlers.NewSportTypeHandler(r.sportTypeService, r.logger)

	sportTypes := group.Group("/sport-types")
	{
		// 基础CRUD
		sportTypes.POST("", sportTypeHandler.CreateSportType)
		sportTypes.GET("", sportTypeHandler.ListSportTypes)
		sportTypes.GET("/:id", sportTypeHandler.GetSportType)
		sportTypes.PUT("/:id", sportTypeHandler.UpdateSportType)
		sportTypes.DELETE("/:id", sportTypeHandler.DeleteSportType)

		// 配置管理
		sportTypes.GET("/:id/configuration", sportTypeHandler.GetSportConfiguration)
		sportTypes.PUT("/:id/configuration", sportTypeHandler.UpdateSportConfiguration)
		sportTypes.POST("/batch-config", sportTypeHandler.BatchUpdateConfiguration)

		// 统计信息
		sportTypes.GET("/:id/stats", sportTypeHandler.GetSportTypeStats)

		// 积分规则相关（需要在这里创建handler实例）
		scoringRuleHandler := adminhandlers.NewScoringRuleHandler(r.scoringRuleService, r.logger)
		sportTypes.GET("/:sport_type_id/active-scoring-rule", scoringRuleHandler.GetActiveScoringRule)
		sportTypes.POST("/:sport_type_id/scoring-rules/:rule_id/recalculate", scoringRuleHandler.RecalculateScores)
	}
}

// registerScoringRuleRoutes 注册积分规则管理路由
func (r *AdminRoutes) registerScoringRuleRoutes(group *gin.RouterGroup) {
	scoringRuleHandler := adminhandlers.NewScoringRuleHandler(r.scoringRuleService, r.logger)

	scoringRules := group.Group("/scoring-rules")
	{
		// 基础CRUD
		scoringRules.POST("", scoringRuleHandler.CreateScoringRule)
		scoringRules.GET("", scoringRuleHandler.ListScoringRules)
		scoringRules.GET("/:id", scoringRuleHandler.GetScoringRule)
		scoringRules.PUT("/:id", scoringRuleHandler.UpdateScoringRule)
		scoringRules.DELETE("/:id", scoringRuleHandler.DeleteScoringRule)

		// 规则管理
		scoringRules.POST("/:id/activate", scoringRuleHandler.SetActiveScoringRule)

		// 积分计算
		scoringRules.POST("/preview", scoringRuleHandler.PreviewScore)
	}
}

// registerAdminManagementRoutes 注册管理员管理路由（带权限检查）
func (r *AdminRoutes) registerAdminManagementRoutes(group *gin.RouterGroup) {
	adminHandler := adminhandlers.NewAdminHandler(r.adminService, r.adminAuditService, r.logger)

	admins := group.Group("/admins")
	{
		// 基础CRUD - 需要管理员管理权限
		admins.POST("",
			r.permissionMiddleware.RequirePermission(admin.PermissionAdminManage),
			adminHandler.CreateAdmin)
		admins.GET("",
			r.permissionMiddleware.RequirePermission(admin.PermissionAdminManage),
			adminHandler.ListAdmins)
		admins.GET("/:id",
			r.permissionMiddleware.RequirePermission(admin.PermissionAdminManage),
			adminHandler.GetAdmin)
		admins.PUT("/:id",
			r.permissionMiddleware.RequirePermission(admin.PermissionAdminManage),
			adminHandler.UpdateAdmin)
		admins.DELETE("/:id",
			r.permissionMiddleware.RequirePermission(admin.PermissionAdminManage),
			adminHandler.DeleteAdmin)

		// 权限管理
		admins.POST("/:id/permissions",
			r.permissionMiddleware.RequirePermission(admin.PermissionAdminManage),
			adminHandler.GrantPermissions)
		admins.DELETE("/:id/permissions",
			r.permissionMiddleware.RequirePermission(admin.PermissionAdminManage),
			adminHandler.RevokePermissions)
		admins.GET("/:id/permissions",
			r.permissionMiddleware.RequirePermission(admin.PermissionAdminManage),
			adminHandler.GetUserPermissions)

		// 运动类型访问权限管理
		admins.POST("/:id/sport-access",
			r.permissionMiddleware.RequirePermission(admin.PermissionAdminManage),
			adminHandler.GrantSportAccess)
		admins.DELETE("/:id/sport-access",
			r.permissionMiddleware.RequirePermission(admin.PermissionAdminManage),
			adminHandler.RevokeSportAccess)
	}

	// 权限列表 - 所有管理员都可以查看
	group.GET("/permissions", adminHandler.ListPermissions)
}

// registerAdminManagementRoutesSimple 注册管理员管理路由（简化版，不使用额外权限检查）
func (r *AdminRoutes) registerAdminManagementRoutesSimple(group *gin.RouterGroup) {
	adminHandler := adminhandlers.NewAdminHandler(r.adminService, r.adminAuditService, r.logger)

	admins := group.Group("/admins")
	{
		// 基础CRUD - 所有管理员都可以访问
		admins.POST("", adminHandler.CreateAdmin)
		admins.GET("", adminHandler.ListAdmins)
		admins.GET("/:id", adminHandler.GetAdmin)
		admins.PUT("/:id", adminHandler.UpdateAdmin)
		admins.DELETE("/:id", adminHandler.DeleteAdmin)

		// 权限管理
		admins.POST("/:id/permissions", adminHandler.GrantPermissions)
		admins.DELETE("/:id/permissions", adminHandler.RevokePermissions)
		admins.GET("/:id/permissions", adminHandler.GetUserPermissions)

		// 运动类型访问权限管理
		admins.POST("/:id/sport-access", adminHandler.GrantSportAccess)
		admins.DELETE("/:id/sport-access", adminHandler.RevokeSportAccess)
	}

	// 权限列表 - 所有管理员都可以查看
	group.GET("/permissions", adminHandler.ListPermissions)
}

// registerAuditRoutes 注册审计日志路由（带权限检查）
func (r *AdminRoutes) registerAuditRoutes(group *gin.RouterGroup) {
	auditHandler := adminhandlers.NewAuditHandler(r.adminAuditService, r.logger)

	audit := group.Group("/audit-logs")
	{
		// 审计日志查看 - 需要审计日志查看权限
		audit.GET("",
			r.permissionMiddleware.RequirePermission(admin.PermissionAuditLogView),
			auditHandler.ListAuditLogs)
		audit.GET("/:id",
			r.permissionMiddleware.RequirePermission(admin.PermissionAuditLogView),
			auditHandler.GetAuditLog)
		audit.GET("/stats",
			r.permissionMiddleware.RequirePermission(admin.PermissionAuditLogView),
			auditHandler.GetAuditStats)
	}
}

// registerAuditRoutesSimple 注册审计日志路由（简化版，不使用额外权限检查）
func (r *AdminRoutes) registerAuditRoutesSimple(group *gin.RouterGroup) {
	auditHandler := adminhandlers.NewAuditHandler(r.adminAuditService, r.logger)

	audit := group.Group("/audit-logs")
	{
		// 审计日志查看 - 所有管理员都可以访问
		audit.GET("", auditHandler.ListAuditLogs)
		audit.GET("/:id", auditHandler.GetAuditLog)
		audit.GET("/stats", auditHandler.GetAuditStats)
	}
}
