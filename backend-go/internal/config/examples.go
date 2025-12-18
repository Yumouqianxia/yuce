package config

import (
	"fmt"
	"log"
	"time"
)

// ExampleBasicUsage 基本使用示例
func ExampleBasicUsage() {
	// 加载默认配置
	config, err := Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	fmt.Printf("Server running on: %s\n", config.Server.GetServerAddr())
	fmt.Printf("Database DSN: %s\n", config.Database.GetDSN())
	fmt.Printf("Redis address: %s\n", config.Redis.GetRedisAddr())
}

// ExampleEnvironmentSpecificConfig 环境特定配置示例
func ExampleEnvironmentSpecificConfig() {
	// 为开发环境加载配置
	config, err := LoadForEnvironment(EnvDevelopment)
	if err != nil {
		log.Fatalf("Failed to load development config: %v", err)
	}

	fmt.Printf("Development mode: %s\n", config.Server.Mode)
	fmt.Printf("Debug features enabled: %v\n", config.Features.EnableSwagger)
}

// ExampleCustomConfigPath 自定义配置路径示例
func ExampleCustomConfigPath() {
	opts := &LoadOptions{
		ConfigPath:   "./custom/path",
		ConfigName:   "myconfig",
		ConfigType:   "yaml",
		EnvPrefix:    "MYAPP",
		SkipValidate: false,
	}

	config, err := Load(opts)
	if err != nil {
		log.Fatalf("Failed to load custom config: %v", err)
	}

	fmt.Printf("Custom config loaded from: %s\n", opts.ConfigPath)
	_ = config
}

// ExampleConfigWatcher 配置监听示例
func ExampleConfigWatcher() {
	loader := NewConfigLoader()

	// 加载初始配置
	config, err := loader.LoadConfig(nil)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 添加配置变更监听器
	loader.AddWatcher(ConfigWatcherFunc(func(oldConfig, newConfig *Config) error {
		fmt.Printf("Config changed: server port %d -> %d\n",
			oldConfig.Server.Port, newConfig.Server.Port)
		return nil
	}))

	// 启动文件监听
	if err := loader.WatchConfig(); err != nil {
		log.Printf("Failed to start config watcher: %v", err)
	}

	// 模拟配置变更
	time.Sleep(1 * time.Second)

	// 停止监听
	if err := loader.StopWatching(); err != nil {
		log.Printf("Failed to stop config watcher: %v", err)
	}

	_ = config
}

// ExampleConfigManager 配置管理器示例
func ExampleConfigManager() {
	// 加载配置
	config, err := Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 创建配置管理器
	manager := NewConfigManager(config)

	// 添加观察者
	manager.AddWatcher(func(newConfig *Config) {
		fmt.Printf("Configuration updated: %s\n", newConfig.Server.GetServerAddr())
	})

	// 更新配置
	newConfig := *config
	newConfig.Server.Port = 9090

	if err := manager.UpdateConfig(&newConfig); err != nil {
		log.Printf("Failed to update config: %v", err)
	}
}

// ExampleConfigValidation 配置验证示例
func ExampleConfigValidation() {
	validator := NewConfigValidator()

	// 创建测试配置
	config := &Config{
		Server: ServerConfig{
			Host:         "localhost",
			Port:         8080,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
			IdleTimeout:  120 * time.Second,
			Mode:         "debug",
		},
		Database: DatabaseConfig{
			Host:            "localhost",
			Port:            3306,
			Username:        "root",
			Password:        "password",
			Database:        "test_db",
			MaxOpenConns:    20,
			MaxIdleConns:    10,
			ConnMaxLifetime: 1 * time.Hour,
		},
		Auth: AuthConfig{
			JWTSecret:           "test-secret-key-with-sufficient-length",
			JWTExpirationHours:  24,
			RefreshTokenExpDays: 30,
			BcryptCost:          12,
		},
	}

	// 验证配置
	if err := validator.Validate(config); err != nil {
		fmt.Printf("Config validation failed: %v\n", err)

		// 格式化验证错误
		errors := FormatValidationErrors(err)
		for _, e := range errors {
			fmt.Printf("Field: %s, Error: %s\n", e.Field, e.Message)
		}
	} else {
		fmt.Println("Configuration is valid")
	}
}

