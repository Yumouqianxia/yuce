package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

// Config 应用配置
type Config struct {
	Server    ServerConfig    `mapstructure:"server" validate:"required"`
	Database  DatabaseConfig  `mapstructure:"database" validate:"required"`
	Redis     RedisConfig     `mapstructure:"redis" validate:"required"`
	Auth      AuthConfig      `mapstructure:"auth" validate:"required"`
	Log       LogConfig       `mapstructure:"log" validate:"required"`
	Features  FeatureConfig   `mapstructure:"features" validate:"required"`
	Cache     CacheConfig     `mapstructure:"cache"`
	External  ExternalConfig  `mapstructure:"external"`
}

// Environment 环境类型
type Environment string

const (
	EnvDevelopment Environment = "development"
	EnvTesting     Environment = "testing"
	EnvStaging     Environment = "staging"
	EnvProduction  Environment = "production"
)

// GetEnvironment 获取当前环境
func GetEnvironment() Environment {
	env := strings.ToLower(os.Getenv("GO_ENV"))
	switch env {
	case "dev", "development":
		return EnvDevelopment
	case "test", "testing":
		return EnvTesting
	case "stage", "staging":
		return EnvStaging
	case "prod", "production":
		return EnvProduction
	default:
		return EnvDevelopment
	}
}

// IsDevelopment 是否为开发环境
func (e Environment) IsDevelopment() bool {
	return e == EnvDevelopment
}

// IsProduction 是否为生产环境
func (e Environment) IsProduction() bool {
	return e == EnvProduction
}

// IsTesting 是否为测试环境
func (e Environment) IsTesting() bool {
	return e == EnvTesting
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host         string        `mapstructure:"host" validate:"required"`
	Port         int           `mapstructure:"port" validate:"required,min=1,max=65535"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout" validate:"required,min=1s"`
	WriteTimeout time.Duration `mapstructure:"write_timeout" validate:"required,min=1s"`
	IdleTimeout  time.Duration `mapstructure:"idle_timeout" validate:"required,min=1s"`
	Mode         string        `mapstructure:"mode" validate:"required,oneof=debug release test"`
	TLS          TLSConfig     `mapstructure:"tls"`
}

