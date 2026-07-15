package v1

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	expertdom "github.com/anthropics/agentsmesh/backend/internal/domain/expert"
	"github.com/anthropics/agentsmesh/backend/internal/domain/expertmarket"
	"github.com/anthropics/agentsmesh/backend/internal/middleware"
	expertSvc "github.com/anthropics/agentsmesh/backend/internal/service/expert"
	specservice "github.com/anthropics/agentsmesh/backend/internal/service/workerspec"
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
		if errors.Is(err, expertSvc.ErrExpertRepublishRequired) {
			apierr.Conflict(
				c,
				apierr.EXPERT_REPUBLISH_REQUIRED,
				"Expert must be republished from a WorkerSpec-backed Pod",
			)
			return
		}
		mapOrchestratorErrorToHTTP(c, err)
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
	var request installMarketApplicationRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		apierr.ValidationError(c, err.Error())
		return
	}
	row, alreadyInstalled, err := h.service.InstallPublishedMarketApplication(
		c.Request.Context(),
		expertSvc.InstallMarketApplicationRequest{
			OrganizationID:  tenant.OrganizationID,
			UserID:          tenant.UserID,
			ModelResourceID: request.ModelResourceID,
			MarketSlug:      c.Param("marketSlug"),
		},
	)
	if err != nil {
		h.installMarketError(c, err)
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

func (h *ExpertHandler) installMarketError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, expertSvc.ErrMarketApplicationNotFound):
		apierr.ResourceNotFound(c, "Market application not found")
	case errors.Is(err, expertSvc.ErrMarketUnavailable):
		apierr.ServiceUnavailable(
			c,
			apierr.SERVICE_UNAVAILABLE,
			"Expert application market is unavailable",
		)
	case errors.Is(err, specservice.ErrInvalidDraft):
		apierr.BadRequest(
			c,
			apierr.INVALID_INPUT,
			"Selected model resource is unavailable or incompatible",
		)
	case errors.Is(err, expertSvc.ErrMarketSnapshotInvalid),
		errors.Is(err, expertSvc.ErrMarketReleaseNotPublished),
		errors.Is(err, expertmarket.ErrConflict),
		errors.Is(err, expertdom.ErrConflict):
		apierr.Conflict(
			c,
			apierr.ALREADY_EXISTS,
			"Market application changed; reload and try again",
		)
	case errors.Is(err, expertdom.ErrNotFound):
		apierr.ResourceNotFound(c, "Installed expert not found")
	default:
		apierr.InternalError(c, "Failed to install market application")
	}
}
