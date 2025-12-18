package handlers

import (
	"strconv"
	"time"

	"backend-go/internal/core/domain"
	"backend-go/internal/core/domain/match"
	"backend-go/pkg/response"
	"github.com/gin-gonic/gin"
)

// MatchHandler 比赛处理器
type MatchHandler struct {
	matchService match.Service
}

// NewMatchHandler 创建比赛处理器
func NewMatchHandler(matchService match.Service) *MatchHandler {
	return &MatchHandler{
		matchService: matchService,
	}
}

// CreateMatch 创建比赛
// @Summary 创建比赛
// @Description 创建新的比赛
// @Tags matches
// @Accept json
// @Produce json
// @Param match body match.CreateMatchRequest true "比赛信息"
// @Success 201 {object} response.Response{data=match.Match}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/matches [post]
func (h *MatchHandler) CreateMatch(c *gin.Context) {
	var req match.CreateMatchRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	// 创建比赛
	m, err := h.matchService.CreateMatch(c.Request.Context(), &req)
	if err != nil {
		switch err {
		case domain.ErrInvalidInput:
			response.BadRequest(c, "Invalid input")
		case domain.ErrInvalidStartTime:
			response.BadRequest(c, "Invalid start time")
		case domain.ErrInvalidTournament:
			response.BadRequest(c, "Invalid tournament")
		default:
			response.InternalError(c, "Failed to create match")
		}
		return
	}

	response.Created(c, "Match created successfully", m)
}

// GetMatch 获取比赛详情
// @Summary 获取比赛详情
// @Description 根据ID获取比赛详情
// @Tags matches
// @Produce json
// @Param id path int true "比赛ID"
// @Success 200 {object} response.Response{data=match.Match}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/matches/{id} [get]
func (h *MatchHandler) GetMatch(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid match ID")
		return
	}

	m, err := h.matchService.GetMatch(c.Request.Context(), uint(id))
	if err != nil {
		if err == domain.ErrMatchNotFound {
			response.NotFound(c, "Match")
		} else {
			response.InternalError(c, "Failed to get match")
		}
		return
	}

	response.OK(c, "Match retrieved successfully", m)
}

// ListMatches 获取比赛列表
// @Summary 获取比赛列表
// @Description 获取比赛列表，支持过滤和分页
// @Tags matches
// @Produce json
// @Param tournament query string false "赛事类型"
// @Param status query string false "比赛状态"
// @Param start_date query string false "开始日期"
// @Param end_date query string false "结束日期"
// @Param limit query int false "限制数量"
// @Param offset query int false "偏移量"
// @Success 200 {object} response.Response{data=[]match.Match}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/matches [get]
func (h *MatchHandler) ListMatches(c *gin.Context) {
	var filter match.ListFilter

	// 解析查询参数
	if tournament := c.Query("tournament"); tournament != "" {
		filter.Tournament = match.Tournament(tournament)
	}

	if status := c.Query("status"); status != "" {
		// 兼容前端的 not_started/finished/live 命名
		switch status {
		case "not_started", "upcoming":
			filter.Status = match.MatchStatusUpcoming
		case "live":
			filter.Status = match.MatchStatusLive
		case "finished":
			filter.Status = match.MatchStatusFinished
		case "cancelled":
			filter.Status = match.MatchStatusCancelled
		default:
			filter.Status = match.MatchStatus(status)
		}
	}

	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if startDate, err := time.Parse("2006-01-02", startDateStr); err == nil {
			filter.StartDate = &startDate
		}
	}

	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if endDate, err := time.Parse("2006-01-02", endDateStr); err == nil {
			filter.EndDate = &endDate
		}
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			filter.Limit = limit
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			filter.Offset = offset
		}
	}

	// 获取比赛列表
	matches, err := h.matchService.ListMatches(c.Request.Context(), filter)
	if err != nil {
		response.InternalError(c, "Failed to get matches")
		return
	}

	response.OK(c, "Matches retrieved successfully", matches)
}

// UpdateMatch 更新比赛信息
// @Summary 更新比赛信息
// @Description 更新比赛信息
// @Tags matches
// @Accept json
// @Produce json
// @Param id path int true "比赛ID"
// @Param match body match.UpdateMatchRequest true "更新信息"
// @Success 200 {object} response.Response{data=match.Match}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/matches/{id} [put]
func (h *MatchHandler) UpdateMatch(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid match ID")
		return
	}

	var req match.UpdateMatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	// 更新比赛
	m, err := h.matchService.UpdateMatch(c.Request.Context(), uint(id), &req)
	if err != nil {
		switch err {
		case domain.ErrMatchNotFound:
			response.NotFound(c, "Match")
		case domain.ErrMatchAlreadyStarted:
			response.BadRequest(c, "Match already started")
		case domain.ErrInvalidStartTime:
			response.BadRequest(c, "Invalid start time")
		case domain.ErrInvalidTournament:
			response.BadRequest(c, "Invalid tournament")
		default:
			response.InternalError(c, "Failed to update match")
		}
		return
	}

	response.OK(c, "Match updated successfully", m)
}

