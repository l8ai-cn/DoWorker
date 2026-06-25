package coordinator

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	runnerDomain "github.com/anthropics/agentsmesh/backend/internal/domain/runner"
	runnersvc "github.com/anthropics/agentsmesh/backend/internal/service/runner"
)

const (
	defaultRunnerWait  = 90 * time.Second
	defaultRunnerPoll  = 2 * time.Second
)

// RunnerSelector picks an online runner for an agent. Satisfied by *runner.Service.
type RunnerSelector interface {
	SelectRunnerWithAffinity(
		ctx context.Context,
		orgID, userID int64,
		agentSlug string,
		hints *runnerDomain.AffinityHints,
		repoHistory map[int64]int,
	) (*runnerDomain.Runner, error)
}

// RunnerLauncher provisions a runner when none is online (auto-harness CreateInstance parity).
type RunnerLauncher interface {
	Launch(ctx context.Context, orgID int64, agentSlug string) error
}

// RunnerEnsurer ensures a capable runner exists before coordinator dispatch.
type RunnerEnsurer struct {
	selector RunnerSelector
	launcher RunnerLauncher
	wait     time.Duration
	poll     time.Duration
	logger   *slog.Logger
}

func NewRunnerEnsurer(selector RunnerSelector, launcher RunnerLauncher, logger *slog.Logger) *RunnerEnsurer {
	if logger == nil {
		logger = slog.Default()
	}
	return &RunnerEnsurer{
		selector: selector,
		launcher: launcher,
		wait:     defaultRunnerWait,
		poll:     defaultRunnerPoll,
		logger:   logger.With("component", "coordinator_runner"),
	}
}

func (e *RunnerEnsurer) Ensure(ctx context.Context, orgID, userID int64, agentSlug string) error {
	if e == nil || e.selector == nil {
		return nil
	}
	if _, err := e.selector.SelectRunnerWithAffinity(ctx, orgID, userID, agentSlug, nil, nil); err == nil {
		return nil
	} else if !errors.Is(err, runnersvc.ErrNoRunnerForAgent) {
		return err
	}
	if e.launcher == nil {
		return runnersvc.ErrNoRunnerForAgent
	}

	e.logger.Info("no online runner; launching provisioner",
		"org_id", orgID, "agent_slug", agentSlug)
	if err := e.launcher.Launch(ctx, orgID, agentSlug); err != nil {
		return fmt.Errorf("runner launcher: %w", err)
	}

	deadline := time.Now().Add(e.wait)
	for time.Now().Before(deadline) {
		if _, err := e.selector.SelectRunnerWithAffinity(ctx, orgID, userID, agentSlug, nil, nil); err == nil {
			e.logger.Info("runner online after provision", "org_id", orgID, "agent_slug", agentSlug)
			return nil
		} else if !errors.Is(err, runnersvc.ErrNoRunnerForAgent) {
			return err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(e.poll):
		}
	}
	return runnersvc.ErrNoRunnerForAgent
}
