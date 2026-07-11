package sessionapi

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	podDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	domain "github.com/anthropics/agentsmesh/backend/internal/domain/agentsession"
	"github.com/gin-gonic/gin"
)

func (d *Deps) handleGetCodexGoal(c *gin.Context) {
	row, _, ok := d.authorizeSession(c, c.Param("id"))
	if !ok {
		return
	}
	if !harnessSupportsCodexGoal(row.AgentSlug) {
		c.JSON(http.StatusOK, gin.H{"goal": nil})
		return
	}
	goal, err := d.Sessions.GetCodexGoal(c.Request.Context(), row.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "read failed"})
		return
	}
	if goal == nil {
		c.JSON(http.StatusOK, gin.H{"goal": nil})
		return
	}
	c.JSON(http.StatusOK, gin.H{"goal": codexGoalWire(goal)})
}

func (d *Deps) handlePutCodexGoal(c *gin.Context) {
	row, pod, ok := d.authorizeSession(c, c.Param("id"))
	if !ok || !d.requireSessionLevel(c, row, levelEdit) {
		return
	}
	if !harnessSupportsCodexGoal(row.AgentSlug) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_input", "detail": "harness mismatch"})
		return
	}
	var body struct {
		Objective   string  `json:"objective"`
		TokenBudget *int64  `json:"token_budget"`
		Status      *string `json:"status"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || strings.TrimSpace(body.Objective) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_input", "detail": "objective required"})
		return
	}
	goal := buildCodexGoal(pod, row.ID, body.Objective, body.TokenBudget, body.Status)
	if err := d.Sessions.SetCodexGoal(c.Request.Context(), row.ID, goal); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "persist failed"})
		return
	}
	d.forwardCodexGoalRPC(c, pod, "goal/set", goal)
	c.JSON(http.StatusOK, gin.H{"goal": codexGoalWire(goal)})
}

func (d *Deps) handlePatchCodexGoalStatus(c *gin.Context) {
	row, pod, ok := d.authorizeSession(c, c.Param("id"))
	if !ok || !d.requireSessionLevel(c, row, levelEdit) {
		return
	}
	if !harnessSupportsCodexGoal(row.AgentSlug) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_input", "detail": "harness mismatch"})
		return
	}
	var body struct {
		Status string `json:"status"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.Status == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	goal, err := d.Sessions.GetCodexGoal(c.Request.Context(), row.ID)
	if err != nil || goal == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "goal not found"})
		return
	}
	goal.Status = body.Status
	now := time.Now().Unix()
	goal.UpdatedAt = &now
	_ = d.Sessions.SetCodexGoal(c.Request.Context(), row.ID, goal)
	method := "goal/resume"
	if body.Status == "paused" {
		method = "goal/pause"
	}
	d.forwardCodexGoalRPC(c, pod, method, goal)
	c.JSON(http.StatusOK, gin.H{"goal": codexGoalWire(goal)})
}

func (d *Deps) handleDeleteCodexGoal(c *gin.Context) {
	row, pod, ok := d.authorizeSession(c, c.Param("id"))
	if !ok || !d.requireSessionLevel(c, row, levelEdit) {
		return
	}
	if !harnessSupportsCodexGoal(row.AgentSlug) {
		c.JSON(http.StatusOK, gin.H{"cleared": false})
		return
	}
	had, _ := d.Sessions.GetCodexGoal(c.Request.Context(), row.ID)
	_ = d.Sessions.SetCodexGoal(c.Request.Context(), row.ID, nil)
	if had != nil {
		d.forwardCodexGoalRPC(c, pod, "goal/clear", had)
	}
	c.JSON(http.StatusOK, gin.H{"cleared": had != nil})
}

func harnessSupportsCodexGoal(agentSlug string) bool {
	switch agentSlug {
	case "codex-native", "native-codex", "doagent":
		return true
	default:
		return strings.Contains(agentSlug, "codex")
	}
}

func buildCodexGoal(pod *podDomain.Pod, sessionID, objective string, budget *int64, status *string) *domain.CodexGoal {
	now := time.Now().Unix()
	st := "active"
	if status != nil && *status != "" {
		st = *status
	}
	return &domain.CodexGoal{
		ThreadID: codexThreadID(pod, sessionID), Objective: strings.TrimSpace(objective),
		Status: st, TokenBudget: budget, CreatedAt: &now, UpdatedAt: &now,
	}
}

func codexThreadID(pod *podDomain.Pod, sessionID string) string {
	if pod != nil && pod.ExternalSessionID != nil && *pod.ExternalSessionID != "" {
		return *pod.ExternalSessionID
	}
	return sessionID
}

func codexGoalWire(goal *domain.CodexGoal) any {
	if goal == nil {
		return nil
	}
	return gin.H{
		"thread_id": goal.ThreadID, "objective": goal.Objective, "status": goal.Status,
		"token_budget": goal.TokenBudget, "tokens_used": goal.TokensUsed,
		"time_used_seconds": goal.TimeUsedSeconds,
		"created_at":        goal.CreatedAt, "updated_at": goal.UpdatedAt,
	}
}

func (d *Deps) forwardCodexGoalRPC(c *gin.Context, pod *podDomain.Pod, method string, goal *domain.CodexGoal) {
	if d.CommandSender == nil || pod == nil || goal == nil {
		return
	}
	payload, _ := json.Marshal(map[string]any{
		"type": "control_request", "subtype": "doagent.rpc",
		"payload": map[string]any{
			"method": method,
			"params": map[string]any{
				"sessionId": goal.ThreadID, "objective": goal.Objective,
				"status": goal.Status, "tokenBudget": goal.TokenBudget,
			},
		},
	})
	_ = d.CommandSender.SendAcpRelay(c.Request.Context(), pod.RunnerID, pod.PodKey, string(payload))
}

func marshalJSON(v any) ([]byte, error) {
	return json.Marshal(v)
}
