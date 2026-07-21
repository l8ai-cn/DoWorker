package v1

import (
	"net/http"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/agentpod"
	"github.com/l8ai-cn/agentcloud/backend/internal/middleware"
	"github.com/l8ai-cn/agentcloud/backend/pkg/apierr"
	"github.com/gin-gonic/gin"
)

type QueuedPodItem struct {
	PodKey        string `json:"pod_key"`
	RunnerID      int64  `json:"runner_id"`
	AgentSlug     string `json:"agent_slug"`
	Alias         string `json:"alias,omitempty"`
	QueuePosition int    `json:"queue_position"`
	CreatedAt     string `json:"created_at"`
	ExpiresAt     string `json:"expires_at"`
}

// ListQueuedPods GET /api/v1/orgs/:slug/pods/queued
func (h *PodHandler) ListQueuedPods(c *gin.Context) {
	tenant := middleware.GetTenant(c)
	ctx := c.Request.Context()
	pods, err := h.podService.ListQueuedPods(ctx, tenant.OrganizationID)
	if err != nil {
		apierr.InternalError(c, "failed to list queued pods")
		return
	}
	items := make([]QueuedPodItem, 0, len(pods))
	for _, pod := range pods {
		item := QueuedPodItem{
			PodKey:    pod.PodKey,
			RunnerID:  pod.RunnerID,
			AgentSlug: pod.AgentSlug,
			CreatedAt: pod.CreatedAt.UTC().Format(timeRFC3339),
		}
		if pod.Alias != nil {
			item.Alias = *pod.Alias
		}
		if h.pendingQueue != nil {
			if pos, err := h.pendingQueue.QueuePosition(ctx, pod.RunnerID, pod.PodKey); err == nil {
				item.QueuePosition = pos
			}
			if exp, err := h.pendingQueue.GetCreatePodExpiry(ctx, pod.PodKey); err == nil {
				item.ExpiresAt = exp.UTC().Format(timeRFC3339)
			}
		}
		items = append(items, item)
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

// CancelQueuedPod DELETE /api/v1/orgs/:slug/pods/:key/queue
func (h *PodHandler) CancelQueuedPod(c *gin.Context) {
	podKey := c.Param("key")
	tenant := middleware.GetTenant(c)
	ctx := c.Request.Context()

	pod, err := h.podService.GetPod(ctx, podKey)
	if err != nil || pod == nil {
		apierr.NotFound(c, apierr.RESOURCE_NOT_FOUND, "Pod not found")
		return
	}
	if pod.OrganizationID != tenant.OrganizationID {
		apierr.ForbiddenAccess(c)
		return
	}
	if pod.Status != agentpod.StatusQueued {
		apierr.Conflict(c, "NOT_QUEUED", "Pod is not queued")
		return
	}
	if !h.canCancelQueuedPod(tenant, pod) {
		apierr.ForbiddenAccess(c)
		return
	}
	if h.podCoordinator == nil {
		apierr.InternalError(c, "pod coordinator unavailable")
		return
	}
	if err := h.podCoordinator.TerminatePod(ctx, podKey); err != nil {
		apierr.InternalError(c, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"pod_key": podKey, "status": agentpod.StatusCompleted})
}

const timeRFC3339 = "2006-01-02T15:04:05Z07:00"
