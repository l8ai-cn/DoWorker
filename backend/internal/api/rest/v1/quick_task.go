package v1

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	control "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
	"github.com/anthropics/agentsmesh/backend/internal/middleware"
	agentsvc "github.com/anthropics/agentsmesh/backend/internal/service/agentpod"
	controlservice "github.com/anthropics/agentsmesh/backend/internal/service/orchestrationcontrol"
	"github.com/anthropics/agentsmesh/backend/pkg/apierr"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type pendingQueueReader interface {
	QueuePosition(ctx context.Context, runnerID int64, podKey string) (int, error)
	GetCreatePodExpiry(ctx context.Context, podKey string) (time.Time, error)
}

// CreateQuickTask POST /api/v1/orgs/:slug/quick-tasks
func (h *PodHandler) CreateQuickTask(c *gin.Context) {
	var req QuickTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierr.ValidationError(c, err.Error())
		return
	}
	planID, err := canonicalQuickTaskPlanID(req.PlanID)
	if err != nil {
		apierr.BadRequest(c, apierr.WORKER_PLAN_INVALID, "plan_id must be a canonical UUID")
		return
	}
	if h.quickTaskPlanAuthorizer == nil ||
		h.quickTaskPlanApplier == nil ||
		h.quickTaskPodReader == nil {
		apierr.ServiceUnavailable(c, apierr.WORKER_APPLY_UNAVAILABLE, "Worker apply service is unavailable")
		return
	}
	tenant := middleware.GetTenant(c)
	if tenant == nil {
		apierr.InternalError(c, "Organization scope is unavailable")
		return
	}
	ctx := c.Request.Context()
	scope := control.Scope{
		OrganizationID:   tenant.OrganizationID,
		OrganizationSlug: slugkit.Slug(tenant.OrganizationSlug),
		ActorID:          tenant.UserID,
	}
	if err := scope.Validate(); err != nil {
		apierr.InternalError(c, "Organization scope is unavailable")
		return
	}
	if err := h.quickTaskPlanAuthorizer.AuthorizeApply(ctx, scope, planID); err != nil {
		mapQuickTaskError(c, err)
		return
	}
	result, err := h.quickTaskPlanApplier.Apply(ctx, scope, planID)
	if err != nil {
		mapQuickTaskError(c, err)
		return
	}
	pod, err := h.quickTaskPodReader.GetPod(ctx, result.PodKey)
	if err != nil || pod == nil || pod.PodKey != result.PodKey {
		apierr.InternalError(c, "Failed to load applied Worker Pod")
		return
	}
	resp := QuickTaskResponse{
		PodKey: result.PodKey,
		Status: pod.Status,
	}
	if pod.Status == agentpod.StatusQueued && h.pendingQueue != nil {
		if pos, posErr := h.pendingQueue.QueuePosition(ctx, result.RunnerID, result.PodKey); posErr == nil {
			resp.QueuePosition = pos
		}
		if exp, expErr := h.pendingQueue.GetCreatePodExpiry(ctx, result.PodKey); expErr == nil {
			resp.ExpiresAt = exp.UTC().Format(time.RFC3339)
		}
	}
	c.JSON(http.StatusAccepted, resp)
}

func canonicalQuickTaskPlanID(value string) (string, error) {
	parsed, err := uuid.Parse(value)
	if err != nil || parsed == uuid.Nil || parsed.String() != value {
		return "", control.ErrInvalid
	}
	return value, nil
}

func mapQuickTaskError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, control.ErrInvalid):
		apierr.BadRequest(c, apierr.WORKER_PLAN_INVALID, "Worker plan is invalid")
	case errors.Is(err, control.ErrNotFound):
		apierr.NotFound(c, apierr.WORKER_PLAN_NOT_FOUND, "Worker plan was not found")
	case errors.Is(err, control.ErrConflict),
		errors.Is(err, control.ErrStale),
		errors.Is(err, control.ErrExpired),
		errors.Is(err, control.ErrConsumed):
		apierr.Conflict(c, apierr.WORKER_PLAN_STATE_CHANGED, "Worker plan state changed; create a new plan")
	case errors.Is(err, controlservice.ErrForbidden):
		apierr.Forbidden(c, apierr.ACCESS_DENIED, "Worker plan access is forbidden")
	case errors.Is(err, controlservice.ErrUnavailable):
		apierr.ServiceUnavailable(c, apierr.WORKER_APPLY_UNAVAILABLE, "Worker apply service is unavailable")
	case errors.Is(err, agentsvc.ErrNoAvailableRunner):
		apierr.Respond(c, http.StatusUnprocessableEntity, "NO_RUNNER_FOR_AGENT", "No runner available for this agent")
	case errors.Is(err, agentpod.ErrQueueFull):
		apierr.Respond(c, http.StatusTooManyRequests, "QUEUE_FULL", "Runner pending queue is full")
	default:
		apierr.InternalError(c, "Failed to apply Worker plan")
	}
}
