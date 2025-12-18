package middleware

import (
	"net/http"
	"strings"

	"backend-go/internal/shared/jwt"
	"backend-go/pkg/response"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware 认证中间件
type AuthMiddleware struct {
	jwtService jwt.JWTService
}

// NewAuthMiddleware 创建认证中间件
func NewAuthMiddleware(jwtService jwt.JWTService) *AuthMiddleware {
	return &AuthMiddleware{
		jwtService: jwtService,
	}
}

// RequireAuth 需要认证的中间件
func (a *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := a.extractToken(c)
		if token == "" {
			response.Error(c, http.StatusUnauthorized, "MISSING_TOKEN", "Authorization token is required")
			c.Abort()
			return
		}

		claims, err := a.jwtService.ValidateToken(token)
		if err != nil {
			response.Error(c, http.StatusUnauthorized, "INVALID_TOKEN", "Invalid or expired token: " + err.Error())
			c.Abort()
			return
		}

		// 验证是否为访问令牌
		if claims.Type != "access" {
			response.Error(c, http.StatusUnauthorized, "INVALID_TOKEN_TYPE", "Access token required")
			c.Abort()
			return
		}

		// 将用户信息存储到上下文中
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("user_role", claims.Role)
		c.Set("claims", claims)

		c.Next()
	}
}

// RequireAdmin 需要管理员权限的中间件
func (a *AuthMiddleware) RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 先执行认证
		a.RequireAuth()(c)
		if c.IsAborted() {
			return
		}

		// 检查管理员权限
		role, exists := c.Get("user_role")
		if !exists || role != "admin" {
			response.Error(c, http.StatusForbidden, "INSUFFICIENT_PERMISSIONS", "Admin privileges required")
			c.Abort()
			return
		}

		c.Next()
	}
}

// OptionalAuth 可选认证中间件（不强制要求认证）
func (a *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := a.extractToken(c)
		if token == "" {
			c.Next()
			return
		}

		claims, err := a.jwtService.ValidateToken(token)
		if err != nil {
			// 令牌无效，但不阻止请求继续
			c.Next()
			return
		}

		// 验证是否为访问令牌
		if claims.Type != "access" {
			c.Next()
			return
		}

		// 将用户信息存储到上下文中
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("user_role", claims.Role)
		c.Set("claims", claims)

		c.Next()
	}
}

// extractToken 从请求中提取令牌
func (a *AuthMiddleware) extractToken(c *gin.Context) string {
	// 从 Authorization header 中提取
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		// Bearer token 格式
		if strings.HasPrefix(authHeader, "Bearer ") {
			return strings.TrimPrefix(authHeader, "Bearer ")
		}
		// 直接返回 token
		return authHeader
	}

	// 从查询参数中提取（用于特殊场景）
	token := c.Query("token")
	if token != "" {
		return token
	}

	// 从 Cookie 中提取
	cookie, err := c.Cookie("access_token")
	if err == nil && cookie != "" {
		return cookie
	}

	return ""
}

// GetCurrentUser 获取当前用户信息的辅助函数
func GetCurrentUser(c *gin.Context) (userID uint, username string, role string, exists bool) {
	userIDVal, userIDExists := c.Get("user_id")
	usernameVal, usernameExists := c.Get("username")
	roleVal, roleExists := c.Get("user_role")

	if !userIDExists || !usernameExists || !roleExists {
		return 0, "", "", false
	}

	userID, ok1 := userIDVal.(uint)
	username, ok2 := usernameVal.(string)
	role, ok3 := roleVal.(string)

	if !ok1 || !ok2 || !ok3 {
		return 0, "", "", false
	}

	return userID, username, role, true
}

// GetCurrentUserID 获取当前用户ID的辅助函数
func GetCurrentUserID(c *gin.Context) (uint, bool) {
	userID, _, _, exists := GetCurrentUser(c)
	return userID, exists
}

// IsCurrentUserAdmin 检查当前用户是否为管理员的辅助函数
func IsCurrentUserAdmin(c *gin.Context) bool {
	_, _, role, exists := GetCurrentUser(c)
	return exists && role == "admin"
}

// GetClaims 获取 JWT Claims 的辅助函数
func GetClaims(c *gin.Context) (*jwt.Claims, bool) {
	claimsVal, exists := c.Get("claims")
	if !exists {
		return nil, false
	}

	claims, ok := claimsVal.(*jwt.Claims)
	return claims, ok
}
