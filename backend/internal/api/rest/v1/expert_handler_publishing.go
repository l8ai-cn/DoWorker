package v1

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/anthropics/agentsmesh/backend/internal/middleware"
	agentpodSvc "github.com/anthropics/agentsmesh/backend/internal/service/agentpod"
	expertSvc "github.com/anthropics/agentsmesh/backend/internal/service/expert"
	"github.com/anthropics/agentsmesh/backend/pkg/apierr"
)

func (h *ExpertHandler) RunExpert(c *gin.Context) {
	tenant := middleware.GetTenant(c)
	var request runExpertRequest
	if err := c.ShouldBindJSON(&request); err != nil &&
		c.Request.ContentLength > 0 {
		apierr.ValidationError(c, err.Error())
		return
	}
	result, err := h.service.Run(
		c.Request.Context(),
		&expertSvc.RunExpertRequest{
			OrganizationID: tenant.OrganizationID,
			UserID:         tenant.UserID,
			ExpertSlug:     c.Param("expertSlug"),
			Alias:          request.Alias,
			PromptOverride: request.PromptOverride,
			Cols:           request.Cols,
			Rows:           request.Rows,
		},
	)
	if err != nil {
		h.runError(c, err)
		return
	}
	if result.Warning != "" {
		c.JSON(
			http.StatusCreated,
			gin.H{"pod": result.Pod, "warning": result.Warning},
		)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"pod": result.Pod})
}

func (h *ExpertHandler) runError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, expertSvc.ErrExpertRepublishRequired),
		errors.Is(err, expertSvc.ErrWorkerSpecSnapshotMismatch),
		errors.Is(err, agentpodSvc.ErrWorkerSpecSnapshotMismatch):
		apierr.Conflict(
			c,
			apierr.EXPERT_REPUBLISH_REQUIRED,
			"Expert must be republished from a valid WorkerSpec-backed Pod",
		)
	case errors.Is(err, expertSvc.ErrWorkerSpecSnapshotUnavailable),
		errors.Is(err, agentpodSvc.ErrWorkerSpecSnapshotUnavailable):
		apierr.ServiceUnavailable(
			c,
			apierr.SERVICE_UNAVAILABLE,
			"WorkerSpec snapshot service is unavailable",
		)
	default:
		mapOrchestratorErrorToHTTP(c, err)
	}
}

func (h *ExpertHandler) PublishFromPod(c *gin.Context) {
	tenant := middleware.GetTenant(c)
	var request publishExpertRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		apierr.ValidationError(c, err.Error())
		return
	}
	row, err := h.service.PublishFromPod(
		c.Request.Context(),
		&expertSvc.PublishFromPodRequest{
			OrganizationID: tenant.OrganizationID,
			UserID:         tenant.UserID,
			PodKey:         c.Param("pod_key"),
			Name:           request.Name,
			Slug:           request.Slug,
			Description:    request.Description,
		},
	)
	if err != nil {
		h.publishError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"expert": row})
}

func (h *ExpertHandler) publishError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, expertSvc.ErrPodAccessDenied):
		apierr.Forbidden(
			c,
			apierr.SOURCE_POD_ACCESS_DENIED,
			"Pod belongs to a different organization",
		)
	case errors.Is(err, expertSvc.ErrPodWorkerSpecSnapshotRequired),
		errors.Is(err, expertSvc.ErrWorkerSpecSnapshotMismatch):
		apierr.Conflict(
			c,
			apierr.EXPERT_WORKER_SPEC_REQUIRED,
			"Pod must have a valid WorkerSpec snapshot before publishing",
		)
	case errors.Is(err, expertSvc.ErrWorkerSpecSnapshotUnavailable):
		apierr.ServiceUnavailable(
			c,
			apierr.SERVICE_UNAVAILABLE,
			"WorkerSpec snapshot service is unavailable",
		)
	default:
		h.validationOrInternal(c, err)
	}
}

func (h *ExpertHandler) InstallMarketApplication(c *gin.Context) {
	tenant := middleware.GetTenant(c)
	row, alreadyInstalled, err := h.service.InstallMarketApplication(
		c.Request.Context(),
		tenant.OrganizationID,
		tenant.UserID,
		c.Param("marketSlug"),
	)
	if err != nil {
		if errors.Is(err, expertSvc.ErrMarketApplicationNotFound) {
			apierr.ResourceNotFound(c, "Market application not found")
			return
		}
		apierr.InternalError(c, "Failed to install market application")
		return
	}
	status := http.StatusCreated
	if alreadyInstalled {
		status = http.StatusOK
	}
	c.JSON(
		status,
		gin.H{"expert": row, "already_installed": alreadyInstalled},
	)
}