// TLSConfig TLS 配置
type TLSConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	CertFile string `mapstructure:"cert_file"`
	KeyFile  string `mapstructure:"key_file"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host            string          `mapstructure:"host" validate:"required"`
	Port            int             `mapstructure:"port" validate:"required,min=1,max=65535"`
	Username        string          `mapstructure:"username" validate:"required"`
	Password        string          `mapstructure:"password"`
	Database        string          `mapstructure:"database" validate:"required"`
	Charset         string          `mapstructure:"charset"`
	Collation       string          `mapstructure:"collation"`
	MaxOpenConns    int             `mapstructure:"max_open_conns" validate:"min=1,max=100"`
	MaxIdleConns    int             `mapstructure:"max_idle_conns" validate:"min=1,max=50"`
	ConnMaxLifetime time.Duration   `mapstructure:"conn_max_lifetime" validate:"min=1m"`
	ConnMaxIdleTime time.Duration   `mapstructure:"conn_max_idle_time"`
	SSL             SSLConfig       `mapstructure:"ssl"`
	Migration       MigrationConfig `mapstructure:"migration"`
}

// SSLConfig SSL 配置
type SSLConfig struct {
	Mode     string `mapstructure:"mode" validate:"oneof=disable require verify-ca verify-full"`
	CertFile string `mapstructure:"cert_file"`
	KeyFile  string `mapstructure:"key_file"`
	CAFile   string `mapstructure:"ca_file"`
}

// MigrationConfig 迁移配置
type MigrationConfig struct {
	Enabled    bool   `mapstructure:"enabled"`
	AutoCreate bool   `mapstructure:"auto_create"`
	Path       string `mapstructure:"path"`
}

// RedisConfig Redis 配置
type RedisConfig struct {
	Host          string        `mapstructure:"host" validate:"required"`
	Port          int           `mapstructure:"port" validate:"required,min=1,max=65535"`
	Password      string        `mapstructure:"password"`
	Database      int           `mapstructure:"database" validate:"min=0,max=15"`
	PoolSize      int           `mapstructure:"pool_size" validate:"min=1,max=100"`
	MinIdleConns  int           `mapstructure:"min_idle_conns" validate:"min=0"`
	MaxRetries    int           `mapstructure:"max_retries" validate:"min=0,max=10"`
	DialTimeout   time.Duration `mapstructure:"dial_timeout" validate:"min=1s"`
	ReadTimeout   time.Duration `mapstructure:"read_timeout" validate:"min=1s"`
	WriteTimeout  time.Duration `mapstructure:"write_timeout" validate:"min=1s"`
	PoolTimeout   time.Duration `mapstructure:"pool_timeout" validate:"min=1s"`
	IdleTimeout   time.Duration `mapstructure:"idle_timeout" validate:"min=1m"`
	MaxConnAge    time.Duration `mapstructure:"max_conn_age"`
	IdleCheckFreq time.Duration `mapstructure:"idle_check_freq"`
	Cluster       ClusterConfig `mapstructure:"cluster"`
}

// ClusterConfig Redis 集群配置
type ClusterConfig struct {
	Enabled   bool     `mapstructure:"enabled"`
	Addresses []string `mapstructure:"addresses"`
}

// AuthConfig 认证配置
type AuthConfig struct {
	JWTSecret           string         `mapstructure:"jwt_secret" validate:"required,min=32"`
	JWTExpirationHours  int            `mapstructure:"jwt_expiration_hours" validate:"required,min=1,max=168"`
	RefreshTokenExpDays int            `mapstructure:"refresh_token_exp_days" validate:"required,min=1,max=365"`
	JWTIssuer           string         `mapstructure:"jwt_issuer" validate:"required"`
	BcryptCost          int            `mapstructure:"bcrypt_cost" validate:"required,min=4,max=31"`
	SessionTimeout      time.Duration  `mapstructure:"session_timeout" validate:"min=5m"`
	MaxLoginAttempts    int            `mapstructure:"max_login_attempts" validate:"min=3,max=10"`
	LockoutDuration     time.Duration  `mapstructure:"lockout_duration" validate:"min=5m"`
	PasswordPolicy      PasswordPolicy `mapstructure:"password_policy"`
}

// PasswordPolicy 密码策略
type PasswordPolicy struct {
	MinLength      int  `mapstructure:"min_length" validate:"min=6,max=128"`
	RequireUpper   bool `mapstructure:"require_upper"`
	RequireLower   bool `mapstructure:"require_lower"`
	RequireNumber  bool `mapstructure:"require_number"`
	RequireSpecial bool `mapstructure:"require_special"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level         string        `mapstructure:"level" validate:"required,oneof=debug info warn error fatal panic"`
	Format        string        `mapstructure:"format" validate:"required,oneof=json text"`
	Output        string        `mapstructure:"output" validate:"required"`
	MaxSize       int           `mapstructure:"max_size" validate:"min=1,max=1000"`
	MaxBackups    int           `mapstructure:"max_backups" validate:"min=0,max=100"`
	MaxAge        int           `mapstructure:"max_age" validate:"min=1,max=365"`
	Compress      bool          `mapstructure:"compress"`
	LocalTime     bool          `mapstructure:"local_time"`
	EnableCaller  bool          `mapstructure:"enable_caller"`
	ServiceName   string        `mapstructure:"service_name"`
	Version       string        `mapstructure:"version"`
	SlowThreshold time.Duration `mapstructure:"slow_threshold"`
}



// FeatureConfig 功能开关配置
type FeatureConfig struct {
	EnableSwagger          bool            `mapstructure:"enable_swagger"`
	EnablePprof            bool            `mapstructure:"enable_pprof"`
	EnableMetrics          bool            `mapstructure:"enable_metrics"`
	EnableCORS             bool            `mapstructure:"enable_cors"`
	EnableRateLimit        bool            `mapstructure:"enable_rate_limit"`
	EnableHealthCheck      bool            `mapstructure:"enable_health_check"`
	EnableGracefulShutdown bool            `mapstructure:"enable_graceful_shutdown"`
	CacheLeaderboard       bool            `mapstructure:"cache_leaderboard"`
	CacheMatchData         bool            `mapstructure:"cache_match_data"`
	RateLimitConfig        RateLimitConfig `mapstructure:"rate_limit"`
	CORSConfig             CORSConfig      `mapstructure:"cors"`
}

