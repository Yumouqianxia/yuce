package handlers

import (
	"strconv"

	"backend-go/internal/adapters/http/middleware"
	"backend-go/internal/core/domain/prediction"
	"backend-go/pkg/response"

	"github.com/gin-gonic/gin"
)

// PredictionHandler 预测处理器
type PredictionHandler struct {
	predictionService prediction.Service
}

// NewPredictionHandler 创建预测处理器
func NewPredictionHandler(predictionService prediction.Service) *PredictionHandler {
	return &PredictionHandler{
		predictionService: predictionService,
	}
}

// CreatePrediction 创建预测
// @Summary 创建预测
// @Description 为指定比赛创建预测
// @Tags predictions
// @Accept json
// @Produce json
// @Param request body prediction.CreatePredictionRequest true "创建预测请求"
// @Success 201 {object} response.Response{data=prediction.Prediction}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 409 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/predictions [post]
// @Security BearerAuth
func (h *PredictionHandler) CreatePrediction(c *gin.Context) {
	// 获取用户ID（上下文存的是字符串）
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		response.Unauthorized(c, "用户未认证")
		return
	}

	var req prediction.CreatePredictionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	pred, err := h.predictionService.CreatePrediction(c.Request.Context(), userID, &req)
	if err != nil {
		if appErr, ok := err.(*response.AppError); ok {
			response.Error(c, appErr.StatusCode, appErr.Message, appErr.Error())
			return
		}
		response.InternalError(c, "创建预测失败: "+err.Error())
		return
	}

	response.Created(c, "预测创建成功", pred)
}

// UpdatePrediction 更新预测
// @Summary 更新预测
// @Description 更新指定的预测
// @Tags predictions
// @Accept json
// @Produce json
// @Param id path int true "预测ID"
// @Param request body prediction.UpdatePredictionRequest true "更新预测请求"
// @Success 200 {object} response.Response{data=prediction.Prediction}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/predictions/{id} [put]
// @Security BearerAuth
func (h *PredictionHandler) UpdatePrediction(c *gin.Context) {
	// 获取用户ID
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		response.Unauthorized(c, "用户未认证")
		return
	}

	// 获取预测ID
	predictionIDStr := c.Param("id")
	predictionID, err := strconv.ParseUint(predictionIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的预测ID")
		return
	}

	var req prediction.UpdatePredictionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	pred, err := h.predictionService.UpdatePrediction(c.Request.Context(), userID, uint(predictionID), &req)
	if err != nil {
		if appErr, ok := err.(*response.AppError); ok {
			response.Error(c, appErr.StatusCode, appErr.Message, appErr.Error())
			return
		}
		response.InternalError(c, "更新预测失败: "+err.Error())
		return
	}

	response.OK(c, "预测更新成功", pred)
}

// GetPrediction 获取预测详情
// @Summary 获取预测详情
// @Description 获取指定预测的详细信息
// @Tags predictions
// @Accept json
// @Produce json
// @Param id path int true "预测ID"
// @Success 200 {object} response.Response{data=prediction.Prediction}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/predictions/{id} [get]
func (h *PredictionHandler) GetPrediction(c *gin.Context) {
	// 获取预测ID
	predictionIDStr := c.Param("id")
	predictionID, err := strconv.ParseUint(predictionIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的预测ID")
		return
	}

	pred, err := h.predictionService.GetPrediction(c.Request.Context(), uint(predictionID))
	if err != nil {
		if appErr, ok := err.(*response.AppError); ok {
			response.Error(c, appErr.StatusCode, appErr.Message, appErr.Error())
			return
		}
		response.InternalError(c, "获取预测失败: "+err.Error())
		return
	}

	response.OK(c, "获取预测成功", pred)
}

// GetPredictionsByMatch 获取比赛的所有预测
// @Summary 获取比赛预测列表
// @Description 获取指定比赛的所有预测
// @Tags predictions
// @Accept json
// @Produce json
// @Param match_id query int true "比赛ID"
// @Success 200 {object} response.Response{data=[]prediction.PredictionWithVotes}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/predictions [get]
func (h *PredictionHandler) GetPredictionsByMatch(c *gin.Context) {
	// 获取比赛ID
	matchIDStr := c.Query("match_id")
	if matchIDStr == "" {
		response.BadRequest(c, "缺少比赛ID参数")
		return
	}

	matchID, err := strconv.ParseUint(matchIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的比赛ID")
		return
	}

	// 获取当前用户ID（可选）
	var userID *uint
	if userIDValue, exists := c.Get("user_id"); exists {
		uid := userIDValue.(uint)
		userID = &uid
	}

	predictions, err := h.predictionService.GetPredictionsByMatch(c.Request.Context(), uint(matchID), userID)
	if err != nil {
		if appErr, ok := err.(*response.AppError); ok {
			response.Error(c, appErr.StatusCode, appErr.Message, appErr.Error())
			return
		}
		response.InternalError(c, "获取预测列表失败: "+err.Error())
		return
	}

	response.OK(c, "获取预测列表成功", predictions)
}

