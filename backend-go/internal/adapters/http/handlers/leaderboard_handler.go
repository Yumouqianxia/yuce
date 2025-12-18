package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"backend-go/internal/core/domain/leaderboard"
	"backend-go/internal/core/domain/scoring"
	"backend-go/pkg/response"
)

// LeaderboardHandler 排行榜处理器
type LeaderboardHandler struct {
	leaderboardService leaderboard.Service
	scoringService     scoring.Service
	logger             *logrus.Logger
}

// NewLeaderboardHandler 创建排行榜处理器
func NewLeaderboardHandler(
	leaderboardService leaderboard.Service,
	scoringService scoring.Service,
	logger *logrus.Logger,
) *LeaderboardHandler {
	return &LeaderboardHandler{
		leaderboardService: leaderboardService,
		scoringService:     scoringService,
		logger:             logger,
	}
}

// GetLeaderboardRequest 获取排行榜请求
type GetLeaderboardRequest struct {
	Tournament string `form:"tournament" binding:"omitempty,oneof=SPRING SUMMER GLOBAL"`
	Limit      int    `form:"limit" binding:"omitempty,min=1,max=100"`
}

// GetLeaderboard 获取排行榜
// @Summary 获取排行榜
// @Description 获取指定锦标赛的排行榜
// @Tags leaderboard
// @Accept json
// @Produce json
// @Param tournament query string false "锦标赛类型" Enums(SPRING,SUMMER,GLOBAL) default(GLOBAL)
// @Param limit query int false "返回数量限制" minimum(1) maximum(100) default(10)
// @Success 200 {object} response.Response{data=[]leaderboard.LeaderboardEntry}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/leaderboard [get]
func (h *LeaderboardHandler) GetLeaderboard(c *gin.Context) {
	var req GetLeaderboardRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		h.logger.WithError(err).Error("绑定排行榜请求参数失败")
		response.Error(c, http.StatusBadRequest, "请求参数无效", err.Error())
		return
	}

	// 设置默认值
	if req.Tournament == "" {
		req.Tournament = string(leaderboard.TournamentGlobal)
	}
	if req.Limit == 0 {
		req.Limit = 10
	}

	h.logger.WithFields(logrus.Fields{
		"tournament": req.Tournament,
		"limit":      req.Limit,
	}).Info("获取排行榜")

	entries, err := h.leaderboardService.GetLeaderboard(c.Request.Context(), req.Tournament, req.Limit)
	if err != nil {
		h.logger.WithError(err).Error("获取排行榜失败")
		response.Error(c, http.StatusInternalServerError, "获取排行榜失败", err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Leaderboard retrieved successfully", entries)
}

// GetUserRank 获取用户排名
// @Summary 获取用户排名
// @Description 获取指定用户在指定锦标赛中的排名信息
// @Tags leaderboard
// @Accept json
// @Produce json
// @Param user_id path int true "用户ID"
// @Param tournament query string false "锦标赛类型" Enums(SPRING,SUMMER,GLOBAL) default(GLOBAL)
// @Success 200 {object} response.Response{data=leaderboard.UserRankInfo}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/leaderboard/users/{user_id}/rank [get]
func (h *LeaderboardHandler) GetUserRank(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "用户ID无效", "")
		return
	}

	tournament := c.DefaultQuery("tournament", string(leaderboard.TournamentGlobal))

	h.logger.WithFields(logrus.Fields{
		"user_id":    userID,
		"tournament": tournament,
	}).Info("获取用户排名")

	rankInfo, err := h.leaderboardService.GetUserRank(c.Request.Context(), uint(userID), tournament)
	if err != nil {
		h.logger.WithError(err).Error("获取用户排名失败")
		response.Error(c, http.StatusInternalServerError, "获取用户排名失败", err.Error())
		return
	}

	if rankInfo == nil {
		response.Error(c, http.StatusNotFound, "用户排名不存在", "")
		return
	}

	response.Success(c, http.StatusOK, "User rank retrieved successfully", rankInfo)
}

