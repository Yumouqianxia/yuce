package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	// RequestIDHeader 请求ID头部名称
	RequestIDHeader = "X-Request-ID"
	// RequestIDKey 请求ID在上下文中的键名
	RequestIDKey = "request_id"
)

// RequestIDGenerator 请求ID生成器接口
type RequestIDGenerator interface {
	Generate() string
}

// UUIDGenerator UUID生成器
type UUIDGenerator struct{}

// Generate 生成UUID格式的请求ID
func (g *UUIDGenerator) Generate() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		// 如果随机数生成失败，使用时间戳作为后备
		return fmt.Sprintf("req_%d", time.Now().UnixNano())
	}

	// 设置版本号和变体
	bytes[6] = (bytes[6] & 0x0f) | 0x40 // Version 4
	bytes[8] = (bytes[8] & 0x3f) | 0x80 // Variant 10

	return fmt.Sprintf("%x-%x-%x-%x-%x",
		bytes[0:4], bytes[4:6], bytes[6:8], bytes[8:10], bytes[10:16])
}

// ShortIDGenerator 短ID生成器
type ShortIDGenerator struct{}

// Generate 生成短格式的请求ID
func (g *ShortIDGenerator) Generate() string {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		return fmt.Sprintf("req_%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(bytes)
}

// TimestampIDGenerator 时间戳ID生成器
type TimestampIDGenerator struct{}

// Generate 生成基于时间戳的请求ID
func (g *TimestampIDGenerator) Generate() string {
	now := time.Now()
	bytes := make([]byte, 4)
	rand.Read(bytes)
	return fmt.Sprintf("%d_%s", now.UnixNano(), hex.EncodeToString(bytes))
}

// RequestIDMiddleware 请求ID中间件
func RequestIDMiddleware(generator RequestIDGenerator) gin.HandlerFunc {
	if generator == nil {
		generator = &UUIDGenerator{}
	}

	return func(c *gin.Context) {
		// 检查请求头中是否已有请求ID
		requestID := c.GetHeader(RequestIDHeader)

		// 如果没有请求ID，生成一个新的
		if requestID == "" {
			requestID = generator.Generate()
		}

		// 设置请求ID到上下文和响应头
		c.Set(RequestIDKey, requestID)
		c.Header(RequestIDHeader, requestID)

		c.Next()
	}
}

// DefaultRequestIDMiddleware 默认请求ID中间件（使用UUID生成器）
func DefaultRequestIDMiddleware() gin.HandlerFunc {
	return RequestIDMiddleware(&UUIDGenerator{})
}

// ShortRequestIDMiddleware 短请求ID中间件
func ShortRequestIDMiddleware() gin.HandlerFunc {
	return RequestIDMiddleware(&ShortIDGenerator{})
}

// TimestampRequestIDMiddleware 时间戳请求ID中间件
func TimestampRequestIDMiddleware() gin.HandlerFunc {
	return RequestIDMiddleware(&TimestampIDGenerator{})
}

// GetRequestID 从上下文获取请求ID
func GetRequestID(c *gin.Context) string {
	if requestID, exists := c.Get(RequestIDKey); exists {
		if id, ok := requestID.(string); ok {
			return id
		}
	}
	return ""
}

// SetRequestID 设置请求ID到上下文
func SetRequestID(c *gin.Context, requestID string) {
	c.Set(RequestIDKey, requestID)
	c.Header(RequestIDHeader, requestID)
}

// RequestIDConfig 请求ID配置
type RequestIDConfig struct {
	Generator     RequestIDGenerator
	HeaderName    string
	ContextKey    string
	SkipPaths     []string
	ForceGenerate bool // 是否强制生成新的请求ID，即使请求头中已有
}

// DefaultRequestIDConfig 默认请求ID配置
func DefaultRequestIDConfig() *RequestIDConfig {
	return &RequestIDConfig{
		Generator:     &UUIDGenerator{},
		HeaderName:    RequestIDHeader,
		ContextKey:    RequestIDKey,
		SkipPaths:     []string{},
		ForceGenerate: false,
	}
}

// RequestIDMiddlewareWithConfig 带配置的请求ID中间件
func RequestIDMiddlewareWithConfig(config *RequestIDConfig) gin.HandlerFunc {
	if config == nil {
		config = DefaultRequestIDConfig()
	}

	return func(c *gin.Context) {
		// 检查是否跳过此路径
		for _, skipPath := range config.SkipPaths {
			if c.Request.URL.Path == skipPath {
				c.Next()
				return
			}
		}

		var requestID string

		// 如果不强制生成，先检查请求头
		if !config.ForceGenerate {
			requestID = c.GetHeader(config.HeaderName)
		}

		// 如果没有请求ID或强制生成，生成一个新的
		if requestID == "" {
			requestID = config.Generator.Generate()
		}

		// 设置请求ID到上下文和响应头
		c.Set(config.ContextKey, requestID)
		c.Header(config.HeaderName, requestID)

		c.Next()
	}
}

// TraceableRequestIDGenerator 可追踪的请求ID生成器
type TraceableRequestIDGenerator struct {
	prefix  string
	counter int64
}

// NewTraceableRequestIDGenerator 创建可追踪的请求ID生成器
func NewTraceableRequestIDGenerator(prefix string) *TraceableRequestIDGenerator {
	return &TraceableRequestIDGenerator{
		prefix:  prefix,
		counter: 0,
	}
}

// Generate 生成可追踪的请求ID
func (g *TraceableRequestIDGenerator) Generate() string {
	g.counter++
	timestamp := time.Now().Unix()
	bytes := make([]byte, 4)
	rand.Read(bytes)

	return fmt.Sprintf("%s_%d_%d_%s",
		g.prefix, timestamp, g.counter, hex.EncodeToString(bytes))
}

// RequestIDValidator 请求ID验证器
type RequestIDValidator interface {
	Validate(requestID string) bool
}

// UUIDValidator UUID验证器
type UUIDValidator struct{}

// Validate 验证UUID格式的请求ID
func (v *UUIDValidator) Validate(requestID string) bool {
	if len(requestID) != 36 {
		return false
	}

	// 简单的UUID格式检查
	for i, char := range requestID {
		switch i {
		case 8, 13, 18, 23:
			if char != '-' {
				return false
			}
		default:
			if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f') || (char >= 'A' && char <= 'F')) {
				return false
			}
		}
	}

	return true
}

