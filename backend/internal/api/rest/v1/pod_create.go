package v1

import (
	"net/http"
	"strings"
	"time"

	"github.com/anthropics/agentsmesh/backend/internal/middleware"
	"github.com/anthropics/agentsmesh/backend/internal/service/agentpod"
	"github.com/anthropics/agentsmesh/backend/pkg/apierr"
	"github.com/gin-gonic/gin"
)

// CreatePodRequest represents pod creation request. After the EnvBundle
// refactor every aspect of pod configuration — including which credential
// bundle to mount — is expressed inside `agentfile_layer`. The legacy
// `credential_profile_id` field has been removed; clients should emit
// `USE_ENV_BUNDLE "name"` in the layer instead.
type CreatePodRequest struct {
	AgentSlug  string  `json:"agent_slug"`  // Required: determines base AgentFile
	RunnerID   int64   `json:"runner_id"`   // Optional: auto-select if omitted
	TicketSlug *string `json:"ticket_slug"` // Optional: associate with ticket
	Alias      *string `json:"alias"`       // Optional: display name (max 100 chars)

	// AgentFile Layer — SSOT for all pod configuration (MODE, CONFIG, REPO,
	// BRANCH, USE_ENV_BUNDLE, PROMPT).
	AgentfileLayer *string `json:"agentfile_layer"`

	// AutomationLevel is the unified permission/automation tier
	// (interactive/auto_edit/autonomous). Empty ⇒ autonomous default, so every
	// Worker is automatable unless the caller downgrades it.
	AutomationLevel string `json:"automation_level"`

	// Platform-level ID references (cannot be expressed as AgentFile declarations)
	RepositoryID *int64 `json:"repository_id,omitempty"`

	// Worker model binding (quota/billing). Bind either a virtual API key
	// (usage attributed to it) or a real ai_models row directly. TokenBudget
	// is an optional per-Worker hint.
	VirtualAPIKeyID *int64 `json:"virtual_api_key_id,omitempty"`
	ModelConfigID   *int64 `json:"model_config_id,omitempty"`
	TokenBudget     *int64 `json:"token_budget,omitempty"`

	// Terminal size (from browser xterm.js)
	Cols int32 `json:"cols"`
	Rows int32 `json:"rows"`

	// Resume related fields
	SourcePodKey       string `json:"source_pod_key"`
	ResumeAgentSession *bool  `json:"resume_agent_session"`

	// Perpetual mode: Runner auto-restarts agent on clean exit
	Perpetual *bool `json:"perpetual"`

	// Knowledge base mounts; win over Agentfile KNOWLEDGE and agent defaults
	KnowledgeMounts []PodKnowledgeMountRequest `json:"knowledge_mounts,omitempty"`

	QueueIfOffline  bool `json:"queue_if_offline"`
	QueueTTLMinutes int  `json:"queue_ttl_minutes"`
}

// PodKnowledgeMountRequest selects one org knowledge base for the new pod.
type PodKnowledgeMountRequest struct {
	Slug string `json:"slug"`
	Mode string `json:"mode,omitempty"` // ro | rw; empty defaults to ro
}

// CreatePod creates a new pod (Worker)
// POST /api/v1/orgs/:slug/pods
// POST /api/v1/ext/orgs/:slug/workers  (external API, API key auth)
// POST /api/v1/ext/orgs/:slug/pods     (legacy alias)
// Supports Resume mode when source_pod_key is provided
func (h *PodHandler) CreatePod(c *gin.Context) {
	var req CreatePodRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierr.ValidationError(c, err.Error())
		return
	}

	tenant := middleware.GetTenant(c)

	// Normalize alias: empty string → nil, validate length
	if req.Alias != nil {
		trimmed := strings.TrimSpace(*req.Alias)
		if trimmed == "" {
			req.Alias = nil
		} else if len(trimmed) > 100 {
			apierr.BadRequest(c, apierr.VALIDATION_FAILED, "Alias must be 100 characters or less")
			return
		} else {
			req.Alias = &trimmed
		}
	}

	// Build orchestration request (protocol adaptation: HTTP → service layer)
	orchReq := &agentpod.OrchestrateCreatePodRequest{
		OrganizationID:     tenant.OrganizationID,
		UserID:             tenant.UserID,
		RunnerID:           req.RunnerID,
		AgentSlug:          req.AgentSlug,
		RepositoryID:       req.RepositoryID,
		TicketSlug:         req.TicketSlug,
		Alias:              req.Alias,
		AgentfileLayer:     req.AgentfileLayer,
		AutomationLevel:    req.AutomationLevel,
		Cols:               req.Cols,
		Rows:               req.Rows,
		SourcePodKey:       req.SourcePodKey,
		ResumeAgentSession: req.ResumeAgentSession,
		Perpetual:          req.Perpetual != nil && *req.Perpetual,
		QueueIfUnavailable: req.QueueIfOffline,
		VirtualAPIKeyID:    req.VirtualAPIKeyID,
		ModelConfigID:      req.ModelConfigID,
		TokenBudget:        req.TokenBudget,
	}
	if req.QueueIfOffline {
		ttl := req.QueueTTLMinutes
		if ttl == 0 {
			ttl = 30
		}
		orchReq.QueueTTL = time.Duration(ttl) * time.Minute
	}
	for _, m := range req.KnowledgeMounts {
		orchReq.KnowledgeMounts = append(orchReq.KnowledgeMounts, agentpod.KnowledgeMountRequest{Slug: m.Slug, Mode: m.Mode})
	}

	result, err := h.orchestrator.CreatePod(c.Request.Context(), orchReq)
	if err != nil {
		mapOrchestratorErrorToHTTP(c, err)
		return
	}

	// Return result with optional warning
	if result.Warning != "" {
		c.JSON(http.StatusCreated, gin.H{
			"pod":     result.Pod,
			"warning": result.Warning,
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"pod": result.Pod})
}