// RateLimitConfig 限流配置
type RateLimitConfig struct {
	RequestsPerSecond int           `mapstructure:"requests_per_second" validate:"min=1,max=10000"`
	BurstSize         int           `mapstructure:"burst_size" validate:"min=1,max=1000"`
	WindowSize        time.Duration `mapstructure:"window_size" validate:"min=1s,max=1h"`
	CleanupInterval   time.Duration `mapstructure:"cleanup_interval" validate:"min=1m,max=1h"`
}

// CORSConfig CORS 配置
type CORSConfig struct {
	AllowedOrigins   []string      `mapstructure:"allowed_origins"`
	AllowedMethods   []string      `mapstructure:"allowed_methods"`
	AllowedHeaders   []string      `mapstructure:"allowed_headers"`
	ExposedHeaders   []string      `mapstructure:"exposed_headers"`
	AllowCredentials bool          `mapstructure:"allow_credentials"`
	MaxAge           time.Duration `mapstructure:"max_age"`
}

// CacheConfig 缓存配置
type CacheConfig struct {
	Leaderboard LeaderboardCacheConfig `mapstructure:"leaderboard"`
	Monitoring  CacheMonitoringConfig  `mapstructure:"monitoring"`
}

// LeaderboardCacheConfig 排行榜缓存配置
type LeaderboardCacheConfig struct {
	CacheExpiration time.Duration `mapstructure:"cache_expiration" validate:"min=1m,max=1h"`
	RefreshInterval time.Duration `mapstructure:"refresh_interval" validate:"min=30s,max=30m"`
}

// CacheMonitoringConfig 缓存监控配置
type CacheMonitoringConfig struct {
	MonitorInterval  time.Duration `mapstructure:"monitor_interval" validate:"min=30s,max=10m"`
	HitRateThreshold float64       `mapstructure:"hit_rate_threshold" validate:"min=50,max=100"`
}

// ExternalConfig 外部服务配置
type ExternalConfig struct {
	Email       EmailConfig       `mapstructure:"email"`
	FileStorage FileStorageConfig `mapstructure:"file_storage"`
	Monitoring  MonitoringConfig  `mapstructure:"monitoring"`
}

// EmailConfig 邮件配置
type EmailConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	Provider string `mapstructure:"provider" validate:"oneof=smtp sendgrid mailgun"`
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port" validate:"min=1,max=65535"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	From     string `mapstructure:"from" validate:"email"`
	TLS      bool   `mapstructure:"tls"`
}

// FileStorageConfig 文件存储配置
type FileStorageConfig struct {
	Provider   string   `mapstructure:"provider" validate:"oneof=local s3 minio"`
	LocalPath  string   `mapstructure:"local_path"`
	S3Config   S3Config `mapstructure:"s3"`
	MaxSize    int64    `mapstructure:"max_size" validate:"min=1024"`
	AllowedExt []string `mapstructure:"allowed_ext"`
}

// S3Config S3 配置
type S3Config struct {
	Region    string `mapstructure:"region"`
	Bucket    string `mapstructure:"bucket"`
	AccessKey string `mapstructure:"access_key"`
	SecretKey string `mapstructure:"secret_key"`
	Endpoint  string `mapstructure:"endpoint"`
	UseSSL    bool   `mapstructure:"use_ssl"`
}

// MonitoringConfig 监控配置
type MonitoringConfig struct {
	Enabled       bool          `mapstructure:"enabled"`
	MetricsURL    string        `mapstructure:"metrics_url"`
	TracingURL    string        `mapstructure:"tracing_url"`
	SampleRate    float64       `mapstructure:"sample_rate" validate:"min=0,max=1"`
	HealthCheck   HealthConfig  `mapstructure:"health_check"`
	Prometheus    PrometheusConfig `mapstructure:"prometheus"`
}

// HealthConfig 健康检查配置
type HealthConfig struct {
	Enabled         bool          `mapstructure:"enabled"`
	Timeout         time.Duration `mapstructure:"timeout"`
	MaxMemoryMB     uint64        `mapstructure:"max_memory_mb"`
	CheckInterval   time.Duration `mapstructure:"check_interval"`
	StartupRetries  int           `mapstructure:"startup_retries"`
	StartupInterval time.Duration `mapstructure:"startup_interval"`
}

