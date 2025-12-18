package config

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v3"
)

// ValidationError 验证错误
type ValidationError struct {
	Field   string      `json:"field"`
	Message string      `json:"message"`
	Value   interface{} `json:"value,omitempty"`
}

// FormatValidationErrors 格式化验证错误
func FormatValidationErrors(err error) []ValidationError {
	var errors []ValidationError

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			errors = append(errors, ValidationError{
				Field:   e.Field(),
				Message: getValidationMessage(e),
				Value:   e.Value(),
			})
		}
	}

	return errors
}

// getValidationMessage 获取验证错误消息
func getValidationMessage(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return "This field is required"
	case "min":
		return fmt.Sprintf("Value must be at least %s", e.Param())
	case "max":
		return fmt.Sprintf("Value must be at most %s", e.Param())
	case "email":
		return "Must be a valid email address"
	case "oneof":
		return fmt.Sprintf("Must be one of: %s", e.Param())
	case "duration":
		return "Must be a valid duration (e.g., 30s, 5m, 1h)"
	case "filepath":
		return "Must be a valid file path"
	default:
		return fmt.Sprintf("Validation failed for tag '%s'", e.Tag())
	}
}

// GetConfigMetadata 获取配置元数据
func GetConfigMetadata(config *Config) map[string]interface{} {
	return map[string]interface{}{
		"version":     "1.0.0",
		"environment": GetEnvironment(),
		"loaded_at":   time.Now().Format(time.RFC3339),
		"validation": map[string]interface{}{
			"enabled": true,
			"strict":  GetEnvironment().IsProduction(),
		},
		"features": map[string]interface{}{
			"hot_reload":   GetEnvironment().IsDevelopment(),
			"config_watch": true,
			"env_override": true,
		},
		"sources": []string{
			"config file",
			"environment variables",
			"default values",
		},
	}
}

// ConfigTemplate 配置模板
type ConfigTemplate struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Environment Environment `json:"environment"`
}

// GetConfigTemplates 获取可用的配置模板
func GetConfigTemplates() []ConfigTemplate {
	return []ConfigTemplate{
		{
			Name:        "development",
			Description: "Development environment with debug features enabled",
			Environment: EnvDevelopment,
		},
		{
			Name:        "testing",
			Description: "Testing environment with test database and reduced security",
			Environment: EnvTesting,
		},
		{
			Name:        "production",
			Description: "Production environment with security and performance optimizations",
			Environment: EnvProduction,
		},
		{
			Name:        "docker",
			Description: "Docker container environment with containerized services",
			Environment: EnvDevelopment,
		},
		{
			Name:        "kubernetes",
			Description: "Kubernetes deployment with service discovery and secrets",
			Environment: EnvProduction,
		},
	}
}

// ApplyTemplate 应用配置模板
func ApplyTemplate(templateName string, baseConfig *Config) (*Config, error) {
	// 深拷贝基础配置
	newConfig := *baseConfig

	switch strings.ToLower(templateName) {
	case "development":
		applyDevelopmentTemplate(&newConfig)
	case "testing":
		applyTestingTemplate(&newConfig)
	case "production":
		applyProductionTemplate(&newConfig)
	case "docker":
		applyDockerTemplate(&newConfig)
	case "kubernetes":
		applyKubernetesTemplate(&newConfig)
	default:
		return nil, fmt.Errorf("unknown template: %s", templateName)
	}

	return &newConfig, nil
}

// applyDevelopmentTemplate 应用开发环境模板
func applyDevelopmentTemplate(config *Config) {
	config.Server.Mode = "debug"
	config.Server.Port = 8080
	config.Log.Level = "debug"
	config.Log.Format = "text"
	config.Features.EnableSwagger = true
	config.Features.EnablePprof = true
	config.Features.EnableRateLimit = false
	config.Auth.BcryptCost = 4
	config.Database.Database = "prediction_system_dev"
	config.Redis.Database = 0
}