// StartMatch 开始比赛
// @Summary 开始比赛
// @Description 将比赛状态设置为进行中
// @Tags matches
// @Produce json
// @Param id path int true "比赛ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/matches/{id}/start [post]
func (h *MatchHandler) StartMatch(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid match ID")
		return
	}

	err = h.matchService.StartMatch(c.Request.Context(), uint(id))
	if err != nil {
		switch err {
		case domain.ErrMatchNotFound:
			response.NotFound(c, "Match")
		case domain.ErrInvalidMatchStatus:
			response.BadRequest(c, "Invalid match status")
		default:
			response.InternalError(c, "Failed to start match")
		}
		return
	}

	response.OK(c, "Match started successfully", nil)
}

// SetResult 设置比赛结果
// @Summary 设置比赛结果
// @Description 设置比赛结果并结束比赛
// @Tags matches
// @Accept json
// @Produce json
// @Param id path int true "比赛ID"
// @Param result body match.SetResultRequest true "比赛结果"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/matches/{id}/result [post]
func (h *MatchHandler) SetResult(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid match ID")
		return
	}

	var req match.SetResultRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	err = h.matchService.SetResult(c.Request.Context(), uint(id), &req)
	if err != nil {
		switch err {
		case domain.ErrMatchNotFound:
			response.NotFound(c, "Match")
		case domain.ErrMatchAlreadyFinished:
			response.BadRequest(c, "Match already finished")
		case domain.ErrInvalidWinner:
			response.BadRequest(c, "Invalid winner")
		default:
			response.InternalError(c, "Failed to set result")
		}
		return
	}

	response.OK(c, "Match result set successfully", nil)
}

// CancelMatch 取消比赛
// @Summary 取消比赛
// @Description 取消比赛
// @Tags matches
// @Produce json
// @Param id path int true "比赛ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/matches/{id}/cancel [post]
func (h *MatchHandler) CancelMatch(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid match ID")
		return
	}

	err = h.matchService.CancelMatch(c.Request.Context(), uint(id))
	if err != nil {
		switch err {
		case domain.ErrMatchNotFound:
			response.NotFound(c, "Match")
		case domain.ErrMatchAlreadyFinished:
			response.BadRequest(c, "Match already finished")
		default:
			response.InternalError(c, "Failed to cancel match")
		}
		return
	}

	response.OK(c, "Match cancelled successfully", nil)
}

// GetUpcomingMatches 获取即将开始的比赛
// @Summary 获取即将开始的比赛
// @Description 获取即将开始的比赛列表
// @Tags matches
// @Produce json
// @Success 200 {object} response.Response{data=[]match.Match}
// @Failure 500 {object} response.Response
// @Router /api/matches/upcoming [get]
func (h *MatchHandler) GetUpcomingMatches(c *gin.Context) {
	matches, err := h.matchService.GetUpcomingMatches(c.Request.Context())
	if err != nil {
		response.InternalError(c, "Failed to get upcoming matches")
		return
	}

	response.OK(c, "Upcoming matches retrieved successfully", matches)
}

// GetLiveMatches 获取正在进行的比赛
// @Summary 获取正在进行的比赛
// @Description 获取正在进行的比赛列表
// @Tags matches
// @Produce json
// @Success 200 {object} response.Response{data=[]match.Match}
// @Failure 500 {object} response.Response
// @Router /api/matches/live [get]
func (h *MatchHandler) GetLiveMatches(c *gin.Context) {
	matches, err := h.matchService.GetLiveMatches(c.Request.Context())
	if err != nil {
		response.InternalError(c, "Failed to get live matches")
		return
	}

	response.OK(c, "Live matches retrieved successfully", matches)
}

// GetFinishedMatches 获取已结束的比赛
// @Summary 获取已结束的比赛
// @Description 获取已结束的比赛列表
// @Tags matches
// @Produce json
// @Param limit query int false "限制数量"
// @Success 200 {object} response.Response{data=[]match.Match}
// @Failure 500 {object} response.Response
// @Router /api/matches/finished [get]
func (h *MatchHandler) GetFinishedMatches(c *gin.Context) {
	limit := 20 // 默认值
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	matches, err := h.matchService.GetFinishedMatches(c.Request.Context(), limit)
	if err != nil {
		response.InternalError(c, "Failed to get finished matches")
		return
	}

	response.OK(c, "Finished matches retrieved successfully", matches)
}
