package redis

import (
	"errors"
	"fmt"
)

// 预定义错误
var (
	// 基础错误
	ErrKeyNotFound  = errors.New("key not found")
	ErrKeyExists    = errors.New("key already exists")
	ErrInvalidKey   = errors.New("invalid key format")
	ErrInvalidValue = errors.New("invalid value")
	ErrExpired      = errors.New("key expired")

	// 连接错误
	ErrConnectionFailed  = errors.New("redis connection failed")
	ErrConnectionTimeout = errors.New("redis connection timeout")
	ErrConnectionClosed  = errors.New("redis connection closed")
	ErrPoolExhausted     = errors.New("redis connection pool exhausted")

	// 操作错误
	ErrOperationFailed   = errors.New("redis operation failed")
	ErrOperationTimeout  = errors.New("redis operation timeout")
	ErrTransactionFailed = errors.New("redis transaction failed")
	ErrScriptError       = errors.New("redis script execution error")

	// 锁错误
	ErrLockFailed  = errors.New("failed to acquire lock")
	ErrLockTimeout = errors.New("lock acquisition timeout")
	ErrLockNotHeld = errors.New("lock not held")
	ErrLockExpired = errors.New("lock expired")

	// 序列化错误
	ErrSerializationFailed   = errors.New("serialization failed")
	ErrDeserializationFailed = errors.New("deserialization failed")
	ErrInvalidJSON           = errors.New("invalid JSON format")

	// 配置错误
	ErrInvalidConfig    = errors.New("invalid redis configuration")
	ErrMissingConfig    = errors.New("missing redis configuration")
	ErrConfigValidation = errors.New("redis configuration validation failed")

	// 集群错误
	ErrClusterDown     = errors.New("redis cluster is down")
	ErrClusterFailover = errors.New("redis cluster failover in progress")
	ErrSlotNotFound    = errors.New("redis cluster slot not found")

	// 健康检查错误
	ErrHealthCheckFailed  = errors.New("redis health check failed")
	ErrUnhealthy          = errors.New("redis is unhealthy")
	ErrMetricsUnavailable = errors.New("redis metrics unavailable")
)

// RedisError Redis 错误包装器
type RedisError struct {
	Op      string // 操作名称
	Key     string // 相关键名
	Err     error  // 原始错误
	Code    string // 错误代码
	Message string // 错误消息
}

// Error 实现 error 接口
func (e *RedisError) Error() string {
	if e.Key != "" {
		return fmt.Sprintf("redis %s operation failed for key '%s': %s", e.Op, e.Key, e.Message)
	}
	return fmt.Sprintf("redis %s operation failed: %s", e.Op, e.Message)
}

// Unwrap 返回原始错误
func (e *RedisError) Unwrap() error {
	return e.Err
}

// Is 检查错误类型
func (e *RedisError) Is(target error) bool {
	return errors.Is(e.Err, target)
}

// NewRedisError 创建 Redis 错误
func NewRedisError(op, key string, err error) *RedisError {
	return &RedisError{
		Op:      op,
		Key:     key,
		Err:     err,
		Message: err.Error(),
	}
}

// NewRedisErrorWithCode 创建带错误代码的 Redis 错误
func NewRedisErrorWithCode(op, key, code string, err error) *RedisError {
	return &RedisError{
		Op:      op,
		Key:     key,
		Err:     err,
		Code:    code,
		Message: err.Error(),
	}
}

// NewRedisErrorWithMessage 创建带自定义消息的 Redis 错误
func NewRedisErrorWithMessage(op, key, message string, err error) *RedisError {
	return &RedisError{
		Op:      op,
		Key:     key,
		Err:     err,
		Message: message,
	}
}

