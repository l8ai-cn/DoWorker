package agentpod

import (
	"context"
	"log/slog"
	"time"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/agentpod"
)

func (s *PodService) MarkInitFailed(ctx context.Context, podKey, errorCode, errorMessage string) error {
	now := time.Now()
	_, err := s.repo.UpdateByKeyAndStatusCounted(ctx, podKey, agentpod.StatusInitializing, map[string]interface{}{
		"status":        agentpod.StatusError,
		"error_code":    errorCode,
		"error_message": errorMessage,
		"finished_at":   now,
	})
	if err != nil {
		slog.ErrorContext(ctx, "failed to mark pod init failed", "pod_key", podKey, "error_code", errorCode, "error", err)
		return err
	}
	slog.WarnContext(ctx, "pod init failed", "pod_key", podKey, "error_code", errorCode, "error_message", errorMessage)
	return nil
}

// MarkDispatchFailed covers both pre-dispatch statuses: a pod created as
// `queued` (enqueue failed afterwards) and one created as `initializing`
// (direct dispatch failed). Without the queued arm such pods would be
// orphaned forever — the expiry sweeper only scans pending rows.
func (s *PodService) MarkDispatchFailed(ctx context.Context, podKey, errorCode, errorMessage string) error {
	now := time.Now()
	updates := map[string]interface{}{
		"status":        agentpod.StatusError,
		"error_code":    errorCode,
		"error_message": errorMessage,
		"finished_at":   now,
	}
	rows, err := s.repo.UpdateByKeyAndStatusCounted(ctx, podKey, agentpod.StatusInitializing, updates)
	if err == nil && rows == 0 {
		rows, err = s.repo.UpdateByKeyAndStatusCounted(ctx, podKey, agentpod.StatusQueued, updates)
	}
	if err != nil {
		slog.ErrorContext(ctx, "failed to mark pod dispatch failed", "pod_key", podKey, "error", err)
		return err
	}
	if rows == 0 {
		slog.WarnContext(ctx, "pod not in pre-dispatch status, dispatch-failure mark skipped", "pod_key", podKey)
	}
	return nil
}

func (s *PodService) MarkQueueExpired(ctx context.Context, podKey, errorCode, errorMessage string) error {
	now := time.Now()
	_, err := s.repo.UpdateByKeyAndStatusCounted(ctx, podKey, agentpod.StatusQueued, map[string]interface{}{
		"status":        agentpod.StatusError,
		"error_code":    errorCode,
		"error_message": errorMessage,
		"finished_at":   now,
	})
	if err != nil {
		slog.ErrorContext(ctx, "failed to mark queued pod expired", "pod_key", podKey, "error", err)
		return err
	}
	slog.WarnContext(ctx, "queued pod expired", "pod_key", podKey, "error_code", errorCode)
	return nil
}
