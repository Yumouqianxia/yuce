package admin

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"backend-go/internal/core/domain/admin"
	"backend-go/internal/core/ports"
	"backend-go/pkg/middleware"
	"backend-go/pkg/response"
)

// AdminHandler 管理员管理处理器
type AdminHandler struct {
	adminService      ports.AdminService
	adminAuditService ports.AdminAuditService
	logger            *logrus.Logger
}

// NewAdminHandler 创建管理员管理处理器
func NewAdminHandler(
	adminService ports.AdminService,
	adminAuditService ports.AdminAuditService,
	logger *logrus.Logger,
) *AdminHandler {
	return &AdminHandler{
		adminService:      adminService,
		adminAuditService: adminAuditService,
		logger:            logger,
	}
}

// CreateAdmin 创建管理员
func (h *AdminHandler) CreateAdmin(c *gin.Context) {
	var req ports.CreateAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request data", err.Error())
		return
	}

	// 检查当前用户权限
	currentUserID, _ := middleware.GetCurrentUserID(c)
	currentAdmin, err := h.adminService.GetAdmin(c.Request.Context(), currentUserID)
	if err != nil {
		response.Error(c, http.StatusForbidden, "Only admins can create other admins", err.Error())
		return
	}

	// 只有系统管理员及以上可以创建管理员
	if !currentAdmin.IsSystemAdmin() {
		response.Error(c, http.StatusForbidden, "System admin level required", "INSUFFICIENT_LEVEL")
		return
	}

	// 超级管理员才能创建系统管理员或超级管理员
	if req.AdminLevel >= admin.AdminLevelSystem && !currentAdmin.IsSuperAdmin() {
		response.Error(c, http.StatusForbidden, "Super admin level required to create system/super admins", "INSUFFICIENT_LEVEL")
		return
	}

	adminUser, err := h.adminService.CreateAdmin(c.Request.Context(), &req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create admin")
		response.Error(c, http.StatusInternalServerError, "Failed to create admin", err.Error())
		return
	}

	response.Success(c, http.StatusCreated, "Admin created successfully", adminUser)
}

