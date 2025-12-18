package handlers

import (
	"net/http"
	"strconv"

	"backend-go/internal/core/services"
	"backend-go/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// AsyncPointsHandler 异步积分计算处理器
type AsyncPointsHandler struct {
	asyncPointsService *services.AsyncPointsService
	logger             *logrus.Logger
}

// NewAsyncPointsHandler 创建异步积分计算处理器
func NewAsyncPointsHandler(asyncPointsService *services.AsyncPointsService, logger *logrus.Logger) *AsyncPointsHandler {
	return &AsyncPointsHandler{
		asyncPointsService: asyncPointsService,
		logger:             logger,
	}
}

// TriggerPointsCalculationRequest 触发积分计算请求
type TriggerPointsCalculationRequest struct {
	MatchID uint  `json:"match_id" binding:"required"`
	RuleID  *uint `json:"rule_id,omitempty"`
}

// TriggerPointsCalculationResponse 触发积分计算响应
type TriggerPointsCalculationResponse struct {
	TaskID  string `json:"task_id"`
	MatchID uint   `json:"match_id"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

// TriggerPointsCalculation 手动触发积分计算
// @Summary 手动触发积分计算
// @Description 为指定比赛手动触发异步积分计算
// @Tags 积分管理
// @Accept json
// @Produce json
// @Param request body TriggerPointsCalculationRequest true "触发请求"
// @Success 200 {object} response.Response{data=TriggerPointsCalculationResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/admin/points/calculate [post]
func (h *AsyncPointsHandler) TriggerPointsCalculation(c *gin.Context) {
	var req TriggerPointsCalculationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "无效的请求参数", err.Error())
		return
	}

	// 触发积分计算
	taskID, err := h.asyncPointsService.QueuePointsCalculation(req.MatchID, req.RuleID)
	if err != nil {
		h.logger.WithError(err).WithField("match_id", req.MatchID).Error("Failed to trigger points calculation")
		response.Error(c, http.StatusInternalServerError, "触发积分计算失败", err.Error())
		return
	}

	resp := TriggerPointsCalculationResponse{
		TaskID:  taskID,
		MatchID: req.MatchID,
		Status:  "queued",
		Message: "积分计算任务已加入队列",
	}

	h.logger.WithFields(logrus.Fields{
		"task_id":  taskID,
		"match_id": req.MatchID,
		"rule_id":  req.RuleID,
	}).Info("Points calculation task triggered manually")

	response.OK(c, "积分计算任务已加入队列", resp)
}

// GetTaskStatus 获取任务状态
// @Summary 获取积分计算任务状态
// @Description 根据任务ID获取积分计算任务的状态
// @Tags 积分管理
// @Produce json
// @Param task_id path string true "任务ID"
// @Success 200 {object} response.Response{data=services.PointsCalculationTask}
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/admin/points/tasks/{task_id} [get]
func (h *AsyncPointsHandler) GetTaskStatus(c *gin.Context) {
	taskID := c.Param("task_id")
	if taskID == "" {
		response.Error(c, http.StatusBadRequest, "任务ID不能为空", "")
		return
	}

	task, err := h.asyncPointsService.GetTaskStatus(taskID)
	if err != nil {
		h.logger.WithError(err).WithField("task_id", taskID).Error("Failed to get task status")
		response.Error(c, http.StatusNotFound, "任务不存在", err.Error())
		return
	}

	response.OK(c, "获取任务状态成功", task)
}

// GetQueueStatus 获取队列状态
// @Summary 获取积分计算队列状态
// @Description 获取积分计算任务队列的整体状态信息
// @Tags 积分管理
// @Produce json
// @Success 200 {object} response.Response{data=map[string]interface{}}
// @Router /api/v1/admin/points/queue/status [get]
func (h *AsyncPointsHandler) GetQueueStatus(c *gin.Context) {
	status := h.asyncPointsService.GetQueueStatus()
	response.OK(c, "获取队列状态成功", status)
}

// BatchTriggerPointsCalculation 批量触发积分计算
type BatchTriggerRequest struct {
	MatchIDs []uint `json:"match_ids" binding:"required"`
	RuleID   *uint  `json:"rule_id,omitempty"`
}

type BatchTriggerResponse struct {
	SuccessCount int                                `json:"success_count"`
	FailedCount  int                                `json:"failed_count"`
	Tasks        []TriggerPointsCalculationResponse `json:"tasks"`
	Errors       []string                           `json:"errors,omitempty"`
}

// BatchTriggerPointsCalculation 批量触发积分计算
// @Summary 批量触发积分计算
// @Description 为多个比赛批量触发异步积分计算
// @Tags 积分管理
// @Accept json
// @Produce json
// @Param request body BatchTriggerRequest true "批量触发请求"
// @Success 200 {object} response.Response{data=BatchTriggerResponse}
// @Failure 400 {object} response.Response
// @Router /api/v1/admin/points/calculate/batch [post]
func (h *AsyncPointsHandler) BatchTriggerPointsCalculation(c *gin.Context) {
	var req BatchTriggerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "无效的请求参数", err.Error())
		return
	}

	if len(req.MatchIDs) == 0 {
		response.Error(c, http.StatusBadRequest, "比赛ID列表不能为空", "")
		return
	}

	if len(req.MatchIDs) > 50 {
		response.Error(c, http.StatusBadRequest, "一次最多只能处理50个比赛", "")
		return
	}

	resp := BatchTriggerResponse{
		Tasks:  make([]TriggerPointsCalculationResponse, 0),
		Errors: make([]string, 0),
	}

	for _, matchID := range req.MatchIDs {
		taskID, err := h.asyncPointsService.QueuePointsCalculation(matchID, req.RuleID)
		if err != nil {
			resp.FailedCount++
			resp.Errors = append(resp.Errors, "比赛 "+strconv.Itoa(int(matchID))+" 触发失败: "+err.Error())
			h.logger.WithError(err).WithField("match_id", matchID).Error("Failed to trigger points calculation in batch")
		} else {
			resp.SuccessCount++
			resp.Tasks = append(resp.Tasks, TriggerPointsCalculationResponse{
				TaskID:  taskID,
				MatchID: matchID,
				Status:  "queued",
				Message: "已加入队列",
			})
		}
	}

	h.logger.WithFields(logrus.Fields{
		"total_matches": len(req.MatchIDs),
		"success_count": resp.SuccessCount,
		"failed_count":  resp.FailedCount,
	}).Info("Batch points calculation triggered")

	response.OK(c, "批量积分计算任务处理完成", resp)
}

// RegisterAsyncPointsRoutes 注册异步积分计算路由
func RegisterAsyncPointsRoutes(r *gin.RouterGroup, handler *AsyncPointsHandler) {
	admin := r.Group("/admin/points")
	{
		admin.POST("/calculate", handler.TriggerPointsCalculation)
		admin.POST("/calculate/batch", handler.BatchTriggerPointsCalculation)
		admin.GET("/tasks/:task_id", handler.GetTaskStatus)
		admin.GET("/queue/status", handler.GetQueueStatus)
	}
}
