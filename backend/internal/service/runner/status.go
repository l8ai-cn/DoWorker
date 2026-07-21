package runner

import (
	"context"
	"log/slog"
	"time"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/runner"
)

func (s *Service) UpdateRunnerStatus(ctx context.Context, runnerID int64, status string) error {
	now := time.Now()
	return s.repo.UpdateFields(ctx, runnerID, map[string]interface{}{
		"status":         status,
		"last_heartbeat": now,
	})
}

func (s *Service) SetRunnerStatus(ctx context.Context, runnerID int64, status string) error {
	return s.UpdateRunnerStatus(ctx, runnerID, status)
}

func (s *Service) IsConnected(runnerID int64) bool {
	_, exists := s.activeRunners.Load(runnerID)
	return exists
}

func (s *Service) MarkConnected(ctx context.Context, runnerID int64) error {
	r, err := s.GetRunner(ctx, runnerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get runner for connect", "runner_id", runnerID, "error", err)
		return err
	}

	r.Status = runner.RunnerStatusOnline

	now := time.Now()
	if err := s.UpdateRunnerStatus(ctx, runnerID, runner.RunnerStatusOnline); err != nil {
		return err
	}
	s.activeMu.Lock()
	s.activeRunners.Store(runnerID, &ActiveRunner{
		Runner:   r,
		LastPing: now,
		PodCount: r.CurrentPods,
	})
	s.activeMu.Unlock()

	slog.InfoContext(ctx, "runner connected", "runner_id", runnerID)
	return nil
}

func (s *Service) MarkDisconnected(ctx context.Context, runnerID int64) error {
	s.activeMu.Lock()
	s.activeRunners.Delete(runnerID)
	s.activeMu.Unlock()
	slog.InfoContext(ctx, "runner disconnected", "runner_id", runnerID)
	return s.UpdateRunnerStatus(ctx, runnerID, runner.RunnerStatusOffline)
}

func (s *Service) RefreshActiveHeartbeat(runnerID int64, currentPods int) {
	now := time.Now()
	s.activeMu.Lock()
	defer s.activeMu.Unlock()

	active, ok := s.activeRunners.Load(runnerID)
	if !ok {
		return
	}
	ar, ok := active.(*ActiveRunner)
	if !ok || ar.Runner == nil {
		return
	}

	updated := *ar.Runner
	updated.Status = runner.RunnerStatusOnline
	updated.LastHeartbeat = &now
	updated.CurrentPods = currentPods
	s.activeRunners.Store(runnerID, &ActiveRunner{
		Runner:   &updated,
		LastPing: now,
		PodCount: currentPods,
	})
}

func (s *Service) UpdateHostInfo(ctx context.Context, runnerID int64, hostInfo map[string]interface{}) error {
	return s.repo.UpdateFields(ctx, runnerID, map[string]interface{}{
		"host_info": hostInfo,
	})
}

func (s *Service) UpdateRunnerVersionAndHostInfo(ctx context.Context, runnerID int64, version string, hostInfo map[string]interface{}) error {
	return s.repo.UpdateFields(ctx, runnerID, map[string]interface{}{
		"runner_version": version,
		"host_info":      hostInfo,
	})
}

func (s *Service) UpdateAvailableAgents(ctx context.Context, runnerID int64, agents []string) error {
	slog.InfoContext(ctx, "runner available agents updated", "runner_id", runnerID, "agents", agents)
	availableAgents := runner.StringSlice(append([]string(nil), agents...))
	if err := s.repo.UpdateFields(ctx, runnerID, map[string]interface{}{
		"available_agents": availableAgents,
	}); err != nil {
		return err
	}

	s.activeMu.Lock()
	defer s.activeMu.Unlock()
	if active, ok := s.activeRunners.Load(runnerID); ok {
		if ar, ok := active.(*ActiveRunner); ok && ar.Runner != nil {
			updated := *ar.Runner
			updated.AvailableAgents = append(runner.StringSlice(nil), availableAgents...)
			s.activeRunners.Store(runnerID, &ActiveRunner{
				Runner:   &updated,
				LastPing: ar.LastPing,
				PodCount: ar.PodCount,
			})
		}
	}
	return nil
}

func (s *Service) UpdateAgentVersions(ctx context.Context, runnerID int64, versions []runner.AgentVersion) error {
	if err := s.repo.UpdateFields(ctx, runnerID, map[string]interface{}{
		"agent_versions": runner.AgentVersionSlice(versions),
	}); err != nil {
		return err
	}

	s.activeMu.Lock()
	defer s.activeMu.Unlock()
	if active, ok := s.activeRunners.Load(runnerID); ok {
		if ar, ok := active.(*ActiveRunner); ok && ar.Runner != nil {
			updated := *ar.Runner
			updated.AgentVersions = runner.AgentVersionSlice(versions)
			s.activeRunners.Store(runnerID, &ActiveRunner{
				Runner:   &updated,
				LastPing: ar.LastPing,
				PodCount: ar.PodCount,
			})
		}
	}
	return nil
}

func (s *Service) MergeAgentVersions(ctx context.Context, runnerID int64, changes map[string]runner.AgentVersion) error {
	if len(changes) == 0 {
		return nil
	}

	r, err := s.GetRunner(ctx, runnerID)
	if err != nil {
		return err
	}

	merged := make(map[string]runner.AgentVersion)
	for _, v := range r.AgentVersions {
		merged[v.Slug] = v
	}

	for slug, change := range changes {
		if change.Version == "" && change.Path == "" {
			delete(merged, slug)
		} else {
			merged[slug] = change
		}
	}

	result := make([]runner.AgentVersion, 0, len(merged))
	for _, v := range merged {
		result = append(result, v)
	}

	return s.UpdateAgentVersions(ctx, runnerID, result)
}

func (s *Service) IncrementPods(ctx context.Context, runnerID int64) error {
	return s.repo.IncrementPods(ctx, runnerID)
}

func (s *Service) DecrementPods(ctx context.Context, runnerID int64) error {
	return s.repo.DecrementPods(ctx, runnerID)
}

type RunnerUpdateFunc func(*runner.Runner)

func (s *Service) SubscribeStatusChanges(ctx context.Context, callback RunnerUpdateFunc) (func(), error) {
	return func() {}, nil
}