// LengthValidator 长度验证器
type LengthValidator struct {
	MinLength int
	MaxLength int
}

// Validate 验证请求ID长度
func (v *LengthValidator) Validate(requestID string) bool {
	length := len(requestID)
	return length >= v.MinLength && length <= v.MaxLength
}

// ValidatingRequestIDMiddleware 带验证的请求ID中间件
func ValidatingRequestIDMiddleware(generator RequestIDGenerator, validator RequestIDValidator) gin.HandlerFunc {
	if generator == nil {
		generator = &UUIDGenerator{}
	}
	if validator == nil {
		validator = &UUIDValidator{}
	}

	return func(c *gin.Context) {
		requestID := c.GetHeader(RequestIDHeader)

		// 验证现有的请求ID
		if requestID != "" && !validator.Validate(requestID) {
			requestID = "" // 无效的请求ID，重新生成
		}

		// 如果没有有效的请求ID，生成一个新的
		if requestID == "" {
			requestID = generator.Generate()
		}

		c.Set(RequestIDKey, requestID)
		c.Header(RequestIDHeader, requestID)

		c.Next()
	}
}

// RequestIDMetrics 请求ID指标
type RequestIDMetrics struct {
	GeneratedCount int64 `json:"generated_count"`
	ReusedCount    int64 `json:"reused_count"`
	InvalidCount   int64 `json:"invalid_count"`
}

// MetricsRequestIDMiddleware 带指标的请求ID中间件
func MetricsRequestIDMiddleware(generator RequestIDGenerator, metrics *RequestIDMetrics) gin.HandlerFunc {
	if generator == nil {
		generator = &UUIDGenerator{}
	}
	if metrics == nil {
		metrics = &RequestIDMetrics{}
	}

	return func(c *gin.Context) {
		requestID := c.GetHeader(RequestIDHeader)

		if requestID != "" {
			// 验证现有请求ID
			validator := &UUIDValidator{}
			if validator.Validate(requestID) {
				metrics.ReusedCount++
			} else {
				metrics.InvalidCount++
				requestID = generator.Generate()
				metrics.GeneratedCount++
			}
		} else {
			requestID = generator.Generate()
			metrics.GeneratedCount++
		}

		c.Set(RequestIDKey, requestID)
		c.Header(RequestIDHeader, requestID)

		c.Next()
	}
}

// GetRequestIDMetrics 获取请求ID指标
func GetRequestIDMetrics(metrics *RequestIDMetrics) map[string]interface{} {
	if metrics == nil {
		return map[string]interface{}{
			"generated_count": 0,
			"reused_count":    0,
			"invalid_count":   0,
		}
	}

	return map[string]interface{}{
		"generated_count": metrics.GeneratedCount,
		"reused_count":    metrics.ReusedCount,
		"invalid_count":   metrics.InvalidCount,
		"total_requests":  metrics.GeneratedCount + metrics.ReusedCount + metrics.InvalidCount,
	}
}
