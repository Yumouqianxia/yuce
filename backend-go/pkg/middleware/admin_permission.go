package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"backend-go/internal/core/domain/admin"
	"backend-go/internal/core/ports"
	"backend-go/pkg/response"
)

// AdminPermissionMiddleware 管理员权限中间件
type AdminPermissionMiddleware struct {
	adminService      ports.AdminService
	adminAuditService ports.AdminAuditService
}

// NewAdminPermissionMiddleware 创建管理员权限中间件
func NewAdminPermissionMiddleware(
	adminService ports.AdminService,
	adminAuditService ports.AdminAuditService,
) *AdminPermissionMiddleware {
	return &AdminPermissionMiddleware{
		adminService:      adminService,
		adminAuditService: adminAuditService,
	}
}

// RequirePermission 需要特定权限的中间件
func (m *AdminPermissionMiddleware) RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := GetCurrentUserID(c)
		if !exists {
			response.Error(c, http.StatusUnauthorized, "Authentication required", "AUTHENTICATION_REQUIRED")
			c.Abort()
			return
		}

		// 检查权限
		hasPermission, err := m.adminService.HasPermission(c.Request.Context(), userID, permission)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, "Failed to check permission", "PERMISSION_CHECK_FAILED: " + err.Error())
			c.Abort()
			return
		}

		if !hasPermission {
			response.Error(c, http.StatusForbidden, "INSUFFICIENT_PERMISSIONS", fmt.Sprintf("Permission required: %s", permission))
			c.Abort()
			return
		}

		// 将权限信息存储到上下文中
		c.Set("required_permission", permission)
		c.Next()
	}
}

// RequireSportAccess 需要运动类型访问权限的中间件
func (m *AdminPermissionMiddleware) RequireSportAccess() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := GetCurrentUserID(c)
		if !exists {
			response.Error(c, http.StatusUnauthorized, "Authentication required", "AUTHENTICATION_REQUIRED")
			c.Abort()
			return
		}

		// 从路径参数获取运动类型ID
		sportTypeIDStr := c.Param("sport_type_id")
		if sportTypeIDStr == "" {
			sportTypeIDStr = c.Param("id") // 有些路由可能使用 id 参数
		}

		if sportTypeIDStr == "" {
			response.Error(c, http.StatusBadRequest, "Sport type ID is required", "MISSING_SPORT_TYPE_ID")
			c.Abort()
			return
		}

		sportTypeID, err := strconv.ParseUint(sportTypeIDStr, 10, 32)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "Invalid sport type ID", "INVALID_SPORT_TYPE_ID: " + err.Error())
			c.Abort()
			return
		}

		// 检查运动类型访问权限
		hasAccess, err := m.adminService.HasSportAccess(c.Request.Context(), userID, uint(sportTypeID))
		if err != nil {
			response.Error(c, http.StatusInternalServerError, "Failed to check sport access", "ACCESS_CHECK_FAILED: " + err.Error())
			c.Abort()
			return
		}

		if !hasAccess {
			response.Error(c, http.StatusForbidden, "Sport type access required", "INSUFFICIENT_SPORT_ACCESS")
			c.Abort()
			return
		}

		// 将运动类型ID存储到上下文中
		c.Set("sport_type_id", uint(sportTypeID))
		c.Next()
	}
}

// RequireAdminLevel 需要特定管理员级别的中间件
func (m *AdminPermissionMiddleware) RequireAdminLevel(level admin.AdminLevel) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := GetCurrentUserID(c)
		if !exists {
			response.Error(c, http.StatusUnauthorized, "Authentication required", "AUTHENTICATION_REQUIRED")
			c.Abort()
			return
		}

		// 获取管理员信息
		adminUser, err := m.adminService.GetAdmin(c.Request.Context(), userID)
		if err != nil {
			response.Error(c, http.StatusForbidden, "Admin privileges required", "NOT_ADMIN: " + err.Error())
			c.Abort()
			return
		}

		if !adminUser.IsActive {
			response.Error(c, http.StatusForbidden, "Admin account is disabled", "ADMIN_DISABLED")
			c.Abort()
			return
		}

		if adminUser.AdminLevel < level {
			response.Error(c, http.StatusForbidden, "INSUFFICIENT_ADMIN_LEVEL", 
				fmt.Sprintf("Admin level %s or higher required", level.GetLevelName()))
			c.Abort()
			return
		}

		// 将管理员信息存储到上下文中
		c.Set("admin_user", adminUser)
		c.Set("admin_level", adminUser.AdminLevel)
		c.Next()
	}
}

