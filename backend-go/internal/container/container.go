package container

import (
	"fmt"
	"time"

	"backend-go/internal/adapters/persistence/mysql"
	"backend-go/internal/adapters/services"
	"backend-go/internal/config"
	"backend-go/internal/core/domain/leaderboard"
	"backend-go/internal/core/domain/match"
	"backend-go/internal/core/domain/prediction"
	"backend-go/internal/core/domain/scoring"
	"backend-go/internal/core/domain/shared"
	"backend-go/internal/core/domain/user"
	"backend-go/internal/core/ports"
	coreServices "backend-go/internal/core/services"
	"backend-go/internal/shared/jwt"
	"backend-go/internal/shared/logger"
	"backend-go/internal/shared/password"
	"backend-go/pkg/database"
	"backend-go/pkg/redis"

	"gorm.io/gorm"
)

// Container 依赖注入容器
type Container struct {
	config             *config.Config
	db                 *gorm.DB
	redisClient        *redis.Client
	userRepo           user.Repository
	userService        user.Service
	matchRepo          match.Repository
	matchService       match.Service
	predictionRepo     prediction.Repository
	voteRepo           prediction.VoteRepository
	scoringRuleRepo    prediction.ScoringRuleRepository
	predictionService  prediction.Service
	leaderboardRepo    leaderboard.Repository
	leaderboardCache   leaderboard.CacheService
	leaderboardService leaderboard.Service
	scoringRepo        scoring.Repository
	scoringCalculator  scoring.Calculator
	scoringService     scoring.Service
	teamRepo           ports.TeamRepository
	teamService        ports.TeamService
	jwtService         jwt.JWTService
	passwordService    password.Service

	// 管理员系统
	adminService         ports.AdminService
	adminAuditService    ports.AdminAuditService
	sportTypeRepo        ports.SportTypeRepository
	sportTypeService     *coreServices.SportTypeService
	sportScoringRuleRepo ports.ScoringRuleRepository
	scoringRuleService   *coreServices.ScoringRuleService
}

// NewContainer 创建容器
func NewContainer(cfg *config.Config) (*Container, error) {
	container := &Container{
		config: cfg,
	}

	// 初始化数据库连接
	if err := container.initDatabase(); err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// 初始化Redis连接
	if err := container.initRedis(); err != nil {
		return nil, fmt.Errorf("failed to initialize redis: %w", err)
	}

	// 初始化服务
	if err := container.initServices(); err != nil {
		return nil, fmt.Errorf("failed to initialize services: %w", err)
	}

	return container, nil
}

