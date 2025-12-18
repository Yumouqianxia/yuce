package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"backend-go/pkg/response"
)

// ErrorHandler 错误处理中间件配置
type ErrorHandler struct {
	logger        *logrus.Logger
	enableStack   bool
	enableRecover bool
	skipPaths     []string
}

// ErrorHandlerOption 错误处理中间件选项
type ErrorHandlerOption func(*ErrorHandler)

// WithLogger 设置日志器
func WithLogger(logger *logrus.Logger) ErrorHandlerOption {
	return func(h *ErrorHandler) {
		h.logger = logger
	}
}

// WithStackTrace 启用堆栈跟踪
func WithStackTrace(enable bool) ErrorHandlerOption {
	return func(h *ErrorHandler) {
		h.enableStack = enable
	}
}

// WithRecover 启用恢复机制
func WithRecover(enable bool) ErrorHandlerOption {
	return func(h *ErrorHandler) {
		h.enableRecover = enable
	}
}

// WithSkipPaths 设置跳过的路径
func WithSkipPaths(paths []string) ErrorHandlerOption {
	return func(h *ErrorHandler) {
		h.skipPaths = paths
	}
}

// NewErrorHandler 创建错误处理中间件
func NewErrorHandler(opts ...ErrorHandlerOption) *ErrorHandler {
	handler := &ErrorHandler{
		logger:        logrus.New(),
		enableStack:   false,
		enableRecover: true,
		skipPaths:     []string{},
	}

	for _, opt := range opts {
		opt(handler)
	}

	return handler
}

// ErrorHandlerMiddleware 错误处理中间件
func (h *ErrorHandler) ErrorHandlerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查是否跳过此路径
		if h.shouldSkipPath(c.Request.URL.Path) {
			c.Next()
			return
		}

		// 恢复机制
		if h.enableRecover {
			defer func() {
				if err := recover(); err != nil {
					h.handlePanic(c, err)
				}
			}()
		}

		// 处理请求
		c.Next()

		// 处理错误
		if len(c.Errors) > 0 {
			h.handleErrors(c)
		}
	}
}

// GlobalErrorHandler 全局错误处理中间件（简化版）
func GlobalErrorHandler() gin.HandlerFunc {
	return NewErrorHandler(
		WithRecover(true),
		WithStackTrace(false),
	).ErrorHandlerMiddleware()
}

// RecoveryMiddleware 恢复中间件
func RecoveryMiddleware(logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 记录panic日志
				stack := debug.Stack()
				logger.WithFields(logrus.Fields{
					"error":      err,
					"stack":      string(stack),
					"path":       c.Request.URL.Path,
					"method":     c.Request.Method,
					"ip":         c.ClientIP(),
					"user_agent": c.Request.UserAgent(),
				}).Error("Panic recovered")

				// 返回500错误
				if !c.Writer.Written() {
					response.InternalError(c, "服务器内部错误")
				}
				c.Abort()
			}
		}()
		c.Next()
	}
}

// handlePanic 处理panic
func (h *ErrorHandler) handlePanic(c *gin.Context, err interface{}) {
	stack := debug.Stack()

	// 记录panic日志
	h.logger.WithFields(logrus.Fields{
		"error":      err,
		"stack":      string(stack),
		"path":       c.Request.URL.Path,
		"method":     c.Request.Method,
		"ip":         c.ClientIP(),
		"user_agent": c.Request.UserAgent(),
		"request_id": getRequestID(c),
	}).Error("Panic recovered")

	// 创建内部错误
	appErr := response.NewInternalError("服务器内部错误")
	if h.enableStack {
		appErr.WithStack()
	}

	// 返回错误响应
	if !c.Writer.Written() {
		response.Error(c, http.StatusInternalServerError, "Internal server error", appErr.Error())
	}
	c.Abort()
}

// handleErrors 处理错误
func (h *ErrorHandler) handleErrors(c *gin.Context) {
	// 获取最后一个错误
	lastError := c.Errors.Last()
	if lastError == nil {
		return
	}

	err := lastError.Err

	// 记录错误日志
	h.logError(c, err)

	// 如果已经写入响应，则不再处理
	if c.Writer.Written() {
		return
	}

	// 处理不同类型的错误
	switch {
	case response.IsAppError(err):
		h.handleAppError(c, err.(*response.AppError))
	default:
		h.handleGenericError(c, err)
	}
}

// handleAppError 处理应用错误
func (h *ErrorHandler) handleAppError(c *gin.Context, err *response.AppError) {
	// 添加堆栈信息（如果启用）
	if h.enableStack && err.Stack == "" {
		err.WithStack()
	}

	response.Error(c, err.StatusCode, err.Message, err.Error())
}

// handleGenericError 处理通用错误
func (h *ErrorHandler) handleGenericError(c *gin.Context, err error) {
	// 创建内部错误
	appErr := response.NewInternalError(err.Error())
	if h.enableStack {
		appErr.WithStack()
	}

	response.Error(c, http.StatusInternalServerError, "Internal server error", appErr.Error())
}

