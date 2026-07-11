package goalloop

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	domain "github.com/anthropics/agentsmesh/backend/internal/domain/goalloop"
	agentpodsvc "github.com/anthropics/agentsmesh/backend/internal/service/agentpod"
	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
)

func (s *Service) Start(ctx context.Context, orgID, userID int64, slug string) (*domain.GoalLoop, error) {
	if !s.executionReady() {
		return nil, ErrExecutionUnavailable
	}
	loop, err := s.GetBySlug(ctx, orgID, slug)
	if err != nil {
		return nil, err
	}
	if loop.Status != domain.StatusDraft && loop.Status != domain.StatusPaused {
		return nil, ErrInvalidState
	}
	now := time.Now()
	if err := s.repo.Update(ctx, loop.ID, map[string]any{
		"status":                        domain.StatusActive,
		"pod_key":                       nil,
		"autopilot_controller_key":      nil,
		"verification_request_id":       nil,
		"verification_exit_code":        nil,
		"verification_output":           nil,
		"verification_output_truncated": false,
		"verification_error":            nil,
		"started_at":                    now,
		"verified_at":                   nil,
		"completed_at":                  nil,
	}); err != nil {
		return nil, err
	}

	podResult, err := s.podCreator.CreatePod(ctx, &agentpodsvc.OrchestrateCreatePodRequest{
		OrganizationID:           loop.OrganizationID,
		UserID:                   userID,
		WorkerSpecSnapshotID:     int64Pointer(loop.WorkerSpecSnapshotID),
		WorkerSpecPromptOverride: stringPointer(loopPrompt(loop)),
		TokenBudget:              loop.TokenBudget,
		Cols:                     120,
		Rows:                     40,
	})
	if err != nil {
		return s.failStart(ctx, loop, fmt.Errorf("pod creation failed: %w", err))
	}
	if err := s.repo.Update(ctx, loop.ID, map[string]any{"pod_key": podResult.Pod.PodKey}); err != nil {
		_ = s.podTerminator.TerminatePod(ctx, podResult.Pod.PodKey)
		return nil, err
	}

	controller, err := s.autopilot.CreateAndStart(ctx, &agentpodsvc.CreateAndStartRequest{
		OrganizationID:      loop.OrganizationID,
		Pod:                 podResult.Pod,
		Prompt:              loopPrompt(loop),
		MaxIterations:       int32(loop.MaxIterations),
		IterationTimeoutSec: int32(min(loop.TimeoutMinutes*60, 900)),
		NoProgressThreshold: int32(loop.NoProgressLimit),
		SameErrorThreshold:  int32(loop.SameErrorLimit),
		ApprovalTimeoutMin:  5,
		KeyPrefix:           fmt.Sprintf("goal-loop-%d", loop.ID),
	})
	if err != nil {
		terminateErr := s.podTerminator.TerminatePod(ctx, podResult.Pod.PodKey)
		startErr := fmt.Errorf("autopilot creation failed: %w", err)
		if terminateErr != nil {
			startErr = errors.Join(startErr, fmt.Errorf("terminate pod: %w", terminateErr))
		}
		return s.failStart(ctx, loop, startErr)
	}
	if err := s.repo.Update(ctx, loop.ID, map[string]any{
		"autopilot_controller_key": controller.AutopilotControllerKey,
	}); err != nil {
		return nil, err
	}
	return s.GetBySlug(ctx, orgID, slug)
}

func (s *Service) Cancel(ctx context.Context, orgID int64, slug string) (*domain.GoalLoop, error) {
	loop, err := s.GetBySlug(ctx, orgID, slug)
	if err != nil {
		return nil, err
	}
	if loop.IsTerminal() {
		return nil, ErrInvalidState
	}
	if err := s.stopPod(ctx, loop); err != nil {
		return nil, err
	}
	now := time.Now()
	if err := s.repo.Update(ctx, loop.ID, map[string]any{
		"status":       domain.StatusCancelled,
		"completed_at": now,
	}); err != nil {
		return nil, err
	}
	return s.GetBySlug(ctx, orgID, slug)
}

