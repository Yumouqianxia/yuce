package response

import (
	"fmt"
	"runtime"
	"strings"
)

// AppError 应用错误
type AppError struct {
	Type       string      `json:"type"`
	Code       string      `json:"code"`
	Message    string      `json:"message"`
	Details    interface{} `json:"details,omitempty"`
	Stack      string      `json:"stack,omitempty"`
	StatusCode int         `json:"status_code"`
	Cause      error       `json:"-"`
}

// Error 实现 error 接口
func (e *AppError) Error() string {
	if e.Details != nil {
		return fmt.Sprintf("[%s] %s: %s (details: %v)", e.Code, e.Type, e.Message, e.Details)
	}
	return fmt.Sprintf("[%s] %s: %s", e.Code, e.Type, e.Message)
}

// Unwrap 返回原始错误
func (e *AppError) Unwrap() error {
	return e.Cause
}

// WithStack 添加堆栈信息
func (e *AppError) WithStack() *AppError {
	e.Stack = getStackTrace()
	return e
}

// WithCause 添加原因错误
func (e *AppError) WithCause(cause error) *AppError {
	e.Cause = cause
	return e
}

// WithDetails 添加详细信息
func (e *AppError) WithDetails(details interface{}) *AppError {
	e.Details = details
	return e
}

// 错误类型常量
const (
	ErrorTypeValidation     = "validation_error"
	ErrorTypeBusiness       = "business_error"
	ErrorTypeAuthentication = "authentication_error"
	ErrorTypeAuthorization  = "authorization_error"
	ErrorTypeNotFound       = "not_found_error"
	ErrorTypeConflict       = "conflict_error"
	ErrorTypeInternal       = "internal_error"
	ErrorTypeExternal       = "external_error"
	ErrorTypeNetwork        = "network_error"
	ErrorTypeDatabase       = "database_error"
	ErrorTypeCache          = "cache_error"
	ErrorTypeRateLimit      = "rate_limit_error"
	ErrorTypeTimeout        = "timeout_error"
	ErrorTypeUnavailable    = "service_unavailable_error"
)

// 错误代码常量
const (
	// 通用错误代码
	CodeInternalError      = "INTERNAL_ERROR"
	CodeBadRequest         = "BAD_REQUEST"
	CodeUnauthorized       = "UNAUTHORIZED"
	CodeForbidden          = "FORBIDDEN"
	CodeNotFound           = "NOT_FOUND"
	CodeConflict           = "CONFLICT"
	CodeValidationFailed   = "VALIDATION_FAILED"
	CodeServiceUnavailable = "SERVICE_UNAVAILABLE"
	CodeTimeout            = "TIMEOUT"
	CodeRateLimit          = "RATE_LIMIT_EXCEEDED"

	// 业务错误代码
	CodeUserNotFound       = "USER_NOT_FOUND"
	CodeUserExists         = "USER_EXISTS"
	CodeInvalidCredentials = "INVALID_CREDENTIALS"
	CodeAccountLocked      = "ACCOUNT_LOCKED"
	CodeTokenExpired       = "TOKEN_EXPIRED"
	CodeTokenInvalid       = "TOKEN_INVALID"

	CodeMatchNotFound = "MATCH_NOT_FOUND"
	CodeMatchStarted  = "MATCH_STARTED"
	CodeMatchFinished = "MATCH_FINISHED"

	CodePredictionNotFound = "PREDICTION_NOT_FOUND"
	CodePredictionExists   = "PREDICTION_EXISTS"
	CodePredictionLocked   = "PREDICTION_LOCKED"

	CodeVoteExists   = "VOTE_EXISTS"
	CodeVoteNotFound = "VOTE_NOT_FOUND"
	CodeSelfVote     = "SELF_VOTE_NOT_ALLOWED"

	CodeScoringRuleNotFound    = "SCORING_RULE_NOT_FOUND"
	CodeNoActiveScoringRule    = "NO_ACTIVE_SCORING_RULE"
	CodeCannotDeleteActiveRule = "CANNOT_DELETE_ACTIVE_RULE"

	// 数据库错误代码
	CodeDatabaseConnection  = "DATABASE_CONNECTION_ERROR"
	CodeDatabaseQuery       = "DATABASE_QUERY_ERROR"
	CodeDatabaseTransaction = "DATABASE_TRANSACTION_ERROR"

	// 缓存错误代码
	CodeCacheConnection = "CACHE_CONNECTION_ERROR"
	CodeCacheOperation  = "CACHE_OPERATION_ERROR"

	// 外部服务错误代码
	CodeExternalService = "EXTERNAL_SERVICE_ERROR"
	CodeNetworkError    = "NETWORK_ERROR"
)