// GetUserPredictions 获取用户的所有预测
// @Summary 获取用户预测列表
// @Description 获取当前用户的所有预测
// @Tags predictions
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=[]prediction.Prediction}
// @Failure 401 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/predictions/my [get]
// @Security BearerAuth
func (h *PredictionHandler) GetUserPredictions(c *gin.Context) {
	// 获取用户ID
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		response.Unauthorized(c, "用户未认证")
		return
	}

	predictions, err := h.predictionService.GetUserPredictions(c.Request.Context(), userID)
	if err != nil {
		if appErr, ok := err.(*response.AppError); ok {
			response.Error(c, appErr.StatusCode, appErr.Message, appErr.Error())
			return
		}
		response.InternalError(c, "获取用户预测失败: "+err.Error())
		return
	}

	response.OK(c, "获取用户预测成功", predictions)
}

// VotePrediction 投票支持预测
// @Summary 投票支持预测
// @Description 为指定预测投票
// @Tags predictions
// @Accept json
// @Produce json
// @Param id path int true "预测ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 409 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/predictions/{id}/vote [post]
// @Security BearerAuth
func (h *PredictionHandler) VotePrediction(c *gin.Context) {
	// 获取用户ID
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		response.Unauthorized(c, "用户未认证")
		return
	}

	// 获取预测ID
	predictionIDStr := c.Param("id")
	predictionID, err := strconv.ParseUint(predictionIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的预测ID")
		return
	}

	err = h.predictionService.VotePrediction(c.Request.Context(), userID, uint(predictionID))
	if err != nil {
		if appErr, ok := err.(*response.AppError); ok {
			response.Error(c, appErr.StatusCode, appErr.Message, appErr.Error())
			return
		}
		response.InternalError(c, "投票失败: "+err.Error())
		return
	}

	response.OK(c, "投票成功", nil)
}

// UnvotePrediction 取消投票
// @Summary 取消投票
// @Description 取消对指定预测的投票
// @Tags predictions
// @Accept json
// @Produce json
// @Param id path int true "预测ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/predictions/{id}/vote [delete]
// @Security BearerAuth
func (h *PredictionHandler) UnvotePrediction(c *gin.Context) {
	// 获取用户ID
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		response.Unauthorized(c, "用户未认证")
		return
	}

	// 获取预测ID
	predictionIDStr := c.Param("id")
	predictionID, err := strconv.ParseUint(predictionIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的预测ID")
		return
	}

	err = h.predictionService.UnvotePrediction(c.Request.Context(), userID, uint(predictionID))
	if err != nil {
		if appErr, ok := err.(*response.AppError); ok {
			response.Error(c, appErr.StatusCode, appErr.Message, appErr.Error())
			return
		}
		response.InternalError(c, "取消投票失败: "+err.Error())
		return
	}

	response.OK(c, "取消投票成功", nil)
}

// ReverifyPrediction 重新验证预测（按预测ID触发所在比赛的积分计算）
func (h *PredictionHandler) ReverifyPrediction(c *gin.Context) {
	// 预测ID
	predictionIDStr := c.Param("id")
	predictionID, err := strconv.ParseUint(predictionIDStr, 10, 32)
	if err != nil {
		// 允许直接用比赛ID触发重算
		response.BadRequest(c, "无效的预测ID或比赛ID")
		return
	}

	ctx := c.Request.Context()

	// 先按预测ID查找
	pred, err := h.predictionService.GetPrediction(ctx, uint(predictionID))
	if err == nil && pred != nil {
		if pred.Match == nil {
			response.InternalError(c, "预测缺少比赛信息")
			return
		}
		if err := h.predictionService.CalculatePoints(ctx, pred.MatchID); err != nil {
			if response.IsErrorCode(err, response.CodeBadRequest) || response.IsValidationError(err) {
				response.BadRequest(c, err.Error())
				return
			}
			if response.IsNotFoundError(err) {
				response.NotFound(c, "Match")
				return
			}
			response.InternalError(c, "重新验证预测失败: "+err.Error())
			return
		}
		response.OK(c, "重新验证预测成功", nil)
		return
	}

	// 若预测不存在，尝试将该ID作为比赛ID重算
	if err := h.predictionService.CalculatePoints(ctx, uint(predictionID)); err != nil {
		if response.IsErrorCode(err, response.CodeBadRequest) || response.IsValidationError(err) {
			response.BadRequest(c, err.Error())
			return
		}
		if response.IsNotFoundError(err) {
			response.NotFound(c, "Prediction or match")
			return
		}
		response.InternalError(c, "重新验证预测失败: "+err.Error())
		return
	}

	response.OK(c, "根据比赛ID重新验证成功", nil)
}

// GetFeaturedPredictions 获取精选预测
// @Summary 获取精选预测
// @Description 获取系统精选的预测列表
// @Tags predictions
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=[]prediction.PredictionWithVotes}
// @Failure 500 {object} response.Response
// @Router /api/v1/predictions/featured [get]
func (h *PredictionHandler) GetFeaturedPredictions(c *gin.Context) {
	predictions, err := h.predictionService.GetFeaturedPredictions(c.Request.Context())
	if err != nil {
		if appErr, ok := err.(*response.AppError); ok {
			response.Error(c, appErr.StatusCode, appErr.Message, appErr.Error())
			return
		}
		response.InternalError(c, "获取精选预测失败: "+err.Error())
		return
	}

	response.OK(c, "获取精选预测成功", predictions)
}