// PrometheusConfig Prometheus配置
type PrometheusConfig struct {
	Enabled       bool     `mapstructure:"enabled"`
	Path          string   `mapstructure:"path"`
	SkipPaths     []string `mapstructure:"skip_paths"`
	NormalizePath bool     `mapstructure:"normalize_path"`
	MaxPathLabels int      `mapstructure:"max_path_labels"`
}

// LoadOptions 配置加载选项
type LoadOptions struct {
	ConfigPath   string
	ConfigName   string
	ConfigType   string
	EnvPrefix    string
	SkipValidate bool
}

// DefaultLoadOptions 默认加载选项
func DefaultLoadOptions() *LoadOptions {
	return &LoadOptions{
		ConfigPath:   "",
		ConfigName:   "config",
		ConfigType:   "yaml",
		EnvPrefix:    "BACKEND",
		SkipValidate: false,
	}
}

// Load 加载配置
func Load(opts ...*LoadOptions) (*Config, error) {
	var options *LoadOptions
	if len(opts) > 0 && opts[0] != nil {
		options = opts[0]
	} else {
		options = DefaultLoadOptions()
	}

	return loadWithOptions(options)
}

// LoadFromFile 从指定文件加载配置
func LoadFromFile(filePath string) (*Config, error) {
	opts := DefaultLoadOptions()
	opts.ConfigPath = filepath.Dir(filePath)
	opts.ConfigName = strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))
	opts.ConfigType = strings.TrimPrefix(filepath.Ext(filePath), ".")

	return loadWithOptions(opts)
}

// LoadForEnvironment 为特定环境加载配置
func LoadForEnvironment(env Environment) (*Config, error) {
	opts := DefaultLoadOptions()
	opts.ConfigName = fmt.Sprintf("config.%s", string(env))

	// 尝试加载环境特定配置，如果不存在则回退到默认配置
	config, err := loadWithOptions(opts)
	if err != nil {
		// 如果环境特定配置不存在，尝试加载默认配置
		opts.ConfigName = "config"
		return loadWithOptions(opts)
	}

	return config, nil
}

// loadWithOptions 使用选项加载配置
func loadWithOptions(opts *LoadOptions) (*Config, error) {
	v := viper.New()

	// 设置配置文件信息
	v.SetConfigName(opts.ConfigName)
	v.SetConfigType(opts.ConfigType)

	// 添加配置文件搜索路径
	if opts.ConfigPath != "" {
		v.AddConfigPath(opts.ConfigPath)
	} else {
		v.AddConfigPath(".")
		v.AddConfigPath("./configs")
		v.AddConfigPath("./config")
		v.AddConfigPath("/etc/backend-go")
		v.AddConfigPath("$HOME/.backend-go")
	}

	// 设置环境变量
	v.SetEnvPrefix(opts.EnvPrefix)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	v.AutomaticEnv()

	// 设置默认值
	setDefaults(v)

	// 读取配置文件
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
		// 配置文件不存在时使用默认值和环境变量
	}

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// 后处理配置
	if err := postProcessConfig(&config); err != nil {
		return nil, fmt.Errorf("failed to post-process config: %w", err)
	}

	// 验证配置
	if !opts.SkipValidate {
		if err := validateConfig(&config); err != nil {
			return nil, fmt.Errorf("config validation failed: %w", err)
		}
	}

	return &config, nil
}

