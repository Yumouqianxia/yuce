package middleware

import (
	"bytes"
	"context"
	"io"
	"strconv"
	"time"

	"backend-go/internal/shared/logger"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// LoggingConfig 日志中间件配置
type LoggingConfig struct {
	// SkipPaths 跳过记录的路径
	SkipPaths []string
	// LogRequestBody 是否记录请求体
	LogRequestBody bool
	// LogResponseBody 是否记录响应体
	LogResponseBody bool
	// MaxBodySize 最大记录的请求/响应体大小
	MaxBodySize int64
	// SlowThreshold 慢请求阈值
	SlowThreshold time.Duration
}

// DefaultLoggingConfig 默认日志配置
func DefaultLoggingConfig() *LoggingConfig {
	return &LoggingConfig{
		SkipPaths: []string{
			"/health",
			"/metrics",
			"/favicon.ico",
		},
		LogRequestBody:  false,
		LogResponseBody: false,
		MaxBodySize:     1024 * 1024, // 1MB
		SlowThreshold:   time.Second,
	}
}

// responseWriter 响应写入器包装
type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *responseWriter) Write(b []byte) (int, error) {
	if w.body != nil {
		w.body.Write(b)
	}
	return w.ResponseWriter.Write(b)
}

// LoggingMiddleware 请求日志中间件
func LoggingMiddleware(config *LoggingConfig) gin.HandlerFunc {
	if config == nil {
		config = DefaultLoggingConfig()
	}

	return func(c *gin.Context) {
		// 检查是否跳过此路径
		path := c.Request.URL.Path
		for _, skipPath := range config.SkipPaths {
			if path == skipPath {
				c.Next()
				return
			}
		}

		// 记录开始时间
		startTime := time.Now()

		// 获取请求ID
		requestID := GetRequestID(c)
		if requestID == "" {
			requestID = "unknown"
		}

		// 创建上下文
		ctx := context.WithValue(c.Request.Context(), logger.RequestIDKey, requestID)
		c.Request = c.Request.WithContext(ctx)

		// 读取请求体
		var requestBody []byte
		if config.LogRequestBody && c.Request.Body != nil {
			requestBody, _ = io.ReadAll(io.LimitReader(c.Request.Body, config.MaxBodySize))
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// 包装响应写入器
		var responseBody *bytes.Buffer
		if config.LogResponseBody {
			responseBody = &bytes.Buffer{}
			c.Writer = &responseWriter{
				ResponseWriter: c.Writer,
				body:          responseBody,
			}
		}

		// 记录请求开始
		logger.WithContext(ctx).WithFields(logrus.Fields{
			"method":     c.Request.Method,
			"path":       path,
			"query":      c.Request.URL.RawQuery,
			"ip":         c.ClientIP(),
			"user_agent": c.Request.UserAgent(),
			"referer":    c.Request.Referer(),
		}).Info("Request started")

		// 处理请求
		c.Next()

		// 计算处理时间
		duration := time.Since(startTime)
		statusCode := c.Writer.Status()

		// 构建日志字段
		fields := logrus.Fields{
			"method":      c.Request.Method,
			"path":        path,
			"query":       c.Request.URL.RawQuery,
			"status_code": statusCode,
			"duration":    duration.String(),
			"duration_ms": duration.Milliseconds(),
			"ip":          c.ClientIP(),
			"user_agent":  c.Request.UserAgent(),
			"referer":     c.Request.Referer(),
			"size":        c.Writer.Size(),
		}

		// 添加请求体
		if config.LogRequestBody && len(requestBody) > 0 {
			fields["request_body"] = string(requestBody)
		}

		// 添加响应体
		if config.LogResponseBody && responseBody != nil && responseBody.Len() > 0 {
			fields["response_body"] = responseBody.String()
		}

		// 添加错误信息
		if len(c.Errors) > 0 {
			fields["errors"] = c.Errors.String()
		}

		// 确定日志级别
		var logLevel logrus.Level
		var message string

		switch {
		case statusCode >= 500:
			logLevel = logrus.ErrorLevel
			message = "Request completed with server error"
		case statusCode >= 400:
			logLevel = logrus.WarnLevel
			message = "Request completed with client error"
		case duration > config.SlowThreshold:
			logLevel = logrus.WarnLevel
			message = "Slow request completed"
		default:
			logLevel = logrus.InfoLevel
			message = "Request completed"
		}

		// 记录日志
		entry := logger.WithContext(ctx).WithFields(fields)
		switch logLevel {
		case logrus.ErrorLevel:
			entry.Error(message)
		case logrus.WarnLevel:
			entry.Warn(message)
		default:
			entry.Info(message)
		}

		// 记录性能指标
		if duration > config.SlowThreshold {
			logger.LogPerformance(logger.Performance{
				Operation: c.Request.Method + " " + path,
				Duration:  duration,
				StartTime: startTime,
				EndTime:   time.Now(),
				Success:   statusCode < 400,
				Error:     c.Errors.String(),
			})
		}
	}
}

// ErrorLoggingMiddleware 错误日志中间件
func ErrorLoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// 记录错误
		for _, err := range c.Errors {
			requestID := GetRequestID(c)
			ctx := context.WithValue(c.Request.Context(), logger.RequestIDKey, requestID)

			logger.WithContext(ctx).WithFields(logrus.Fields{
				"method":     c.Request.Method,
				"path":       c.Request.URL.Path,
				"status":     c.Writer.Status(),
				"error_type": err.Type,
			}).WithError(err.Err).Error("Request error occurred")
		}
	}
}