// NewAppError 创建应用错误
func NewAppError(errorType, code, message string, statusCode int) *AppError {
	return &AppError{
		Type:       errorType,
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
	}
}

// 预定义错误创建函数

// NewBadRequestError 创建400错误
func NewBadRequestError(message string, details interface{}) *AppError {
	return &AppError{
		Type:       ErrorTypeValidation,
		Code:       CodeBadRequest,
		Message:    message,
		Details:    details,
		StatusCode: 400,
	}
}

// NewUnauthorizedError 创建401错误
func NewUnauthorizedError(message string) *AppError {
	return &AppError{
		Type:       ErrorTypeAuthentication,
		Code:       CodeUnauthorized,
		Message:    message,
		StatusCode: 401,
	}
}

// NewForbiddenError 创建403错误
func NewForbiddenError(message string) *AppError {
	return &AppError{
		Type:       ErrorTypeAuthorization,
		Code:       CodeForbidden,
		Message:    message,
		StatusCode: 403,
	}
}

// NewNotFoundError 创建404错误
func NewNotFoundError(message string) *AppError {
	return &AppError{
		Type:       ErrorTypeNotFound,
		Code:       CodeNotFound,
		Message:    message,
		StatusCode: 404,
	}
}

// NewConflictError 创建409错误
func NewConflictError(message string, details interface{}) *AppError {
	return &AppError{
		Type:       ErrorTypeConflict,
		Code:       CodeConflict,
		Message:    message,
		Details:    details,
		StatusCode: 409,
	}
}

// NewValidationError 创建422错误
func NewValidationError(message string, details interface{}) *AppError {
	return &AppError{
		Type:       ErrorTypeValidation,
		Code:       CodeValidationFailed,
		Message:    message,
		Details:    details,
		StatusCode: 422,
	}
}

// NewInternalError 创建500错误
func NewInternalError(message string) *AppError {
	return &AppError{
		Type:       ErrorTypeInternal,
		Code:       CodeInternalError,
		Message:    message,
		StatusCode: 500,
	}
}

// NewServiceUnavailableError 创建503错误
func NewServiceUnavailableError(message string) *AppError {
	return &AppError{
		Type:       ErrorTypeUnavailable,
		Code:       CodeServiceUnavailable,
		Message:    message,
		StatusCode: 503,
	}
}

// 业务错误创建函数

// NewUserNotFoundError 用户不存在错误
func NewUserNotFoundError(userID interface{}) *AppError {
	return &AppError{
		Type:       ErrorTypeBusiness,
		Code:       CodeUserNotFound,
		Message:    "用户不存在",
		Details:    map[string]interface{}{"user_id": userID},
		StatusCode: 404,
	}
}

// NewUserExistsError 用户已存在错误
func NewUserExistsError(identifier string) *AppError {
	return &AppError{
		Type:       ErrorTypeBusiness,
		Code:       CodeUserExists,
		Message:    "用户已存在",
		Details:    map[string]interface{}{"identifier": identifier},
		StatusCode: 409,
	}
}

// NewInvalidCredentialsError 无效凭据错误
func NewInvalidCredentialsError() *AppError {
	return &AppError{
		Type:       ErrorTypeAuthentication,
		Code:       CodeInvalidCredentials,
		Message:    "用户名或密码错误",
		StatusCode: 401,
	}
}

// NewAccountLockedError 账户锁定错误
func NewAccountLockedError(unlockTime interface{}) *AppError {
	return &AppError{
		Type:       ErrorTypeAuthentication,
		Code:       CodeAccountLocked,
		Message:    "账户已被锁定",
		Details:    map[string]interface{}{"unlock_time": unlockTime},
		StatusCode: 423,
	}
}

// NewTokenExpiredError 令牌过期错误
func NewTokenExpiredError() *AppError {
	return &AppError{
		Type:       ErrorTypeAuthentication,
		Code:       CodeTokenExpired,
		Message:    "令牌已过期",
		StatusCode: 401,
	}
}

// NewTokenInvalidError 令牌无效错误
func NewTokenInvalidError() *AppError {
	return &AppError{
		Type:       ErrorTypeAuthentication,
		Code:       CodeTokenInvalid,
		Message:    "令牌无效",
		StatusCode: 401,
	}
}

// NewMatchNotFoundError 比赛不存在错误
func NewMatchNotFoundError(matchID interface{}) *AppError {
	return &AppError{
		Type:       ErrorTypeBusiness,
		Code:       CodeMatchNotFound,
		Message:    "比赛不存在",
		Details:    map[string]interface{}{"match_id": matchID},
		StatusCode: 404,
	}
}