// UpdateAdmin 更新管理员
func (h *AdminHandler) UpdateAdmin(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid user ID", err.Error())
		return
	}

	var req ports.UpdateAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request data", err.Error())
		return
	}

	// 检查当前用户权限
	currentUserID, _ := middleware.GetCurrentUserID(c)
	currentAdmin, err := h.adminService.GetAdmin(c.Request.Context(), currentUserID)
	if err != nil {
		response.Error(c, http.StatusForbidden, "Only admins can update other admins", err.Error())
		return
	}

	// 获取目标管理员信息
	targetAdmin, err := h.adminService.GetAdmin(c.Request.Context(), uint(userID))
	if err != nil {
		response.Error(c, http.StatusNotFound, "Admin not found", err.Error())
		return
	}

	// 权限检查：只有系统管理员及以上可以更新管理员
	if !currentAdmin.IsSystemAdmin() {
		response.Error(c, http.StatusForbidden, "System admin level required", "INSUFFICIENT_LEVEL")
		return
	}

	// 权限检查：不能修改比自己级别高的管理员
	if targetAdmin.AdminLevel >= currentAdmin.AdminLevel && currentUserID != uint(userID) {
		response.Error(c, http.StatusForbidden, "Cannot modify admin with equal or higher level", "INSUFFICIENT_LEVEL")
		return
	}

	// 权限检查：只有超级管理员可以设置系统管理员或超级管理员级别
	if req.AdminLevel != nil && *req.AdminLevel >= admin.AdminLevelSystem && !currentAdmin.IsSuperAdmin() {
		response.Error(c, http.StatusForbidden, "Super admin level required to set system/super admin level", "INSUFFICIENT_LEVEL")
		return
	}

	adminUser, err := h.adminService.UpdateAdmin(c.Request.Context(), uint(userID), &req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to update admin")
		response.Error(c, http.StatusInternalServerError, "Failed to update admin", err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Admin updated successfully", adminUser)
}

// DeleteAdmin 删除管理员
func (h *AdminHandler) DeleteAdmin(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid user ID", err.Error())
		return
	}

	// 检查当前用户权限
	currentUserID, _ := middleware.GetCurrentUserID(c)
	currentAdmin, err := h.adminService.GetAdmin(c.Request.Context(), currentUserID)
	if err != nil {
		response.Error(c, http.StatusForbidden, "Only admins can delete other admins", err.Error())
		return
	}

	// 不能删除自己
	if currentUserID == uint(userID) {
		response.Error(c, http.StatusBadRequest, "Cannot delete your own admin account", "CANNOT_DELETE_SELF")
		return
	}

	// 获取目标管理员信息
	targetAdmin, err := h.adminService.GetAdmin(c.Request.Context(), uint(userID))
	if err != nil {
		response.Error(c, http.StatusNotFound, "Admin not found", err.Error())
		return
	}

	// 权限检查：只有系统管理员及以上可以删除管理员
	if !currentAdmin.IsSystemAdmin() {
		response.Error(c, http.StatusForbidden, "System admin level required", "INSUFFICIENT_LEVEL")
		return
	}

	// 权限检查：不能删除比自己级别高的管理员
	if targetAdmin.AdminLevel >= currentAdmin.AdminLevel {
		response.Error(c, http.StatusForbidden, "Cannot delete admin with equal or higher level", "INSUFFICIENT_LEVEL")
		return
	}

	if err := h.adminService.DeleteAdmin(c.Request.Context(), uint(userID)); err != nil {
		h.logger.WithError(err).Error("Failed to delete admin")
		response.Error(c, http.StatusInternalServerError, "Failed to delete admin", err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Admin deleted successfully", nil)
}

// GetAdmin 获取管理员信息
func (h *AdminHandler) GetAdmin(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid user ID", err.Error())
		return
	}

	adminUser, err := h.adminService.GetAdmin(c.Request.Context(), uint(userID))
	if err != nil {
		response.Error(c, http.StatusNotFound, "Admin not found", err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Admin retrieved successfully", adminUser)
}

// ListAdmins 获取管理员列表
func (h *AdminHandler) ListAdmins(c *gin.Context) {
	var req ports.ListAdminsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request parameters", err.Error())
		return
	}

	result, err := h.adminService.ListAdmins(c.Request.Context(), &req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list admins")
		response.Error(c, http.StatusInternalServerError, "Failed to list admins", err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Admins retrieved successfully", result)
}

// GrantPermissions 授予权限
func (h *AdminHandler) GrantPermissions(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid user ID", err.Error())
		return
	}

	var req struct {
		Permissions []string `json:"permissions" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request data", err.Error())
		return
	}

	// 检查权限
	if err := h.checkPermissionManagementAccess(c, uint(userID)); err != nil {
		response.Error(c, http.StatusForbidden, err.Error(), "PERMISSION_DENIED")
		return
	}

	if err := h.adminService.GrantPermissions(c.Request.Context(), uint(userID), req.Permissions); err != nil {
		h.logger.WithError(err).Error("Failed to grant permissions")
		response.Error(c, http.StatusInternalServerError, "Failed to grant permissions", err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Permissions granted successfully", nil)
}

// RevokePermissions 撤销权限
func (h *AdminHandler) RevokePermissions(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid user ID", err.Error())
		return
	}

	var req struct {
		Permissions []string `json:"permissions" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request data", err.Error())
		return
	}

	// 检查权限
	if err := h.checkPermissionManagementAccess(c, uint(userID)); err != nil {
		response.Error(c, http.StatusForbidden, err.Error(), "PERMISSION_DENIED")
		return
	}

	if err := h.adminService.RevokePermissions(c.Request.Context(), uint(userID), req.Permissions); err != nil {
		h.logger.WithError(err).Error("Failed to revoke permissions")
		response.Error(c, http.StatusInternalServerError, "Failed to revoke permissions", err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Permissions revoked successfully", nil)
}

// GrantSportAccess 授予运动类型访问权限
func (h *AdminHandler) GrantSportAccess(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid user ID", err.Error())
		return
	}

	var req struct {
		SportTypes []uint `json:"sport_types" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request data", err.Error())
		return
	}

	// 检查权限
	if err := h.checkPermissionManagementAccess(c, uint(userID)); err != nil {
		response.Error(c, http.StatusForbidden, err.Error(), "PERMISSION_DENIED")
		return
	}

	if err := h.adminService.GrantSportAccess(c.Request.Context(), uint(userID), req.SportTypes); err != nil {
		h.logger.WithError(err).Error("Failed to grant sport access")
		response.Error(c, http.StatusInternalServerError, "Failed to grant sport access", err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Sport access granted successfully", nil)
}

// RevokeSportAccess 撤销运动类型访问权限
func (h *AdminHandler) RevokeSportAccess(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid user ID", err.Error())
		return
	}

	var req struct {
		SportTypes []uint `json:"sport_types" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request data", err.Error())
		return
	}

	// 检查权限
	if err := h.checkPermissionManagementAccess(c, uint(userID)); err != nil {
		response.Error(c, http.StatusForbidden, err.Error(), "PERMISSION_DENIED")
		return
	}

	if err := h.adminService.RevokeSportAccess(c.Request.Context(), uint(userID), req.SportTypes); err != nil {
		h.logger.WithError(err).Error("Failed to revoke sport access")
		response.Error(c, http.StatusInternalServerError, "Failed to revoke sport access", err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Sport access revoked successfully", nil)
}

// ListPermissions 获取权限列表
func (h *AdminHandler) ListPermissions(c *gin.Context) {
	permissions, err := h.adminService.ListPermissions(c.Request.Context())
	if err != nil {
		h.logger.WithError(err).Error("Failed to list permissions")
		response.Error(c, http.StatusInternalServerError, "Failed to list permissions", err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Permissions retrieved successfully", permissions)
}

// GetUserPermissions 获取用户权限列表
func (h *AdminHandler) GetUserPermissions(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid user ID", err.Error())
		return
	}

	permissions, err := h.adminService.GetUserPermissions(c.Request.Context(), uint(userID))
	if err != nil {
		h.logger.WithError(err).Error("Failed to get user permissions")
		response.Error(c, http.StatusInternalServerError, "Failed to get user permissions", err.Error())
		return
	}

	response.Success(c, http.StatusOK, "User permissions retrieved successfully", map[string]interface{}{
		"user_id":     userID,
		"permissions": permissions,
	})
}

// checkPermissionManagementAccess 检查权限管理访问权限
func (h *AdminHandler) checkPermissionManagementAccess(c *gin.Context, targetUserID uint) error {
	currentUserID, _ := middleware.GetCurrentUserID(c)
	currentAdmin, err := h.adminService.GetAdmin(c.Request.Context(), currentUserID)
	if err != nil {
		return fmt.Errorf("only admins can manage permissions")
	}

	// 获取目标管理员信息
	targetAdmin, err := h.adminService.GetAdmin(c.Request.Context(), targetUserID)
	if err != nil {
		return fmt.Errorf("target admin not found")
	}

	// 权限检查：只有系统管理员及以上可以管理权限
	if !currentAdmin.IsSystemAdmin() {
		return fmt.Errorf("system admin level required")
	}

	// 权限检查：不能管理比自己级别高的管理员权限
	if targetAdmin.AdminLevel >= currentAdmin.AdminLevel && currentUserID != targetUserID {
		return fmt.Errorf("cannot manage permissions for admin with equal or higher level")
	}

	return nil
}
