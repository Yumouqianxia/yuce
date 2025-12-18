package logger

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	log           *logrus.Logger
	defaultFields logrus.Fields
)

// LogConfig 日志配置
type LogConfig struct {
	Level      string `json:"level"`
	Format     string `json:"format"`
	Output     string `json:"output"`
	MaxSize    int    `json:"max_size"`
	MaxBackups int    `json:"max_backups"`
	MaxAge     int    `json:"max_age"`
	Compress   bool   `json:"compress"`
	LocalTime  bool   `json:"local_time"`
}

// ContextKey 上下文键类型
type ContextKey string

const (
	// RequestIDKey 请求ID上下文键
	RequestIDKey ContextKey = "request_id"
	// UserIDKey 用户ID上下文键
	UserIDKey ContextKey = "user_id"
	// TraceIDKey 追踪ID上下文键
	TraceIDKey ContextKey = "trace_id"
)

// CustomFormatter 自定义格式化器
type CustomFormatter struct {
	logrus.JSONFormatter
	ServiceName string
	Version     string
}

// Format 格式化日志条目
func (f *CustomFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	// 添加服务信息
	if f.ServiceName != "" {
		entry.Data["service"] = f.ServiceName
	}
	if f.Version != "" {
		entry.Data["version"] = f.Version
	}

	// 添加调用者信息
	if entry.HasCaller() {
		entry.Data["caller"] = fmt.Sprintf("%s:%d", 
			filepath.Base(entry.Caller.File), entry.Caller.Line)
		entry.Data["function"] = entry.Caller.Function
	}

	// 添加环境信息
	entry.Data["hostname"], _ = os.Hostname()
	entry.Data["pid"] = os.Getpid()

	return f.JSONFormatter.Format(entry)
}

// Init 初始化日志器
func Init(level string) {
	InitWithConfig(&LogConfig{
		Level:      level,
		Format:     "json",
		Output:     "stdout",
		MaxSize:    100,
		MaxBackups: 3,
		MaxAge:     28,
		Compress:   true,
		LocalTime:  true,
	})
}

// InitWithConfig 使用配置初始化日志器
func InitWithConfig(config *LogConfig) {
	log = logrus.New()

	// 设置日志级别
	level, err := logrus.ParseLevel(config.Level)
	if err != nil {
		level = logrus.InfoLevel
	}
	log.SetLevel(level)

	// 设置调用者报告
	log.SetReportCaller(true)

	// 设置格式化器
	if config.Format == "json" {
		log.SetFormatter(&CustomFormatter{
			JSONFormatter: logrus.JSONFormatter{
				TimestampFormat: time.RFC3339,
				PrettyPrint:     false,
			},
			ServiceName: "prediction-system",
			Version:     "1.0.0",
		})
	} else {
		log.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: time.RFC3339,
			DisableColors:   config.Output != "stdout",
		})
	}

	// 设置输出
	var output io.Writer
	switch config.Output {
	case "stdout":
		output = os.Stdout
	case "stderr":
		output = os.Stderr
	case "file":
		// 使用 lumberjack 进行日志轮转
		output = &lumberjack.Logger{
			Filename:   "logs/app.log",
			MaxSize:    config.MaxSize,    // MB
			MaxBackups: config.MaxBackups,
			MaxAge:     config.MaxAge,     // days
			Compress:   config.Compress,
			LocalTime:  config.LocalTime,
		}
	default:
		// 如果是文件路径
		if strings.Contains(config.Output, "/") || strings.Contains(config.Output, "\\") {
			// 确保目录存在
			dir := filepath.Dir(config.Output)
			if err := os.MkdirAll(dir, 0755); err != nil {
				log.Errorf("Failed to create log directory: %v", err)
				output = os.Stdout
			} else {
				output = &lumberjack.Logger{
					Filename:   config.Output,
					MaxSize:    config.MaxSize,
					MaxBackups: config.MaxBackups,
					MaxAge:     config.MaxAge,
					Compress:   config.Compress,
					LocalTime:  config.LocalTime,
				}
			}
		} else {
			output = os.Stdout
		}
	}

	log.SetOutput(output)

	// 设置默认字段
	defaultFields = logrus.Fields{
		"service": "prediction-system",
		"version": "1.0.0",
	}
}

// Debug 调试日志
func Debug(args ...interface{}) {
	if log != nil {
		log.WithFields(defaultFields).Debug(args...)
	}
}

// Debugf 格式化调试日志
func Debugf(format string, args ...interface{}) {
	if log != nil {
		log.WithFields(defaultFields).Debugf(format, args...)
	}
}

// Info 信息日志
func Info(args ...interface{}) {
	if log != nil {
		log.WithFields(defaultFields).Info(args...)
	}
}

// Infof 格式化信息日志
func Infof(format string, args ...interface{}) {
	if log != nil {
		log.WithFields(defaultFields).Infof(format, args...)
	}
}