// initDatabase 初始化数据库连接
func (c *Container) initDatabase() error {
	dbConfig := database.Config{
		Host:            c.config.Database.Host,
		Port:            c.config.Database.Port,
		User:            c.config.Database.Username,
		Password:        c.config.Database.Password,
		DBName:          c.config.Database.Database,
		Charset:         c.config.Database.Charset,
		MaxIdleConns:    c.config.Database.MaxIdleConns,
		MaxOpenConns:    c.config.Database.MaxOpenConns,
		ConnMaxLifetime: c.config.Database.ConnMaxLifetime,
	}

	db, err := database.NewConnection(dbConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	c.db = db
	logger.Info("Database connection established")
	return nil
}

// initRedis 初始化Redis连接
func (c *Container) initRedis() error {
	redisConfig := &config.RedisConfig{
		Host:         c.config.Redis.Host,
		Port:         c.config.Redis.Port,
		Password:     c.config.Redis.Password,
		Database:     c.config.Redis.Database,
		PoolSize:     c.config.Redis.PoolSize,
		MinIdleConns: c.config.Redis.MinIdleConns,
		MaxRetries:   c.config.Redis.MaxRetries,
		DialTimeout:  c.config.Redis.DialTimeout,
		ReadTimeout:  c.config.Redis.ReadTimeout,
		WriteTimeout: c.config.Redis.WriteTimeout,
		PoolTimeout:  c.config.Redis.PoolTimeout,
		IdleTimeout:  c.config.Redis.IdleTimeout,
		MaxConnAge:   c.config.Redis.MaxConnAge,
	}

	client, err := redis.NewClient(redisConfig, logger.GetLogger())
	if err != nil {
		return fmt.Errorf("failed to connect to redis: %w", err)
	}

	c.redisClient = client
	logger.Info("Redis connection established")
	return nil
}

// initServices 初始化服务
func (c *Container) initServices() error {
	// 初始化密码服务
	c.passwordService = password.NewService(password.Config{
		Cost: c.config.Auth.BcryptCost,
	})

	// 初始化 JWT 服务
	c.jwtService = jwt.NewJWTService(jwt.Config{
		SecretKey:       c.config.Auth.JWTSecret,
		AccessTokenTTL:  time.Duration(c.config.Auth.JWTExpirationHours) * time.Hour,
		RefreshTokenTTL: time.Duration(c.config.Auth.RefreshTokenExpDays) * 24 * time.Hour,
		Issuer:          c.config.Auth.JWTIssuer,
	})

	// 初始化仓储
	c.userRepo = mysql.NewUserRepository(c.db, c.passwordService)
	c.matchRepo = mysql.NewMatchRepository(c.db)
	c.predictionRepo = mysql.NewPredictionRepository(c.db)
	c.voteRepo = mysql.NewVoteRepository(c.db)
	// TODO: Implement prediction.ScoringRuleRepository
	// c.scoringRuleRepo = mysql.NewPredictionScoringRuleRepository(c.db)
	c.leaderboardRepo = mysql.NewLeaderboardRepository(c.db)
	c.scoringRepo = mysql.NewScoringRepository(c.db)
	c.teamRepo = mysql.NewTeamRepository(c.db)

	// 管理员系统仓储
	c.sportTypeRepo = mysql.NewSportTypeRepository(c.db)
	c.sportScoringRuleRepo = mysql.NewSportScoringRuleRepository(c.db)

	// 初始化缓存服务
	cacheService := redis.NewCacheService(c.redisClient)
	// 用于排行榜领域的缓存（适配器层实现）
	c.leaderboardCache = services.NewLeaderboardCacheService(cacheService)
	// 用于用户服务的排行榜缓存（核心服务实现）
	userLeaderboardCache := coreServices.NewLeaderboardCacheService(
		c.userRepo,
		cacheService,
		coreServices.LeaderboardCacheConfig{
			CacheExpiration: c.config.Cache.Leaderboard.CacheExpiration,
			RefreshInterval: c.config.Cache.Leaderboard.RefreshInterval,
		},
	)

	// 初始化积分计算器
	c.scoringCalculator = services.NewScoringCalculator()

	// 初始化服务
	c.userService = coreServices.NewUserService(
		c.userRepo,
		c.jwtService,
		c.passwordService,
		userLeaderboardCache,
		coreServices.Config{
			MaxLoginAttempts: c.config.Auth.MaxLoginAttempts,
			LockoutDuration:  c.config.Auth.LockoutDuration,
		},
	)
	// Match service requires cache and event bus; pass nils if not available
	var matchCache *coreServices.MatchCacheService
	var eventBus shared.EventBus
	c.matchService = coreServices.NewMatchService(c.matchRepo, matchCache, eventBus, logger.GetLogger())
	c.predictionService = coreServices.NewPredictionService(
		c.predictionRepo,
		c.voteRepo,
		c.matchRepo,
		c.userRepo,
		c.scoringRuleRepo,
		eventBus,
	)
	c.leaderboardService = services.NewLeaderboardService(
		c.leaderboardRepo,
		c.leaderboardCache,
		logger.GetLogger(),
	)
	c.scoringService = services.NewScoringService(
		c.predictionRepo,
		c.scoringRuleRepo,
		c.userRepo,
		c.matchRepo,
		c.scoringRepo,
		c.scoringCalculator,
		logger.GetLogger(),
	)
	c.teamService = coreServices.NewTeamService(c.teamRepo)

	// 初始化管理员系统服务
	dbWrapper := &database.DB{DB: c.db}
	c.adminService = coreServices.NewAdminService(dbWrapper)
	c.adminAuditService = coreServices.NewAdminAuditService(dbWrapper)
	c.sportTypeService = coreServices.NewSportTypeService(c.sportTypeRepo, logger.GetLogger())
	c.scoringRuleService = coreServices.NewScoringRuleService(
		c.sportScoringRuleRepo,
		coreServices.NewDefaultScoreCalculator(logger.GetLogger()),
		logger.GetLogger(),
	)

	logger.Info("Services initialized successfully")
	return nil
}

// GetUserService 获取用户服务
func (c *Container) GetUserService() user.Service {
	return c.userService
}

// GetMatchService 获取比赛服务
func (c *Container) GetMatchService() match.Service {
	return c.matchService
}

// GetPredictionService 获取预测服务
func (c *Container) GetPredictionService() prediction.Service {
	return c.predictionService
}

// GetLeaderboardService 获取排行榜服务
func (c *Container) GetLeaderboardService() leaderboard.Service {
	return c.leaderboardService
}

// GetTeamService 获取战队服务
func (c *Container) GetTeamService() ports.TeamService {
	return c.teamService
}

// GetScoringService 获取积分计算服务
func (c *Container) GetScoringService() scoring.Service {
	return c.scoringService
}

// GetAdminService 获取管理员服务
func (c *Container) GetAdminService() ports.AdminService {
	return c.adminService
}

// GetAdminAuditService 获取管理员审计服务
func (c *Container) GetAdminAuditService() ports.AdminAuditService {
	return c.adminAuditService
}

// GetSportTypeService 获取运动类型服务
func (c *Container) GetSportTypeService() *coreServices.SportTypeService {
	return c.sportTypeService
}

// GetScoringRuleService 获取积分规则服务
func (c *Container) GetScoringRuleService() *coreServices.ScoringRuleService {
	return c.scoringRuleService
}

// GetDB 获取数据库连接
func (c *Container) GetDB() *gorm.DB {
	return c.db
}

// GetRedisClient 获取Redis客户端
func (c *Container) GetRedisClient() *redis.Client {
	return c.redisClient
}

// Close 关闭容器资源
func (c *Container) Close() error {
	var err error

	if c.redisClient != nil {
		if closeErr := c.redisClient.Close(); closeErr != nil {
			err = closeErr
		}
	}

	if c.db != nil {
		sqlDB, dbErr := c.db.DB()
		if dbErr != nil {
			return dbErr
		}
		if closeErr := sqlDB.Close(); closeErr != nil {
			err = closeErr
		}
	}

	return err
}
