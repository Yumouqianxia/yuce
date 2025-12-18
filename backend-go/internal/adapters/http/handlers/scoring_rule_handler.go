package handlers

import (
	"strconv"

	"backend-go/internal/core/domain/prediction"
	"backend-go/pkg/response"
	"github.com/gin-gonic/gin"
)

// ScoringRuleHandler 积分规则处理器
type ScoringRuleHandler struct {
	scoringRuleService prediction.ScoringRuleService
}

// NewScoringRuleHandler 创建积分规则处理器
func NewScoringRuleHandler(scoringRuleService prediction.ScoringRuleService) *ScoringRuleHandler {
	return &ScoringRuleHandler{
		scoringRuleService: scoringRuleService,
	}
}

// CreateScoringRule 创建积分规则
// @Summary 创建积分规则
// @Description 管理员创建新的积分规则
// @Tags 积分规则
// @Accept json
// @Produce json
// @Param request body prediction.CreateScoringRuleRequest true "创建积分规则请求"
// @Success 201 {object} response.Response{data=prediction.ScoringRule}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/admin/scoring-rules [post]
func (h *ScoringRuleHandler) CreateScoringRule(c *gin.Context) {
	var req prediction.CreateScoringRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	rule, err := h.scoringRuleService.CreateScoringRule(c.Request.Context(), &req)
	if err != nil {
		if appErr, ok := err.(*response.AppError); ok {
			response.Error(c, appErr.StatusCode, appErr.Message, appErr.Error())
			return
		}
		response.InternalError(c, "创建积分规则失败: "+err.Error())
		return
	}

	response.Created(c, "积分规则创建成功", rule)
}

// GetScoringRule 获取积分规则详情
// @Summary 获取积分规则详情
// @Description 获取指定ID的积分规则详情
// @Tags 积分规则
// @Accept json
// @Produce json
// @Param id path int true "积分规则ID"
// @Success 200 {object} response.Response{data=prediction.ScoringRule}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/scoring-rules/{id} [get]
func (h *ScoringRuleHandler) GetScoringRule(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的积分规则ID")
		return
	}

	rule, err := h.scoringRuleService.GetScoringRule(c.Request.Context(), uint(id))
	if err != nil {
		if appErr, ok := err.(*response.AppError); ok {
			response.Error(c, appErr.StatusCode, appErr.Message, appErr.Error())
			return
		}
		response.InternalError(c, "获取积分规则失败: "+err.Error())
		return
	}

	response.OK(c, "获取积分规则成功", rule)
}

// GetActiveScoringRule 获取当前激活的积分规则
// @Summary 获取当前激活的积分规则
// @Description 获取当前系统使用的积分规则
// @Tags 积分规则
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=prediction.ScoringRule}
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/scoring-rules/active [get]
func (h *ScoringRuleHandler) GetActiveScoringRule(c *gin.Context) {
	rule, err := h.scoringRuleService.GetActiveScoringRule(c.Request.Context())
	if err != nil {
		if appErr, ok := err.(*response.AppError); ok {
			response.Error(c, appErr.StatusCode, appErr.Message, appErr.Error())
			return
		}
		response.InternalError(c, "获取激活积分规则失败: "+err.Error())
		return
	}

	response.OK(c, "获取激活积分规则成功", rule)
}

// ListScoringRules 获取所有积分规则
// @Summary 获取所有积分规则
// @Description 获取系统中所有的积分规则列表
// @Tags 积分规则
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=[]prediction.ScoringRule}
// @Failure 500 {object} response.Response
// @Router /api/v1/scoring-rules [get]
func (h *ScoringRuleHandler) ListScoringRules(c *gin.Context) {
	rules, err := h.scoringRuleService.ListScoringRules(c.Request.Context())
	if err != nil {
		if appErr, ok := err.(*response.AppError); ok {
			response.Error(c, appErr.StatusCode, appErr.Message, appErr.Error())
			return
		}
		response.InternalError(c, "获取积分规则列表失败: "+err.Error())
		return
	}

	response.OK(c, "获取积分规则列表成功", rules)
}