// NewMatchStartedError 比赛已开始错误
func NewMatchStartedError(matchID interface{}) *AppError {
	return &AppError{
		Type:       ErrorTypeBusiness,
		Code:       CodeMatchStarted,
		Message:    "比赛已开始，无法修改预测",
		Details:    map[string]interface{}{"match_id": matchID},
		StatusCode: 400,
	}
}

// NewPredictionNotFoundError 预测不存在错误
func NewPredictionNotFoundError(predictionID interface{}) *AppError {
	return &AppError{
		Type:       ErrorTypeBusiness,
		Code:       CodePredictionNotFound,
		Message:    "预测不存在",
		Details:    map[string]interface{}{"prediction_id": predictionID},
		StatusCode: 404,
	}
}

// NewPredictionExistsError 预测已存在错误
func NewPredictionExistsError(userID, matchID interface{}) *AppError {
	return &AppError{
		Type:       ErrorTypeBusiness,
		Code:       CodePredictionExists,
		Message:    "预测已存在",
		Details:    map[string]interface{}{"user_id": userID, "match_id": matchID},
		StatusCode: 409,
	}
}

// NewVoteExistsError 投票已存在错误
func NewVoteExistsError(userID, predictionID interface{}) *AppError {
	return &AppError{
		Type:       ErrorTypeBusiness,
		Code:       CodeVoteExists,
		Message:    "已经投过票了",
		Details:    map[string]interface{}{"user_id": userID, "prediction_id": predictionID},
		StatusCode: 409,
	}
}

// NewSelfVoteError 自己投票错误
func NewSelfVoteError() *AppError {
	return &AppError{
		Type:       ErrorTypeBusiness,
		Code:       CodeSelfVote,
		Message:    "不能给自己的预测投票",
		StatusCode: 400,
	}
}

// NewDatabaseError 数据库错误
func NewDatabaseError(operation string, cause error) *AppError {
	return &AppError{
		Type:       ErrorTypeDatabase,
		Code:       CodeDatabaseQuery,
		Message:    fmt.Sprintf("数据库操作失败: %s", operation),
		Details:    map[string]interface{}{"operation": operation},
		StatusCode: 500,
		Cause:      cause,
	}
}

// NewCacheError 缓存错误
func NewCacheError(operation string, cause error) *AppError {
	return &AppError{
		Type:       ErrorTypeCache,
		Code:       CodeCacheOperation,
		Message:    fmt.Sprintf("缓存操作失败: %s", operation),
		Details:    map[string]interface{}{"operation": operation},
		StatusCode: 500,
		Cause:      cause,
	}
}

// NewExternalServiceError 外部服务错误
func NewExternalServiceError(service string, cause error) *AppError {
	return &AppError{
		Type:       ErrorTypeExternal,
		Code:       CodeExternalService,
		Message:    fmt.Sprintf("外部服务调用失败: %s", service),
		Details:    map[string]interface{}{"service": service},
		StatusCode: 502,
		Cause:      cause,
	}
}

// NewTimeoutError 超时错误
func NewTimeoutError(operation string) *AppError {
	return &AppError{
		Type:       ErrorTypeTimeout,
		Code:       CodeTimeout,
		Message:    fmt.Sprintf("操作超时: %s", operation),
		Details:    map[string]interface{}{"operation": operation},
		StatusCode: 408,
	}
}

// NewRateLimitError 限流错误
func NewRateLimitError(limit int, window string) *AppError {
	return &AppError{
		Type:       ErrorTypeRateLimit,
		Code:       CodeRateLimit,
		Message:    "请求过于频繁，请稍后再试",
		Details:    map[string]interface{}{"limit": limit, "window": window},
		StatusCode: 429,
	}
}

// 错误包装函数

// WrapError 包装错误
func WrapError(err error, errorType, code, message string, statusCode int) *AppError {
	appErr := &AppError{
		Type:       errorType,
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
		Cause:      err,
	}
	return appErr
}

// WrapDatabaseError 包装数据库错误
func WrapDatabaseError(err error, operation string) *AppError {
	return WrapError(err, ErrorTypeDatabase, CodeDatabaseQuery,
		fmt.Sprintf("数据库操作失败: %s", operation), 500)
}

// WrapCacheError 包装缓存错误
func WrapCacheError(err error, operation string) *AppError {
	return WrapError(err, ErrorTypeCache, CodeCacheOperation,
		fmt.Sprintf("缓存操作失败: %s", operation), 500)
}

// WrapValidationError 包装验证错误
func WrapValidationError(err error, field string) *AppError {
	return WrapError(err, ErrorTypeValidation, CodeValidationFailed,
		fmt.Sprintf("字段验证失败: %s", field), 422)
}

// 错误检查函数

// IsAppError 检查是否为应用错误
func IsAppError(err error) bool {
	_, ok := err.(*AppError)
	return ok
}

