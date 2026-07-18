package v1

import (
	"errors"
	"net/http"

	expertdomain "github.com/anthropics/agentsmesh/backend/internal/domain/expert"
	workerspecdomain "github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	"github.com/anthropics/agentsmesh/backend/internal/middleware"
	"github.com/anthropics/agentsmesh/backend/pkg/apierr"
	"github.com/anthropics/agentsmesh/backend/pkg/policy"
	"github.com/gin-gonic/gin"
)

type podWorkerExpert struct {
	ID   int64  `json:"id"`
	Slug string `json:"slug"`
	Name string `json:"name"`
}

type podWorkerContext struct {
	SnapshotID int64            `json:"snapshot_id"`
	Alias      string           `json:"alias"`
	Expert     *podWorkerExpert `json:"expert,omitempty"`
	SkillSlugs []string         `json:"skill_slugs"`
}

func (h *PodHandler) GetPodWorkerContext(c *gin.Context) {
	if h.workerSpecs == nil {
		apierr.ServiceUnavailable(c, "worker_context_unavailable", "Worker context is not available")
		return
	}
	ctx := c.Request.Context()
	pod, err := h.podService.GetPod(ctx, c.Param("key"))
	if err != nil {
		apierr.ResourceNotFound(c, "Pod not found")
		return
	}
	tenant := middleware.GetTenant(c)
	subject := policy.NewSubject(tenant.OrganizationID, tenant.UserID, tenant.UserRole)
	resource := h.podResourceWithGrants(ctx, pod.PodKey, pod.OrganizationID, pod.CreatedByID)
	if !policy.PodPolicy.AllowRead(subject, resource) {
		apierr.ForbiddenAccess(c)
		return
	}
	if pod.WorkerSpecSnapshotID == nil {
		apierr.ResourceNotFound(c, "Worker snapshot not found")
		return
	}
	snapshot, err := h.workerSpecs.GetByID(ctx, tenant.OrganizationID, *pod.WorkerSpecSnapshotID)
	if errors.Is(err, workerspecdomain.ErrNotFound) {
		apierr.ResourceNotFound(c, "Worker snapshot not found")
		return
	}
	if err != nil {
		apierr.InternalError(c, "Failed to load Worker context")
		return
	}
	response := podWorkerContext{
		SnapshotID: snapshot.ID,
		Alias:      snapshot.Spec.Metadata.Alias,
		SkillSlugs: workerSnapshotSkillSlugs(snapshot),
	}
	if sourceID := snapshot.Spec.Metadata.SourceExpertID; sourceID != nil && h.experts != nil {
		expert, loadErr := h.experts.GetByID(ctx, tenant.OrganizationID, *sourceID)
		switch {
		case loadErr == nil:
			response.Expert = &podWorkerExpert{ID: expert.ID, Slug: expert.Slug, Name: expert.Name}
		case !errors.Is(loadErr, expertdomain.ErrNotFound):
			apierr.InternalError(c, "Failed to load Worker expert")
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{"worker": response})
}

func workerSnapshotSkillSlugs(snapshot workerspecdomain.Snapshot) []string {
	slugs := make([]string, 0, len(snapshot.Spec.Workspace.SkillPackages))
	for _, skill := range snapshot.Spec.Workspace.SkillPackages {
		slugs = append(slugs, skill.Slug)
	}
	return slugs
}
