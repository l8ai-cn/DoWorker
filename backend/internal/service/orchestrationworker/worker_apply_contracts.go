package orchestrationworker

import (
	"context"
	"time"

	control "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationcontrol"
	controlservice "github.com/l8ai-cn/agentcloud/backend/internal/service/orchestrationcontrol"
)

type WorkerLaunchProjection = controlservice.WorkerLaunchProjection
type WorkerApplyMutation = controlservice.WorkerApplyMutation
type WorkerApplyBuilder = controlservice.WorkerApplyBuilder
type AppliedWorker = controlservice.AppliedWorker

type WorkerApplyRepository interface {
	RunWorkerApplyTransaction(
		context.Context,
		control.Scope,
		string,
		WorkerApplyBuilder,
	) (AppliedWorker, error)
	ClaimWorkerLaunch(
		context.Context,
		control.Scope,
		int64,
		time.Duration,
		string,
	) (WorkerLaunchClaim, error)
	ReleaseWorkerLaunch(
		context.Context,
		control.Scope,
		WorkerLaunchClaim,
		string,
	) error
	CompleteWorkerLaunch(
		context.Context,
		control.Scope,
		WorkerLaunchClaim,
		WorkerPodLaunch,
		time.Duration,
	) (AppliedWorker, error)
}

type WorkerLaunchClaim = controlservice.WorkerLaunchClaim
type WorkerPodLaunch = controlservice.WorkerPodLaunch

type WorkerPodLauncher interface {
	MaterializeWorkerPod(
		context.Context,
		WorkerLaunchClaim,
	) (WorkerPodLaunch, error)
}

type WorkerDispatchNotifier interface {
	TriggerWorkerDispatch(int64)
}
