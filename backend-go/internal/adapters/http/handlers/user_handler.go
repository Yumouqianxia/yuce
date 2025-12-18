package handlers

import (
	"net/http"
	"strconv"

	"backend-go/internal/core/domain/user"
	"backend-go/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// UserHandler 用户管理处理器
type UserHandler struct {
	userService user.Service
	db          *gorm.DB
	logger      *logrus.Logger
}

// NewUserHandler 创建用户处理器
func NewUserHandler(userService user.Service, db *gorm.DB, logger *logrus.Logger) *UserHandler {
	return &UserHandler{
		userService: userService,
		db:          db,
		logger:      logger,
	}
}

// ListUsers 获取用户列表
func (h *UserHandler) ListUsers(c *gin.Context) {
	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	// 从数据库获取用户列表
	var users []user.User
	var total int64

	// 获取总数
	if err := h.db.Model(&user.User{}).Count(&total).Error; err != nil {
		h.logger.WithError(err).Error("获取用户总数失败")
		response.Error(c, http.StatusInternalServerError, "Failed to get user count", err.Error())
		return
	}

	// 获取用户列表
	if err := h.db.Offset(offset).Limit(pageSize).Find(&users).Error; err != nil {
		h.logger.WithError(err).Error("获取用户列表失败")
		response.Error(c, http.StatusInternalServerError, "Failed to get users", err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Users retrieved successfully", gin.H{
		"users": users,
		"total": total,
		"page":  page,
		"page_size": pageSize,
	})
}

// GetUser 获取用户详情
func (h *UserHandler) GetUser(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid user ID", err.Error())
		return
	}

	foundUser, err := h.userService.GetProfile(c.Request.Context(), uint(userID))
	if err != nil {
		h.logger.WithError(err).Error("获取用户详情失败")
		response.Error(c, http.StatusNotFound, "User not found", err.Error())
		return
	}

	response.Success(c, http.StatusOK, "User retrieved successfully", foundUser)
}

// UpdateUser 更新用户信息
func (h *UserHandler) UpdateUser(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid user ID", err.Error())
		return
	}

	var updateData struct {
		Nickname *string     `json:"nickname"`
		Email    *string     `json:"email"`
		Role     *user.UserRole `json:"role"`
		Points   *int        `json:"points"`
	}

	if err := c.ShouldBindJSON(&updateData); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request data", err.Error())
		return
	}

	// 获取用户
	var foundUser user.User
	if err := h.db.First(&foundUser, userID).Error; err != nil {
		response.Error(c, http.StatusNotFound, "User not found", err.Error())
		return
	}

	// 更新字段
	updates := make(map[string]interface{})
	if updateData.Nickname != nil {
		updates["nickname"] = *updateData.Nickname
	}
	if updateData.Email != nil {
		updates["email"] = *updateData.Email
	}
	if updateData.Role != nil {
		updates["role"] = *updateData.Role
	}
	if updateData.Points != nil {
		updates["points"] = *updateData.Points
	}

	if err := h.db.Model(&foundUser).Updates(updates).Error; err != nil{
		h.logger.WithError(err).Error("更新用户失败")
		response.Error(c, http.StatusInternalServerError, "Failed to update user", err.Error())
		return
	}

	// 重新获取更新后的用户
	if err := h.db.First(&foundUser, userID).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get updated user", err.Error())
		return
	}

	response.Success(c, http.StatusOK, "User updated successfully", foundUser)
}

// DeleteUser 删除用户
func (h *UserHandler) DeleteUser(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid user ID", err.Error())
		return
	}

	// 删除用户
	if err := h.db.Delete(&user.User{}, userID).Error; err != nil{
		h.logger.WithError(err).Error("删除用户失败")
		response.Error(c, http.StatusInternalServerError, "Failed to delete user", err.Error())
		return
	}

	response.Success(c, http.StatusOK, "User deleted successfully", nil)
}

// ResetPassword 重置/修改用户密码
func (h *UserHandler) ResetPassword(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid user ID", err.Error())
		return
	}

	var req struct {
		Password string `json:"password" binding:"required,min=6"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request data", err.Error())
		return
	}

	if err := h.userService.ChangePassword(c.Request.Context(), uint(userID), req.Password); err != nil {
		h.logger.WithError(err).Error("重置用户密码失败")
		response.Error(c, http.StatusInternalServerError, "Failed to reset password", err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Password reset successfully", gin.H{"user_id": userID})
}