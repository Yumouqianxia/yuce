package admin

import (
	"net/http"
	"strconv"

	"backend-go/internal/core/ports"
	"backend-go/internal/core/types"
	"backend-go/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// ScoringRuleHandler 积分规则管理处理器
type ScoringRuleHandler struct {
	scoringRuleService ports.ScoringRuleService
	logger             *logrus.Logger
}

// NewScoringRuleHandler 创建积分规则处理器实例
func NewScoringRuleHandler(scoringRuleService ports.ScoringRuleService, logger *logrus.Logger) *ScoringRuleHandler {
	return &ScoringRuleHandler{
		scoringRuleService: scoringRuleService,
		logger:             logger,
	}
}

// CreateScoringRule 创建积分规则
// @Summary 创建积分规则
// @Description 为指定运动类型创建积分规则
// @Tags admin-scoring-rules
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body ports.CreateScoringRuleRequest true "创建积分规则请求"
// @Success 201 {object} response.Response{data=sport.ScoringRule}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/admin/scoring-rules [post]
func (h *ScoringRuleHandler) CreateScoringRule(c *gin.Context) {
	var req ports.CreateScoringRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Invalid request body")
		response.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	rule, err := h.scoringRuleService.CreateScoringRule(c.Request.Context(), &req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create scoring rule")
		response.Error(c, http.StatusInternalServerError, "Failed to create scoring rule", err.Error())
		return
	}

	response.Success(c, http.StatusCreated, "Scoring rule created successfully", rule)
}

// GetScoringRule 获取积分规则详情
// @Summary 获取积分规则详情
// @Description 根据ID获取积分规则详情
// @Tags admin-scoring-rules
// @Security BearerAuth
// @Produce json
// @Param id path int true "积分规则ID"
// @Success 200 {object} response.Response{data=sport.ScoringRule}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/admin/scoring-rules/{id} [get]
func (h *ScoringRuleHandler) GetScoringRule(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid scoring rule ID", err.Error())
		return
	}

	rule, err := h.scoringRuleService.GetScoringRule(c.Request.Context(), uint(id))
	if err != nil {
		h.logger.WithError(err).WithField("id", id).Error("Failed to get scoring rule")
		response.Error(c, http.StatusNotFound, "Scoring rule not found", err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Scoring rule retrieved successfully", rule)
}

// UpdateScoringRule 更新积分规则
// @Summary 更新积分规则
// @Description 更新积分规则信息
// @Tags admin-scoring-rules
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "积分规则ID"
// @Param request body ports.UpdateScoringRuleRequest true "更新积分规则请求"
// @Success 200 {object} response.Response{data=sport.ScoringRule}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/admin/scoring-rules/{id} [put]
func (h *ScoringRuleHandler) UpdateScoringRule(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid scoring rule ID", err.Error())
		return
	}

	var req ports.UpdateScoringRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Invalid request body")
		response.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	rule, err := h.scoringRuleService.UpdateScoringRule(c.Request.Context(), uint(id), &req)
	if err != nil {
		h.logger.WithError(err).WithField("id", id).Error("Failed to update scoring rule")
		response.Error(c, http.StatusInternalServerError, "Failed to update scoring rule", err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Scoring rule updated successfully", rule)
}

// DeleteScoringRule 删除积分规则
// @Summary 删除积分规则
// @Description 删除积分规则（仅限非激活状态）
// @Tags admin-scoring-rules
// @Security BearerAuth
// @Produce json
// @Param id path int true "积分规则ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/admin/scoring-rules/{id} [delete]
func (h *ScoringRuleHandler) DeleteScoringRule(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid scoring rule ID", err.Error())
		return
	}

	err = h.scoringRuleService.DeleteScoringRule(c.Request.Context(), uint(id))
	if err != nil {
		h.logger.WithError(err).WithField("id", id).Error("Failed to delete scoring rule")
		response.Error(c, http.StatusInternalServerError, "Failed to delete scoring rule", err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Scoring rule deleted successfully", nil)
}

// ListScoringRules 获取积分规则列表
// @Summary 获取积分规则列表
// @Description 获取积分规则列表，支持分页和过滤
// @Tags admin-scoring-rules
// @Security BearerAuth
// @Produce json
// @Param sport_type_id query int false "运动类型ID"
// @Param is_active query bool false "是否激活"
// @Param order_by query string false "排序字段" Enums(name, created_at)
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} response.Response{data=ports.ListScoringRulesResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Router /api/v1/admin/scoring-rules [get]
func (h *ScoringRuleHandler) ListScoringRules(c *gin.Context) {
	var req ports.ListScoringRulesRequest

	// 解析查询参数
	if sportTypeIDStr := c.Query("sport_type_id"); sportTypeIDStr != "" {
		sportTypeID, err := strconv.ParseUint(sportTypeIDStr, 10, 32)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "Invalid sport_type_id parameter", err.Error())
			return
		}
		sportTypeIDUint := uint(sportTypeID)
		req.SportTypeID = &sportTypeIDUint
	}

	if isActiveStr := c.Query("is_active"); isActiveStr != "" {
		isActive, err := strconv.ParseBool(isActiveStr)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "Invalid is_active parameter", err.Error())
			return
		}
		req.IsActive = &isActive
	}

	req.OrderBy = c.Query("order_by")

	if pageStr := c.Query("page"); pageStr != "" {
		page, err := strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			response.Error(c, http.StatusBadRequest, "Invalid page parameter", err.Error())
			return
		}
		req.Page = page
	} else {
		req.Page = 1
	}

	if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
		pageSize, err := strconv.Atoi(pageSizeStr)
		if err != nil || pageSize < 1 || pageSize > 100 {
			response.Error(c, http.StatusBadRequest, "Invalid page_size parameter", err.Error())
			return
		}
		req.PageSize = pageSize
	} else {
		req.PageSize = 20
	}

	result, err := h.scoringRuleService.ListScoringRules(c.Request.Context(), &req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list scoring rules")
		response.Error(c, http.StatusInternalServerError, "Failed to list scoring rules", err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Scoring rules retrieved successfully", result)
}

// GetActiveScoringRule 获取激活的积分规则
// @Summary 获取激活的积分规则
// @Description 获取指定运动类型的激活积分规则
// @Tags admin-scoring-rules
// @Security BearerAuth
// @Produce json
// @Param sport_type_id path int true "运动类型ID"
// @Success 200 {object} response.Response{data=sport.ScoringRule}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/admin/sport-types/{sport_type_id}/active-scoring-rule [get]
func (h *ScoringRuleHandler) GetActiveScoringRule(c *gin.Context) {
	sportTypeID, err := strconv.ParseUint(c.Param("sport_type_id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid sport type ID", err.Error())
		return
	}

	rule, err := h.scoringRuleService.GetActiveScoringRule(c.Request.Context(), uint(sportTypeID))
	if err != nil {
		h.logger.WithError(err).WithField("sport_type_id", sportTypeID).Error("Failed to get active scoring rule")
		response.Error(c, http.StatusNotFound, "Active scoring rule not found", err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Active scoring rule retrieved successfully", rule)
}

// SetActiveScoringRule 设置激活的积分规则
// @Summary 设置激活的积分规则
// @Description 设置指定积分规则为激活状态
// @Tags admin-scoring-rules
// @Security BearerAuth
// @Produce json
// @Param id path int true "积分规则ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/admin/scoring-rules/{id}/activate [post]
func (h *ScoringRuleHandler) SetActiveScoringRule(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid scoring rule ID", err.Error())
		return
	}

	err = h.scoringRuleService.SetActiveScoringRule(c.Request.Context(), uint(id))
	if err != nil {
		h.logger.WithError(err).WithField("id", id).Error("Failed to set scoring rule as active")
		response.Error(c, http.StatusInternalServerError, "Failed to set scoring rule as active", err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Scoring rule activated successfully", nil)
}

// PreviewScore 预览积分计算
// @Summary 预览积分计算
// @Description 根据积分规则预览积分计算结果
// @Tags admin-scoring-rules
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body services.PreviewScoreRequest true "积分预览请求"
// @Success 200 {object} response.Response{data=services.ScoreBreakdown}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Router /api/v1/admin/scoring-rules/preview [post]
func (h *ScoringRuleHandler) PreviewScore(c *gin.Context) {
	var req types.PreviewScoreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Invalid request body")
		response.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	breakdown, err := h.scoringRuleService.PreviewScore(c.Request.Context(), &req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to preview score")
		response.Error(c, http.StatusInternalServerError, "Failed to preview score", err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Score preview calculated successfully", breakdown)
}

// RecalculateScores 批量重算积分
// @Summary 批量重算积分
// @Description 使用新的积分规则批量重算指定运动类型的积分
// @Tags admin-scoring-rules
// @Security BearerAuth
// @Produce json
// @Param sport_type_id path int true "运动类型ID"
// @Param rule_id path int true "积分规则ID"
// @Success 200 {object} response.Response{data=ports.RecalculateResult}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Router /api/v1/admin/sport-types/{sport_type_id}/scoring-rules/{rule_id}/recalculate [post]
func (h *ScoringRuleHandler) RecalculateScores(c *gin.Context) {
	sportTypeID, err := strconv.ParseUint(c.Param("sport_type_id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid sport type ID", err.Error())
		return
	}

	ruleID, err := strconv.ParseUint(c.Param("rule_id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid scoring rule ID", err.Error())
		return
	}

	result, err := h.scoringRuleService.RecalculateScores(c.Request.Context(), uint(sportTypeID), uint(ruleID))
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"sport_type_id": sportTypeID,
			"rule_id":       ruleID,
		}).Error("Failed to recalculate scores")
		response.Error(c, http.StatusInternalServerError, "Failed to recalculate scores", err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Scores recalculated successfully", result)
}
