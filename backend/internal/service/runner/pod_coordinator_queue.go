package runner

import (
	"context"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/agentpod"
	runnerDomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/runner"
	runnerv1 "github.com/l8ai-cn/agentcloud/proto/gen/go/runner/v1"
)

type CreatePodQueueOpts = agentpod.CreatePodQueueOpts

func (pc *PodCoordinator) SetPendingQueue(q *PendingCommandQueue) {
	pc.pendingQueue = q
}

func (pc *PodCoordinator) SetPendingDrainer(d *PendingCommandDrainer) {
	pc.pendingDrainer = d
}

func (pc *PodCoordinator) PendingDrainer() *PendingCommandDrainer {
	return pc.pendingDrainer
}

func (pc *PodCoordinator) CreatePodOrQueue(
	ctx context.Context,
	runnerID int64,
	cmd *runnerv1.CreatePodCommand,
	opts CreatePodQueueOpts,
) error {
	if pc.shouldDispatchNow(ctx, runnerID) {
		return pc.CreatePod(ctx, runnerID, cmd)
	}
	if !opts.Queue || pc.pendingQueue == nil || !pc.pendingQueue.Enabled() {
		if pc.connChecker != nil && !pc.connChecker.IsConnected(runnerID) {
			return ErrRunnerNotConnected
		}
		return ErrRunnerNotConnected
	}
	_, err := pc.pendingQueue.EnqueueCreatePod(ctx, opts.OrgID, runnerID, cmd.PodKey, cmd, opts.TTL)
	if err != nil {
		return err
	}
	return agentpod.ErrPodQueued
}

func (pc *PodCoordinator) ShouldDispatchNow(ctx context.Context, runnerID int64) bool {
	return pc.shouldDispatchNow(ctx, runnerID)
}

func (pc *PodCoordinator) shouldDispatchNow(ctx context.Context, runnerID int64) bool {
	if pc.connChecker == nil || !pc.connChecker.IsConnected(runnerID) {
		return false
	}
	run, err := pc.runnerRepo.GetByID(ctx, runnerID)
	if err != nil || run == nil {
		return false
	}
	return run.CurrentPods < run.MaxConcurrentPods
}

func (pc *PodCoordinator) SetConnectionChecker(c ConnectionChecker) {
	pc.connChecker = c
}

func (pc *PodCoordinator) triggerPendingDrain(runnerID int64) {
	if pc.pendingDrainer != nil {
		pc.pendingDrainer.DrainRunner(runnerID)
	}
}

func (pc *PodCoordinator) GetRunnerRepo() runnerDomain.RunnerRepository {
	return pc.runnerRepo
}
