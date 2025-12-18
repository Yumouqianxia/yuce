package admin

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"backend-go/internal/core/ports"
	"backend-go/pkg/response"
)

// AuditHandler 审计日志处理器
type AuditHandler struct {
	adminAuditService ports.AdminAuditService
	logger            *logrus.Logger
}

// NewAuditHandler 创建审计日志处理器
func NewAuditHandler(
	adminAuditService ports.AdminAuditService,
	logger *logrus.Logger,
) *AuditHandler {
	return &AuditHandler{
		adminAuditService: adminAuditService,
		logger:            logger,
	}
}

// GetAuditLog 获取审计日志详情
func (h *AuditHandler) GetAuditLog(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid audit log ID", err.Error())
		return
	}

	auditLog, err := h.adminAuditService.GetAuditLog(c.Request.Context(), uint(id))
	if err != nil {
		response.Error(c, http.StatusNotFound, "Audit log not found", err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Audit log retrieved successfully", auditLog)
}

// ListAuditLogs 获取审计日志列表
func (h *AuditHandler) ListAuditLogs(c *gin.Context) {
	var req ports.ListAuditLogsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request parameters", err.Error())
		return
	}

	result, err := h.adminAuditService.ListAuditLogs(c.Request.Context(), &req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list audit logs")
		response.Error(c, http.StatusInternalServerError, "Failed to list audit logs", err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Audit logs retrieved successfully", result)
}

// GetAuditStats 获取审计统计
func (h *AuditHandler) GetAuditStats(c *gin.Context) {
	var req ports.AuditStatsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request parameters", err.Error())
		return
	}

	stats, err := h.adminAuditService.GetAuditStats(c.Request.Context(), &req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get audit stats")
		response.Error(c, http.StatusInternalServerError, "Failed to get audit stats", err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Audit stats retrieved successfully", stats)
}