// setDefaults 设置默认配置值
func setDefaults(v *viper.Viper) {
	env := GetEnvironment()

	// 服务器默认配置
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.read_timeout", "30s")
	v.SetDefault("server.write_timeout", "30s")
	v.SetDefault("server.idle_timeout", "120s")
	if env.IsDevelopment() {
		v.SetDefault("server.mode", "debug")
	} else {
		v.SetDefault("server.mode", "release")
	}
	v.SetDefault("server.tls.enabled", false)

	// 数据库默认配置
	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", 3306)
	v.SetDefault("database.username", "root")
	v.SetDefault("database.password", "")
	if env.IsTesting() {
		v.SetDefault("database.database", "prediction_system_test")
	} else if env.IsDevelopment() {
		v.SetDefault("database.database", "prediction_system_dev")
	} else {
		v.SetDefault("database.database", "prediction_system")
	}
	v.SetDefault("database.charset", "utf8mb4")
	v.SetDefault("database.collation", "utf8mb4_unicode_ci")
	v.SetDefault("database.max_open_conns", 20)
	v.SetDefault("database.max_idle_conns", 10)
	v.SetDefault("database.conn_max_lifetime", "1h")
	v.SetDefault("database.conn_max_idle_time", "30m")
	v.SetDefault("database.ssl.mode", "disable")
	v.SetDefault("database.migration.enabled", true)
	v.SetDefault("database.migration.auto_create", env.IsDevelopment())

	// Redis 默认配置
	v.SetDefault("redis.host", "localhost")
	v.SetDefault("redis.port", 6379)
	v.SetDefault("redis.password", "")
	if env.IsTesting() {
		v.SetDefault("redis.database", 1)
	} else {
		v.SetDefault("redis.database", 0)
	}
	v.SetDefault("redis.pool_size", 10)
	v.SetDefault("redis.min_idle_conns", 5)
	v.SetDefault("redis.max_retries", 3)
	v.SetDefault("redis.dial_timeout", "5s")
	v.SetDefault("redis.read_timeout", "3s")
	v.SetDefault("redis.write_timeout", "3s")
	v.SetDefault("redis.pool_timeout", "4s")
	v.SetDefault("redis.idle_timeout", "5m")
	v.SetDefault("redis.max_conn_age", "30m")
	v.SetDefault("redis.idle_check_freq", "1m")
	v.SetDefault("redis.cluster.enabled", false)

	// 认证默认配置
	v.SetDefault("auth.jwt_secret", generateDefaultSecret())
	v.SetDefault("auth.jwt_expiration_hours", 24)
	v.SetDefault("auth.refresh_token_exp_days", 30)
	v.SetDefault("auth.jwt_issuer", "prediction-system")
	if env.IsDevelopment() {
		v.SetDefault("auth.bcrypt_cost", 4) // 开发环境使用较低成本
	} else {
		v.SetDefault("auth.bcrypt_cost", 12)
	}
	v.SetDefault("auth.session_timeout", "24h")
	v.SetDefault("auth.max_login_attempts", 5)
	v.SetDefault("auth.lockout_duration", "15m")
	v.SetDefault("auth.password_policy.min_length", 8)
	v.SetDefault("auth.password_policy.require_upper", true)
	v.SetDefault("auth.password_policy.require_lower", true)
	v.SetDefault("auth.password_policy.require_number", true)
	v.SetDefault("auth.password_policy.require_special", false)

	// 日志默认配置
	if env.IsDevelopment() {
		v.SetDefault("log.level", "debug")
		v.SetDefault("log.format", "text")
	} else {
		v.SetDefault("log.level", "info")
		v.SetDefault("log.format", "json")
	}
	v.SetDefault("log.output", "stdout")
	v.SetDefault("log.max_size", 100)
	v.SetDefault("log.max_backups", 3)
	v.SetDefault("log.max_age", 28)
	v.SetDefault("log.compress", true)
	v.SetDefault("log.local_time", true)
	v.SetDefault("log.enable_caller", true)
	v.SetDefault("log.service_name", "prediction-system")
	v.SetDefault("log.version", "1.0.0")
	v.SetDefault("log.slow_threshold", "1s")



	// 功能开关默认配置
	v.SetDefault("features.enable_swagger", env.IsDevelopment())
	v.SetDefault("features.enable_pprof", env.IsDevelopment())
	v.SetDefault("features.enable_metrics", true)
	v.SetDefault("features.enable_cors", true)
	v.SetDefault("features.enable_rate_limit", !env.IsDevelopment())
	v.SetDefault("features.enable_health_check", true)
	v.SetDefault("features.enable_graceful_shutdown", true)
	v.SetDefault("features.cache_leaderboard", true)
	v.SetDefault("features.cache_match_data", true)

	// 限流配置
	v.SetDefault("features.rate_limit.requests_per_second", 100)
	v.SetDefault("features.rate_limit.burst_size", 200)
	v.SetDefault("features.rate_limit.window_size", "1m")
	v.SetDefault("features.rate_limit.cleanup_interval", "5m")

	// CORS 配置
	v.SetDefault("features.cors.allowed_origins", []string{"*"})
	v.SetDefault("features.cors.allowed_methods", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})
	v.SetDefault("features.cors.allowed_headers", []string{"*"})
	v.SetDefault("features.cors.allow_credentials", true)
	v.SetDefault("features.cors.max_age", "12h")

	// 缓存默认配置
	v.SetDefault("cache.leaderboard.cache_expiration", "5m")
	v.SetDefault("cache.leaderboard.refresh_interval", "2m")
	v.SetDefault("cache.monitoring.monitor_interval", "1m")
	v.SetDefault("cache.monitoring.hit_rate_threshold", 90.0)

	// 外部服务默认配置
	v.SetDefault("external.email.enabled", false)
	v.SetDefault("external.email.provider", "smtp")
	v.SetDefault("external.email.host", "localhost")
	v.SetDefault("external.email.port", 587)
	v.SetDefault("external.email.from", "noreply@example.com")
	v.SetDefault("external.file_storage.provider", "local")
	v.SetDefault("external.file_storage.local_path", "./uploads")
	v.SetDefault("external.file_storage.max_size", 10*1024*1024) // 10MB
	v.SetDefault("external.file_storage.allowed_ext", []string{".jpg", ".jpeg", ".png", ".gif"})
	
	// 监控默认配置
	v.SetDefault("external.monitoring.enabled", !env.IsDevelopment())
	v.SetDefault("external.monitoring.sample_rate", 0.1)
	v.SetDefault("external.monitoring.health_check.enabled", true)
	v.SetDefault("external.monitoring.health_check.timeout", "10s")
	v.SetDefault("external.monitoring.health_check.max_memory_mb", 1024)
	v.SetDefault("external.monitoring.health_check.check_interval", "30s")
	v.SetDefault("external.monitoring.health_check.startup_retries", 10)
	v.SetDefault("external.monitoring.health_check.startup_interval", "5s")
	v.SetDefault("external.monitoring.prometheus.enabled", true)
	v.SetDefault("external.monitoring.prometheus.path", "/metrics")
	v.SetDefault("external.monitoring.prometheus.skip_paths", []string{"/metrics", "/health", "/favicon.ico"})
	v.SetDefault("external.monitoring.prometheus.normalize_path", true)
	v.SetDefault("external.monitoring.prometheus.max_path_labels", 100)
}