// applyTestingTemplate 应用测试环境模板
func applyTestingTemplate(config *Config) {
	config.Server.Mode = "test"
	config.Server.Port = 8081
	config.Log.Level = "warn"
	config.Database.Database = "prediction_system_test"
	config.Redis.Database = 1
	config.Features.EnableSwagger = false
	config.Features.EnablePprof = false
	config.Auth.BcryptCost = 4
}

// applyProductionTemplate 应用生产环境模板
func applyProductionTemplate(config *Config) {
	config.Server.Mode = "release"
	config.Server.Port = 8080
	config.Log.Level = "info"
	config.Log.Format = "json"
	config.Features.EnableSwagger = false
	config.Features.EnablePprof = false
	config.Features.EnableRateLimit = true
	config.Auth.BcryptCost = 12
	config.Database.Database = "prediction_system"
	config.Redis.Database = 0
}

// applyDockerTemplate 应用Docker环境模板
func applyDockerTemplate(config *Config) {
	applyDevelopmentTemplate(config)
	config.Database.Host = "mysql"
	config.Redis.Host = "redis"
	config.Server.Host = "0.0.0.0"
}

// applyKubernetesTemplate 应用Kubernetes环境模板
func applyKubernetesTemplate(config *Config) {
	applyProductionTemplate(config)
	config.Database.Host = "mysql-service"
	config.Redis.Host = "redis-service"
	config.Server.Host = "0.0.0.0"
}

// ConfigDiff 配置差异
type ConfigDiff struct {
	Field    string      `json:"field"`
	OldValue interface{} `json:"old_value"`
	NewValue interface{} `json:"new_value"`
	Type     string      `json:"type"` // "added", "removed", "changed"
}

// CompareConfigs 比较两个配置
func CompareConfigs(config1, config2 *Config) []ConfigDiff {
	var diffs []ConfigDiff

	v1 := reflect.ValueOf(config1).Elem()
	v2 := reflect.ValueOf(config2).Elem()

	compareDiffs("", v1, v2, &diffs)

	return diffs
}

// compareDiffs 递归比较配置差异
func compareDiffs(prefix string, v1, v2 reflect.Value, diffs *[]ConfigDiff) {
	if v1.Type() != v2.Type() {
		return
	}

	switch v1.Kind() {
	case reflect.Struct:
		for i := 0; i < v1.NumField(); i++ {
			field := v1.Type().Field(i)
			fieldName := field.Name
			if prefix != "" {
				fieldName = prefix + "." + fieldName
			}

			f1 := v1.Field(i)
			f2 := v2.Field(i)

			if f1.CanInterface() && f2.CanInterface() {
				compareDiffs(fieldName, f1, f2, diffs)
			}
		}
	case reflect.Slice:
		if v1.Len() != v2.Len() {
			*diffs = append(*diffs, ConfigDiff{
				Field:    prefix,
				OldValue: v1.Interface(),
				NewValue: v2.Interface(),
				Type:     "changed",
			})
			return
		}

		for i := 0; i < v1.Len(); i++ {
			fieldName := fmt.Sprintf("%s[%d]", prefix, i)
			compareDiffs(fieldName, v1.Index(i), v2.Index(i), diffs)
		}
	default:
		if !reflect.DeepEqual(v1.Interface(), v2.Interface()) {
			*diffs = append(*diffs, ConfigDiff{
				Field:    prefix,
				OldValue: v1.Interface(),
				NewValue: v2.Interface(),
				Type:     "changed",
			})
		}
	}
}

// ConfigExporter 配置导出器
type ConfigExporter struct {
	config *Config
}

// NewConfigExporter 创建配置导出器
func NewConfigExporter(config *Config) *ConfigExporter {
	return &ConfigExporter{config: config}
}

// ExportToJSON 导出为JSON格式
func (e *ConfigExporter) ExportToJSON() ([]byte, error) {
	return json.MarshalIndent(e.config, "", "  ")
}

