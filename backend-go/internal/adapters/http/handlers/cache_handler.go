package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"backend-go/internal/core/services"
	"backend-go/pkg/response"
)

// CacheHandler 缓存管理处理器
type CacheHandler struct {
	leaderboardCache    services.LeaderboardCacheService
	invalidationService services.LeaderboardInvalidationService
	monitoringService   services.CacheMonitoringService
}

// NewCacheHandler 创建缓存管理处理器
func NewCacheHandler(
	leaderboardCache services.LeaderboardCacheService,
	invalidationService services.LeaderboardInvalidationService,
	monitoringService services.CacheMonitoringService,
) *CacheHandler {
	return &CacheHandler{
		leaderboardCache:    leaderboardCache,
		invalidationService: invalidationService,
		monitoringService:   monitoringService,
	}
}

// GetCacheStats 获取缓存统计信息
// @Summary 获取缓存统计信息
// @Description 获取排行榜缓存的统计信息，包括命中率、请求数等
// @Tags 缓存管理
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=services.CacheStats}
// @Failure 500 {object} response.Response
// @Router /api/v1/admin/cache/stats [get]
func (h *CacheHandler) GetCacheStats(c *gin.Context) {
	stats := h.leaderboardCache.GetCacheStats()
	response.Success(c, http.StatusOK, "Cache stats retrieved successfully", stats)
}

// GetCacheMetrics 获取详细缓存指标
// @Summary 获取详细缓存指标
// @Description 获取包括Redis状态、系统健康等详细缓存指标
// @Tags 缓存管理
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=services.CacheMetrics}
// @Failure 500 {object} response.Response
// @Router /api/v1/admin/cache/metrics [get]
func (h *CacheHandler) GetCacheMetrics(c *gin.Context) {
	metrics := h.monitoringService.GetMetrics()
	response.Success(c, http.StatusOK, "Cache metrics retrieved successfully", metrics)
}

// GetCacheReport 获取缓存详细报告
// @Summary 获取缓存详细报告
// @Description 获取包括历史统计、告警、优化建议的详细缓存报告
// @Tags 缓存管理
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=services.CacheReport}
// @Failure 500 {object} response.Response
// @Router /api/v1/admin/cache/report [get]
func (h *CacheHandler) GetCacheReport(c *gin.Context) {
	report := h.monitoringService.GetDetailedReport()
	response.Success(c, http.StatusOK, "Cache report retrieved successfully", report)
}

// InvalidateLeaderboard 使排行榜缓存失效
// @Summary 使排行榜缓存失效
// @Description 手动使指定锦标赛的排行榜缓存失效
// @Tags 缓存管理
// @Accept json
// @Produce json
// @Param tournament query string true "锦标赛名称"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/admin/cache/leaderboard/invalidate [post]
func (h *CacheHandler) InvalidateLeaderboard(c *gin.Context) {
	tournament := c.Query("tournament")
	if tournament == "" {
		response.Error(c, http.StatusBadRequest, "tournament parameter is required", "")
		return
	}

	if err := h.leaderboardCache.InvalidateLeaderboard(c.Request.Context(), tournament); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to invalidate cache", err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Cache invalidated successfully", gin.H{
		"message":    "Cache invalidated successfully",
		"tournament": tournament,
	})
}

// RefreshLeaderboard 刷新排行榜缓存
// @Summary 刷新排行榜缓存
// @Description 手动刷新指定锦标赛的排行榜缓存
// @Tags 缓存管理
// @Accept json
// @Produce json
// @Param tournament query string true "锦标赛名称"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/admin/cache/leaderboard/refresh [post]
func (h *CacheHandler) RefreshLeaderboard(c *gin.Context) {
	tournament := c.Query("tournament")
	if tournament == "" {
		response.Error(c, http.StatusBadRequest, "tournament parameter is required", "")
		return
	}

	if err := h.leaderboardCache.RefreshCache(c.Request.Context(), tournament); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to refresh cache", err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Cache refreshed successfully", gin.H{
		"message":    "Cache refreshed successfully",
		"tournament": tournament,
	})
}

