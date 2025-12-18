package middleware

import (
	"strconv"
	"strings"

	"backend-go/internal/core/domain/user"
	"backend-go/internal/shared/logger"
	"backend-go/pkg/response"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware 认证中间件
type AuthMiddleware struct {
	userService user.Service
}

// NewAuthMiddleware 创建认证中间件
func NewAuthMiddleware(userService user.Service) *AuthMiddleware {
	return &AuthMiddleware{
		userService: userService,
	}
}

// RequireAuth 需要认证的中间件
func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从 Authorization 头获取令牌
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "Authorization header is required")
			c.Abort()
			return
		}

		// 检查 Bearer 前缀
		if !strings.HasPrefix(authHeader, "Bearer ") {
			response.Unauthorized(c, "Invalid authorization header format")
			c.Abort()
			return
		}

		// 提取令牌
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == "" {
			response.Unauthorized(c, "Token is required")
			c.Abort()
			return
		}

		// 验证令牌
		foundUser, err := m.userService.ValidateToken(c.Request.Context(), token)
		if err != nil {
			logger.Warnf("Invalid token: %v", err)
			response.Unauthorized(c, "Invalid or expired token")
			c.Abort()
			return
		}

		// 将用户信息存储到上下文中
		c.Set("user_id", strconv.FormatUint(uint64(foundUser.ID), 10))
		c.Set("username", foundUser.Username)
		c.Set("user_role", string(foundUser.Role))
		c.Set("user", foundUser)

		c.Next()
	}
}

// RequireRole 需要特定角色的中间件
func (m *AuthMiddleware) RequireRole(roles ...user.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 先检查是否已认证
		userRole, exists := c.Get("user_role")
		if !exists {
			response.Unauthorized(c, "Authentication required")
			c.Abort()
			return
		}

		// 检查角色
		currentRole := user.UserRole(userRole.(string))
		for _, requiredRole := range roles {
			if currentRole == requiredRole {
				c.Next()
				return
			}
		}

		response.Forbidden(c, "Insufficient permissions")
		c.Abort()
	}
}

// RequireAdmin 需要管理员权限的中间件
func (m *AuthMiddleware) RequireAdmin() gin.HandlerFunc {
	return m.RequireRole(user.UserRoleAdmin)
}

// RequireSuperAdmin 超级管理员占位校验（当前无等级字段时默认等同管理员，可后续收紧）
func (m *AuthMiddleware) RequireSuperAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists || userRole.(string) != string(user.UserRoleAdmin) {
			response.Forbidden(c, "Insufficient permissions")
			c.Abort()
			return
		}

		// 预留超级管理员标识：若上游已设置 is_super_admin=false，则阻断；未设置则视为通过
		if isSuper, ok := c.Get("is_super_admin"); ok {
			if isSuperBool, ok := isSuper.(bool); ok && !isSuperBool {
				response.Forbidden(c, "Super admin required")
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// OptionalAuth 可选认证中间件（不强制要求认证）
func (m *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从 Authorization 头获取令牌
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		// 检查 Bearer 前缀
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.Next()
			return
		}

		// 提取令牌
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == "" {
			c.Next()
			return
		}

		// 验证令牌
		foundUser, err := m.userService.ValidateToken(c.Request.Context(), token)
		if err != nil {
			// 令牌无效，但不阻止请求继续
			c.Next()
			return
		}

		// 将用户信息存储到上下文中
		c.Set("user_id", strconv.FormatUint(uint64(foundUser.ID), 10))
		c.Set("username", foundUser.Username)
		c.Set("user_role", string(foundUser.Role))
		c.Set("user", foundUser)

		c.Next()
	}
}

// GetCurrentUser 从上下文获取当前用户
func GetCurrentUser(c *gin.Context) (*user.User, bool) {
	userInterface, exists := c.Get("user")
	if !exists {
		return nil, false
	}

	currentUser, ok := userInterface.(*user.User)
	return currentUser, ok
}

// GetCurrentUserID 从上下文获取当前用户ID
func GetCurrentUserID(c *gin.Context) (uint, bool) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}

	userID, err := strconv.ParseUint(userIDStr.(string), 10, 32)
	if err != nil {
		return 0, false
	}

	return uint(userID), true
}

// IsAdmin 检查当前用户是否为管理员
func IsAdmin(c *gin.Context) bool {
	userRole, exists := c.Get("user_role")
	if !exists {
		return false
	}

	return userRole.(string) == string(user.UserRoleAdmin)
}