// generateDefaultSecret 生成默认密钥（仅用于开发环境）
func generateDefaultSecret() string {
	env := GetEnvironment()
	if env.IsProduction() {
		// 生产环境必须设置环境变量
		return ""
	}
	return "dev-jwt-secret-key-change-this-in-production"
}

// postProcessConfig 配置后处理
func postProcessConfig(config *Config) error {
	env := GetEnvironment()

	// 生产环境安全检查
	if env.IsProduction() {
		if config.Auth.JWTSecret == "" || config.Auth.JWTSecret == "dev-jwt-secret-key-change-this-in-production" {
			return fmt.Errorf("JWT secret must be set in production environment")
		}

		if config.Server.Mode == "debug" {
			config.Server.Mode = "release"
		}

		if config.Features.EnablePprof {
			config.Features.EnablePprof = false
		}

		if config.Features.EnableSwagger {
			config.Features.EnableSwagger = false
		}
	}

	// 设置依赖关系
	if config.Features.EnableRateLimit && config.Features.RateLimitConfig.RequestsPerSecond == 0 {
		config.Features.RateLimitConfig.RequestsPerSecond = 100
	}

	// 验证 TLS 配置
	if config.Server.TLS.Enabled {
		if config.Server.TLS.CertFile == "" || config.Server.TLS.KeyFile == "" {
			return fmt.Errorf("TLS cert_file and key_file are required when TLS is enabled")
		}

		if _, err := os.Stat(config.Server.TLS.CertFile); os.IsNotExist(err) {
			return fmt.Errorf("TLS cert file not found: %s", config.Server.TLS.CertFile)
		}

		if _, err := os.Stat(config.Server.TLS.KeyFile); os.IsNotExist(err) {
			return fmt.Errorf("TLS key file not found: %s", config.Server.TLS.KeyFile)
		}
	}

	return nil
}

