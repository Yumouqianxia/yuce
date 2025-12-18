package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"backend-go/internal/adapters/http/middleware"
	"backend-go/internal/core/domain/user"
	"backend-go/internal/shared/logger"
	"backend-go/pkg/response"

	"github.com/gin-gonic/gin"
)

// AuthHandler 认证处理器
type AuthHandler struct {
	userService user.Service
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler(userService user.Service) *AuthHandler {
	return &AuthHandler{
		userService: userService,
	}
}

// RegisterRequest 注册请求结构
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50" example:"johndoe"`
	Email    string `json:"email" binding:"required,email" example:"john@example.com"`
	Password string `json:"password" binding:"required,min=6" example:"password123"`
	Nickname string `json:"nickname" binding:"max=50" example:"John Doe"`
}

// LoginRequest 登录请求结构
type LoginRequest struct {
	Username string `json:"username" binding:"required" example:"johndoe"`
	Password string `json:"password" binding:"required" example:"password123"`
}

// RegisterResponse 注册响应结构
type RegisterResponse struct {
	ID        uint   `json:"id" example:"1"`
	Username  string `json:"username" example:"johndoe"`
	Email     string `json:"email" example:"john@example.com"`
	Nickname  string `json:"nickname" example:"John Doe"`
	Avatar    string `json:"avatar" example:"/api/uploads/avatar/demo.jpg"`
	Points    int    `json:"points" example:"0"`
	Role      string `json:"role" example:"user"`
	CreatedAt string `json:"createdAt" example:"2023-01-01T00:00:00Z"`
	UpdatedAt string `json:"updatedAt" example:"2023-01-01T00:00:00Z"`
}

// LoginResponse 登录响应结构
type LoginResponse struct {
	User         RegisterResponse `json:"user"`
	AccessToken  string           `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	RefreshToken string           `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	ExpiresIn    int64            `json:"expires_in" example:"3600"`
}

// RefreshTokenRequest 刷新令牌请求结构
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

// Register 用户注册
// @Summary 用户注册
// @Description 创建新用户账户
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "注册信息"
// @Success 201 {object} response.Response{data=RegisterResponse} "注册成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 409 {object} response.Response "用户名或邮箱已存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBind(&req); err != nil {
		logger.Warnf("Invalid register request: %v", err)
		response.Error(c, http.StatusBadRequest, "Invalid request parameters", err.Error())
		return
	}

	// 转换为领域请求
	domainReq := &user.RegisterRequest{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
		Nickname: req.Nickname,
	}

	// 调用服务
	newUser, err := h.userService.Register(c.Request.Context(), domainReq)
	if err != nil {
		// 根据错误类型返回不同的状态码
		if strings.Contains(err.Error(), "already exists") {
			response.Error(c, http.StatusConflict, "User already exists", err.Error())
			return
		}
		if strings.Contains(err.Error(), "validation failed") || strings.Contains(err.Error(), "password validation failed") {
			response.Error(c, http.StatusBadRequest, "Validation failed", err.Error())
			return
		}

		logger.Errorf("Failed to register user: %v", err)
		response.Error(c, http.StatusInternalServerError, "Failed to register user", "Internal server error")
		return
	}

	// 构造响应
	resp := RegisterResponse{
		ID:        newUser.ID,
		Username:  newUser.Username,
		Email:     newUser.Email,
		Nickname:  newUser.Nickname,
		Points:    newUser.Points,
		Role:      string(newUser.Role),
		CreatedAt: newUser.CreatedAt.Format(time.RFC3339),
		UpdatedAt: newUser.UpdatedAt.Format(time.RFC3339),
	}

	response.Success(c, http.StatusCreated, "User registered successfully", resp)
}

// ChangePassword 校验旧密码后修改
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok || userID == 0 {
		response.Error(c, http.StatusUnauthorized, "无效的用户ID", "")
		return
	}

	var req struct {
		CurrentPassword    string `json:"currentPassword" form:"currentPassword" binding:"required"`
		NewPassword        string `json:"newPassword" form:"newPassword" binding:"required"`
		NewPasswordConfirm string `json:"newPasswordConfirm" form:"newPasswordConfirm"`
	}
	if err := c.ShouldBind(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}

	if req.NewPasswordConfirm != "" && req.NewPasswordConfirm != req.NewPassword {
		response.Error(c, http.StatusBadRequest, "两次输入的新密码不一致", "")
		return
	}

	if err := h.userService.ChangePasswordWithVerify(c.Request.Context(), userID, req.CurrentPassword, req.NewPassword); err != nil {
		response.Error(c, http.StatusBadRequest, "修改密码失败", err.Error())
		return
	}

	response.Success(c, http.StatusOK, "密码修改成功", nil)
}

// UpdateProfile 更新个人资料（昵称/头像）
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "无效的用户ID", "")
		return
	}

	var req user.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}

	updated, err := h.userService.UpdateProfile(c.Request.Context(), userID, &req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "更新失败", err.Error())
		return
	}

	resp := RegisterResponse{
		ID:        updated.ID,
		Username:  updated.Username,
		Email:     updated.Email,
		Nickname:  updated.Nickname,
		Avatar:    updated.Avatar,
		Points:    updated.Points,
		Role:      string(updated.Role),
		CreatedAt: updated.CreatedAt.Format(time.RFC3339),
		UpdatedAt: updated.UpdatedAt.Format(time.RFC3339),
	}

	response.Success(c, http.StatusOK, "Profile updated", resp)
}

