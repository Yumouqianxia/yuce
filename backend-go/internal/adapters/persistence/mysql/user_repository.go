package mysql

import (
	"context"
	"errors"
	"fmt"

	"backend-go/internal/core/domain/user"
	"backend-go/internal/shared/password"
	"gorm.io/gorm"
)

// UserRepository MySQL 用户仓储实现
type UserRepository struct {
	db              *gorm.DB
	passwordService password.Service
}

// NewUserRepository 创建用户仓储
func NewUserRepository(db *gorm.DB, passwordService password.Service) user.Repository {
	return &UserRepository{
		db:              db,
		passwordService: passwordService,
	}
}

// Create 创建用户
func (r *UserRepository) Create(ctx context.Context, u *user.User) error {
	if u == nil {
		return errors.New("user cannot be nil")
	}

	// 验证必填字段
	if u.Username == "" {
		return errors.New("username is required")
	}
	if u.Email == "" {
		return errors.New("email is required")
	}
	if u.Password == "" {
		return errors.New("password is required")
	}

	// 检查用户名是否已存在
	exists, err := r.ExistsByUsername(ctx, u.Username)
	if err != nil {
		return fmt.Errorf("failed to check username existence: %w", err)
	}
	if exists {
		return errors.New("username already exists")
	}

	// 检查邮箱是否已存在
	exists, err = r.ExistsByEmail(ctx, u.Email)
	if err != nil {
		return fmt.Errorf("failed to check email existence: %w", err)
	}
	if exists {
		return errors.New("email already exists")
	}

	// 哈希密码
	hashedPassword, err := r.passwordService.HashPassword(u.Password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}
	u.Password = hashedPassword

	// 设置默认值
	if u.Role == "" {
		u.Role = user.UserRoleUser
	}
	if u.Points == 0 {
		u.Points = 0
	}

	// 创建用户
	if err := r.db.WithContext(ctx).Create(u).Error; err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetByID 根据 ID 获取用户
func (r *UserRepository) GetByID(ctx context.Context, id uint) (*user.User, error) {
	if id == 0 {
		return nil, errors.New("invalid user ID")
	}

	var u user.User
	if err := r.db.WithContext(ctx).First(&u, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	return &u, nil
}

// GetByUsername 根据用户名获取用户
func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*user.User, error) {
	if username == "" {
		return nil, errors.New("username cannot be empty")
	}

	var u user.User
	if err := r.db.WithContext(ctx).Where("username = ?", username).First(&u).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}

	return &u, nil
}

// GetByEmail 根据邮箱获取用户
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	if email == "" {
		return nil, errors.New("email cannot be empty")
	}

	var u user.User
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&u).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return &u, nil
}

// Update 更新用户信息
func (r *UserRepository) Update(ctx context.Context, u *user.User) error {
	if u == nil {
		return errors.New("user cannot be nil")
	}
	if u.ID == 0 {
		return errors.New("user ID is required")
	}

	// 检查用户是否存在
	_, err := r.GetByID(ctx, u.ID)
	if err != nil {
		return err
	}

	// 如果更新了用户名，检查是否重复
	if u.Username != "" {
		var existingUser user.User
		err := r.db.WithContext(ctx).Where("username = ? AND id != ?", u.Username, u.ID).First(&existingUser).Error
		if err == nil {
			return errors.New("username already exists")
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("failed to check username uniqueness: %w", err)
		}
	}

	// 如果更新了邮箱，检查是否重复
	if u.Email != "" {
		var existingUser user.User
		err := r.db.WithContext(ctx).Where("email = ? AND id != ?", u.Email, u.ID).First(&existingUser).Error
		if err == nil {
			return errors.New("email already exists")
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("failed to check email uniqueness: %w", err)
		}
	}

	// 如果更新了密码，需要哈希
	if u.Password != "" {
		hashedPassword, err := r.passwordService.HashPassword(u.Password)
		if err != nil {
			return fmt.Errorf("failed to hash password: %w", err)
		}
		u.Password = hashedPassword
	}

	// 更新用户
	if err := r.db.WithContext(ctx).Save(u).Error; err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// UpdatePoints 更新用户积分
func (r *UserRepository) UpdatePoints(ctx context.Context, userID uint, points int) error {
	if userID == 0 {
		return errors.New("invalid user ID")
	}

	// 使用原子操作更新积分
	result := r.db.WithContext(ctx).Model(&user.User{}).
		Where("id = ?", userID).
		Update("points", gorm.Expr("points + ?", points))

	if result.Error != nil {
		return fmt.Errorf("failed to update user points: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.New("user not found")
	}

	return nil
}

// GetLeaderboard 获取排行榜
func (r *UserRepository) GetLeaderboard(ctx context.Context, tournament string, limit int) ([]user.LeaderboardEntry, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	var entries []user.LeaderboardEntry

	// 构建查询
	query := r.db.WithContext(ctx).
		Model(&user.User{}).
		Select("id as user_id, username, nickname, avatar, points, ROW_NUMBER() OVER (ORDER BY points DESC) as rank").
		Order("points DESC").
		Limit(limit)

	// 如果指定了锦标赛，这里可以根据业务需求添加过滤条件
	// 目前先返回全局排行榜
	if err := query.Scan(&entries).Error; err != nil {
		return nil, fmt.Errorf("failed to get leaderboard: %w", err)
	}

	// 设置锦标赛字段
	for i := range entries {
		entries[i].Tournament = tournament
		entries[i].Rank = i + 1 // 重新计算排名，确保从1开始
	}

	return entries, nil
}

// ExistsByUsername 检查用户名是否存在
func (r *UserRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	if username == "" {
		return false, errors.New("username cannot be empty")
	}

	var count int64
	if err := r.db.WithContext(ctx).Model(&user.User{}).Where("username = ?", username).Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check username existence: %w", err)
	}

	return count > 0, nil
}

// ExistsByEmail 检查邮箱是否存在
func (r *UserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	if email == "" {
		return false, errors.New("email cannot be empty")
	}

	var count int64
	if err := r.db.WithContext(ctx).Model(&user.User{}).Where("email = ?", email).Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check email existence: %w", err)
	}

	return count > 0, nil
}

// Delete 删除用户
func (r *UserRepository) Delete(ctx context.Context, id uint) error {
	if id == 0 {
		return errors.New("invalid user ID")
	}

	// 检查用户是否存在
	_, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// 软删除用户
	if err := r.db.WithContext(ctx).Delete(&user.User{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

// ValidatePassword 验证用户密码
func (r *UserRepository) ValidatePassword(ctx context.Context, userID uint, password string) (bool, error) {
	u, err := r.GetByID(ctx, userID)
	if err != nil {
		return false, err
	}

	return r.passwordService.ValidatePassword(u.Password, password), nil
}

// ChangePassword 修改用户密码
func (r *UserRepository) ChangePassword(ctx context.Context, userID uint, newPassword string) error {
	if userID == 0 {
		return errors.New("invalid user ID")
	}
	if newPassword == "" {
		return errors.New("new password cannot be empty")
	}

	// 验证密码强度
	if err := r.passwordService.ValidatePasswordStrength(newPassword); err != nil {
		return fmt.Errorf("password validation failed: %w", err)
	}

	// 哈希新密码
	hashedPassword, err := r.passwordService.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// 更新密码
	result := r.db.WithContext(ctx).Model(&user.User{}).
		Where("id = ?", userID).
		Update("password", hashedPassword)

	if result.Error != nil {
		return fmt.Errorf("failed to update password: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.New("user not found")
	}

	return nil
}