// validateConfig 验证配置
func validateConfig(config *Config) error {
	validate := validator.New()

	// 注册自定义验证器
	if err := registerCustomValidators(validate); err != nil {
		return fmt.Errorf("failed to register custom validators: %w", err)
	}

	// 验证配置结构
	if err := validate.Struct(config); err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}

	// 业务逻辑验证
	if err := validateBusinessLogic(config); err != nil {
		return fmt.Errorf("business logic validation failed: %w", err)
	}

	return nil
}

// registerCustomValidators 注册自定义验证器
func registerCustomValidators(validate *validator.Validate) error {
	// 注册时间格式验证器
	if err := validate.RegisterValidation("duration", validateDuration); err != nil {
		return err
	}

	// 注册文件路径验证器
	if err := validate.RegisterValidation("filepath", validateFilePath); err != nil {
		return err
	}

	return nil
}

// validateDuration 验证时间格式
func validateDuration(fl validator.FieldLevel) bool {
	_, err := time.ParseDuration(fl.Field().String())
	return err == nil
}

// validateFilePath 验证文件路径
func validateFilePath(fl validator.FieldLevel) bool {
	path := fl.Field().String()
	if path == "" {
		return true // 空路径由 required 验证器处理
	}

	// 检查路径是否有效
	_, err := filepath.Abs(path)
	return err == nil
}

// validateBusinessLogic 业务逻辑验证
func validateBusinessLogic(config *Config) error {
	env := GetEnvironment()

	// 生产环境特殊验证
	if env.IsProduction() {
		if config.Auth.JWTSecret == "" {
			return fmt.Errorf("JWT secret is required in production")
		}

		if len(config.Auth.JWTSecret) < 32 {
			return fmt.Errorf("JWT secret must be at least 32 characters in production")
		}

		if config.Auth.BcryptCost < 10 {
			return fmt.Errorf("bcrypt cost should be at least 10 in production")
		}
	}

	// 数据库配置验证
	if config.Database.MaxIdleConns > config.Database.MaxOpenConns {
		return fmt.Errorf("max_idle_conns (%d) cannot be greater than max_open_conns (%d)",
			config.Database.MaxIdleConns, config.Database.MaxOpenConns)
	}

	// Redis 配置验证
	if config.Redis.MinIdleConns > config.Redis.PoolSize {
		return fmt.Errorf("redis min_idle_conns (%d) cannot be greater than pool_size (%d)",
			config.Redis.MinIdleConns, config.Redis.PoolSize)
	}

	// WebSocket 配置验证 (暂时注释掉，因为WebSocket配置未定义)
	// if config.WebSocket.PingPeriod >= config.WebSocket.PongWait {
	// 	return fmt.Errorf("websocket ping_period (%v) must be less than pong_wait (%v)",
	// 		config.WebSocket.PingPeriod, config.WebSocket.PongWait)
	// }

	// 限流配置验证
	if config.Features.EnableRateLimit {
		if config.Features.RateLimitConfig.BurstSize < config.Features.RateLimitConfig.RequestsPerSecond {
			return fmt.Errorf("rate limit burst_size (%d) should be >= requests_per_second (%d)",
				config.Features.RateLimitConfig.BurstSize, config.Features.RateLimitConfig.RequestsPerSecond)
		}
	}

	return nil
}

// GetDSN 获取数据库连接字符串
func (c *DatabaseConfig) GetDSN() string {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&collation=%s&parseTime=True&loc=Local",
		c.Username, c.Password, c.Host, c.Port, c.Database, c.Charset, c.Collation)

	// 添加 SSL 配置
	if c.SSL.Mode != "" && c.SSL.Mode != "disable" {
		dsn += "&tls=" + c.SSL.Mode
	}

	return dsn
}

// GetRedisAddr 获取 Redis 地址
func (c *RedisConfig) GetRedisAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// GetServerAddr 获取服务器地址
func (c *ServerConfig) GetServerAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}



// IsSecure 检查是否启用 TLS
func (c *ServerConfig) IsSecure() bool {
	return c.TLS.Enabled
}

// GetScheme 获取协议方案
func (c *ServerConfig) GetScheme() string {
	if c.IsSecure() {
		return "https"
	}
	return "http"
}