// Login 用户登录
// @Summary 用户登录
// @Description 用户登录获取访问令牌
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body LoginRequest true "登录信息"
// @Success 200 {object} response.Response{data=LoginResponse} "登录成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 401 {object} response.Response "用户名或密码错误"
// @Failure 423 {object} response.Response "账户被锁定"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBind(&req); err != nil {
		logger.Warnf("Invalid login request: %v", err)
		response.Error(c, http.StatusBadRequest, "Invalid request parameters", err.Error())
		return
	}

	// 转换为领域请求
	domainReq := &user.LoginRequest{
		Username: req.Username,
		Password: req.Password,
	}

	// 调用服务
	authResp, err := h.userService.Login(c.Request.Context(), domainReq)
	if err != nil {
		// 根据错误类型返回不同的状态码
		if strings.Contains(err.Error(), "invalid username or password") {
			response.Error(c, http.StatusUnauthorized, "Authentication failed", "Invalid username or password")
			return
		}
		if strings.Contains(err.Error(), "account locked") {
			response.Error(c, http.StatusLocked, "Account locked", err.Error())
			return
		}
		if strings.Contains(err.Error(), "validation failed") {
			response.Error(c, http.StatusBadRequest, "Validation failed", err.Error())
			return
		}

		logger.Errorf("Failed to login user: %v", err)
		response.Error(c, http.StatusInternalServerError, "Login failed", "Internal server error")
		return
	}

	// 构造响应
	resp := LoginResponse{
		User: RegisterResponse{
			ID:        authResp.User.ID,
			Username:  authResp.User.Username,
			Email:     authResp.User.Email,
			Nickname:  authResp.User.Nickname,
			Avatar:    authResp.User.Avatar,
			Points:    authResp.User.Points,
			Role:      string(authResp.User.Role),
			CreatedAt: authResp.User.CreatedAt.Format(time.RFC3339),
			UpdatedAt: authResp.User.UpdatedAt.Format(time.RFC3339),
		},
		AccessToken:  authResp.AccessToken,
		RefreshToken: authResp.RefreshToken,
		ExpiresIn:    authResp.ExpiresIn,
	}

	response.Success(c, http.StatusOK, "Login successful", resp)
}

// RefreshToken 刷新访问令牌
// @Summary 刷新访问令牌
// @Description 使用刷新令牌获取新的访问令牌
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body RefreshTokenRequest true "刷新令牌"
// @Success 200 {object} response.Response{data=LoginResponse} "刷新成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 401 {object} response.Response "刷新令牌无效"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warnf("Invalid refresh token request: %v", err)
		response.Error(c, http.StatusBadRequest, "Invalid request parameters", err.Error())
		return
	}

	// 调用服务
	authResp, err := h.userService.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		if strings.Contains(err.Error(), "invalid") || strings.Contains(err.Error(), "expired") {
			response.Error(c, http.StatusUnauthorized, "Invalid refresh token", err.Error())
			return
		}

		logger.Errorf("Failed to refresh token: %v", err)
		response.Error(c, http.StatusInternalServerError, "Token refresh failed", "Internal server error")
		return
	}

	// 构造响应
	resp := LoginResponse{
		User: RegisterResponse{
			ID:        authResp.User.ID,
			Username:  authResp.User.Username,
			Email:     authResp.User.Email,
			Nickname:  authResp.User.Nickname,
			Avatar:    authResp.User.Avatar,
			Points:    authResp.User.Points,
			Role:      string(authResp.User.Role),
			CreatedAt: authResp.User.CreatedAt.Format(time.RFC3339),
			UpdatedAt: authResp.User.UpdatedAt.Format(time.RFC3339),
		},
		AccessToken:  authResp.AccessToken,
		RefreshToken: authResp.RefreshToken,
		ExpiresIn:    authResp.ExpiresIn,
	}

	response.Success(c, http.StatusOK, "Token refreshed successfully", resp)
}

// GetProfile 获取用户资料
// @Summary 获取用户资料
// @Description 获取当前登录用户的资料信息
// @Tags 用户
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=RegisterResponse} "获取成功"
// @Failure 401 {object} response.Response "未授权"
// @Failure 404 {object} response.Response "用户不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /auth/profile [get]
func (h *AuthHandler) GetProfile(c *gin.Context) {
	// 从中间件获取用户ID
	userIDStr, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Unauthorized", "User ID not found in context")
		return
	}

	userID, err := strconv.ParseUint(userIDStr.(string), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid user ID", err.Error())
		return
	}

	// 调用服务
	foundUser, err := h.userService.GetProfile(c.Request.Context(), uint(userID))
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			response.Error(c, http.StatusNotFound, "User not found", err.Error())
			return
		}

		logger.Errorf("Failed to get user profile: %v", err)
		response.Error(c, http.StatusInternalServerError, "Failed to get profile", "Internal server error")
		return
	}

	// 构造响应
	resp := RegisterResponse{
		ID:        foundUser.ID,
		Username:  foundUser.Username,
		Email:     foundUser.Email,
		Nickname:  foundUser.Nickname,
		Avatar:    foundUser.Avatar,
		Points:    foundUser.Points,
		Role:      string(foundUser.Role),
		CreatedAt: foundUser.CreatedAt.Format(time.RFC3339),
		UpdatedAt: foundUser.UpdatedAt.Format(time.RFC3339),
	}

	response.Success(c, http.StatusOK, "Profile retrieved successfully", resp)
}

// Logout 用户登出
// @Summary 用户登出
// @Description 用户登出（客户端应删除本地令牌）
// @Tags 认证
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response "登出成功"
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	// JWT 是无状态的，服务端不需要做特殊处理
	// 客户端应该删除本地存储的令牌
	// 在生产环境中，可以考虑将令牌加入黑名单（需要 Redis 支持）

	response.Success(c, http.StatusOK, "Logout successful", nil)
}
