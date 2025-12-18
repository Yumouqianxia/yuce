package user

import (
	"context"
)

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
	Nickname string `json:"nickname" validate:"max=50"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// UpdateProfileRequest 更新个人资料请求
type UpdateProfileRequest struct {
	Nickname string `json:"nickname" validate:"max=50"`
	Avatar   string `json:"avatar" validate:"max=255"`
}

// AuthResponse 认证响应
type AuthResponse struct {
	User         *User  `json:"user"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

// Service 用户服务接口
type Service interface {
	// Register 用户注册
	Register(ctx context.Context, req *RegisterRequest) (*User, error)

	// Login 用户登录
	Login(ctx context.Context, req *LoginRequest) (*AuthResponse, error)

	// GetProfile 获取用户资料
	GetProfile(ctx context.Context, userID uint) (*User, error)

	// UpdateProfile 更新用户资料
	UpdateProfile(ctx context.Context, userID uint, req *UpdateProfileRequest) (*User, error)

	// GetLeaderboard 获取排行榜
	GetLeaderboard(ctx context.Context, tournament string) ([]LeaderboardEntry, error)

	// RefreshToken 刷新令牌
	RefreshToken(ctx context.Context, refreshToken string) (*AuthResponse, error)

	// ValidateToken 验证令牌
	ValidateToken(ctx context.Context, token string) (*User, error)

	// ChangePassword 重置/修改用户密码（管理员场景）
	ChangePassword(ctx context.Context, userID uint, newPassword string) error

	// ChangePasswordWithVerify 校验当前密码后修改（用户自助）
	ChangePasswordWithVerify(ctx context.Context, userID uint, currentPassword, newPassword string) error
}