// GetBaseURL 获取基础 URL
func (c *ServerConfig) GetBaseURL() string {
	return fmt.Sprintf("%s://%s", c.GetScheme(), c.GetServerAddr())
}

// Validate 验证密码策略
func (p *PasswordPolicy) Validate(password string) error {
	if len(password) < p.MinLength {
		return fmt.Errorf("password must be at least %d characters long", p.MinLength)
	}

	if p.RequireUpper && !containsUpper(password) {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}

	if p.RequireLower && !containsLower(password) {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}

	if p.RequireNumber && !containsNumber(password) {
		return fmt.Errorf("password must contain at least one number")
	}

	if p.RequireSpecial && !containsSpecial(password) {
		return fmt.Errorf("password must contain at least one special character")
	}

	return nil
}

// 辅助函数
func containsUpper(s string) bool {
	for _, r := range s {
		if r >= 'A' && r <= 'Z' {
			return true
		}
	}
	return false
}

func containsLower(s string) bool {
	for _, r := range s {
		if r >= 'a' && r <= 'z' {
			return true
		}
	}
	return false
}

func containsNumber(s string) bool {
	for _, r := range s {
		if r >= '0' && r <= '9' {
			return true
		}
	}
	return false
}

func containsSpecial(s string) bool {
	special := "!@#$%^&*()_+-=[]{}|;:,.<>?"
	for _, r := range s {
		for _, s := range special {
			if r == s {
				return true
			}
		}
	}
	return false
}

// ConfigManager 配置管理器
type ConfigManager struct {
	config   *Config
	watchers []func(*Config)
}

// NewConfigManager 创建配置管理器
func NewConfigManager(config *Config) *ConfigManager {
	return &ConfigManager{
		config:   config,
		watchers: make([]func(*Config), 0),
	}
}

// GetConfig 获取配置
func (cm *ConfigManager) GetConfig() *Config {
	return cm.config
}

// UpdateConfig 更新配置
func (cm *ConfigManager) UpdateConfig(newConfig *Config) error {
	if err := validateConfig(newConfig); err != nil {
		return err
	}

	cm.config = newConfig

	// 通知观察者
	for _, watcher := range cm.watchers {
		watcher(newConfig)
	}

	return nil
}

// AddWatcher 添加配置变更观察者
func (cm *ConfigManager) AddWatcher(watcher func(*Config)) {
	cm.watchers = append(cm.watchers, watcher)
}

// ReloadFromFile 从文件重新加载配置
func (cm *ConfigManager) ReloadFromFile(filePath string) error {
	newConfig, err := LoadFromFile(filePath)
	if err != nil {
		return err
	}

	return cm.UpdateConfig(newConfig)
}

// GetConfigSummary 获取配置摘要（用于日志和调试）
func GetConfigSummary(config *Config) map[string]interface{} {
	return map[string]interface{}{
		"environment":    GetEnvironment(),
		"server_addr":    config.Server.GetServerAddr(),
		"server_mode":    config.Server.Mode,
		"database_host":  config.Database.Host,
		"database_name":  config.Database.Database,
		"redis_addr":     config.Redis.GetRedisAddr(),
		"log_level":      config.Log.Level,
		"features": map[string]bool{
			"swagger":    config.Features.EnableSwagger,
			"pprof":      config.Features.EnablePprof,
			"metrics":    config.Features.EnableMetrics,
			"cors":       config.Features.EnableCORS,
			"rate_limit": config.Features.EnableRateLimit,
		},
	}
}

// GetJWTConfig 获取 JWT 服务配置
func (c *AuthConfig) GetJWTConfig() map[string]interface{} {
	return map[string]interface{}{
		"secret_key":        c.JWTSecret,
		"access_token_ttl":  time.Duration(c.JWTExpirationHours) * time.Hour,
		"refresh_token_ttl": time.Duration(c.RefreshTokenExpDays) * 24 * time.Hour,
		"issuer":            c.JWTIssuer,
	}
}

// GetPasswordConfig 获取密码服务配置
func (c *AuthConfig) GetPasswordConfig() map[string]interface{} {
	return map[string]interface{}{
		"cost": c.BcryptCost,
	}
}