// Warn 警告日志
func Warn(args ...interface{}) {
	if log != nil {
		log.WithFields(defaultFields).Warn(args...)
	}
}

// Warnf 格式化警告日志
func Warnf(format string, args ...interface{}) {
	if log != nil {
		log.WithFields(defaultFields).Warnf(format, args...)
	}
}

// Error 错误日志
func Error(args ...interface{}) {
	if log != nil {
		log.WithFields(defaultFields).Error(args...)
	}
}

// Errorf 格式化错误日志
func Errorf(format string, args ...interface{}) {
	if log != nil {
		log.WithFields(defaultFields).Errorf(format, args...)
	}
}

// Fatal 致命错误日志
func Fatal(args ...interface{}) {
	if log != nil {
		log.WithFields(defaultFields).Fatal(args...)
	}
}

// Fatalf 格式化致命错误日志
func Fatalf(format string, args ...interface{}) {
	if log != nil {
		log.WithFields(defaultFields).Fatalf(format, args...)
	}
}

// WithField 添加字段
func WithField(key string, value interface{}) *logrus.Entry {
	if log != nil {
		return log.WithFields(defaultFields).WithField(key, value)
	}
	return nil
}

// WithFields 添加多个字段
func WithFields(fields logrus.Fields) *logrus.Entry {
	if log != nil {
		// 合并默认字段和自定义字段
		mergedFields := make(logrus.Fields)
		for k, v := range defaultFields {
			mergedFields[k] = v
		}
		for k, v := range fields {
			mergedFields[k] = v
		}
		return log.WithFields(mergedFields)
	}
	return nil
}

// WithContext 从上下文创建日志条目
func WithContext(ctx context.Context) *logrus.Entry {
	if log == nil {
		return nil
	}

	fields := make(logrus.Fields)
	for k, v := range defaultFields {
		fields[k] = v
	}

	// 从上下文中提取字段
	if requestID := ctx.Value(RequestIDKey); requestID != nil {
		fields["request_id"] = requestID
	}
	if userID := ctx.Value(UserIDKey); userID != nil {
		fields["user_id"] = userID
	}
	if traceID := ctx.Value(TraceIDKey); traceID != nil {
		fields["trace_id"] = traceID
	}

	return log.WithFields(fields)
}

// WithError 添加错误字段
func WithError(err error) *logrus.Entry {
	if log != nil {
		return log.WithFields(defaultFields).WithError(err)
	}
	return nil
}

// GetLogger 获取日志器实例
func GetLogger() *logrus.Logger {
	return log
}

// SetDefaultFields 设置默认字段
func SetDefaultFields(fields logrus.Fields) {
	defaultFields = fields
}

// AddDefaultField 添加默认字段
func AddDefaultField(key string, value interface{}) {
	if defaultFields == nil {
		defaultFields = make(logrus.Fields)
	}
	defaultFields[key] = value
}

