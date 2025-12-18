package admin

import (
	"net/http"
	"strconv"

	"backend-go/internal/core/ports"
	"backend-go/internal/shared/response"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// SportTypeHandler 运动类型管理处理器
type SportTypeHandler struct {
	sportTypeService ports.SportTypeService
	logger           *logrus.Logger
}

// NewSportTypeHandler 创建运动类型处理器实例
func NewSportTypeHandler(sportTypeService ports.SportTypeService, logger *logrus.Logger) *SportTypeHandler {
	return &SportTypeHandler{
		sportTypeService: sportTypeService,
		logger:           logger,
	}
}

// CreateSportType 创建运动类型
// @Summary 创建运动类型
// @Description 创建新的运动类型
// @Tags admin-sport-types
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body ports.CreateSportTypeRequest true "创建运动类型请求"
// @Success 201 {object} response.Response{data=sport.SportType}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/admin/sport-types [post]
func (h *SportTypeHandler) CreateSportType(c *gin.Context) {
	var req ports.CreateSportTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Invalid request body")
		response.Error(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err)
		return
	}

	sportType, err := h.sportTypeService.CreateSportType(c.Request.Context(), &req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create sport type")
		response.Error(c, http.StatusInternalServerError, "CREATE_FAILED", "Failed to create sport type", err)
		return
	}

	response.Success(c, http.StatusCreated, "Sport type created successfully", sportType)
}

// GetSportType 获取运动类型详情
// @Summary 获取运动类型详情
// @Description 根据ID获取运动类型详情
// @Tags admin-sport-types
// @Security BearerAuth
// @Produce json
// @Param id path int true "运动类型ID"
// @Success 200 {object} response.Response{data=sport.SportType}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/admin/sport-types/{id} [get]
func (h *SportTypeHandler) GetSportType(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", "Invalid sport type ID", err)
		return
	}

	sportType, err := h.sportTypeService.GetSportType(c.Request.Context(), uint(id))
	if err != nil {
		h.logger.WithError(err).WithField("id", id).Error("Failed to get sport type")
		response.Error(c, http.StatusNotFound, "NOT_FOUND", "Sport type not found", err)
		return
	}

	response.Success(c, http.StatusOK, "Sport type retrieved successfully", sportType)
}

// UpdateSportType 更新运动类型
// @Summary 更新运动类型
// @Description 更新运动类型信息
// @Tags admin-sport-types
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "运动类型ID"
// @Param request body ports.UpdateSportTypeRequest true "更新运动类型请求"
// @Success 200 {object} response.Response{data=sport.SportType}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/admin/sport-types/{id} [put]
func (h *SportTypeHandler) UpdateSportType(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", "Invalid sport type ID", err)
		return
	}

	var req ports.UpdateSportTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Invalid request body")
		response.Error(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err)
		return
	}

	sportType, err := h.sportTypeService.UpdateSportType(c.Request.Context(), uint(id), &req)
	if err != nil {
		h.logger.WithError(err).WithField("id", id).Error("Failed to update sport type")
		response.Error(c, http.StatusInternalServerError, "UPDATE_FAILED", "Failed to update sport type", err)
		return
	}

	response.Success(c, http.StatusOK, "Sport type updated successfully", sportType)
}

// DeleteSportType 删除运动类型
// @Summary 删除运动类型
// @Description 删除运动类型
// @Tags admin-sport-types
// @Security BearerAuth
// @Produce json
// @Param id path int true "运动类型ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/admin/sport-types/{id} [delete]
func (h *SportTypeHandler) DeleteSportType(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", "Invalid sport type ID", err)
		return
	}

	err = h.sportTypeService.DeleteSportType(c.Request.Context(), uint(id))
	if err != nil {
		h.logger.WithError(err).WithField("id", id).Error("Failed to delete sport type")
		response.Error(c, http.StatusInternalServerError, "DELETE_FAILED", "Failed to delete sport type", err)
		return
	}

	response.Success(c, http.StatusOK, "Sport type deleted successfully", nil)
}