// ExampleFeatureFlags 功能开关示例
func ExampleFeatureFlags() {
	config, err := Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 检查功能开关
	if config.Features.EnableSwagger {
		fmt.Println("Swagger documentation is enabled")
	}

	if config.Features.EnablePprof {
		fmt.Println("Performance profiling is enabled")
	}

	if config.Features.EnableRateLimit {
		fmt.Printf("Rate limiting enabled: %d requests/second\n",
			config.Features.RateLimitConfig.RequestsPerSecond)
	}

	if config.Features.EnableCORS {
		fmt.Printf("CORS enabled for origins: %v\n",
			config.Features.CORSConfig.AllowedOrigins)
	}
}

// ExampleExternalServices 外部服务配置示例
func ExampleExternalServices() {
	config, err := Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 邮件服务配置
	if config.External.Email.Enabled {
		fmt.Printf("Email service enabled: %s:%d\n",
			config.External.Email.Host, config.External.Email.Port)
	}

	// 文件存储配置
	switch config.External.FileStorage.Provider {
	case "local":
		fmt.Printf("Using local file storage: %s\n",
			config.External.FileStorage.LocalPath)
	case "s3":
		fmt.Printf("Using S3 storage: %s/%s\n",
			config.External.FileStorage.S3Config.Region,
			config.External.FileStorage.S3Config.Bucket)
	}

	// 监控配置
	if config.External.Monitoring.Enabled {
		fmt.Printf("Monitoring enabled - Metrics: %s, Tracing: %s\n",
			config.External.Monitoring.MetricsURL,
			config.External.Monitoring.TracingURL)
	}
}

// ExamplePasswordPolicy 密码策略示例
func ExamplePasswordPolicy() {
	config, err := Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	policy := &config.Auth.PasswordPolicy

	// 测试密码
	passwords := []string{
		"weak",
		"Password123",
		"Password123!",
		"verylongpasswordwithoutspecialchars123",
	}

	for _, password := range passwords {
		if err := policy.Validate(password); err != nil {
			fmt.Printf("Password '%s' is invalid: %v\n", password, err)
		} else {
			fmt.Printf("Password '%s' is valid\n", password)
		}
	}
}

// ExampleConfigHealth 配置健康检查示例
func ExampleConfigHealth() {
	config, err := Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 获取配置健康状态
	health := GetConfigHealth(config)

	fmt.Printf("Config health status: %s\n", health["status"])

	if issues, exists := health["issues"]; exists {
		fmt.Printf("Issues found: %v\n", issues)
	}

	// 获取配置摘要
	summary := GetConfigSummary(config)
	fmt.Printf("Config summary: %+v\n", summary)
}

// ExampleDynamicConfigUpdate 动态配置更新示例
func ExampleDynamicConfigUpdate() {
	loader := NewConfigLoader()

	// 加载初始配置
	config, err := loader.LoadConfig(nil)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	fmt.Printf("Initial server port: %d\n", config.Server.Port)

	// 动态更新配置值
	loader.SetConfigValue("server.port", 9090)
	loader.SetConfigValue("features.enable_swagger", true)

	// 合并配置更新
	updates := map[string]interface{}{
		"server.host":           "0.0.0.0",
		"log.level":             "info",
		"features.enable_pprof": false,
	}

	if err := loader.MergeConfig(updates); err != nil {
		log.Printf("Failed to merge config: %v", err)
	} else {
		updatedConfig := loader.GetConfig()
		fmt.Printf("Updated server port: %d\n", updatedConfig.Server.Port)
		fmt.Printf("Updated server host: %s\n", updatedConfig.Server.Host)
	}
}

// ExampleConfigSnapshot 配置快照示例
func ExampleConfigSnapshot() {
	loader := NewConfigLoader()

	// 加载配置
	_, err := loader.LoadConfig(nil)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 创建快照
	snapshot := loader.CreateSnapshot()
	fmt.Printf("Created snapshot at timestamp: %d\n", snapshot.Timestamp)
	fmt.Printf("Config file: %s\n", snapshot.File)

	// 修改配置
	loader.SetConfigValue("server.port", 9090)

	// 从快照恢复
	if err := loader.RestoreFromSnapshot(snapshot); err != nil {
		log.Printf("Failed to restore from snapshot: %v", err)
	} else {
		restoredConfig := loader.GetConfig()
		fmt.Printf("Restored server port: %d\n", restoredConfig.Server.Port)
	}
}
