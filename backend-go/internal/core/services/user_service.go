package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"backend-go/internal/core/domain/user"
	"backend-go/internal/shared/jwt"
	"backend-go/internal/shared/logger"
	"backend-go/internal/shared/password"
)

// userService 用户服务实现
type userService struct {
	userRepo         user.Repository
	jwtService       jwt.JWTService
	passwordService  password.Service
	leaderboardCache LeaderboardCacheService
	loginAttempts    map[string]*LoginAttempt // 简单的内存存储，生产环境应使用 Redis
}

// LoginAttempt 登录尝试记录
type LoginAttempt struct {
	Count       int
	LastTry     time.Time
	LockedUntil *time.Time
}

// Config 用户服务配置
type Config struct {
	MaxLoginAttempts int           `mapstructure:"max_login_attempts"`
	LockoutDuration  time.Duration `mapstructure:"lockout_duration"`
}

// NewUserService 创建用户服务
func NewUserService(
	userRepo user.Repository,
	jwtService jwt.JWTService,
	passwordService password.Service,
	leaderboardCache LeaderboardCacheService,
	config Config,
) user.Service {
	if config.MaxLoginAttempts == 0 {
		config.MaxLoginAttempts = 5
	}
	if config.LockoutDuration == 0 {
		config.LockoutDuration = 15 * time.Minute
	}

	return &userService{
		userRepo:         userRepo,
		jwtService:       jwtService,
		passwordService:  passwordService,
		leaderboardCache: leaderboardCache,
		loginAttempts:    make(map[string]*LoginAttempt),
	}
}