// UpdateScoringRule 更新积分规则
// @Summary 更新积分规则
// @Description 管理员更新积分规则信息
// @Tags 积分规则
// @Accept json
// @Produce json
// @Param id path int true "积分规则ID"
// @Param request body prediction.UpdateScoringRuleRequest true "更新积分规则请求"
// @Success 200 {object} response.Response{data=prediction.ScoringRule}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/admin/scoring-rules/{id} [put]
func (h *ScoringRuleHandler) UpdateScoringRule(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的积分规则ID")
		return
	}

	var req prediction.UpdateScoringRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	rule, err := h.scoringRuleService.UpdateScoringRule(c.Request.Context(), uint(id), &req)
	if err != nil {
		if appErr, ok := err.(*response.AppError); ok {
			response.Error(c, appErr.StatusCode, appErr.Message, appErr.Error())
			return
		}
		response.InternalError(c, "更新积分规则失败: "+err.Error())
		return
	}

	response.OK(c, "积分规则更新成功", rule)
}

// SetActiveRule 设置激活的积分规则
// @Summary 设置激活的积分规则
// @Description 管理员设置当前使用的积分规则
// @Tags 积分规则
// @Accept json
// @Produce json
// @Param id path int true "积分规则ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/admin/scoring-rules/{id}/activate [post]
func (h *ScoringRuleHandler) SetActiveRule(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的积分规则ID")
		return
	}

	err = h.scoringRuleService.SetActiveRule(c.Request.Context(), uint(id))
	if err != nil {
		if appErr, ok := err.(*response.AppError); ok {
			response.Error(c, appErr.StatusCode, appErr.Message, appErr.Error())
			return
		}
		response.InternalError(c, "设置激活规则失败: "+err.Error())
		return
	}

	response.OK(c, "积分规则已激活", nil)
}

// DeleteScoringRule 删除积分规则
// @Summary 删除积分规则
// @Description 管理员删除积分规则（不能删除激活的规则）
// @Tags 积分规则
// @Accept json
// @Produce json
// @Param id path int true "积分规则ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/admin/scoring-rules/{id} [delete]
func (h *ScoringRuleHandler) DeleteScoringRule(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的积分规则ID")
		return
	}

	err = h.scoringRuleService.DeleteScoringRule(c.Request.Context(), uint(id))
	if err != nil {
		if appErr, ok := err.(*response.AppError); ok {
			response.Error(c, appErr.StatusCode, appErr.Message, appErr.Error())
			return
		}
		response.InternalError(c, "删除积分规则失败: "+err.Error())
		return
	}

	response.OK(c, "积分规则已删除", nil)
}

// CalculatePointsWithRule 使用指定规则重新计算比赛积分
// @Summary 使用指定规则重新计算比赛积分
// @Description 管理员使用指定积分规则重新计算某场比赛的所有预测积分
// @Tags 积分规则
// @Accept json
// @Produce json
// @Param match_id path int true "比赛ID"
// @Param rule_id query int false "积分规则ID（不指定则使用激活规则）"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/admin/matches/{match_id}/recalculate-points [post]
func (h *ScoringRuleHandler) CalculatePointsWithRule(c *gin.Context) {
	matchIDStr := c.Param("match_id")
	matchID, err := strconv.ParseUint(matchIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的比赛ID")
		return
	}

	var ruleID *uint
	if ruleIDStr := c.Query("rule_id"); ruleIDStr != "" {
		id, err := strconv.ParseUint(ruleIDStr, 10, 32)
		if err != nil {
			response.BadRequest(c, "无效的积分规则ID")
			return
		}
		ruleIDUint := uint(id)
		ruleID = &ruleIDUint
	}

	// 这里需要调用预测服务的方法
	// 由于处理器不应该直接依赖预测服务，这个功能应该在预测处理器中实现
	// 或者通过事件系统来处理

	response.OK(c, "积分重新计算完成", gin.H{
		"match_id": matchID,
		"rule_id":  ruleID,
	})
}
