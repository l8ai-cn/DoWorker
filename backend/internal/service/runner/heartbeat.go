package runner

import (
	"context"
	"time"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/runner"
)

func (s *Service) Heartbeat(ctx context.Context, runnerID int64, currentPods int) error {
	now := time.Now()
	return s.repo.UpdateFields(ctx, runnerID, map[string]interface{}{
		"last_heartbeat": now,
		"current_pods":   currentPods,
		"status":         runner.RunnerStatusOnline,
	})
}

type HeartbeatPodInfo struct {
	PodKey      string `json:"pod_key"`
	Status      string `json:"status,omitempty"`
	AgentStatus string `json:"agent_status,omitempty"`
}

func (s *Service) UpdateHeartbeatWithPods(ctx context.Context, runnerID int64, pods []HeartbeatPodInfo, version string) error {
	now := time.Now()
	updates := map[string]interface{}{
		"last_heartbeat": now,
		"current_pods":   len(pods),
		"status":         runner.RunnerStatusOnline,
	}
	if version != "" {
		updates["runner_version"] = version
	}

	if err := s.repo.UpdateFields(ctx, runnerID, updates); err != nil {
		return err
	}

	r, err := s.repo.GetByID(ctx, runnerID)
	if err != nil {
		return err
	}
	if r == nil {
		return ErrRunnerNotFound
	}

	cached := r
	s.activeMu.Lock()
	if active, ok := s.activeRunners.Load(runnerID); ok {
		if ar, ok := active.(*ActiveRunner); ok && ar.Runner != nil {
			updated := *ar.Runner
			updated.Status = r.Status
			updated.LastHeartbeat = r.LastHeartbeat
			updated.CurrentPods = r.CurrentPods
			updated.RunnerVersion = r.RunnerVersion
			cached = &updated
		}
	}
	s.activeRunners.Store(runnerID, &ActiveRunner{
		Runner:   cached,
		LastPing: now,
		PodCount: len(pods),
	})
	s.activeMu.Unlock()
	return nil
}

func (s *Service) MarkOfflineRunners(ctx context.Context, timeout time.Duration) error {
	threshold := time.Now().Add(-timeout)
	return s.repo.MarkOfflineRunners(ctx, threshold)
}
