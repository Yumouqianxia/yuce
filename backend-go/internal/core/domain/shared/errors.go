package shared

import (
	"fmt"
	"net/http"
)

// AppError 应用错误类型
type AppError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// Error 实现 error 接口
func (e *AppError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s: %s", e.Message, e.Details)
	}
	return e.Message
}

// NewAppError 创建应用错误
func NewAppError(code int, message string, details ...string) *AppError {
	err := &AppError{
		Code:    code,
		Message: message,
	}
	if len(details) > 0 {
		err.Details = details[0]
	}
	return err
}

// 预定义错误
var (
	// 通用错误
	ErrInternalServer = NewAppError(http.StatusInternalServerError, "内部服务器错误")
	ErrBadRequest     = NewAppError(http.StatusBadRequest, "请求参数错误")
	ErrUnauthorized   = NewAppError(http.StatusUnauthorized, "未授权访问")
	ErrForbidden      = NewAppError(http.StatusForbidden, "禁止访问")
	ErrNotFound       = NewAppError(http.StatusNotFound, "资源不存在")
	ErrConflict       = NewAppError(http.StatusConflict, "资源冲突")

	// 用户相关错误
	ErrUserNotFound    = NewAppError(http.StatusNotFound, "用户不存在")
	ErrUserExists      = NewAppError(http.StatusConflict, "用户已存在")
	ErrInvalidPassword = NewAppError(http.StatusUnauthorized, "密码错误")
	ErrInvalidToken    = NewAppError(http.StatusUnauthorized, "令牌无效")
	ErrTokenExpired    = NewAppError(http.StatusUnauthorized, "令牌已过期")
	ErrUsernameExists  = NewAppError(http.StatusConflict, "用户名已存在")
	ErrEmailExists     = NewAppError(http.StatusConflict, "邮箱已存在")

	// 比赛相关错误
	ErrMatchNotFound    = NewAppError(http.StatusNotFound, "比赛不存在")
	ErrMatchStarted     = NewAppError(http.StatusBadRequest, "比赛已开始")
	ErrMatchFinished    = NewAppError(http.StatusBadRequest, "比赛已结束")
	ErrMatchCancelled   = NewAppError(http.StatusBadRequest, "比赛已取消")
	ErrInvalidMatchTime = NewAppError(http.StatusBadRequest, "比赛时间无效")

	// 预测相关错误
	ErrPredictionNotFound    = NewAppError(http.StatusNotFound, "预测不存在")
	ErrPredictionExists      = NewAppError(http.StatusConflict, "预测已存在")
	ErrPredictionNotAllowed  = NewAppError(http.StatusBadRequest, "不允许预测")
	ErrPredictionNotEditable = NewAppError(http.StatusBadRequest, "预测不可编辑")
	ErrSelfVoteNotAllowed    = NewAppError(http.StatusBadRequest, "不能给自己的预测投票")

	// 投票相关错误
	ErrVoteNotFound = NewAppError(http.StatusNotFound, "投票不存在")
	ErrVoteExists   = NewAppError(http.StatusConflict, "已投票")

	// 数据库相关错误
	ErrDatabaseConnection = NewAppError(http.StatusInternalServerError, "数据库连接错误")
	ErrDatabaseQuery      = NewAppError(http.StatusInternalServerError, "数据库查询错误")

	// 缓存相关错误
	ErrCacheConnection = NewAppError(http.StatusInternalServerError, "缓存连接错误")
	ErrCacheOperation  = NewAppError(http.StatusInternalServerError, "缓存操作错误")

	// 验证相关错误
	ErrValidationFailed = NewAppError(http.StatusBadRequest, "数据验证失败")
)

// IsAppError 检查是否为应用错误
func IsAppError(err error) (*AppError, bool) {
	if appErr, ok := err.(*AppError); ok {
		return appErr, true
	}
	return nil, false
}

// WrapError 包装错误为应用错误
func WrapError(err error, code int, message string) *AppError {
	return NewAppError(code, message, err.Error())
}
