package user

import (
	"context"
)

// Repository 用户仓储接口
type Repository interface {
	// Create 创建用户
	Create(ctx context.Context, user *User) error

	// GetByID 根据 ID 获取用户
	GetByID(ctx context.Context, id uint) (*User, error)

	// GetByUsername 根据用户名获取用户
	GetByUsername(ctx context.Context, username string) (*User, error)

	// GetByEmail 根据邮箱获取用户
	GetByEmail(ctx context.Context, email string) (*User, error)

	// Update 更新用户信息
	Update(ctx context.Context, user *User) error

	// UpdatePoints 更新用户积分
	UpdatePoints(ctx context.Context, userID uint, points int) error

	// GetLeaderboard 获取排行榜
	GetLeaderboard(ctx context.Context, tournament string, limit int) ([]LeaderboardEntry, error)

	// ExistsByUsername 检查用户名是否存在
	ExistsByUsername(ctx context.Context, username string) (bool, error)

	// ExistsByEmail 检查邮箱是否存在
	ExistsByEmail(ctx context.Context, email string) (bool, error)

	// Delete 删除用户
	Delete(ctx context.Context, id uint) error

	// ValidatePassword 验证用户密码
	ValidatePassword(ctx context.Context, userID uint, password string) (bool, error)

	// ChangePassword 修改用户密码
	ChangePassword(ctx context.Context, userID uint, newPassword string) error
}
