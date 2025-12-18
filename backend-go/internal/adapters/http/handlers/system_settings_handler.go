package handlers

import (
	"net/http"
	"time"

	"backend-go/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// SystemSettings 系统设置模型
type SystemSettings struct {
	ID                     uint      `json:"id" gorm:"primaryKey"`
	SiteName               string    `json:"siteName" gorm:"size:200;not null;default:'预测系统'"`
	AllowRegistration      bool      `json:"allowRegistration" gorm:"default:true"`
	EnableLeaderboard      bool      `json:"enableLeaderboard" gorm:"default:true"`
	PredictionDeadlineHours int      `json:"predictionDeadlineHours" gorm:"default:1"`
	CreatedAt              time.Time `json:"createdAt"`
	UpdatedAt              time.Time `json:"updatedAt"`
}

// TableName 指定系统设置表名
func (SystemSettings) TableName() string {
	return "system_settings"
}

// SystemSettingsHandler 系统设置处理器
type SystemSettingsHandler struct {
	db     *gorm.DB
	logger *logrus.Logger
}

// NewSystemSettingsHandler 创建处理器
func NewSystemSettingsHandler(db *gorm.DB, logger *logrus.Logger) *SystemSettingsHandler {
	return &SystemSettingsHandler{
		db:     db,
		logger: logger,
	}
}

// getOrCreateSettings 获取或创建默认设置
func (h *SystemSettingsHandler) getOrCreateSettings() (*SystemSettings, error) {
	var settings SystemSettings
	if err := h.db.First(&settings).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			settings = SystemSettings{
				SiteName:               "预测系统",
				AllowRegistration:      true,
				EnableLeaderboard:      true,
				PredictionDeadlineHours: 1,
			}
			if createErr := h.db.Create(&settings).Error; createErr != nil {
				return nil, createErr
			}
		} else {
			return nil, err
		}
	}
	return &settings, nil
}

// GetSettings 获取系统设置
func (h *SystemSettingsHandler) GetSettings(c *gin.Context) {
	settings, err := h.getOrCreateSettings()
	if err != nil {
		h.logger.WithError(err).Error("获取系统设置失败")
		response.Error(c, http.StatusInternalServerError, "Failed to get system settings", err.Error())
		return
	}

	response.Success(c, http.StatusOK, "System settings retrieved successfully", settings)
}

// UpdateSettings 更新系统设置
func (h *SystemSettingsHandler) UpdateSettings(c *gin.Context) {
	var req struct {
		SiteName               *string `json:"siteName"`
		AllowRegistration      *bool   `json:"allowRegistration"`
		EnableLeaderboard      *bool   `json:"enableLeaderboard"`
		PredictionDeadlineHours *int   `json:"predictionDeadlineHours"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request data", err.Error())
		return
	}

	settings, err := h.getOrCreateSettings()
	if err != nil {
		h.logger.WithError(err).Error("获取系统设置失败")
		response.Error(c, http.StatusInternalServerError, "Failed to get system settings", err.Error())
		return
	}

	updates := make(map[string]interface{})
	if req.SiteName != nil {
		updates["site_name"] = *req.SiteName
	}
	if req.AllowRegistration != nil {
		updates["allow_registration"] = *req.AllowRegistration
	}
	if req.EnableLeaderboard != nil {
		updates["enable_leaderboard"] = *req.EnableLeaderboard
	}
	if req.PredictionDeadlineHours != nil {
		updates["prediction_deadline_hours"] = *req.PredictionDeadlineHours
	}

	if len(updates) == 0 {
		response.Success(c, http.StatusOK, "System settings updated successfully", settings)
		return
	}

	if err := h.db.Model(settings).Updates(updates).Error; err != nil {
		h.logger.WithError(err).Error("更新系统设置失败")
		response.Error(c, http.StatusInternalServerError, "Failed to update system settings", err.Error())
		return
	}

	// 重新获取更新后的设置
	if err := h.db.First(settings).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to load updated settings", err.Error())
		return
	}

	response.Success(c, http.StatusOK, "System settings updated successfully", settings)
}

