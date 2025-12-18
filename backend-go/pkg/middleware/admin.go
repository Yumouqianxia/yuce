package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"backend-go/internal/core/domain/user"
	"backend-go/pkg/response"
)

// AdminMiddleware 管理员权限中间件
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从上下文中获取用户信息（由 AuthMiddleware 设置）
		userValue, exists := c.Get("user")
		if !exists {
			response.Error(c, http.StatusUnauthorized, "Authentication required", "User not found in context")
			c.Abort()
			return
		}

		currentUser, ok := userValue.(*user.User)
		if !ok {
			response.Error(c, http.StatusInternalServerError, "Invalid user context", "Failed to cast user from context")
			c.Abort()
			return
		}

		// 检查用户是否为管理员
		if !currentUser.IsAdmin() {
			response.Error(c, http.StatusForbidden, "Admin privileges required", "User does not have admin role")
			c.Abort()
			return
		}

		// 继续处理请求
		c.Next()
	}
}
