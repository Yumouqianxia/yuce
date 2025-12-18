package domain

import "errors"

// 领域错误定义
var (
	// 用户相关错误
	ErrUserNotFound       = errors.New("user not found")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidPassword    = errors.New("invalid password")
	ErrPasswordTooWeak    = errors.New("password too weak")
	ErrUserNotActive      = errors.New("user not active")
	ErrInsufficientPoints = errors.New("insufficient points")

	// 比赛相关错误
	ErrMatchNotFound         = errors.New("match not found")
	ErrMatchAlreadyStarted   = errors.New("match already started")
	ErrMatchNotActive        = errors.New("match not active")
	ErrMatchAlreadyCompleted = errors.New("match already completed")
	ErrMatchAlreadyFinished  = errors.New("match already finished")
	ErrInvalidWinner         = errors.New("invalid winner")
	ErrInvalidScore          = errors.New("invalid score")
	ErrInvalidMatchStatus    = errors.New("invalid match status")
	ErrInvalidStartTime      = errors.New("invalid start time")
	ErrInvalidTournament     = errors.New("invalid tournament")

	// 预测相关错误
	ErrPredictionNotFound         = errors.New("prediction not found")
	ErrPredictionAlreadyExists    = errors.New("prediction already exists")
	ErrCannotModifyPrediction     = errors.New("cannot modify prediction")
	ErrPredictionAlreadyProcessed = errors.New("prediction already processed")
	ErrTooManyModifications       = errors.New("too many modifications")
	ErrModificationNotAllowed     = errors.New("modification not allowed")

	// 投票相关错误
	ErrVoteNotFound            = errors.New("vote not found")
	ErrVoteAlreadyExists       = errors.New("vote already exists")
	ErrCannotVoteOwnPrediction = errors.New("cannot vote for own prediction")
	ErrVoteNotAllowed          = errors.New("vote not allowed")
	ErrDailyVoteLimitExceeded  = errors.New("daily vote limit exceeded")

	// 认证相关错误
	ErrUnauthorized  = errors.New("unauthorized")
	ErrForbidden     = errors.New("forbidden")
	ErrTokenExpired  = errors.New("token expired")
	ErrInvalidToken  = errors.New("invalid token")
	ErrTokenNotFound = errors.New("token not found")

	// 业务规则错误
	ErrInvalidInput          = errors.New("invalid input")
	ErrValidationFailed      = errors.New("validation failed")
	ErrBusinessRuleViolation = errors.New("business rule violation")
	ErrResourceNotFound      = errors.New("resource not found")
	ErrResourceConflict      = errors.New("resource conflict")
	ErrOperationNotAllowed   = errors.New("operation not allowed")

	// 系统错误
	ErrInternalServer     = errors.New("internal server error")
	ErrDatabaseConnection = errors.New("database connection error")
	ErrCacheConnection    = errors.New("cache connection error")
	ErrServiceUnavailable = errors.New("service unavailable")
	ErrTimeout            = errors.New("operation timeout")
)

// ErrorCode 错误码类型
type ErrorCode string

const (
	// 用户相关错误码
	CodeUserNotFound       ErrorCode = "USER_NOT_FOUND"
	CodeUserAlreadyExists  ErrorCode = "USER_ALREADY_EXISTS"
	CodeInvalidCredentials ErrorCode = "INVALID_CREDENTIALS"
	CodeInvalidPassword    ErrorCode = "INVALID_PASSWORD"
	CodePasswordTooWeak    ErrorCode = "PASSWORD_TOO_WEAK"
	CodeUserNotActive      ErrorCode = "USER_NOT_ACTIVE"
	CodeInsufficientPoints ErrorCode = "INSUFFICIENT_POINTS"

	// 比赛相关错误码
	CodeMatchNotFound         ErrorCode = "MATCH_NOT_FOUND"
	CodeMatchAlreadyStarted   ErrorCode = "MATCH_ALREADY_STARTED"
	CodeMatchNotActive        ErrorCode = "MATCH_NOT_ACTIVE"
	CodeMatchAlreadyCompleted ErrorCode = "MATCH_ALREADY_COMPLETED"
	CodeMatchAlreadyFinished  ErrorCode = "MATCH_ALREADY_FINISHED"
	CodeInvalidWinner         ErrorCode = "INVALID_WINNER"
	CodeInvalidScore          ErrorCode = "INVALID_SCORE"
	CodeInvalidMatchStatus    ErrorCode = "INVALID_MATCH_STATUS"
	CodeInvalidStartTime      ErrorCode = "INVALID_START_TIME"
	CodeInvalidTournament     ErrorCode = "INVALID_TOURNAMENT"

	// 预测相关错误码
	CodePredictionNotFound         ErrorCode = "PREDICTION_NOT_FOUND"
	CodePredictionAlreadyExists    ErrorCode = "PREDICTION_ALREADY_EXISTS"
	CodeCannotModifyPrediction     ErrorCode = "CANNOT_MODIFY_PREDICTION"
	CodePredictionAlreadyProcessed ErrorCode = "PREDICTION_ALREADY_PROCESSED"
	CodeTooManyModifications       ErrorCode = "TOO_MANY_MODIFICATIONS"
	CodeModificationNotAllowed     ErrorCode = "MODIFICATION_NOT_ALLOWED"

	// 投票相关错误码
	CodeVoteNotFound            ErrorCode = "VOTE_NOT_FOUND"
	CodeVoteAlreadyExists       ErrorCode = "VOTE_ALREADY_EXISTS"
	CodeCannotVoteOwnPrediction ErrorCode = "CANNOT_VOTE_OWN_PREDICTION"
	CodeVoteNotAllowed          ErrorCode = "VOTE_NOT_ALLOWED"
	CodeDailyVoteLimitExceeded  ErrorCode = "DAILY_VOTE_LIMIT_EXCEEDED"

	// 认证相关错误码
	CodeUnauthorized  ErrorCode = "UNAUTHORIZED"
	CodeForbidden     ErrorCode = "FORBIDDEN"
	CodeTokenExpired  ErrorCode = "TOKEN_EXPIRED"
	CodeInvalidToken  ErrorCode = "INVALID_TOKEN"
	CodeTokenNotFound ErrorCode = "TOKEN_NOT_FOUND"

	// 业务规则错误码
	CodeInvalidInput          ErrorCode = "INVALID_INPUT"
	CodeValidationFailed      ErrorCode = "VALIDATION_FAILED"
	CodeBusinessRuleViolation ErrorCode = "BUSINESS_RULE_VIOLATION"
	CodeResourceNotFound      ErrorCode = "RESOURCE_NOT_FOUND"
	CodeResourceConflict      ErrorCode = "RESOURCE_CONFLICT"
	CodeOperationNotAllowed   ErrorCode = "OPERATION_NOT_ALLOWED"

	// 系统错误码
	CodeInternalServer     ErrorCode = "INTERNAL_SERVER_ERROR"
	CodeDatabaseConnection ErrorCode = "DATABASE_CONNECTION_ERROR"
	CodeCacheConnection    ErrorCode = "CACHE_CONNECTION_ERROR"
	CodeServiceUnavailable ErrorCode = "SERVICE_UNAVAILABLE"
	CodeTimeout            ErrorCode = "TIMEOUT"
)