// ListSportTypes 获取运动类型列表
// @Summary 获取运动类型列表
// @Description 获取运动类型列表，支持分页和过滤
// @Tags admin-sport-types
// @Security BearerAuth
// @Produce json
// @Param category query string false "运动类别" Enums(esports, traditional)
// @Param is_active query bool false "是否启用"
// @Param order_by query string false "排序字段" Enums(name, code, sort_order, created_at)
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} response.Response{data=ports.ListSportTypesResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Router /api/v1/admin/sport-types [get]
func (h *SportTypeHandler) ListSportTypes(c *gin.Context) {
	var req ports.ListSportTypesRequest

	// 解析查询参数
	req.Category = c.Query("category")
	req.OrderBy = c.Query("order_by")

	if isActiveStr := c.Query("is_active"); isActiveStr != "" {
		isActive, err := strconv.ParseBool(isActiveStr)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "INVALID_PARAM", "Invalid is_active parameter", err)
			return
		}
		req.IsActive = &isActive
	}

	if pageStr := c.Query("page"); pageStr != "" {
		page, err := strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			response.Error(c, http.StatusBadRequest, "INVALID_PARAM", "Invalid page parameter", err)
			return
		}
		req.Page = page
	} else {
		req.Page = 1
	}

	if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
		pageSize, err := strconv.Atoi(pageSizeStr)
		if err != nil || pageSize < 1 || pageSize > 100 {
			response.Error(c, http.StatusBadRequest, "INVALID_PARAM", "Invalid page_size parameter", err)
			return
		}
		req.PageSize = pageSize
	} else {
		req.PageSize = 20
	}

	result, err := h.sportTypeService.ListSportTypes(c.Request.Context(), &req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list sport types")
		response.Error(c, http.StatusInternalServerError, "LIST_FAILED", "Failed to list sport types", err)
		return
	}

	response.Success(c, http.StatusOK, "Sport types retrieved successfully", result)
}

// GetSportConfiguration 获取运动配置
// @Summary 获取运动配置
// @Description 获取指定运动类型的配置
// @Tags admin-sport-types
// @Security BearerAuth
// @Produce json
// @Param id path int true "运动类型ID"
// @Success 200 {object} response.Response{data=sport.SportConfiguration}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/admin/sport-types/{id}/configuration [get]
func (h *SportTypeHandler) GetSportConfiguration(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", "Invalid sport type ID", err)
		return
	}

	config, err := h.sportTypeService.GetSportConfiguration(c.Request.Context(), uint(id))
	if err != nil {
		h.logger.WithError(err).WithField("sport_type_id", id).Error("Failed to get sport configuration")
		response.Error(c, http.StatusNotFound, "NOT_FOUND", "Sport configuration not found", err)
		return
	}

	response.Success(c, http.StatusOK, "Sport configuration retrieved successfully", config)
}

// UpdateSportConfiguration 更新运动配置
// @Summary 更新运动配置
// @Description 更新指定运动类型的配置
// @Tags admin-sport-types
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "运动类型ID"
// @Param request body ports.UpdateSportConfigurationRequest true "更新运动配置请求"
// @Success 200 {object} response.Response{data=sport.SportConfiguration}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/admin/sport-types/{id}/configuration [put]
func (h *SportTypeHandler) UpdateSportConfiguration(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", "Invalid sport type ID", err)
		return
	}

	var req ports.UpdateSportConfigurationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Invalid request body")
		response.Error(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err)
		return
	}

	config, err := h.sportTypeService.UpdateSportConfiguration(c.Request.Context(), uint(id), &req)
	if err != nil {
		h.logger.WithError(err).WithField("sport_type_id", id).Error("Failed to update sport configuration")
		response.Error(c, http.StatusInternalServerError, "UPDATE_FAILED", "Failed to update sport configuration", err)
		return
	}

	response.Success(c, http.StatusOK, "Sport configuration updated successfully", config)
}

// BatchUpdateConfiguration 批量更新配置
// @Summary 批量更新运动配置
// @Description 批量更新多个运动类型的配置
// @Tags admin-sport-types
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body ports.BatchUpdateConfigRequest true "批量更新配置请求"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Router /api/v1/admin/sport-types/batch-config [post]
func (h *SportTypeHandler) BatchUpdateConfiguration(c *gin.Context) {
	var req ports.BatchUpdateConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Invalid request body")
		response.Error(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err)
		return
	}

	err := h.sportTypeService.BatchUpdateConfiguration(c.Request.Context(), &req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to batch update configurations")
		response.Error(c, http.StatusInternalServerError, "BATCH_UPDATE_FAILED", "Failed to batch update configurations", err)
		return
	}

	response.Success(c, http.StatusOK, "Configurations updated successfully", nil)
}

// GetSportTypeStats 获取运动类型统计信息
// @Summary 获取运动类型统计信息
// @Description 获取指定运动类型的统计信息
// @Tags admin-sport-types
// @Security BearerAuth
// @Produce json
// @Param id path int true "运动类型ID"
// @Success 200 {object} response.Response{data=ports.SportTypeStats}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/admin/sport-types/{id}/stats [get]
func (h *SportTypeHandler) GetSportTypeStats(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", "Invalid sport type ID", err)
		return
	}

	stats, err := h.sportTypeService.GetSportTypeStats(c.Request.Context(), uint(id))
	if err != nil {
		h.logger.WithError(err).WithField("sport_type_id", id).Error("Failed to get sport type stats")
		response.Error(c, http.StatusInternalServerError, "STATS_FAILED", "Failed to get sport type stats", err)
		return
	}

	response.Success(c, http.StatusOK, "Sport type stats retrieved successfully", stats)
}