package handlers

import (
	"net/http"
	"strconv"

	"backend-go/internal/core/ports"
	"backend-go/pkg/response"
	"github.com/gin-gonic/gin"
)

// TeamHandler 战队管理
type TeamHandler struct {
	service ports.TeamService
}

// NewTeamHandler 创建处理器
func NewTeamHandler(service ports.TeamService) *TeamHandler {
	return &TeamHandler{service: service}
}

// ListTeams 列出战队
func (h *TeamHandler) ListTeams(c *gin.Context) {
	includeInactive := c.Query("all") == "1" || c.Query("include_inactive") == "1"
	teams, err := h.service.ListTeams(c.Request.Context(), includeInactive)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取战队列表失败", err.Error())
		return
	}
	response.Success(c, http.StatusOK, "ok", teams)
}

// CreateTeam 创建战队
func (h *TeamHandler) CreateTeam(c *gin.Context) {
	var req struct {
		Name      string `json:"name" binding:"required"`
		ShortName string `json:"shortName"`
		LogoURL   string `json:"logoUrl"`
		IsActive  bool   `json:"isActive"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err.Error())
		return
	}

	team, err := h.service.CreateTeam(c.Request.Context(), &ports.CreateTeamRequest{
		Name:      req.Name,
		ShortName: req.ShortName,
		LogoURL:   req.LogoURL,
		IsActive:  req.IsActive,
	})
	if err != nil {
		response.Error(c, http.StatusBadRequest, "创建战队失败", err.Error())
		return
	}
	response.Success(c, http.StatusCreated, "创建成功", team)
}

// GetTeam 获取战队
func (h *TeamHandler) GetTeam(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无效的ID", "")
		return
	}
	team, err := h.service.GetTeam(c.Request.Context(), uint(id))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取战队失败", err.Error())
		return
	}
	if team == nil {
		response.Error(c, http.StatusNotFound, "战队不存在", "")
		return
	}
	response.Success(c, http.StatusOK, "ok", team)
}

// UpdateTeam 更新战队
func (h *TeamHandler) UpdateTeam(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无效的ID", "")
		return
	}
	var req struct {
		Name      string `json:"name" binding:"required"`
		ShortName string `json:"shortName"`
		LogoURL   string `json:"logoUrl"`
		IsActive  bool   `json:"isActive"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err.Error())
		return
	}
	team, err := h.service.UpdateTeam(c.Request.Context(), uint(id), &ports.UpdateTeamRequest{
		Name:      req.Name,
		ShortName: req.ShortName,
		LogoURL:   req.LogoURL,
		IsActive:  req.IsActive,
	})
	if err != nil {
		response.Error(c, http.StatusBadRequest, "更新战队失败", err.Error())
		return
	}
	response.Success(c, http.StatusOK, "更新成功", team)
}

// DeleteTeam 删除战队
func (h *TeamHandler) DeleteTeam(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无效的ID", "")
		return
	}
	if err := h.service.DeleteTeam(c.Request.Context(), uint(id)); err != nil {
		response.Error(c, http.StatusInternalServerError, "删除战队失败", err.Error())
		return
	}
	response.Success(c, http.StatusOK, "删除成功", nil)
}