// Register 用户注册
func (s *userService) Register(ctx context.Context, req *user.RegisterRequest) (*user.User, error) {
	if req == nil {
		return nil, errors.New("register request cannot be nil")
	}

	// 验证输入
	if err := s.validateRegisterRequest(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// 验证密码强度
	if err := s.passwordService.ValidatePasswordStrength(req.Password); err != nil {
		return nil, fmt.Errorf("password validation failed: %w", err)
	}

	// 检查用户名是否已存在
	exists, err := s.userRepo.ExistsByUsername(ctx, req.Username)
	if err != nil {
		logger.Error("Failed to check username existence: %v", err)
		return nil, errors.New("failed to check username availability")
	}
	if exists {
		return nil, errors.New("username already exists")
	}

	// 检查邮箱是否已存在
	exists, err = s.userRepo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		logger.Error("Failed to check email existence: %v", err)
		return nil, errors.New("failed to check email availability")
	}
	if exists {
		return nil, errors.New("email already exists")
	}

	// 创建用户实体
	newUser := &user.User{
		Username: strings.TrimSpace(req.Username),
		Email:    strings.ToLower(strings.TrimSpace(req.Email)),
		Password: req.Password, // 将在仓储层进行哈希
		Nickname: strings.TrimSpace(req.Nickname),
		Role:     user.UserRoleUser,
		Points:   0,
	}

	// 如果没有提供昵称，使用用户名
	if newUser.Nickname == "" {
		newUser.Nickname = newUser.Username
	}

	// 创建用户
	if err := s.userRepo.Create(ctx, newUser); err != nil {
		logger.Error("Failed to create user: %v", err)
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	logger.Infof("User registered successfully: %s (ID: %d)", newUser.Username, newUser.ID)
	return newUser, nil
}

// Login 用户登录
func (s *userService) Login(ctx context.Context, req *user.LoginRequest) (*user.AuthResponse, error) {
	if req == nil {
		return nil, errors.New("login request cannot be nil")
	}

	// 验证输入
	if err := s.validateLoginRequest(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// 检查登录尝试限制
	if err := s.checkLoginAttempts(req.Username); err != nil {
		return nil, err
	}

	// 获取用户（支持用户名或邮箱登录）
	var foundUser *user.User
	var err error

	// 判断是邮箱还是用户名
	if strings.Contains(req.Username, "@") {
		foundUser, err = s.userRepo.GetByEmail(ctx, req.Username)
	} else {
		foundUser, err = s.userRepo.GetByUsername(ctx, req.Username)
	}

	if err != nil {
		s.recordFailedLogin(req.Username)
		logger.Warnf("Login attempt with invalid username/email: %s", req.Username)
		return nil, errors.New("invalid username or password")
	}

	// 验证密码
	if !s.passwordService.ValidatePassword(foundUser.Password, req.Password) {
		s.recordFailedLogin(req.Username)
		logger.Warnf("Login attempt with invalid password for user: %s", foundUser.Username)
		return nil, errors.New("invalid username or password")
	}

	// 登录成功，清除失败记录
	s.clearLoginAttempts(req.Username)

	// 生成 JWT 令牌
	tokenPair, err := s.jwtService.GenerateToken(foundUser.ID, foundUser.Username, string(foundUser.Role))
	if err != nil {
		logger.Errorf("Failed to generate JWT token for user %s: %v", foundUser.Username, err)
		return nil, errors.New("failed to generate authentication token")
	}

	logger.Infof("User logged in successfully: %s (ID: %d)", foundUser.Username, foundUser.ID)

	return &user.AuthResponse{
		User:         foundUser,
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
	}, nil
}

// GetProfile 获取用户资料
func (s *userService) GetProfile(ctx context.Context, userID uint) (*user.User, error) {
	if userID == 0 {
		return nil, errors.New("invalid user ID")
	}

	foundUser, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		logger.Errorf("Failed to get user profile for ID %d: %v", userID, err)
		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}

	return foundUser, nil
}

// UpdateProfile 更新用户资料
func (s *userService) UpdateProfile(ctx context.Context, userID uint, req *user.UpdateProfileRequest) (*user.User, error) {
	if userID == 0 {
		return nil, errors.New("invalid user ID")
	}
	if req == nil {
		return nil, errors.New("update request cannot be nil")
	}

	// 获取现有用户
	existingUser, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// 更新字段
	if req.Nickname != "" {
		existingUser.Nickname = strings.TrimSpace(req.Nickname)
	}
	if req.Avatar != "" {
		existingUser.Avatar = strings.TrimSpace(req.Avatar)
	}

	// 保存更新
	if err := s.userRepo.Update(ctx, existingUser); err != nil {
		logger.Errorf("Failed to update user profile for ID %d: %v", userID, err)
		return nil, fmt.Errorf("failed to update user profile: %w", err)
	}

	logger.Infof("User profile updated successfully: %s (ID: %d)", existingUser.Username, existingUser.ID)
	return existingUser, nil
}

// GetLeaderboard 获取排行榜
func (s *userService) GetLeaderboard(ctx context.Context, tournament string) ([]user.LeaderboardEntry, error) {
	// 使用缓存服务获取排行榜
	entries, err := s.leaderboardCache.GetLeaderboard(ctx, tournament)
	if err != nil {
		logger.Errorf("Failed to get leaderboard for tournament %s: %v", tournament, err)
		return nil, fmt.Errorf("failed to get leaderboard: %w", err)
	}

	return entries, nil
}

// RefreshToken 刷新令牌
func (s *userService) RefreshToken(ctx context.Context, refreshToken string) (*user.AuthResponse, error) {
	if refreshToken == "" {
		return nil, errors.New("refresh token cannot be empty")
	}

	// 验证刷新令牌
	claims, err := s.jwtService.ValidateToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	if claims.Type != "refresh" {
		return nil, errors.New("invalid token type")
	}

	// 获取用户信息
	foundUser, err := s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// 生成新的令牌对
	tokenPair, err := s.jwtService.RefreshTokenWithUserInfo(refreshToken, foundUser.Username, string(foundUser.Role))
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	return &user.AuthResponse{
		User:         foundUser,
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
	}, nil
}

// ValidateToken 验证令牌
func (s *userService) ValidateToken(ctx context.Context, token string) (*user.User, error) {
	if token == "" {
		return nil, errors.New("token cannot be empty")
	}

	// 验证令牌
	claims, err := s.jwtService.ValidateToken(token)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	if claims.Type != "access" {
		return nil, errors.New("invalid token type")
	}

	// 获取用户信息
	foundUser, err := s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	return foundUser, nil
}

// ChangePassword 重置/修改用户密码（管理员调用）
func (s *userService) ChangePassword(ctx context.Context, userID uint, newPassword string) error {
	if userID == 0 {
		return errors.New("invalid user ID")
	}
	if newPassword == "" {
		return errors.New("new password is required")
	}

	// 调用仓储执行密码强度校验、哈希与更新
	if err := s.userRepo.ChangePassword(ctx, userID, newPassword); err != nil {
		return fmt.Errorf("failed to change password: %w", err)
	}

	logger.Infof("Password changed for user ID: %d", userID)
	return nil
}

// ChangePasswordWithVerify 校验当前密码后修改（用户自助）
func (s *userService) ChangePasswordWithVerify(ctx context.Context, userID uint, currentPassword, newPassword string) error {
	if userID == 0 {
		return errors.New("invalid user ID")
	}
	if currentPassword == "" || newPassword == "" {
		return errors.New("password cannot be empty")
	}

	// 获取用户
	foundUser, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// 校验当前密码
	if !s.passwordService.ValidatePassword(foundUser.Password, currentPassword) {
		return errors.New("current password is incorrect")
	}

	// 更新为新密码
	return s.ChangePassword(ctx, userID, newPassword)
}

// validateRegisterRequest 验证注册请求
func (s *userService) validateRegisterRequest(req *user.RegisterRequest) error {
	if req.Username == "" {
		return errors.New("username is required")
	}
	if len(req.Username) < 3 {
		return errors.New("username must be at least 3 characters long")
	}
	if len(req.Username) > 50 {
		return errors.New("username must be no more than 50 characters long")
	}

	if req.Email == "" {
		return errors.New("email is required")
	}
	if !strings.Contains(req.Email, "@") {
		return errors.New("invalid email format")
	}

	if req.Password == "" {
		return errors.New("password is required")
	}
	if len(req.Password) < 6 {
		return errors.New("password must be at least 6 characters long")
	}

	if len(req.Nickname) > 50 {
		return errors.New("nickname must be no more than 50 characters long")
	}

	return nil
}

// validateLoginRequest 验证登录请求
func (s *userService) validateLoginRequest(req *user.LoginRequest) error {
	if req.Username == "" {
		return errors.New("username is required")
	}
	if req.Password == "" {
		return errors.New("password is required")
	}
	return nil
}

// checkLoginAttempts 检查登录尝试限制
func (s *userService) checkLoginAttempts(username string) error {
	attempt, exists := s.loginAttempts[username]
	if !exists {
		return nil
	}

	// 检查是否仍在锁定期内
	if attempt.LockedUntil != nil && time.Now().Before(*attempt.LockedUntil) {
		return fmt.Errorf("account locked due to too many failed login attempts, try again after %v",
			attempt.LockedUntil.Sub(time.Now()).Round(time.Second))
	}

	// 如果锁定期已过，清除记录
	if attempt.LockedUntil != nil && time.Now().After(*attempt.LockedUntil) {
		delete(s.loginAttempts, username)
	}

	return nil
}

// recordFailedLogin 记录失败的登录尝试
func (s *userService) recordFailedLogin(username string) {
	now := time.Now()
	attempt, exists := s.loginAttempts[username]

	if !exists {
		s.loginAttempts[username] = &LoginAttempt{
			Count:   1,
			LastTry: now,
		}
		return
	}

	attempt.Count++
	attempt.LastTry = now

	// 如果达到最大尝试次数，锁定账户
	if attempt.Count >= 5 { // 可配置
		lockUntil := now.Add(15 * time.Minute) // 可配置
		attempt.LockedUntil = &lockUntil
		logger.Warnf("Account locked due to too many failed login attempts: %s", username)
	}
}

// clearLoginAttempts 清除登录尝试记录
func (s *userService) clearLoginAttempts(username string) {
	delete(s.loginAttempts, username)
}