// GetLeaderboardStats 获取排行榜统计
// @Summary 获取排行榜统计
// @Description 获取指定锦标赛的排行榜统计信息
// @Tags leaderboard
// @Accept json
// @Produce json
// @Param tournament query string false "锦标赛类型" Enums(SPRING,SUMMER,GLOBAL) default(GLOBAL)
// @Success 200 {object} response.Response{data=leaderboard.LeaderboardStats}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/leaderboard/stats [get]
func (h *LeaderboardHandler) GetLeaderboardStats(c *gin.Context) {
	tournament := c.DefaultQuery("tournament", string(leaderboard.TournamentGlobal))

	h.logger.WithField("tournament", tournament).Info("获取排行榜统计")

	stats, err := h.leaderboardService.GetLeaderboardStats(c.Request.Context(), tournament)
	if err != nil {
		h.logger.WithError(err).Error("获取排行榜统计失败")
		response.Error(c, http.StatusInternalServerError, "获取排行榜统计失败", err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Leaderboard stats retrieved successfully", stats)
}

// RefreshLeaderboard 刷新排行榜缓存
// @Summary 刷新排行榜缓存
// @Description 手动刷新指定锦标赛的排行榜缓存
// @Tags leaderboard
// @Accept json
// @Produce json
// @Param tournament query string false "锦标赛类型" Enums(SPRING,SUMMER,GLOBAL) default(GLOBAL)
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/leaderboard/refresh [post]
func (h *LeaderboardHandler) RefreshLeaderboard(c *gin.Context) {
	tournament := c.DefaultQuery("tournament", string(leaderboard.TournamentGlobal))

	h.logger.WithField("tournament", tournament).Info("刷新排行榜缓存")

	err := h.leaderboardService.RefreshLeaderboard(c.Request.Context(), tournament)
	if err != nil {
		h.logger.WithError(err).Error("刷新排行榜缓存失败")
		response.Error(c, http.StatusInternalServerError, "刷新排行榜缓存失败", err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Leaderboard cache refreshed", gin.H{"message": "排行榜缓存刷新成功"})
}

// CalculateMatchPointsRequest 计算比赛积分请求
type CalculateMatchPointsRequest struct {
	RuleID *uint `json:"rule_id" binding:"omitempty"`
}

// CalculateMatchPoints 计算比赛积分
// @Summary 计算比赛积分
// @Description 计算指定比赛结束后的所有预测积分
// @Tags leaderboard
// @Accept json
// @Produce json
// @Param match_id path int true "比赛ID"
// @Param request body CalculateMatchPointsRequest false "计算请求"
// @Success 200 {object} response.Response{data=scoring.MatchPointsCalculation}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/leaderboard/matches/{match_id}/calculate-points [post]
func (h *LeaderboardHandler) CalculateMatchPoints(c *gin.Context) {
	matchIDStr := c.Param("match_id")
	matchID, err := strconv.ParseUint(matchIDStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "比赛ID无效", "")
		return
	}

	var req CalculateMatchPointsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("绑定计算积分请求参数失败")
		response.Error(c, http.StatusBadRequest, "请求参数无效", err.Error())
		return
	}

	h.logger.WithFields(logrus.Fields{
		"match_id": matchID,
		"rule_id":  req.RuleID,
	}).Info("计算比赛积分")

	calculation, err := h.scoringService.CalculateMatchPointsWithRule(c.Request.Context(), uint(matchID), req.RuleID)
	if err != nil {
		h.logger.WithError(err).Error("计算比赛积分失败")
		response.Error(c, http.StatusInternalServerError, "计算比赛积分失败", err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Match points calculated successfully", calculation)
}

// GetPointsHistory 获取用户积分历史
// @Summary 获取用户积分历史
// @Description 获取指定用户的积分变化历史
// @Tags leaderboard
// @Accept json
// @Produce json
// @Param user_id path int true "用户ID"
// @Param tournament query string false "锦标赛类型" Enums(SPRING,SUMMER,GLOBAL)
// @Success 200 {object} response.Response{data=[]scoring.PointsUpdateEvent}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/leaderboard/users/{user_id}/points-history [get]
func (h *LeaderboardHandler) GetPointsHistory(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "用户ID无效", "")
		return
	}

	tournament := c.Query("tournament")

	h.logger.WithFields(logrus.Fields{
		"user_id":    userID,
		"tournament": tournament,
	}).Info("获取用户积分历史")

	history, err := h.scoringService.GetPointsHistory(c.Request.Context(), uint(userID), tournament)
	if err != nil {
		h.logger.WithError(err).Error("获取用户积分历史失败")
		response.Error(c, http.StatusInternalServerError, "获取用户积分历史失败", err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Points history retrieved successfully", history)
}

// GetUsersAroundRank 获取排名周围的用户
// @Summary 获取排名周围的用户
// @Description 获取指定排名周围的用户列表
// @Tags leaderboard
// @Accept json
// @Produce json
// @Param rank path int true "排名"
// @Param tournament query string false "锦标赛类型" Enums(SPRING,SUMMER,GLOBAL) default(GLOBAL)
// @Param radius query int false "范围半径" minimum(1) maximum(20) default(5)
// @Success 200 {object} response.Response{data=[]leaderboard.LeaderboardEntry}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/leaderboard/ranks/{rank}/around [get]
func (h *LeaderboardHandler) GetUsersAroundRank(c *gin.Context) {
	rankStr := c.Param("rank")
	rank, err := strconv.Atoi(rankStr)
	if err != nil || rank <= 0 {
		response.Error(c, http.StatusBadRequest, "排名无效", "")
		return
	}

	tournament := c.DefaultQuery("tournament", string(leaderboard.TournamentGlobal))
	radiusStr := c.DefaultQuery("radius", "5")
	radius, err := strconv.Atoi(radiusStr)
	if err != nil || radius <= 0 || radius > 20 {
		radius = 5
	}

	h.logger.WithFields(logrus.Fields{
		"rank":       rank,
		"tournament": tournament,
		"radius":     radius,
	}).Info("获取排名周围的用户")

	entries, err := h.leaderboardService.GetUsersAroundRank(c.Request.Context(), tournament, rank, radius)
	if err != nil {
		h.logger.WithError(err).Error("获取排名周围的用户失败")
		response.Error(c, http.StatusInternalServerError, "获取排名周围的用户失败", err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Users around rank retrieved successfully", entries)
}