func (s *Service) Verify(ctx context.Context, orgID int64, slug string) (*domain.GoalLoop, error) {
	loop, err := s.GetBySlug(ctx, orgID, slug)
	if err != nil {
		return nil, err
	}
	if loop.Status != domain.StatusActive && loop.Status != domain.StatusVerifying {
		return nil, ErrInvalidState
	}
	if err := s.beginVerification(ctx, loop); err != nil {
		return nil, err
	}
	return s.GetBySlug(ctx, orgID, slug)
}

func (s *Service) HandlePodStatus(ctx context.Context, podKey, status string) error {
	loop, err := s.repo.GetByPodKey(ctx, podKey)
	if errors.Is(err, domain.ErrNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	if loop.IsTerminal() || loop.Status == domain.StatusPaused {
		return nil
	}
	if status == agentpod.StatusCompleted {
		return s.beginVerification(ctx, loop)
	}
	if status == agentpod.StatusError || status == agentpod.StatusTerminated {
		return s.escalate(ctx, loop, "pod stopped before verification", nil)
	}
	return nil
}

func (s *Service) HandleAutopilotStatus(ctx context.Context, autopilotKey, phase string) error {
	loop, err := s.repo.GetByAutopilotControllerKey(ctx, autopilotKey)
	if errors.Is(err, domain.ErrNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	if loop.IsTerminal() || loop.Status == domain.StatusPaused || loop.Status == domain.StatusVerifying {
		return nil
	}
	switch phase {
	case agentpod.AutopilotPhaseCompleted:
		return s.beginVerification(ctx, loop)
	case agentpod.AutopilotPhasePaused, agentpod.AutopilotPhaseWaitingApproval,
		agentpod.AutopilotPhaseUserTakeover, agentpod.AutopilotPhaseMaxIterations,
		agentpod.AutopilotPhaseFailed, agentpod.AutopilotPhaseStopped:
		return s.escalate(ctx, loop, "autopilot stopped: "+phase, nil)
	default:
		return nil
	}
}

func (s *Service) ExpireTimedOut(ctx context.Context, now time.Time) error {
	loops, err := s.repo.ListTimedOut(ctx, now)
	if err != nil {
		return err
	}
	for _, loop := range loops {
		if err := s.escalate(ctx, loop, "runtime budget exhausted", nil); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) HandleVerificationResult(
	ctx context.Context,
	runnerID int64,
	result *runnerv1.VerificationResultEvent,
) error {
	if result == nil || strings.TrimSpace(result.GetRequestId()) == "" {
		return ErrInvalidInput
	}
	loop, err := s.repo.GetByVerificationRequestID(ctx, result.GetRequestId())
	if errors.Is(err, domain.ErrNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	if loop.Status != domain.StatusVerifying || loop.PodKey == nil || *loop.PodKey != result.GetPodKey() {
		return nil
	}
	pod, err := s.podLookup.GetPod(ctx, *loop.PodKey)
	if err != nil {
		return err
	}
	if pod.OrganizationID != loop.OrganizationID || pod.RunnerID != runnerID {
		return ErrInvalidInput
	}
	now := time.Now()
	output := truncateOutput(result.GetOutput())
	updates := map[string]any{
		"verification_exit_code":        int(result.GetExitCode()),
		"verification_output":           output,
		"verification_output_truncated": result.GetOutputTruncated() || len(result.GetOutput()) > len(output),
		"verification_error":            nullableString(result.GetError()),
		"verified_at":                   now,
	}
	if result.GetError() == "" && result.GetExitCode() == 0 {
		updates["status"] = domain.StatusCompleted
		updates["completed_at"] = now
		if err := s.repo.Update(ctx, loop.ID, updates); err != nil {
			return err
		}
		return s.stopPod(ctx, loop)
	}
	return s.escalate(ctx, loop, verificationFailureReason(result), updates)
}

func (s *Service) beginVerification(ctx context.Context, loop *domain.GoalLoop) error {
	if loop.Status == domain.StatusVerifying || loop.IsTerminal() || loop.Status == domain.StatusPaused {
		return nil
	}
	if !s.verificationReady() || loop.PodKey == nil {
		return ErrVerificationPending
	}
	pod, err := s.podLookup.GetPod(ctx, *loop.PodKey)
	if err != nil {
		return err
	}
	if pod.OrganizationID != loop.OrganizationID || pod.RunnerID <= 0 {
		return ErrInvalidInput
	}
	requestID := uuid.NewString()
	if err := s.repo.Update(ctx, loop.ID, map[string]any{
		"status":                  domain.StatusVerifying,
		"verification_request_id": requestID,
		"verification_error":      nil,
	}); err != nil {
		return err
	}
	if err := s.verificationSender.SendRunVerification(ctx, pod.RunnerID, &runnerv1.RunVerificationCommand{
		RequestId:      requestID,
		PodKey:         pod.PodKey,
		Command:        loop.VerificationCommand,
		TimeoutSeconds: int32(min(loop.TimeoutMinutes*60, 900)),
	}); err != nil {
		return s.escalate(ctx, loop, fmt.Sprintf("verification dispatch failed: %v", err), nil)
	}
	return nil
}

func (s *Service) failStart(ctx context.Context, loop *domain.GoalLoop, cause error) (*domain.GoalLoop, error) {
	now := time.Now()
	if err := s.repo.Update(ctx, loop.ID, map[string]any{
		"status":             domain.StatusFailed,
		"verification_error": cause.Error(),
		"completed_at":       now,
	}); err != nil {
		return nil, errors.Join(cause, err)
	}
	return nil, cause
}

func (s *Service) escalate(ctx context.Context, loop *domain.GoalLoop, reason string, extra map[string]any) error {
	now := time.Now()
	updates := map[string]any{
		"verification_error": reason,
	}
	for key, value := range extra {
		updates[key] = value
	}
	if loop.EscalationPolicy == domain.EscalationPause {
		updates["status"] = domain.StatusPaused
	} else {
		updates["status"] = domain.StatusFailed
		updates["completed_at"] = now
	}
	if err := s.repo.Update(ctx, loop.ID, updates); err != nil {
		return err
	}
	return s.stopPod(ctx, loop)
}

func (s *Service) stopPod(ctx context.Context, loop *domain.GoalLoop) error {
	if s.podLookup == nil || s.podTerminator == nil || loop.PodKey == nil {
		return nil
	}
	pod, err := s.podLookup.GetPod(ctx, *loop.PodKey)
	if err != nil {
		return err
	}
	if !pod.IsActive() {
		return nil
	}
	return s.podTerminator.TerminatePod(ctx, pod.PodKey)
}

func (s *Service) executionReady() bool {
	return s.podCreator != nil &&
		s.podLookup != nil &&
		s.podTerminator != nil &&
		s.autopilot != nil &&
		s.verificationSender != nil
}

func (s *Service) verificationReady() bool {
	return s.podLookup != nil && s.verificationSender != nil
}

func loopPrompt(loop *domain.GoalLoop) string {
	return fmt.Sprintf(
		"Objective:\n%s\n\nAcceptance criteria:\n%s\n\nWork until the criteria are implemented. "+
			"Do not treat a textual completion claim as evidence; an external verification command will decide the result.",
		loop.Objective,
		formatCriteria(loop.AcceptanceCriteria),
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

func verificationFailureReason(result *runnerv1.VerificationResultEvent) string {
	if result.GetError() != "" {
		return "verification failed: " + result.GetError()
	}
	return fmt.Sprintf("verification exited with code %d", result.GetExitCode())
}

func truncateOutput(output string) string {
	const maxBytes = 64 << 10
	if len(output) <= maxBytes {
		return output
	}
	return output[:maxBytes]
}

func nullableString(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return value
}

func int64Pointer(value int64) *int64 {
	return &value
}

func stringPointer(value string) *string {
	return &value
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