// AuditMiddleware 审计中间件
func (m *AdminPermissionMiddleware) AuditMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 只对管理员API进行审计
		if !strings.HasPrefix(c.Request.URL.Path, "/api/v1/admin") {
			c.Next()
			return
		}

		userID, exists := GetCurrentUserID(c)
		if !exists {
			c.Next()
			return
		}

		start := time.Now()

		// 记录请求体（对于修改操作）
		var requestBody interface{}
		if c.Request.Method != "GET" && c.Request.Method != "DELETE" {
			if c.Request.ContentLength > 0 {
				bodyBytes, _ := c.GetRawData()
				if len(bodyBytes) > 0 {
					json.Unmarshal(bodyBytes, &requestBody)
					// 重新设置请求体供后续处理使用
					c.Request.Body = io.NopCloser(strings.NewReader(string(bodyBytes)))
				}
			}
		}

		// 记录修改前的数据（对于PUT/PATCH操作）
		var oldValues interface{}
		if (c.Request.Method == "PUT" || c.Request.Method == "PATCH") && c.Param("id") != "" {
			oldValues = m.captureOldValues(c)
		}

		c.Next()

		// 记录审计日志
		go m.logAuditAsync(c, userID, start, requestBody, oldValues)
	}
}

// logAuditAsync 异步记录审计日志
func (m *AdminPermissionMiddleware) logAuditAsync(c *gin.Context, userID uint, start time.Time, requestBody, oldValues interface{}) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	duration := time.Since(start).Milliseconds()
	status := admin.AuditStatusSuccess
	errorMsg := ""

	// 根据HTTP状态码确定操作状态
	if c.Writer.Status() >= 400 {
		status = admin.AuditStatusFailed
		if c.Writer.Status() >= 500 {
			errorMsg = "Internal server error"
		} else {
			errorMsg = "Client error"
		}
	}

	// 确定操作类型
	action := m.getActionFromRequest(c)
	resource := m.getResourceFromPath(c.Request.URL.Path)
	resourceID := c.Param("id")

	// 记录审计日志
	req := &ports.LogActionRequest{
		AdminUserID: userID,
		Action:      action,
		Resource:    resource,
		ResourceID:  resourceID,
		Method:      c.Request.Method,
		Path:        c.Request.URL.Path,
		IPAddress:   c.ClientIP(),
		UserAgent:   c.Request.UserAgent(),
		OldValues:   oldValues,
		NewValues:   requestBody,
		Status:      status,
		ErrorMsg:    errorMsg,
		Duration:    duration,
	}

	if err := m.adminAuditService.LogAction(ctx, req); err != nil {
		// 审计日志记录失败，记录到应用日志中
		fmt.Printf("Failed to log audit: %v\n", err)
	}
}

// getActionFromRequest 从请求中获取操作类型
func (m *AdminPermissionMiddleware) getActionFromRequest(c *gin.Context) string {
	method := c.Request.Method
	path := c.Request.URL.Path

	switch method {
	case "POST":
		if strings.Contains(path, "/batch-") {
			return "batch_create"
		}
		return "create"
	case "GET":
		if c.Param("id") != "" {
			return "view"
		}
		return "list"
	case "PUT":
		return "update"
	case "PATCH":
		return "partial_update"
	case "DELETE":
		return "delete"
	default:
		return strings.ToLower(method)
	}
}

// getResourceFromPath 从路径中获取资源类型
func (m *AdminPermissionMiddleware) getResourceFromPath(path string) string {
	// 移除 /api/v1/admin/ 前缀
	path = strings.TrimPrefix(path, "/api/v1/admin/")
	
	// 获取第一个路径段作为资源类型
	parts := strings.Split(path, "/")
	if len(parts) > 0 {
		return parts[0]
	}
	
	return "unknown"
}

// captureOldValues 捕获修改前的数据
func (m *AdminPermissionMiddleware) captureOldValues(c *gin.Context) interface{} {
	// 这里可以根据不同的资源类型实现不同的数据捕获逻辑
	// 为了简化，这里返回nil，实际实现中可以根据需要查询数据库
	return nil
}

// GetCurrentAdminUser 获取当前管理员用户信息的辅助函数
func GetCurrentAdminUser(c *gin.Context) (*admin.AdminUser, bool) {
	adminUserVal, exists := c.Get("admin_user")
	if !exists {
		return nil, false
	}

	adminUser, ok := adminUserVal.(*admin.AdminUser)
	return adminUser, ok
}

// GetCurrentAdminLevel 获取当前管理员级别的辅助函数
func GetCurrentAdminLevel(c *gin.Context) (admin.AdminLevel, bool) {
	levelVal, exists := c.Get("admin_level")
	if !exists {
		return 0, false
	}

	level, ok := levelVal.(admin.AdminLevel)
	return level, ok
}

// GetRequiredPermission 获取当前请求所需权限的辅助函数
func GetRequiredPermission(c *gin.Context) (string, bool) {
	permissionVal, exists := c.Get("required_permission")
	if !exists {
		return "", false
	}

	permission, ok := permissionVal.(string)
	return permission, ok
}

// GetSportTypeID 获取当前运动类型ID的辅助函数
func GetSportTypeID(c *gin.Context) (uint, bool) {
	sportTypeIDVal, exists := c.Get("sport_type_id")
	if !exists {
		return 0, false
	}

	sportTypeID, ok := sportTypeIDVal.(uint)
	return sportTypeID, ok
}