// PrewarmCache 预热缓存
// @Summary 预热缓存
// @Description 预热所有锦标赛的排行榜缓存
// @Tags 缓存管理
// @Accept json
// @Produce json
// @Success 200 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/admin/cache/prewarm [post]
func (h *CacheHandler) PrewarmCache(c *gin.Context) {
	if err := h.leaderboardCache.PrewarmCache(c.Request.Context()); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to prewarm cache", err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Cache prewarmed successfully", gin.H{
		"message": "Cache prewarmed successfully",
	})
}

// BatchInvalidate 批量使缓存失效
// @Summary 批量使缓存失效
// @Description 批量使多个锦标赛的排行榜缓存失效
// @Tags 缓存管理
// @Accept json
// @Produce json
// @Param request body BatchInvalidateRequest true "批量失效请求"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/admin/cache/batch-invalidate [post]
func (h *CacheHandler) BatchInvalidate(c *gin.Context) {
	var req BatchInvalidateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	if len(req.Tournaments) == 0 {
		response.Error(c, http.StatusBadRequest, "tournaments list cannot be empty", "")
		return
	}

	if err := h.invalidationService.BatchInvalidate(c.Request.Context(), req.Tournaments); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to batch invalidate", err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Batch invalidation completed successfully", gin.H{
		"message":     "Batch invalidation completed successfully",
		"tournaments": req.Tournaments,
	})
}

// InvalidateOnPointsUpdate 积分更新时使缓存失效
// @Summary 积分更新时使缓存失效
// @Description 当用户积分更新时使相关排行榜缓存失效
// @Tags 缓存管理
// @Accept json
// @Produce json
// @Param user_id query int true "用户ID"
// @Param tournament query string true "锦标赛名称"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/admin/cache/invalidate-points [post]
func (h *CacheHandler) InvalidateOnPointsUpdate(c *gin.Context) {
	userIDStr := c.Query("user_id")
	tournament := c.Query("tournament")

	if userIDStr == "" || tournament == "" {
		response.Error(c, http.StatusBadRequest, "user_id and tournament parameters are required", "")
		return
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid user_id", err.Error())
		return
	}

	if err := h.invalidationService.InvalidateOnPointsUpdate(c.Request.Context(), uint(userID), tournament); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to invalidate on points update", err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Cache invalidated on points update", gin.H{
		"message":    "Cache invalidated on points update",
		"user_id":    userID,
		"tournament": tournament,
	})
}

// CheckHitRate 检查命中率
// @Summary 检查命中率
// @Description 检查当前缓存命中率是否达到阈值
// @Tags 缓存管理
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=HitRateCheckResponse}
// @Failure 500 {object} response.Response
// @Router /api/v1/admin/cache/hit-rate [get]
func (h *CacheHandler) CheckHitRate(c *gin.Context) {
	isHealthy := h.monitoringService.CheckHitRateThreshold()
	stats := h.leaderboardCache.GetCacheStats()

	response.Success(c, http.StatusOK, "Hit rate checked", HitRateCheckResponse{
		IsHealthy:   isHealthy,
		CurrentRate: stats.HitRate,
		Threshold:   90.0, // 可以从配置中获取
		Stats:       stats,
	})
}

// BatchInvalidateRequest 批量失效请求
type BatchInvalidateRequest struct {
	Tournaments []string `json:"tournaments" binding:"required"`
}

// HitRateCheckResponse 命中率检查响应
type HitRateCheckResponse struct {
	IsHealthy   bool                `json:"is_healthy"`
	CurrentRate float64             `json:"current_rate"`
	Threshold   float64             `json:"threshold"`
	Stats       services.CacheStats `json:"stats"`
}