// LogEntry 结构化日志条目
type LogEntry struct {
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Timestamp time.Time              `json:"timestamp"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
	Error     string                 `json:"error,omitempty"`
	Stack     string                 `json:"stack,omitempty"`
}

// LogWithStack 记录带堆栈信息的错误
func LogWithStack(level logrus.Level, err error, message string) {
	if log == nil {
		return
	}

	// 获取调用栈
	stack := make([]byte, 4096)
	length := runtime.Stack(stack, false)
	
	entry := log.WithFields(defaultFields).WithError(err)
	entry.Data["stack"] = string(stack[:length])
	
	switch level {
	case logrus.DebugLevel:
		entry.Debug(message)
	case logrus.InfoLevel:
		entry.Info(message)
	case logrus.WarnLevel:
		entry.Warn(message)
	case logrus.ErrorLevel:
		entry.Error(message)
	case logrus.FatalLevel:
		entry.Fatal(message)
	case logrus.PanicLevel:
		entry.Panic(message)
	}
}

// ErrorWithStack 记录带堆栈的错误日志
func ErrorWithStack(err error, message string) {
	LogWithStack(logrus.ErrorLevel, err, message)
}

// FatalWithStack 记录带堆栈的致命错误日志
func FatalWithStack(err error, message string) {
	LogWithStack(logrus.FatalLevel, err, message)
}

// Performance 性能日志
type Performance struct {
	Operation string        `json:"operation"`
	Duration  time.Duration `json:"duration"`
	StartTime time.Time     `json:"start_time"`
	EndTime   time.Time     `json:"end_time"`
	Success   bool          `json:"success"`
	Error     string        `json:"error,omitempty"`
}

// LogPerformance 记录性能日志
func LogPerformance(perf Performance) {
	if log == nil {
		return
	}

	fields := logrus.Fields{
		"operation":  perf.Operation,
		"duration":   perf.Duration.String(),
		"duration_ms": perf.Duration.Milliseconds(),
		"start_time": perf.StartTime.Format(time.RFC3339),
		"end_time":   perf.EndTime.Format(time.RFC3339),
		"success":    perf.Success,
	}

	if perf.Error != "" {
		fields["error"] = perf.Error
	}

	entry := log.WithFields(defaultFields).WithFields(fields)
	
	if perf.Success {
		entry.Info("Operation completed")
	} else {
		entry.Error("Operation failed")
	}
}

// Timer 计时器
type Timer struct {
	operation string
	startTime time.Time
	fields    logrus.Fields
}

// StartTimer 开始计时
func StartTimer(operation string) *Timer {
	return &Timer{
		operation: operation,
		startTime: time.Now(),
		fields:    make(logrus.Fields),
	}
}

// AddField 添加字段到计时器
func (t *Timer) AddField(key string, value interface{}) *Timer {
	t.fields[key] = value
	return t
}

// Stop 停止计时并记录日志
func (t *Timer) Stop() {
	t.StopWithSuccess(true, "")
}

// StopWithError 停止计时并记录错误
func (t *Timer) StopWithError(err error) {
	t.StopWithSuccess(false, err.Error())
}

// StopWithSuccess 停止计时并记录结果
func (t *Timer) StopWithSuccess(success bool, errorMsg string) {
	duration := time.Since(t.startTime)
	
	perf := Performance{
		Operation: t.operation,
		Duration:  duration,
		StartTime: t.startTime,
		EndTime:   time.Now(),
		Success:   success,
		Error:     errorMsg,
	}

	// 添加自定义字段
	if log != nil {
		fields := make(logrus.Fields)
		for k, v := range defaultFields {
			fields[k] = v
		}
		for k, v := range t.fields {
			fields[k] = v
		}
		fields["operation"] = perf.Operation
		fields["duration"] = perf.Duration.String()
		fields["duration_ms"] = perf.Duration.Milliseconds()
		fields["success"] = perf.Success

		entry := log.WithFields(fields)
		
		if success {
			entry.Infof("Operation '%s' completed in %v", t.operation, duration)
		} else {
			entry.WithField("error", errorMsg).Errorf("Operation '%s' failed after %v", t.operation, duration)
		}
	}
}

// Audit 审计日志
type AuditLog struct {
	UserID    string                 `json:"user_id"`
	Action    string                 `json:"action"`
	Resource  string                 `json:"resource"`
	Result    string                 `json:"result"`
	Timestamp time.Time              `json:"timestamp"`
	IP        string                 `json:"ip,omitempty"`
	UserAgent string                 `json:"user_agent,omitempty"`
	Details   map[string]interface{} `json:"details,omitempty"`
}

// LogAudit 记录审计日志
func LogAudit(audit AuditLog) {
	if log == nil {
		return
	}

	audit.Timestamp = time.Now()
	
	fields := logrus.Fields{
		"audit":      true,
		"user_id":    audit.UserID,
		"action":     audit.Action,
		"resource":   audit.Resource,
		"result":     audit.Result,
		"timestamp":  audit.Timestamp.Format(time.RFC3339),
	}

	if audit.IP != "" {
		fields["ip"] = audit.IP
	}
	if audit.UserAgent != "" {
		fields["user_agent"] = audit.UserAgent
	}
	if audit.Details != nil {
		fields["details"] = audit.Details
	}

	log.WithFields(defaultFields).WithFields(fields).Info("Audit log")
}

// Security 安全日志
type SecurityLog struct {
	Event     string                 `json:"event"`
	Severity  string                 `json:"severity"`
	UserID    string                 `json:"user_id,omitempty"`
	IP        string                 `json:"ip,omitempty"`
	UserAgent string                 `json:"user_agent,omitempty"`
	Details   map[string]interface{} `json:"details,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// LogSecurity 记录安全日志
func LogSecurity(security SecurityLog) {
	if log == nil {
		return
	}

	security.Timestamp = time.Now()
	
	fields := logrus.Fields{
		"security":  true,
		"event":     security.Event,
		"severity":  security.Severity,
		"timestamp": security.Timestamp.Format(time.RFC3339),
	}

	if security.UserID != "" {
		fields["user_id"] = security.UserID
	}
	if security.IP != "" {
		fields["ip"] = security.IP
	}
	if security.UserAgent != "" {
		fields["user_agent"] = security.UserAgent
	}
	if security.Details != nil {
		fields["details"] = security.Details
	}

	entry := log.WithFields(defaultFields).WithFields(fields)
	
	switch strings.ToLower(security.Severity) {
	case "low":
		entry.Info("Security event")
	case "medium":
		entry.Warn("Security event")
	case "high", "critical":
		entry.Error("Security event")
	default:
		entry.Info("Security event")
	}
}
