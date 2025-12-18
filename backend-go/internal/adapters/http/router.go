package http

import (
	"time"

	"backend-go/internal/adapters/http/handlers"
	"backend-go/internal/adapters/http/middleware"
	"backend-go/internal/adapters/http/routes"
	"backend-go/internal/core/domain/leaderboard"
	"backend-go/internal/core/domain/match"
	"backend-go/internal/core/domain/prediction"
	"backend-go/internal/core/domain/scoring"
	"backend-go/internal/core/domain/user"
	"backend-go/internal/core/ports"
	"backend-go/internal/shared/logger"
	"backend-go/pkg/middleware/cors"
	requestid "backend-go/pkg/middleware/request_id"
	"backend-go/pkg/response"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"
)

// RouterConfig 路由配置
type RouterConfig struct {
	UserService        user.Service
	MatchService       match.Service
	PredictionService  prediction.Service
	LeaderboardService leaderboard.Service
	ScoringService     scoring.Service
	TeamService        ports.TeamService

	// 管理员系统服务
	AdminService       ports.AdminService
	AdminAuditService  ports.AdminAuditService
	SportTypeService   ports.SportTypeService
	ScoringRuleService ports.ScoringRuleService

	// 数据库连接（用于简单的管理功能）
	DB *gorm.DB
}

// SetupRouter 设置路由
func SetupRouter(config RouterConfig) *gin.Engine {
	// 设置 Gin 模式
	gin.SetMode(gin.ReleaseMode)

	// 创建路由器
	router := gin.New()

	// 添加全局中间件
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(requestid.RequestID())
	router.Use(cors.CORS())
	// 静态资源（头像等）
	router.Static("/uploads", "./uploads")

	// 添加限流中间件
	router.Use(middleware.RateLimit(100, time.Minute)) // 每分钟100个请求

	// 健康检查端点
	router.GET("/health", func(c *gin.Context) {
		response.OK(c, "Service is healthy", gin.H{
			"status":    "ok",
			"timestamp": time.Now().Unix(),
			"service":   "prediction-system-api",
		})
	})

	// API 主路由组（/api）
	api := router.Group("/api")

	// 注册认证路由
	authRoutes := routes.NewAuthRoutes(config.UserService)
	authRoutes.RegisterRoutes(api)

	// 兼容前端老路径（未带 /api 前缀的直接路由）
	authNoPrefix := router.Group("/")
	authRoutes.RegisterRoutes(authNoPrefix)

	// 注册比赛路由
	matchRoutes := routes.NewMatchRoutes(config.MatchService, authRoutes.GetAuthMiddleware())
	matchRoutes.RegisterRoutes(api)

	// 上传路由
	uploadRoutes := routes.NewUploadRoutes(authRoutes.GetAuthMiddleware(), "./uploads")
	uploadRoutes.RegisterRoutes(api)

	// 注册战队路由
	if config.TeamService != nil {
		teamRoutes := routes.NewTeamRoutes(config.TeamService, authRoutes.GetAuthMiddleware())
		teamRoutes.RegisterRoutes(api)
	}

	// 注册预测路由
	predictionRoutes := routes.NewPredictionRoutes(config.PredictionService, authRoutes.GetAuthMiddleware())
	predictionRoutes.RegisterRoutes(api)

	// 注册排行榜路由
	leaderboardHandler := handlers.NewLeaderboardHandler(
		config.LeaderboardService,
		config.ScoringService,
		logger.GetLogger(),
	)
	routes.RegisterLeaderboardRoutes(api, leaderboardHandler, authRoutes.GetAuthMiddleware())

	// 注册管理员路由（暂时禁用，因为服务未完全实现）
	// if config.AdminService != nil && config.SportTypeService != nil {
	// 	adminRoutes := routes.NewAdminRoutes(
	// 		config.SportTypeService,
	// 		config.ScoringRuleService,
	// 		config.AdminService,
	// 		config.AdminAuditService,
	// 		logger.GetLogger(),
	// 	)
	// 	adminRoutes.RegisterRoutes(router, authRoutes.GetAuthMiddleware())
	// }

	// 注册简单的管理API（用户和公告管理）
	if config.DB != nil {
		// 自动迁移基础管理表，避免缺表导致500
		if err := config.DB.AutoMigrate(&handlers.Announcement{}, &handlers.SystemSettings{}); err != nil {
			logger.GetLogger().WithError(err).Error("Failed to auto migrate admin tables")
		}

		announcementHandler := handlers.NewAnnouncementHandler(config.DB, logger.GetLogger())

		// 公告公共读取接口（无登录，仅最新）
		publicAPI := router.Group("/api")
		{
			publicAPI.GET("/announcements/latest", announcementHandler.GetLatestAnnouncement)
		}

		adminAPI := router.Group("/api")
		adminAPI.Use(authRoutes.GetAuthMiddleware().RequireAuth())
		adminAPI.Use(authRoutes.GetAuthMiddleware().RequireAdmin())

		// 用户管理
		userHandler := handlers.NewUserHandler(config.UserService, config.DB, logger.GetLogger())
		users := adminAPI.Group("/users")
		{
			users.GET("", userHandler.ListUsers)
			users.GET("/:id", userHandler.GetUser)
			users.PUT("/:id", userHandler.UpdateUser)

			// 高危操作：删除用户、重置密码需要更高权限（占位）
			usersWithSuper := users.Group("")
			usersWithSuper.Use(authRoutes.GetAuthMiddleware().RequireSuperAdmin())
			usersWithSuper.DELETE("/:id", userHandler.DeleteUser)
			usersWithSuper.POST("/:id/password", userHandler.ResetPassword)
		}

		// 公告管理
		announcements := adminAPI.Group("/announcements")
		{
			announcements.GET("", announcementHandler.ListAnnouncements)
			announcements.GET("/:id", announcementHandler.GetAnnouncement)
			// 写操作提升权限（占位）
			annWrite := announcements.Group("")
			annWrite.Use(authRoutes.GetAuthMiddleware().RequireSuperAdmin())
			annWrite.POST("", announcementHandler.CreateAnnouncement)
			annWrite.PUT("/:id", announcementHandler.UpdateAnnouncement)
			annWrite.DELETE("/:id", announcementHandler.DeleteAnnouncement)
		}

		// 系统设置
		systemSettingsHandler := handlers.NewSystemSettingsHandler(config.DB, logger.GetLogger())
		admin := adminAPI.Group("/admin")
		{
			admin.GET("/settings", systemSettingsHandler.GetSettings)
			admin.Use(authRoutes.GetAuthMiddleware().RequireSuperAdmin())
			admin.POST("/settings", systemSettingsHandler.UpdateSettings)
		}
	}

	// Swagger UI 路由 - 带自定义配置
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler,
		ginSwagger.URL("swagger/doc.json"), // 指定 OpenAPI 规范文件的 URL
		ginSwagger.DeepLinking(true),       // 启用深度链接
		ginSwagger.DocExpansion("list"),    // 默认展开级别
	))

	// 提供 OpenAPI 规范的直接访问
	router.GET("/api/docs", func(c *gin.Context) {
		response.OK(c, "API Documentation", gin.H{
			"swagger_ui":   "/swagger/index.html",
			"openapi_json": "/swagger/doc.json",
			"version":      "1.0.0",
			"description":  "预测系统 API 文档",
		})
	})

	// 404 处理
	router.NoRoute(func(c *gin.Context) {
		response.NotFound(c, "Endpoint not found")
	})

	return router
}