// DomainError 领域错误
type DomainError struct {
	Code    ErrorCode              `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// Error 实现 error 接口
func (e *DomainError) Error() string {
	return e.Message
}

// NewDomainError 创建领域错误
func NewDomainError(code ErrorCode, message string, details ...map[string]interface{}) *DomainError {
	var detailsMap map[string]interface{}
	if len(details) > 0 {
		detailsMap = details[0]
	}

	return &DomainError{
		Code:    code,
		Message: message,
		Details: detailsMap,
	}
}

// GetErrorCode 获取错误对应的错误码
func GetErrorCode(err error) ErrorCode {
	switch err {
	case ErrUserNotFound:
		return CodeUserNotFound
	case ErrUserAlreadyExists:
		return CodeUserAlreadyExists
	case ErrInvalidCredentials:
		return CodeInvalidCredentials
	case ErrInvalidPassword:
		return CodeInvalidPassword
	case ErrPasswordTooWeak:
		return CodePasswordTooWeak
	case ErrUserNotActive:
		return CodeUserNotActive
	case ErrInsufficientPoints:
		return CodeInsufficientPoints
	case ErrMatchNotFound:
		return CodeMatchNotFound
	case ErrMatchAlreadyStarted:
		return CodeMatchAlreadyStarted
	case ErrMatchNotActive:
		return CodeMatchNotActive
	case ErrMatchAlreadyCompleted:
		return CodeMatchAlreadyCompleted
	case ErrMatchAlreadyFinished:
		return CodeMatchAlreadyFinished
	case ErrInvalidWinner:
		return CodeInvalidWinner
	case ErrInvalidScore:
		return CodeInvalidScore
	case ErrInvalidMatchStatus:
		return CodeInvalidMatchStatus
	case ErrInvalidStartTime:
		return CodeInvalidStartTime
	case ErrInvalidTournament:
		return CodeInvalidTournament
	case ErrPredictionNotFound:
		return CodePredictionNotFound
	case ErrPredictionAlreadyExists:
		return CodePredictionAlreadyExists
	case ErrCannotModifyPrediction:
		return CodeCannotModifyPrediction
	case ErrPredictionAlreadyProcessed:
		return CodePredictionAlreadyProcessed
	case ErrTooManyModifications:
		return CodeTooManyModifications
	case ErrModificationNotAllowed:
		return CodeModificationNotAllowed
	case ErrVoteNotFound:
		return CodeVoteNotFound
	case ErrVoteAlreadyExists:
		return CodeVoteAlreadyExists
	case ErrCannotVoteOwnPrediction:
		return CodeCannotVoteOwnPrediction
	case ErrVoteNotAllowed:
		return CodeVoteNotAllowed
	case ErrDailyVoteLimitExceeded:
		return CodeDailyVoteLimitExceeded
	case ErrUnauthorized:
		return CodeUnauthorized
	case ErrForbidden:
		return CodeForbidden
	case ErrTokenExpired:
		return CodeTokenExpired
	case ErrInvalidToken:
		return CodeInvalidToken
	case ErrTokenNotFound:
		return CodeTokenNotFound
	case ErrInvalidInput:
		return CodeInvalidInput
	case ErrValidationFailed:
		return CodeValidationFailed
	case ErrBusinessRuleViolation:
		return CodeBusinessRuleViolation
	case ErrResourceNotFound:
		return CodeResourceNotFound
	case ErrResourceConflict:
		return CodeResourceConflict
	case ErrOperationNotAllowed:
		return CodeOperationNotAllowed
	case ErrInternalServer:
		return CodeInternalServer
	case ErrDatabaseConnection:
		return CodeDatabaseConnection
	case ErrCacheConnection:
		return CodeCacheConnection
	case ErrServiceUnavailable:
		return CodeServiceUnavailable
	case ErrTimeout:
		return CodeTimeout
	default:
		return CodeInternalServer
	}
}