// 错误代码常量
const (
	// 连接相关错误代码
	ErrCodeConnectionFailed  = "CONN_FAILED"
	ErrCodeConnectionTimeout = "CONN_TIMEOUT"
	ErrCodePoolExhausted     = "POOL_EXHAUSTED"

	// 操作相关错误代码
	ErrCodeKeyNotFound      = "KEY_NOT_FOUND"
	ErrCodeKeyExists        = "KEY_EXISTS"
	ErrCodeOperationTimeout = "OP_TIMEOUT"
	ErrCodeOperationFailed  = "OP_FAILED"

	// 锁相关错误代码
	ErrCodeLockFailed  = "LOCK_FAILED"
	ErrCodeLockTimeout = "LOCK_TIMEOUT"
	ErrCodeLockNotHeld = "LOCK_NOT_HELD"

	// 数据相关错误代码
	ErrCodeInvalidData           = "INVALID_DATA"
	ErrCodeSerializationFailed   = "SERIALIZATION_FAILED"
	ErrCodeDeserializationFailed = "DESERIALIZATION_FAILED"

	// 配置相关错误代码
	ErrCodeInvalidConfig = "INVALID_CONFIG"
	ErrCodeMissingConfig = "MISSING_CONFIG"

	// 集群相关错误代码
	ErrCodeClusterDown     = "CLUSTER_DOWN"
	ErrCodeClusterFailover = "CLUSTER_FAILOVER"

	// 健康检查相关错误代码
	ErrCodeHealthCheckFailed = "HEALTH_CHECK_FAILED"
	ErrCodeUnhealthy         = "UNHEALTHY"
)

// 错误分类函数

// IsConnectionError 检查是否为连接错误
func IsConnectionError(err error) bool {
	if err == nil {
		return false
	}

	var redisErr *RedisError
	if errors.As(err, &redisErr) {
		switch redisErr.Code {
		case ErrCodeConnectionFailed, ErrCodeConnectionTimeout, ErrCodePoolExhausted:
			return true
		}
	}

	return errors.Is(err, ErrConnectionFailed) ||
		errors.Is(err, ErrConnectionTimeout) ||
		errors.Is(err, ErrPoolExhausted)
}

// IsOperationError 检查是否为操作错误
func IsOperationError(err error) bool {
	if err == nil {
		return false
	}

	var redisErr *RedisError
	if errors.As(err, &redisErr) {
		switch redisErr.Code {
		case ErrCodeOperationFailed, ErrCodeOperationTimeout:
			return true
		}
	}

	return errors.Is(err, ErrOperationFailed) ||
		errors.Is(err, ErrOperationTimeout)
}

// IsKeyError 检查是否为键相关错误
func IsKeyError(err error) bool {
	if err == nil {
		return false
	}

	var redisErr *RedisError
	if errors.As(err, &redisErr) {
		switch redisErr.Code {
		case ErrCodeKeyNotFound, ErrCodeKeyExists:
			return true
		}
	}

	return errors.Is(err, ErrKeyNotFound) ||
		errors.Is(err, ErrKeyExists)
}

// IsLockError 检查是否为锁相关错误
func IsLockError(err error) bool {
	if err == nil {
		return false
	}

	var redisErr *RedisError
	if errors.As(err, &redisErr) {
		switch redisErr.Code {
		case ErrCodeLockFailed, ErrCodeLockTimeout, ErrCodeLockNotHeld:
			return true
		}
	}

	return errors.Is(err, ErrLockFailed) ||
		errors.Is(err, ErrLockTimeout) ||
		errors.Is(err, ErrLockNotHeld)
}

// IsSerializationError 检查是否为序列化错误
func IsSerializationError(err error) bool {
	if err == nil {
		return false
	}

	var redisErr *RedisError
	if errors.As(err, &redisErr) {
		switch redisErr.Code {
		case ErrCodeSerializationFailed, ErrCodeDeserializationFailed:
			return true
		}
	}

	return errors.Is(err, ErrSerializationFailed) ||
		errors.Is(err, ErrDeserializationFailed)
}

