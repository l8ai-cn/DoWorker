package v1

import (
	"encoding/json"
	"net/http"

	sessionDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentsession"
	"github.com/anthropics/agentsmesh/backend/internal/middleware"
	"github.com/anthropics/agentsmesh/backend/pkg/apierr"
	"github.com/gin-gonic/gin"
)

// CreatePodRequest retains rejected fresh-create fields so the cutover fails
// explicitly instead of silently dropping runtime configuration.
type CreatePodRequest struct {
	AgentSlug          string                     `json:"agent_slug"`
	RunnerID           int64                      `json:"runner_id"`
	TicketSlug         *string                    `json:"ticket_slug"`
	Alias              *string                    `json:"alias"`
	AgentfileLayer     *string                    `json:"agentfile_layer"`
	AutomationLevel    string                     `json:"automation_level"`
	RepositoryID       *int64                     `json:"repository_id,omitempty"`
	ModelResourceID    *int64                     `json:"model_resource_id,omitempty"`
	TokenBudget        *int64                     `json:"token_budget,omitempty"`
	Cols               int32                      `json:"cols"`
	Rows               int32                      `json:"rows"`
	SourcePodKey       string                     `json:"source_pod_key"`
	ResumeAgentSession *bool                      `json:"resume_agent_session"`
	Perpetual          *bool                      `json:"perpetual"`
	KnowledgeMounts    []PodKnowledgeMountRequest `json:"knowledge_mounts,omitempty"`
	QueueIfOffline     bool                       `json:"queue_if_offline"`
	QueueTTLMinutes    int                        `json:"queue_ttl_minutes"`
}

type PodKnowledgeMountRequest struct {
	Slug string `json:"slug"`
	Mode string `json:"mode,omitempty"`
}

// CreatePod resumes a Worker from immutable source lineage.
// POST /api/v1/orgs/:slug/pods
// POST /api/v1/ext/orgs/:slug/workers  (external API, API key auth)
// POST /api/v1/ext/orgs/:slug/pods     (route alias)
func (h *PodHandler) CreatePod(c *gin.Context) {
	var req CreatePodRequest
	raw, err := c.GetRawData()
	if err != nil {
		apierr.ValidationError(c, err.Error())
		return
	}
	if field, ok := legacyPodCreateModelField(raw); ok {
		apierr.BadRequest(c, apierr.VALIDATION_FAILED, field+" is no longer supported; use model_resource_id")
		return
	}
	if err := json.Unmarshal(raw, &req); err != nil {
		apierr.ValidationError(c, err.Error())
		return
	}

	if req.SourcePodKey == "" {
		apierr.Conflict(
			c,
			apierr.WORKER_RESOURCE_APPLY_REQUIRED,
			"Fresh Worker creation must use orchestration validate-plan-apply",
		)
		return
	}
	if hasRESTResumeRuntimeOverrides(req) {
		apierr.Conflict(
			c,
			apierr.WORKER_RESUME_LINEAGE_ONLY,
			"Worker resume accepts source lineage only; runtime overrides are not supported",
		)
		return
	}
	if h.orchestrator == nil {
		apierr.ServiceUnavailable(
			c,
			apierr.SERVICE_UNAVAILABLE,
			"Pod orchestrator is not configured",
		)
		return
	}

	tenant := middleware.GetTenant(c)
	orchReq := buildRESTResumePodRequest(req, tenant)
	result, err := h.orchestrator.CreatePod(c.Request.Context(), orchReq)
	if err != nil {
		mapOrchestratorErrorToHTTP(c, err)
		return
	}

	if result.Warning != "" {
		c.JSON(http.StatusCreated, gin.H{
			"pod":     result.Pod,
			"warning": result.Warning,
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"pod": result.Pod})
}
