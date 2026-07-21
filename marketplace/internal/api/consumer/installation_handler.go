package consumer

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	actorapi "github.com/l8ai-cn/agentcloud/marketplace/internal/api/actor"
	"github.com/l8ai-cn/agentcloud/marketplace/internal/service"
	"github.com/gin-gonic/gin"
)

type InstallationOrchestrator interface {
	CreatePlan(
		context.Context,
		service.CreateInstallationPlanCommand,
	) (service.InstallationPlanResult, error)
	Apply(
		context.Context,
		service.ApplyInstallationCommand,
	) (service.ApplyResult, error)
	GetOperation(context.Context, string, int64) (service.ApplyResult, error)
}

type InstallationHandler struct {
	orchestration InstallationOrchestrator
}

func NewInstallationHandler(orchestration InstallationOrchestrator) *InstallationHandler {
	return &InstallationHandler{orchestration: orchestration}
}

func (h *InstallationHandler) RegisterRoutes(router *gin.RouterGroup) {
	router.POST("/markets/:marketSlug/listings/:listingSlug/plans", h.createPlan)
	operations := router.Group("/installation-operations/:operationID")
	operations.POST("/apply", h.apply)
	operations.GET("", h.getOperation)
}

type createPlanRequest struct {
	ListingVersionID             string          `json:"listing_version_id"`
	TargetPlatformOrganizationID string          `json:"target_platform_organization_id"`
	RequestedConfiguration       json.RawMessage `json:"requested_configuration"`
}

func (h *InstallationHandler) createPlan(c *gin.Context) {
	current, ok := actorapi.FromContext(c)
	if !ok {
		writeInstallationError(c, service.ErrInvalidInstallationRequest)
		return
	}
	var request createPlanRequest
	if c.ShouldBindJSON(&request) != nil {
		writeInstallationError(c, service.ErrInvalidInstallationRequest)
		return
	}
	versionID, versionErr := strconv.ParseInt(request.ListingVersionID, 10, 64)
	targetOrgID, orgErr := strconv.ParseInt(request.TargetPlatformOrganizationID, 10, 64)
	if versionErr != nil || orgErr != nil {
		writeInstallationError(c, service.ErrInvalidInstallationRequest)
		return
	}
	result, err := h.orchestration.CreatePlan(c, service.CreateInstallationPlanCommand{
		MarketSlug: c.Param("marketSlug"), ListingSlug: c.Param("listingSlug"),
		ListingVersionID: versionID, TargetOrganizationID: targetOrgID,
		ActorUserID:            current.UserID,
		RequestedConfiguration: request.RequestedConfiguration,
	})
	if err != nil {
		writeInstallationError(c, err)
		return
	}
	c.JSON(http.StatusCreated, mapPlanResult(result))
}

type applyRequest struct {
	PlanID     string `json:"plan_id"`
	PlanDigest string `json:"plan_digest"`
}

func (h *InstallationHandler) apply(c *gin.Context) {
	current, ok := actorapi.FromContext(c)
	if !ok {
		writeInstallationError(c, service.ErrInvalidInstallationRequest)
		return
	}
	var request applyRequest
	if c.ShouldBindJSON(&request) != nil {
		writeInstallationError(c, service.ErrInvalidInstallationRequest)
		return
	}
	result, err := h.orchestration.Apply(c, service.ApplyInstallationCommand{
		OperationID: c.Param("operationID"), PlanID: request.PlanID,
		PlanDigest:     request.PlanDigest,
		IdempotencyKey: strings.TrimSpace(c.GetHeader("Idempotency-Key")),
		ActorUserID:    current.UserID,
	})
	if err != nil {
		if errors.Is(err, service.ErrTargetOrganizationForbidden) ||
			errors.Is(err, service.ErrRuntimeAuthorizationFailed) {
			writeInstallationError(c, err)
			return
		}
		if result.Status == service.ApplyFailed {
			c.JSON(http.StatusBadGateway, mapApplyResult(result))
			return
		}
		writeInstallationError(c, err)
		return
	}
	c.JSON(http.StatusOK, mapApplyResult(result))
}

func (h *InstallationHandler) getOperation(c *gin.Context) {
	current, ok := actorapi.FromContext(c)
	if !ok {
		writeInstallationError(c, service.ErrInvalidInstallationRequest)
		return
	}
	result, err := h.orchestration.GetOperation(
		c,
		c.Param("operationID"),
		current.UserID,
	)
	if err != nil {
		writeInstallationError(c, err)
		return
	}
	c.JSON(http.StatusOK, mapApplyResult(result))
}

func writeInstallationError(c *gin.Context, err error) {
	status, code, message := http.StatusInternalServerError,
		"INTERNAL_ERROR", "市场服务暂时不可用"
	switch {
	case errors.Is(err, service.ErrInvalidInstallationRequest):
		status, code, message = http.StatusBadRequest,
			"INVALID_INSTALLATION_REQUEST", "安装请求无效"
	case errors.Is(err, service.ErrApprovalRequired):
		status, code, message = http.StatusConflict,
			"APPROVAL_REQUIRED", "此应用需要先申请授权"
	case errors.Is(err, service.ErrGrantRequired):
		status, code, message = http.StatusForbidden,
			"GRANT_REQUIRED", "此应用仅支持管理员授权"
	case errors.Is(err, service.ErrTargetOrganizationForbidden):
		status, code, message = http.StatusForbidden,
			"TARGET_ORGANIZATION_FORBIDDEN", "你无权在这个组织中启用应用"
	case errors.Is(err, service.ErrQuotaAccountNotFound):
		status, code, message = http.StatusConflict,
			"QUOTA_ACCOUNT_NOT_FOUND", "目标组织尚未配置市场额度"
	case errors.Is(err, service.ErrQuotaInsufficient):
		status, code, message = http.StatusConflict,
			"QUOTA_INSUFFICIENT", "市场额度不足"
	case errors.Is(err, service.ErrPlanExpired):
		status, code, message = http.StatusConflict,
			"PLAN_EXPIRED", "安装计划已过期，请重新检查"
	case errors.Is(err, service.ErrPlanMismatch):
		status, code, message = http.StatusConflict,
			"PLAN_MISMATCH", "安装计划已经变化，请重新检查"
	case errors.Is(err, service.ErrApplicationAlreadyInstalled):
		status, code, message = http.StatusConflict,
			"APPLICATION_ALREADY_INSTALLED", "此应用已在目标组织中启用"
	case errors.Is(err, service.ErrOperationNotFound):
		status, code, message = http.StatusNotFound,
			"INSTALLATION_OPERATION_NOT_FOUND", "找不到安装操作"
	case errors.Is(err, service.ErrListingNotFound):
		status, code, message = http.StatusNotFound,
			"LISTING_NOT_AVAILABLE", "此内容当前不可获取"
	}
	c.JSON(status, gin.H{"error": gin.H{"code": code, "message": message}})
}
