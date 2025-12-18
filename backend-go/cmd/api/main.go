package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	httpAdapter "backend-go/internal/adapters/http"
	"backend-go/internal/config"
	"backend-go/internal/container"
	"backend-go/internal/shared/logger"
	"backend-go/internal/shared/monitoring"

	// Swagger imports
	"backend-go/docs"
	_ "backend-go/docs"
)

// @title 预测系统 API
// @version 1.0
// @description 基于 Go + MySQL + Redis 的高性能预测系统后端 API，支持体育比赛预测、用户投票、排行榜和实时更新功能
// @description
// @description ## 功能特性
// @description - 🏆 体育比赛预测系统
// @description - 👥 用户注册和认证
// @description - 🗳️ 预测投票功能
// @description - 📊 实时排行榜
// @description - ⚡ WebSocket 实时通信
// @description - 🚀 高性能缓存策略
// @description
// @description ## 技术栈
// @description - **后端**: Go + Gin + GORM
// @description - **数据库**: MySQL 8.0
// @description - **缓存**: Redis 6.0
// @description - **认证**: JWT
// @description
// @description ## API 版本管理
// @description 当前版本: v1.0，支持向后兼容
// @termsOfService http://swagger.io/terms/

// @contact.name API Support Team
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name MIT License
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1
// @schemes http https

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description 输入 "Bearer " + JWT令牌进行身份验证

// @externalDocs.description OpenAPI 规范
// @externalDocs.url https://swagger.io/resources/open-api/

// setupSwaggerInfo 动态设置 Swagger 信息
func setupSwaggerInfo(cfg *config.Config) {
	// 根据环境动态设置主机和协议
	docs.SwaggerInfo.Host = fmt.Sprintf("localhost:%d", cfg.Server.Port)
	docs.SwaggerInfo.Schemes = []string{"http"}

	// 设置版本信息
	docs.SwaggerInfo.Version = "1.0.0"
	docs.SwaggerInfo.BasePath = "/api/v1"

	// 根据环境设置不同的描述
	if config.GetEnvironment().IsDevelopment() {
		docs.SwaggerInfo.Title = "预测系统 API (开发环境)"
		docs.SwaggerInfo.Description = "开发环境的预测系统 API - 包含调试信息和测试端点"
	} else if config.GetEnvironment().IsProduction() {
		docs.SwaggerInfo.Title = "预测系统 API"
		docs.SwaggerInfo.Description = "生产环境的预测系统 API - 高性能体育比赛预测平台"
	}

	logger.Info("Swagger UI available at: %s://%s/swagger/index.html",
		docs.SwaggerInfo.Schemes[0], docs.SwaggerInfo.Host)
}

func main() {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化日志
	logConfig := &logger.LogConfig{
		Level:      cfg.Log.Level,
		Format:     cfg.Log.Format,
		Output:     cfg.Log.Output,
		MaxSize:    cfg.Log.MaxSize,
		MaxBackups: cfg.Log.MaxBackups,
		MaxAge:     cfg.Log.MaxAge,
		Compress:   cfg.Log.Compress,
		LocalTime:  cfg.Log.LocalTime,
	}
	logger.InitWithConfig(logConfig)
	logger.Info("Starting API server...")

	// 动态配置 Swagger 信息
	setupSwaggerInfo(cfg)

	// 初始化依赖注入容器
	container, err := container.NewContainer(cfg)
	if err != nil {
		logger.Fatal("Failed to initialize container: %v", err)
	}
	defer container.Close()

	// 初始化监控服务
	monitoringService := monitoring.NewMonitoringService(cfg)
	if err := monitoringService.Initialize(container.GetDB(), container.GetRedisClient().GetRedisClient()); err != nil {
		logger.Fatal("Failed to initialize monitoring service: %v", err)
	}

	// 执行启动探针
	if err := monitoringService.StartupProbe(); err != nil {
		logger.Fatal("Startup probe failed: %v", err)
	}

	// 设置路由
	router := httpAdapter.SetupRouter(httpAdapter.RouterConfig{
		UserService:        container.GetUserService(),
		MatchService:       container.GetMatchService(),
		PredictionService:  container.GetPredictionService(),
		LeaderboardService: container.GetLeaderboardService(),
		ScoringService:     container.GetScoringService(),
		TeamService:        container.GetTeamService(),

		// 管理员系统服务
		AdminService:       container.GetAdminService(),
		AdminAuditService:  container.GetAdminAuditService(),
		SportTypeService:   container.GetSportTypeService(),
		ScoringRuleService: container.GetScoringRuleService(),

		// 数据库连接（用于简单的管理功能）
		DB: container.GetDB(),
	})

	// 设置监控中间件和路由
	monitoringService.SetupMiddleware(router)
	monitoringService.SetupRoutes(router)

	// 创建 HTTP 服务器
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.Server.IdleTimeout) * time.Second,
	}

	// 启动服务器
	go func() {
		logger.Info("Server listening on port %d", cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server: %v", err)
		}
	}()

	// 等待中断信号以优雅关闭服务器
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// 优雅关闭服务器，等待现有连接完成
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown: %v", err)
	}

	logger.Info("Server exited")
}