// IsErrorType 检查错误类型
func IsErrorType(err error, errorType string) bool {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Type == errorType
	}
	return false
}

// IsErrorCode 检查错误代码
func IsErrorCode(err error, code string) bool {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Code == code
	}
	return false
}

// IsValidationError 检查是否为验证错误
func IsValidationError(err error) bool {
	return IsErrorType(err, ErrorTypeValidation)
}

// IsBusinessError 检查是否为业务错误
func IsBusinessError(err error) bool {
	return IsErrorType(err, ErrorTypeBusiness)
}

// IsAuthenticationError 检查是否为认证错误
func IsAuthenticationError(err error) bool {
	return IsErrorType(err, ErrorTypeAuthentication)
}

// IsAuthorizationError 检查是否为授权错误
func IsAuthorizationError(err error) bool {
	return IsErrorType(err, ErrorTypeAuthorization)
}

// IsNotFoundError 检查是否为未找到错误
func IsNotFoundError(err error) bool {
	return IsErrorType(err, ErrorTypeNotFound)
}

// IsConflictError 检查是否为冲突错误
func IsConflictError(err error) bool {
	return IsErrorType(err, ErrorTypeConflict)
}

// IsInternalError 检查是否为内部错误
func IsInternalError(err error) bool {
	return IsErrorType(err, ErrorTypeInternal)
}

// IsDatabaseError 检查是否为数据库错误
func IsDatabaseError(err error) bool {
	return IsErrorType(err, ErrorTypeDatabase)
}

// IsCacheError 检查是否为缓存错误
func IsCacheError(err error) bool {
	return IsErrorType(err, ErrorTypeCache)
}

// IsExternalError 检查是否为外部服务错误
func IsExternalError(err error) bool {
	return IsErrorType(err, ErrorTypeExternal)
}

// IsTimeoutError 检查是否为超时错误
func IsTimeoutError(err error) bool {
	return IsErrorType(err, ErrorTypeTimeout)
}

// IsRateLimitError 检查是否为限流错误
func IsRateLimitError(err error) bool {
	return IsErrorType(err, ErrorTypeRateLimit)
}

// 辅助函数

// getStackTrace 获取堆栈跟踪
func getStackTrace() string {
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(3, pcs[:])

	var sb strings.Builder
	frames := runtime.CallersFrames(pcs[:n])

	for {
		frame, more := frames.Next()
		sb.WriteString(fmt.Sprintf("%s:%d %s\n", frame.File, frame.Line, frame.Function))
		if !more {
			break
		}
	}

	return sb.String()
}

// ErrorSummary 错误摘要
type ErrorSummary struct {
	Type     string      `json:"type"`
	Code     string      `json:"code"`
	Message  string      `json:"message"`
	Count    int         `json:"count"`
	LastSeen int64       `json:"last_seen"`
	Details  interface{} `json:"details,omitempty"`
}

// GetErrorSummary 获取错误摘要
func GetErrorSummary(err error) *ErrorSummary {
	if appErr, ok := err.(*AppError); ok {
		return &ErrorSummary{
			Type:    appErr.Type,
			Code:    appErr.Code,
			Message: appErr.Message,
			Details: appErr.Details,
		}
	}

	return &ErrorSummary{
		Type:    ErrorTypeInternal,
		Code:    CodeInternalError,
		Message: err.Error(),
	}
}

// 积分规则相关错误变量
var (
	ErrScoringRuleNotFound    = NewNotFoundError("积分规则不存在")
	ErrNoActiveScoringRule    = NewNotFoundError("没有激活的积分规则")
	ErrCannotDeleteActiveRule = NewBadRequestError("不能删除激活的积分规则", nil)
)

// NewScoringRuleNotFoundError 积分规则不存在错误
func NewScoringRuleNotFoundError(ruleID interface{}) *AppError {
	return &AppError{
		Type:       ErrorTypeBusiness,
		Code:       CodeScoringRuleNotFound,
		Message:    "积分规则不存在",
		Details:    map[string]interface{}{"rule_id": ruleID},
		StatusCode: 404,
	}
}

// NewNoActiveScoringRuleError 没有激活积分规则错误
func NewNoActiveScoringRuleError() *AppError {
	return &AppError{
		Type:       ErrorTypeBusiness,
		Code:       CodeNoActiveScoringRule,
		Message:    "没有激活的积分规则",
		StatusCode: 404,
	}
}

// NewCannotDeleteActiveRuleError 不能删除激活规则错误
func NewCannotDeleteActiveRuleError() *AppError {
	return &AppError{
		Type:       ErrorTypeBusiness,
		Code:       CodeCannotDeleteActiveRule,
		Message:    "不能删除激活的积分规则",
		StatusCode: 400,
	}
}
