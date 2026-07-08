package v1

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	"github.com/anthropics/agentsmesh/backend/internal/middleware"
	agentsvc "github.com/anthropics/agentsmesh/backend/internal/service/agentpod"
	"github.com/anthropics/agentsmesh/backend/pkg/apierr"
	"github.com/gin-gonic/gin"
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
	prompt := strings.TrimSpace(req.Prompt)
	if prompt == "" || len(prompt) > quickTaskPromptMaxLen {
		apierr.BadRequest(c, apierr.VALIDATION_FAILED, "prompt must be 1..10000 characters")
		return
	}
	ttlMinutes := req.QueueTTLMinutes
	if ttlMinutes == 0 {
		ttlMinutes = 30
	}
	if ttlMinutes < 1 || ttlMinutes > 1440 {
		apierr.BadRequest(c, apierr.VALIDATION_FAILED, "queue_ttl_minutes must be between 1 and 1440")
		return
	}
	tenant := middleware.GetTenant(c)
	ctx := c.Request.Context()

	agentSlug := strings.TrimSpace(req.AgentSlug)
	if agentSlug == "" {
		slug, err := h.runnerService.FirstAvailableAgentSlug(ctx, tenant.OrganizationID, tenant.UserID)
		if err != nil {
			apierr.Respond(c, http.StatusUnprocessableEntity, "NO_RUNNER_FOR_AGENT", "No runner available for any agent in this organization")
			return
		}
		agentSlug = slug
	}

	runnerID := req.RunnerID
	if runnerID == 0 {
		selected, err := h.runnerService.SelectRunnerWithAffinity(ctx, tenant.OrganizationID, tenant.UserID, agentSlug, nil, nil)
		if err != nil {
			fallback, fbErr := h.runnerService.MostRecentRunnerForAgent(ctx, tenant.OrganizationID, tenant.UserID, agentSlug)
			if fbErr != nil {
				apierr.Respond(c, http.StatusUnprocessableEntity, "NO_RUNNER_FOR_AGENT", "No runner available for this agent")
				return
			}
			runnerID = fallback.ID
		} else {
			runnerID = selected.ID
		}
	}

	layer := buildQuickTaskAgentfileLayer(prompt)
	var alias *string
	if trimmed := strings.TrimSpace(req.Alias); trimmed != "" {
		if len(trimmed) > quickTaskAliasMaxLen {
			apierr.BadRequest(c, apierr.VALIDATION_FAILED, "alias must be at most 100 characters")
			return
		}
		alias = &trimmed
	}

	orchReq := &agentsvc.OrchestrateCreatePodRequest{
		OrganizationID:     tenant.OrganizationID,
		UserID:             tenant.UserID,
		RunnerID:           runnerID,
		AgentSlug:          agentSlug,
		RepositoryID:       req.RepositoryID,
		Alias:              alias,
		AgentfileLayer:     &layer,
		Cols:               120,
		Rows:               40,
		QueueIfUnavailable: true,
		QueueTTL:           time.Duration(ttlMinutes) * time.Minute,
	}

	result, err := h.orchestrator.CreatePod(ctx, orchReq)
	if err != nil {
		mapQuickTaskError(c, err)
		return
	}

	resp := QuickTaskResponse{PodKey: result.Pod.PodKey}
	if result.Queued {
		resp.Status = agentpod.StatusQueued
		if h.pendingQueue != nil {
			if pos, posErr := h.pendingQueue.QueuePosition(ctx, runnerID, result.Pod.PodKey); posErr == nil {
				resp.QueuePosition = pos
			}
			if exp, expErr := h.pendingQueue.GetCreatePodExpiry(ctx, result.Pod.PodKey); expErr == nil {
				resp.ExpiresAt = exp.UTC().Format(time.RFC3339)
			}
		}
	} else {
		resp.Status = agentpod.StatusInitializing
	}
	c.JSON(http.StatusAccepted, resp)
}

func mapQuickTaskError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, agentsvc.ErrConfigBuildFailed):
		apierr.Respond(c, http.StatusInternalServerError, "CONFIG_BUILD_FAILED", "Failed to build pod configuration")
	case errors.Is(err, agentsvc.ErrMissingAgentSlug):
		apierr.NotFound(c, "AGENT_NOT_FOUND", "Agent not found")
	case errors.Is(err, agentsvc.ErrNoAvailableRunner):
		apierr.Respond(c, http.StatusUnprocessableEntity, "NO_RUNNER_FOR_AGENT", "No runner available for this agent")
	case errors.Is(err, agentpod.ErrQueueFull):
		apierr.Respond(c, http.StatusTooManyRequests, "QUEUE_FULL", "Runner pending queue is full")
	default:
		mapOrchestratorErrorToHTTP(c, err)
	}
}