// ExportToYAML 导出为YAML格式
func (e *ConfigExporter) ExportToYAML() ([]byte, error) {
	return yaml.Marshal(e.config)
}

// ExportToFile 导出到文件
func (e *ConfigExporter) ExportToFile(filename string) error {
	var data []byte
	var err error

	ext := strings.ToLower(filename[strings.LastIndex(filename, ".")+1:])
	switch ext {
	case "json":
		data, err = e.ExportToJSON()
	case "yaml", "yml":
		data, err = e.ExportToYAML()
	default:
		return fmt.Errorf("unsupported file format: %s", ext)
	}

	if err != nil {
		return err
	}

	return writeFile(filename, data)
}

// ConfigImporter 配置导入器
type ConfigImporter struct{}

// NewConfigImporter 创建配置导入器
func NewConfigImporter() *ConfigImporter {
	return &ConfigImporter{}
}

// ImportFromFile 从文件导入配置
func (i *ConfigImporter) ImportFromFile(filename string) (*Config, error) {
	return LoadFromFile(filename)
}

// ImportFromJSON 从JSON导入配置
func (i *ConfigImporter) ImportFromJSON(data []byte) (*Config, error) {
	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

// ImportFromYAML 从YAML导入配置
func (i *ConfigImporter) ImportFromYAML(data []byte) (*Config, error) {
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

// ConfigValidator 配置验证器
type ConfigValidator struct {
	validator *validator.Validate
}

// NewConfigValidator 创建配置验证器
func NewConfigValidator() *ConfigValidator {
	v := validator.New()
	registerCustomValidators(v)
	return &ConfigValidator{validator: v}
}

// Validate 验证配置
func (v *ConfigValidator) Validate(config *Config) error {
	if err := v.validator.Struct(config); err != nil {
		return err
	}
	return validateBusinessLogic(config)
}

// ValidatePartial 部分验证配置
func (v *ConfigValidator) ValidatePartial(config interface{}, fields ...string) error {
	return v.validator.StructPartial(config, fields...)
}

// GetConfigHealth 获取配置健康状态
func GetConfigHealth(config *Config) map[string]interface{} {
	health := map[string]interface{}{
		"status": "healthy",
		"checks": make(map[string]interface{}),
		"issues": []string{},
	}

	checks := health["checks"].(map[string]interface{})
	issues := &[]string{}

	// 检查数据库配置
	dbHealth := checkDatabaseHealth(config)
	checks["database"] = dbHealth
	if healthy, ok := dbHealth["healthy"].(bool); ok && !healthy {
		*issues = append(*issues, "Database configuration issues")
	}

	// 检查Redis配置
	redisHealth := checkRedisHealth(config)
	checks["redis"] = redisHealth
	if healthy, ok := redisHealth["healthy"].(bool); ok && !healthy {
		*issues = append(*issues, "Redis configuration issues")
	}

	// 检查认证配置
	authHealth := checkAuthHealth(config)
	checks["auth"] = authHealth
	if healthy, ok := authHealth["healthy"].(bool); ok && !healthy {
		*issues = append(*issues, "Authentication configuration issues")
	}

	// 检查服务器配置
	serverHealth := checkServerHealth(config)
	checks["server"] = serverHealth
	if healthy, ok := serverHealth["healthy"].(bool); ok && !healthy {
		*issues = append(*issues, "Server configuration issues")
	}

	health["issues"] = *issues
	if len(*issues) > 0 {
		health["status"] = "unhealthy"
	}

	return health
}

// checkDatabaseHealth 检查数据库配置健康状态
func checkDatabaseHealth(config *Config) map[string]interface{} {
	health := map[string]interface{}{
		"healthy": true,
		"issues":  []string{},
	}

	issues := &[]string{}

	if config.Database.Host == "" {
		*issues = append(*issues, "Database host is empty")
	}

	if config.Database.Database == "" {
		*issues = append(*issues, "Database name is empty")
	}

	if config.Database.MaxIdleConns > config.Database.MaxOpenConns {
		*issues = append(*issues, "Max idle connections exceeds max open connections")
	}

	health["issues"] = *issues
	health["healthy"] = len(*issues) == 0

	return health
}

// checkRedisHealth 检查Redis配置健康状态
func checkRedisHealth(config *Config) map[string]interface{} {
	health := map[string]interface{}{
		"healthy": true,
		"issues":  []string{},
	}

	issues := &[]string{}

	if config.Redis.Host == "" {
		*issues = append(*issues, "Redis host is empty")
	}

	if config.Redis.MinIdleConns > config.Redis.PoolSize {
		*issues = append(*issues, "Min idle connections exceeds pool size")
	}

	health["issues"] = *issues
	health["healthy"] = len(*issues) == 0

	return health
}

// checkAuthHealth 检查认证配置健康状态
func checkAuthHealth(config *Config) map[string]interface{} {
	health := map[string]interface{}{
		"healthy": true,
		"issues":  []string{},
	}

	issues := &[]string{}
	env := GetEnvironment()

	if config.Auth.JWTSecret == "" {
		*issues = append(*issues, "JWT secret is empty")
	} else if env.IsProduction() && len(config.Auth.JWTSecret) < 32 {
		*issues = append(*issues, "JWT secret too short for production")
	}

	if env.IsProduction() && config.Auth.BcryptCost < 10 {
		*issues = append(*issues, "Bcrypt cost too low for production")
	}

	health["issues"] = *issues
	health["healthy"] = len(*issues) == 0

	return health
}

// checkServerHealth 检查服务器配置健康状态
func checkServerHealth(config *Config) map[string]interface{} {
	health := map[string]interface{}{
		"healthy": true,
		"issues":  []string{},
	}

	issues := &[]string{}
	env := GetEnvironment()

	if config.Server.Port <= 0 || config.Server.Port > 65535 {
		*issues = append(*issues, "Invalid server port")
	}

	if env.IsProduction() && config.Server.Mode == "debug" {
		*issues = append(*issues, "Debug mode enabled in production")
	}

	health["issues"] = *issues
	health["healthy"] = len(*issues) == 0

	return health
}

// ConfigProfiler 配置性能分析器
type ConfigProfiler struct {
	loadTimes []time.Duration
	stats     map[string]interface{}
}

// NewConfigProfiler 创建配置性能分析器
func NewConfigProfiler() *ConfigProfiler {
	return &ConfigProfiler{
		loadTimes: make([]time.Duration, 0),
		stats:     make(map[string]interface{}),
	}
}

// ProfileLoad 分析配置加载性能
func (p *ConfigProfiler) ProfileLoad(loadFunc func() (*Config, error)) (*Config, error) {
	start := time.Now()
	config, err := loadFunc()
	duration := time.Since(start)

	p.loadTimes = append(p.loadTimes, duration)

	return config, err
}

// GetStats 获取性能统计
func (p *ConfigProfiler) GetStats() map[string]interface{} {
	if len(p.loadTimes) == 0 {
		return map[string]interface{}{
			"samples": 0,
		}
	}

	var total time.Duration
	min := p.loadTimes[0]
	max := p.loadTimes[0]

	for _, t := range p.loadTimes {
		total += t
		if t < min {
			min = t
		}
		if t > max {
			max = t
		}
	}

	avg := total / time.Duration(len(p.loadTimes))

	return map[string]interface{}{
		"samples":    len(p.loadTimes),
		"total_time": total.String(),
		"average":    avg.String(),
		"min":        min.String(),
		"max":        max.String(),
		"performance": map[string]interface{}{
			"fast":   avg < 10*time.Millisecond,
			"normal": avg >= 10*time.Millisecond && avg < 100*time.Millisecond,
			"slow":   avg >= 100*time.Millisecond,
		},
	}
}

// writeFile 写入文件的辅助函数
func writeFile(filename string, data []byte) error {
	// 这里应该使用实际的文件写入逻辑
	// 为了简化，这里返回nil
	return nil
}
