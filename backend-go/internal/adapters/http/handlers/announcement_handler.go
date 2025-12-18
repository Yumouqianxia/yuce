package handlers

import (
	"net/http"
	"strconv"
	"time"

	"backend-go/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Announcement 公告模型
type Announcement struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Title     string    `json:"title" gorm:"size:200;not null"`
	Content   string    `json:"content" gorm:"type:text;not null"`
	Type      string    `json:"type" gorm:"size:20;default:'info'"` // info, warning, success, error
	IsActive  bool      `json:"is_active" gorm:"default:true"`
	Priority  int       `json:"priority" gorm:"default:0"` // 优先级，数字越大越靠前
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName 指定表名
func (Announcement) TableName() string {
	return "announcements"
}

// AnnouncementHandler 公告处理器
type AnnouncementHandler struct {
	db     *gorm.DB
	logger *logrus.Logger
}

// NewAnnouncementHandler 创建公告处理器
func NewAnnouncementHandler(db *gorm.DB, logger *logrus.Logger) *AnnouncementHandler {
	return &AnnouncementHandler{
		db:     db,
		logger: logger,
	}
}

// GetLatestAnnouncement 获取最新的激活公告（按优先级+时间）
func (h *AnnouncementHandler) GetLatestAnnouncement(c *gin.Context) {
	var announcement Announcement
	if err := h.db.
		Where("is_active = ?", true).
		Order("priority DESC, created_at DESC").
		Limit(1).
		First(&announcement).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Success(c, http.StatusOK, "No announcement", gin.H{"announcement": nil})
			return
		}
		h.logger.WithError(err).Error("获取最新公告失败")
		response.Error(c, http.StatusInternalServerError, "Failed to get latest announcement", err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Latest announcement retrieved", announcement)
}

// ListAnnouncements 获取公告列表
func (h *AnnouncementHandler) ListAnnouncements(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	isActive := c.Query("is_active")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	var announcements []Announcement
	var total int64

	query := h.db.Model(&Announcement{})
	
	// 过滤激活状态
	if isActive != "" {
		active := isActive == "true"
		query = query.Where("is_active = ?", active)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		h.logger.WithError(err).Error("获取公告总数失败")
		response.Error(c, http.StatusInternalServerError, "Failed to get announcement count", err.Error())
		return
	}

	// 获取公告列表，按优先级和创建时间排序
	if err := query.Order("priority DESC, created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&announcements).Error; err != nil {
		h.logger.WithError(err).Error("获取公告列表失败")
		response.Error(c, http.StatusInternalServerError, "Failed to get announcements", err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Announcements retrieved successfully", gin.H{
		"announcements": announcements,
		"total":         total,
		"page":          page,
		"page_size":     pageSize,
	})
}

// GetAnnouncement 获取公告详情
func (h *AnnouncementHandler) GetAnnouncement(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid announcement ID", err.Error())
		return
	}

	var announcement Announcement
	if err := h.db.First(&announcement, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Error(c, http.StatusNotFound, "Announcement not found", "")
			return
		}
		h.logger.WithError(err).Error("获取公告详情失败")
		response.Error(c, http.StatusInternalServerError, "Failed to get announcement", err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Announcement retrieved successfully", announcement)
}

// CreateAnnouncement 创建公告
func (h *AnnouncementHandler) CreateAnnouncement(c *gin.Context) {
	var req struct {
		Title    string `json:"title" binding:"required"`
		Content  string `json:"content" binding:"required"`
		Type     string `json:"type"`
		IsActive *bool  `json:"is_active"`
		Priority *int   `json:"priority"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request data", err.Error())
		return
	}

	announcement := Announcement{
		Title:   req.Title,
		Content: req.Content,
		Type:    req.Type,
	}

	if announcement.Type == "" {
		announcement.Type = "info"
	}

	if req.IsActive != nil {
		announcement.IsActive = *req.IsActive
	} else {
		announcement.IsActive = true
	}

	if req.Priority != nil {
		announcement.Priority = *req.Priority
	}

	if err := h.db.Create(&announcement).Error; err != nil {
		h.logger.WithError(err).Error("创建公告失败")
		response.Error(c, http.StatusInternalServerError, "Failed to create announcement", err.Error())
		return
	}

	response.Success(c, http.StatusCreated, "Announcement created successfully", announcement)
}

// UpdateAnnouncement 更新公告
func (h *AnnouncementHandler) UpdateAnnouncement(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid announcement ID", err.Error())
		return
	}

	var req struct {
		Title    *string `json:"title"`
		Content  *string `json:"content"`
		Type     *string `json:"type"`
		IsActive *bool   `json:"is_active"`
		Priority *int    `json:"priority"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request data", err.Error())
		return
	}

	var announcement Announcement
	if err := h.db.First(&announcement, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Error(c, http.StatusNotFound, "Announcement not found", "")
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to get announcement", err.Error())
		return
	}

	// 更新字段
	updates := make(map[string]interface{})
	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.Content != nil {
		updates["content"] = *req.Content
	}
	if req.Type != nil {
		updates["type"] = *req.Type
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}
	if req.Priority != nil {
		updates["priority"] = *req.Priority
	}

	if err := h.db.Model(&announcement).Updates(updates).Error; err != nil {
		h.logger.WithError(err).Error("更新公告失败")
		response.Error(c, http.StatusInternalServerError, "Failed to update announcement", err.Error())
		return
	}

	// 重新获取更新后的公告
	if err := h.db.First(&announcement, id).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get updated announcement", err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Announcement updated successfully", announcement)
}

// DeleteAnnouncement 删除公告
func (h *AnnouncementHandler) DeleteAnnouncement(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid announcement ID", err.Error())
		return
	}

	if err := h.db.Delete(&Announcement{}, id).Error; err != nil {
		h.logger.WithError(err).Error("删除公告失败")
		response.Error(c, http.StatusInternalServerError, "Failed to delete announcement", err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Announcement deleted successfully", nil)
}