// logError 记录错误日志
func (h *ErrorHandler) logError(c *gin.Context, err error) {
	fields := logrus.Fields{
		"error":      err.Error(),
		"path":       c.Request.URL.Path,
		"method":     c.Request.Method,
		"ip":         c.ClientIP(),
		"user_agent": c.Request.UserAgent(),
		"request_id": getRequestID(c),
	}

	// 添加应用错误的详细信息
	if appErr, ok := err.(*response.AppError); ok {
		fields["error_type"] = appErr.Type
		fields["error_code"] = appErr.Code
		if appErr.Details != nil {
			fields["error_details"] = appErr.Details
		}
		if appErr.Cause != nil {
			fields["error_cause"] = appErr.Cause.Error()
		}
	}

	// 根据错误类型选择日志级别
	switch {
	case response.IsInternalError(err):
		h.logger.WithFields(fields).Error("Internal server error")
	case response.IsExternalError(err):
		h.logger.WithFields(fields).Warn("External service error")
	case response.IsValidationError(err):
		h.logger.WithFields(fields).Info("Validation error")
	case response.IsAuthenticationError(err):
		h.logger.WithFields(fields).Warn("Authentication error")
	case response.IsAuthorizationError(err):
		h.logger.WithFields(fields).Warn("Authorization error")
	case response.IsNotFoundError(err):
		h.logger.WithFields(fields).Info("Resource not found")
	case response.IsConflictError(err):
		h.logger.WithFields(fields).Info("Resource conflict")
	default:
		h.logger.WithFields(fields).Error("Unhandled error")
	}
}

// shouldSkipPath 检查是否应该跳过路径
func (h *ErrorHandler) shouldSkipPath(path string) bool {
	for _, skipPath := range h.skipPaths {
		if strings.HasPrefix(path, skipPath) {
			return true
		}
	}
	return false
}

// getRequestID 获取请求ID
func getRequestID(c *gin.Context) string {
	if requestID, exists := c.Get("request_id"); exists {
		if id, ok := requestID.(string); ok {
			return id
		}
	}
	return ""
}

// ErrorReporter 错误报告器接口
type ErrorReporter interface {
	Report(c *gin.Context, err error)
}

// LogErrorReporter 日志错误报告器
type LogErrorReporter struct {
	logger *logrus.Logger
}

// NewLogErrorReporter 创建日志错误报告器
func NewLogErrorReporter(logger *logrus.Logger) *LogErrorReporter {
	return &LogErrorReporter{logger: logger}
}

// Report 报告错误
func (r *LogErrorReporter) Report(c *gin.Context, err error) {
	fields := logrus.Fields{
		"error":      err.Error(),
		"path":       c.Request.URL.Path,
		"method":     c.Request.Method,
		"ip":         c.ClientIP(),
		"user_agent": c.Request.UserAgent(),
		"timestamp":  time.Now().Unix(),
	}

	if appErr, ok := err.(*response.AppError); ok {
		fields["error_type"] = appErr.Type
		fields["error_code"] = appErr.Code
		fields["status_code"] = appErr.StatusCode
	}

	r.logger.WithFields(fields).Error("Error reported")
}

// MetricsErrorReporter 指标错误报告器
type MetricsErrorReporter struct {
	// 这里可以集成 Prometheus 或其他指标系统
}

// NewMetricsErrorReporter 创建指标错误报告器
func NewMetricsErrorReporter() *MetricsErrorReporter {
	return &MetricsErrorReporter{}
}

// Report 报告错误到指标系统
func (r *MetricsErrorReporter) Report(c *gin.Context, err error) {
	// TODO: 实现指标收集
	// 例如：增加错误计数器、记录错误类型分布等
}

// ErrorReportingMiddleware 错误报告中间件
func ErrorReportingMiddleware(reporters ...ErrorReporter) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// 如果有错误，报告给所有报告器
		if len(c.Errors) > 0 {
			lastError := c.Errors.Last()
			if lastError != nil {
				for _, reporter := range reporters {
					reporter.Report(c, lastError.Err)
				}
			}
		}
	}
}

// ValidationErrorHandler 验证错误处理器
func ValidationErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// 处理验证错误
		if len(c.Errors) > 0 {
			for _, ginErr := range c.Errors {
				if ginErr.Type == gin.ErrorTypeBind {
					// 处理绑定错误
					validationErr := response.NewValidationError(
						"请求参数验证失败",
						map[string]string{"field": ginErr.Error()},
					)
					response.Error(c, http.StatusUnprocessableEntity, "Validation failed", validationErr.Error())
					c.Abort()
					return
				}
			}
		}
	}
}

// TimeoutErrorHandler 超时错误处理器
func TimeoutErrorHandler(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 设置超时上下文
		ctx, cancel := c.Request.Context(), func() {}
		if timeout > 0 {
			ctx, cancel = c.Request.Context(), cancel
		}
		defer cancel()

		// 替换请求上下文
		c.Request = c.Request.WithContext(ctx)

		// 处理请求
		done := make(chan struct{})
		go func() {
			defer close(done)
			c.Next()
		}()

		select {
		case <-done:
			// 请求正常完成
		case <-ctx.Done():
			// 请求超时
			if !c.Writer.Written() {
				timeoutErr := response.NewTimeoutError("请求处理超时")
				response.Error(c, http.StatusRequestTimeout, "Request timeout", timeoutErr.Error())
			}
			c.Abort()
		}
	}
}

// NotFoundHandler 404处理器
func NotFoundHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		response.NotFound(c, fmt.Sprintf("路径 %s 不存在", c.Request.URL.Path))
	}
}

// MethodNotAllowedHandler 405处理器
func MethodNotAllowedHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		response.Error(c, http.StatusMethodNotAllowed, "Method not allowed", fmt.Sprintf("方法 %s 不被允许", c.Request.Method))
	}
}

// CORSErrorHandler CORS错误处理器
func CORSErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// 检查CORS错误
		if c.Writer.Status() == http.StatusForbidden {
			origin := c.Request.Header.Get("Origin")
			if origin != "" {
				corsErr := response.NewForbiddenError("CORS策略不允许此请求")
				corsErr.WithDetails(map[string]string{"origin": origin})
				response.Error(c, http.StatusForbidden, "CORS policy violation", corsErr.Error())
				c.Abort()
			}
		}
	}
}