// RecoveryLoggingMiddleware 恢复日志中间件
func RecoveryLoggingMiddleware() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		requestID := GetRequestID(c)
		ctx := context.WithValue(c.Request.Context(), logger.RequestIDKey, requestID)

		logger.WithContext(ctx).WithFields(logrus.Fields{
			"method":   c.Request.Method,
			"path":     c.Request.URL.Path,
			"ip":       c.ClientIP(),
			"panic":    recovered,
		}).Error("Panic recovered")

		c.AbortWithStatus(500)
	})
}

// AuditMiddleware 审计日志中间件
func AuditMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 只对写操作进行审计
		if c.Request.Method == "GET" || c.Request.Method == "HEAD" || c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}

		startTime := time.Now()
		c.Next()

		// 获取用户ID（如果有）
		userID := ""
		if uid, exists := c.Get("user_id"); exists {
			if id, ok := uid.(string); ok {
				userID = id
			} else if id, ok := uid.(uint); ok {
				userID = strconv.FormatUint(uint64(id), 10)
			}
		}

		// 记录审计日志
		logger.LogAudit(logger.AuditLog{
			UserID:    userID,
			Action:    c.Request.Method,
			Resource:  c.Request.URL.Path,
			Result:    getResultFromStatus(c.Writer.Status()),
			IP:        c.ClientIP(),
			UserAgent: c.Request.UserAgent(),
			Details: map[string]interface{}{
				"duration_ms": time.Since(startTime).Milliseconds(),
				"status_code": c.Writer.Status(),
				"query":       c.Request.URL.RawQuery,
			},
		})
	}
}

// getResultFromStatus 根据状态码获取结果
func getResultFromStatus(status int) string {
	switch {
	case status >= 200 && status < 300:
		return "success"
	case status >= 400 && status < 500:
		return "client_error"
	case status >= 500:
		return "server_error"
	default:
		return "unknown"
	}
}

// SecurityLoggingMiddleware 安全日志中间件
func SecurityLoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		status := c.Writer.Status()
		
		// 记录安全相关事件
		switch {
		case status == 401:
			logger.LogSecurity(logger.SecurityLog{
				Event:     "unauthorized_access",
				Severity:  "medium",
				IP:        c.ClientIP(),
				UserAgent: c.Request.UserAgent(),
				Details: map[string]interface{}{
					"path":   c.Request.URL.Path,
					"method": c.Request.Method,
				},
			})
		case status == 403:
			logger.LogSecurity(logger.SecurityLog{
				Event:     "forbidden_access",
				Severity:  "high",
				IP:        c.ClientIP(),
				UserAgent: c.Request.UserAgent(),
				Details: map[string]interface{}{
					"path":   c.Request.URL.Path,
					"method": c.Request.Method,
				},
			})
		case status == 429:
			logger.LogSecurity(logger.SecurityLog{
				Event:     "rate_limit_exceeded",
				Severity:  "medium",
				IP:        c.ClientIP(),
				UserAgent: c.Request.UserAgent(),
				Details: map[string]interface{}{
					"path":   c.Request.URL.Path,
					"method": c.Request.Method,
				},
			})
		}
	}
}

// StructuredLoggingMiddleware 结构化日志中间件
func StructuredLoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 设置结构化日志字段
		requestID := GetRequestID(c)
		
		// 添加到上下文
		ctx := context.WithValue(c.Request.Context(), logger.RequestIDKey, requestID)
		c.Request = c.Request.WithContext(ctx)

		// 添加到响应头
		c.Header("X-Request-ID", requestID)

		c.Next()
	}
}