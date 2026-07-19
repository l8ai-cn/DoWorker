package sessionapi

import (
	"context"
	"errors"
	"net/http"

	podDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	domain "github.com/anthropics/agentsmesh/backend/internal/domain/agentsession"
	"github.com/anthropics/agentsmesh/backend/internal/middleware"
	"github.com/anthropics/agentsmesh/backend/internal/service/agentpod"
	runnerservice "github.com/anthropics/agentsmesh/backend/internal/service/runner"
	"github.com/gin-gonic/gin"
)

func (d *Deps) handleSwitchAgent(c *gin.Context) {
	row, pod, ok := d.authorizeSession(c, c.Param("id"))
	if !ok {
		return
	}
	tenant := middleware.GetTenant(c)
	if !d.requireSessionLevel(c, row, levelEdit) {
		return
	}
	var body struct {
		AgentID         string                 `json:"agent_id"`
		ModelResourceID *int64                 `json:"model_resource_id"`
		WorkerSpec      *sessionWorkerSpecBody `json:"worker_spec"`
		AutomationLevel string                 `json:"automation_level"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.AgentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "agent_id required"})
		return
	}
	if body.AgentID == row.AgentSlug {
		if hasSessionWorkerConfigChange(
			body.ModelResourceID,
			body.WorkerSpec,
			body.AutomationLevel,
		) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": rejectSameAgentWorkerConfigChangeMessage(),
				"code":  "validation_failed",
			})
			return
		}
		c.JSON(http.StatusOK, d.sessionWire(row, pod, row.RunnerNodeID))
		return
	}
	if d.sessionIsBusy(c.Param("id"), pod) {
		c.JSON(http.StatusConflict, gin.H{
			"error": gin.H{"code": "session_busy", "message": "Session is busy"},
		})
		return
	}
	newPod, err := d.rebuildSessionPod(c, row, pod, body.AgentID, tenant.OrganizationSlug, sessionWorkerPlanInput{
		WorkerSpec:      body.WorkerSpec,
		WorkerTypeSlug:  body.AgentID,
		ModelResourceID: body.ModelResourceID,
		AgentfileLayer:  acpAgentfileLayer(),
		AutomationLevel: body.AutomationLevel,
	})
	if err != nil {
		writeSessionPodError(c, err)
		return
	}
	row.AgentSlug = body.AgentID
	row.PodKey = newPod.PodKey
	if d.Updates != nil {
		d.Updates.NotifyChanged(row.ID)
	}
	c.JSON(http.StatusOK, d.sessionWire(row, newPod, row.RunnerNodeID))
}

func (d *Deps) sessionIsBusy(sessionID string, pod *podDomain.Pod) bool {
	if d.Hub != nil {
		if _, active := d.Hub.ActiveResponse(sessionID); active {
			return true
		}
	}
	switch mapSessionStatus(pod) {
	case "running", "waiting", "launching":
		return true
	default:
		return false
	}
}

func (d *Deps) rebuildSessionPod(
	c *gin.Context,
	row *domain.Session,
	pod *podDomain.Pod,
	agentSlug string,
	orgSlug string,
	plans ...sessionWorkerPlanInput,
) (*podDomain.Pod, error) {
	if d.PodOrchestrator == nil || d.Sessions == nil || d.PodCoordinator == nil {
		return nil, errSwitchUnavailable
	}
	runnerID := int64(0)
	if pod != nil {
		runnerID = pod.RunnerID
	}
	var orchReq *agentpod.OrchestrateCreatePodRequest
	if agentSlug == row.AgentSlug {
		snapshotID, snapshotErr := sessionSnapshotSource(row, pod)
		if snapshotErr != nil {
			return nil, snapshotErr
		}
		orchReq = buildSessionSnapshotRebuildPodRequest(row, runnerID, snapshotID)
	} else {
		if len(plans) == 0 {
			return nil, invalidSessionWorkerPlan("worker_spec", "is required")
		}
		draft, draftErr := d.buildFreshWorkerPlan(
			c.Request.Context(),
			row.OrganizationID,
			row.UserID,
			orgSlug,
			plans[0],
		)
		if draftErr != nil {
			return nil, draftErr
		}
		orchReq = buildSessionPlanRebuildPodRequest(row, runnerID, draft)
	}
	result, err := d.PodOrchestrator.CreatePod(c.Request.Context(), orchReq)
	if err != nil {
		return nil, err
	}
	if err := d.Sessions.UpdateAgentAndPod(c.Request.Context(), row.ID, agentSlug, result.Pod.PodKey); err != nil {
		cleanupErr := d.terminateCreatedSessionPod(c.Request.Context(), result.Pod.PodKey)
		return nil, joinSessionCompensationFailure(err, cleanupErr)
	}
	materialized := result
	result, err = d.PodOrchestrator.DispatchDeferredPod(c.Request.Context(), orchReq, materialized)
	if err != nil {
		return nil, d.restorePreviousSessionPod(c.Request.Context(), row, materialized.Pod.PodKey, err)
	}
	if pod != nil && pod.PodKey != "" {
		if err := d.terminateCreatedSessionPod(c.Request.Context(), pod.PodKey); err != nil {
			if d.Updates != nil {
				d.Updates.NotifyChanged(row.ID)
			}
			return nil, joinSessionCompensationFailure(
				errors.New("old session pod termination failed"),
				err,
			)
		}
	}
	if d.Stream != nil {
		d.Stream.PublishPodStatus(c.Request.Context(), result.Pod.PodKey, result.Pod.Status, result.Pod.AgentStatus)
	}
	return result.Pod, nil
}

func (d *Deps) restorePreviousSessionPod(
	ctx context.Context,
	row *domain.Session,
	newPodKey string,
	cause error,
) error {
	cleanupCtx, cancel := sessionCompensationContext(ctx)
	defer cancel()
	restoreErr := d.Sessions.UpdateAgentAndPod(
		cleanupCtx,
		row.ID,
		row.AgentSlug,
		row.PodKey,
	)
	newPodErr := d.terminateSessionPod(cleanupCtx, newPodKey)
	if d.Updates != nil {
		d.Updates.NotifyChanged(row.ID)
	}
	return joinSessionCompensationFailure(cause, errors.Join(restoreErr, newPodErr))
}

var errSwitchUnavailable = errors.New("switch unavailable")

func writeSessionPodError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, errSessionCompensationFailed):
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "session cleanup failed",
			"code":  "session_compensation_failed",
		})
	case errors.Is(err, agentpod.ErrNoAvailableRunner):
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "runner unavailable", "code": "runner_unavailable"})
	case errors.Is(err, agentpod.ErrRunnerDispatchFailed):
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error(), "code": "runner_dispatch_failed"})
	case errors.Is(err, runnerservice.ErrRunnerNotConnected), errors.Is(err, runnerservice.ErrRunnerOffline):
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "runner unavailable", "code": "runner_unavailable"})
	case errors.Is(err, errSwitchUnavailable):
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "unavailable"})
	default:
		writeOrchestratorError(c, err)
	}
}
