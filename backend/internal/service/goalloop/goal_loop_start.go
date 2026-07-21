package goalloop

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	domain "github.com/l8ai-cn/agentcloud/backend/internal/domain/goalloop"
	agentpodsvc "github.com/l8ai-cn/agentcloud/backend/internal/service/agentpod"
)

func (s *Service) Start(ctx context.Context, orgID, userID int64, slug string) (*domain.GoalLoop, error) {
	if err := s.ValidateExecutionReady(); err != nil {
		return nil, err
	}
	loop, err := s.GetBySlug(ctx, orgID, slug)
	if err != nil {
		return nil, err
	}
	if !loop.HasCompleteResourceBinding() {
		return nil, ErrInvalidState
	}
	if loop.Status != domain.StatusDraft && loop.Status != domain.StatusPaused {
		return nil, ErrInvalidState
	}
	if err := s.ValidateWorkerSnapshotForExecution(
		ctx, loop.OrganizationID, userID, loop.WorkerSpecSnapshotID,
	); err != nil {
		return nil, err
	}
	now := time.Now()
	claimed, err := s.repo.TransitionStatus(ctx, loop.ID, []string{
		domain.StatusDraft,
		domain.StatusPaused,
	}, map[string]any{
		"status": domain.StatusActive, "pod_key": nil, "autopilot_controller_key": nil,
		"current_iteration": 0, "no_progress_count": 0, "same_error_count": 0,
		"last_progress_fingerprint": nil, "last_error_fingerprint": nil,
		"retry_prompt_command_id": nil, "retry_prompt_created_at": nil,
		"verification_request_id": nil, "verification_exit_code": nil,
		"verification_output": nil, "verification_output_truncated": false,
		"verification_error": nil, "started_at": now, "verified_at": nil, "completed_at": nil,
	})
	if err != nil {
		return nil, err
	}
	if !claimed {
		return nil, ErrInvalidState
	}

	podResult, err := s.podCreator.CreatePod(ctx, &agentpodsvc.OrchestrateCreatePodRequest{
		OrganizationID: loop.OrganizationID, UserID: userID,
		WorkerSpecSnapshotID:     int64Pointer(loop.WorkerSpecSnapshotID),
		WorkerSpecPromptOverride: stringPointer(loopPrompt(loop)),
		TokenBudget:              loop.TokenBudget, Cols: 120, Rows: 40,
	})
	if err != nil {
		return s.failStart(ctx, loop, fmt.Errorf("pod creation failed: %w", err))
	}
	if podResult == nil || podResult.Pod == nil {
		return s.failStart(ctx, loop, fmt.Errorf("pod creation returned nil pod"))
	}
	persisted, err := s.persistStartKey(ctx, loop, "pod_key", podResult.Pod.PodKey)
	if err != nil {
		return s.failAfterPod(ctx, loop, podResult.Pod.PodKey, "persist pod key", err)
	}
	if !persisted {
		return s.abortCancelledStart(ctx, loop, podResult.Pod.PodKey)
	}
	return s.GetBySlug(ctx, orgID, slug)
}

func (s *Service) persistStartKey(
	ctx context.Context, loop *domain.GoalLoop, key, value string,
) (bool, error) {
	return s.repo.TransitionStatus(
		ctx,
		loop.ID,
		[]string{domain.StatusActive},
		map[string]any{key: value},
	)
}

func (s *Service) executionReady() bool {
	return s.podCreator != nil && s.podLookup != nil && s.podTerminator != nil &&
		s.verificationSender != nil && s.promptSender != nil && s.promptSender.Enabled()
}

func loopPrompt(loop *domain.GoalLoop) string {
	return fmt.Sprintf(
		"Objective:\n%s\n\nAcceptance criteria:\n%s\n\nWork until the criteria are implemented. "+
			"Do not treat a textual completion claim as evidence; an external verification command will decide the result.",
		loop.Objective, formatCriteria(loop.AcceptanceCriteria),
	)
}

func formatCriteria(raw []byte) string {
	var criteria []string
	_ = json.Unmarshal(raw, &criteria)
	lines := make([]string, 0, len(criteria))
	for _, criterion := range criteria {
		if value := strings.TrimSpace(criterion); value != "" {
			lines = append(lines, "- "+value)
		}
	}
	return strings.Join(lines, "\n")
}

func int64Pointer(value int64) *int64    { return &value }
func stringPointer(value string) *string { return &value }

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