// IsConfigError 检查是否为配置错误
func IsConfigError(err error) bool {
	if err == nil {
		return false
	}

	var redisErr *RedisError
	if errors.As(err, &redisErr) {
		switch redisErr.Code {
		case ErrCodeInvalidConfig, ErrCodeMissingConfig:
			return true
		}
	}

	return errors.Is(err, ErrInvalidConfig) ||
		errors.Is(err, ErrMissingConfig)
}

// IsClusterError 检查是否为集群错误
func IsClusterError(err error) bool {
	if err == nil {
		return false
	}

	var redisErr *RedisError
	if errors.As(err, &redisErr) {
		switch redisErr.Code {
		case ErrCodeClusterDown, ErrCodeClusterFailover:
			return true
		}
	}

	return errors.Is(err, ErrClusterDown) ||
		errors.Is(err, ErrClusterFailover)
}

// IsHealthError 检查是否为健康检查错误
func IsHealthError(err error) bool {
	if err == nil {
		return false
	}

	var redisErr *RedisError
	if errors.As(err, &redisErr) {
		switch redisErr.Code {
		case ErrCodeHealthCheckFailed, ErrCodeUnhealthy:
			return true
		}
	}

	return errors.Is(err, ErrHealthCheckFailed) ||
		errors.Is(err, ErrUnhealthy)
}

// IsRetryableError 检查错误是否可重试
func IsRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// 连接错误通常可重试
	if IsConnectionError(err) {
		return true
	}

	// 集群故障转移可重试
	if IsClusterError(err) {
		return true
	}

	// 操作超时可重试
	if errors.Is(err, ErrOperationTimeout) {
		return true
	}

	// 锁获取失败可重试
	if errors.Is(err, ErrLockFailed) {
		return true
	}

	return false
}

// IsFatalError 检查错误是否为致命错误
func IsFatalError(err error) bool {
	if err == nil {
		return false
	}

	// 配置错误是致命的
	if IsConfigError(err) {
		return true
	}

	// 序列化错误通常是致命的
	if IsSerializationError(err) {
		return true
	}

	// 无效键格式是致命的
	if errors.Is(err, ErrInvalidKey) {
		return true
	}

	return false
}

// WrapError 包装错误
func WrapError(op, key string, err error) error {
	if err == nil {
		return nil
	}

	// 如果已经是 RedisError，直接返回
	var redisErr *RedisError
	if errors.As(err, &redisErr) {
		return err
	}

	return NewRedisError(op, key, err)
}

// WrapErrorWithCode 包装错误并添加错误代码
func WrapErrorWithCode(op, key, code string, err error) error {
	if err == nil {
		return nil
	}

	return NewRedisErrorWithCode(op, key, code, err)
}

// ErrorSummary 错误摘要
type ErrorSummary struct {
	Type      string `json:"type"`
	Code      string `json:"code"`
	Message   string `json:"message"`
	Operation string `json:"operation"`
	Key       string `json:"key,omitempty"`
	Retryable bool   `json:"retryable"`
	Fatal     bool   `json:"fatal"`
}

// GetErrorSummary 获取错误摘要
func GetErrorSummary(err error) *ErrorSummary {
	if err == nil {
		return nil
	}

	summary := &ErrorSummary{
		Message:   err.Error(),
		Retryable: IsRetryableError(err),
		Fatal:     IsFatalError(err),
	}

	var redisErr *RedisError
	if errors.As(err, &redisErr) {
		summary.Operation = redisErr.Op
		summary.Key = redisErr.Key
		summary.Code = redisErr.Code
	}

	// 确定错误类型
	switch {
	case IsConnectionError(err):
		summary.Type = "connection"
	case IsOperationError(err):
		summary.Type = "operation"
	case IsKeyError(err):
		summary.Type = "key"
	case IsLockError(err):
		summary.Type = "lock"
	case IsSerializationError(err):
		summary.Type = "serialization"
	case IsConfigError(err):
		summary.Type = "config"
	case IsClusterError(err):
		summary.Type = "cluster"
	case IsHealthError(err):
		summary.Type = "health"
	default:
		summary.Type = "unknown"
	}

	return summary
}
