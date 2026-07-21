package main

import (
	"context"

	control "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationcontrol"
	"github.com/l8ai-cn/agentcloud/backend/internal/service/agentpod"
	workerplanner "github.com/l8ai-cn/agentcloud/backend/internal/service/orchestrationworker"
	runnerv1 "github.com/l8ai-cn/agentcloud/proto/gen/go/runner/v1"
	"google.golang.org/protobuf/proto"
)

type workerPodOrchestrator interface {
	CreatePod(
		context.Context,
		*agentpod.OrchestrateCreatePodRequest,
	) (*agentpod.OrchestrateCreatePodResult, error)
}

type workerCommandPayloadSealer interface {
	SealPayload([]byte) ([]byte, error)
}

type orchestrationWorkerPodLauncher struct {
	orchestrator workerPodOrchestrator
	sealer       workerCommandPayloadSealer
}

func newOrchestrationWorkerPodLauncher(
	orchestrator workerPodOrchestrator,
	sealer workerCommandPayloadSealer,
) *orchestrationWorkerPodLauncher {
	return &orchestrationWorkerPodLauncher{orchestrator: orchestrator, sealer: sealer}
}

func (launcher *orchestrationWorkerPodLauncher) MaterializeWorkerPod(
	ctx context.Context,
	claim workerplanner.WorkerLaunchClaim,
) (workerplanner.WorkerPodLaunch, error) {
	if launcher == nil || launcher.orchestrator == nil ||
		launcher.sealer == nil {
		return workerplanner.WorkerPodLaunch{}, control.ErrInvalid
	}
	request := &agentpod.OrchestrateCreatePodRequest{
		OrganizationID:              claim.OrganizationID,
		UserID:                      claim.ActorID,
		Alias:                       optionalWorkerAlias(claim.Alias),
		WorkerSpecSnapshotID:        int64Pointer(claim.WorkerSpecSnapshotID),
		WorkerSpecPromptOverride:    cloneWorkerPrompt(claim.Prompt),
		OrchestrationWorkerLaunchID: int64Pointer(claim.LaunchID),
		DeferRunnerDispatch:         true,
		QueueIfUnavailable:          true,
	}
	result, err := launcher.orchestrator.CreatePod(ctx, request)
	if err != nil {
		return workerplanner.WorkerPodLaunch{}, err
	}
	if result == nil || result.Pod == nil ||
		result.DeferredCreateCommand == nil {
		return workerplanner.WorkerPodLaunch{}, control.ErrCorrupt
	}
	payload, err := proto.Marshal(&runnerv1.ServerMessage{
		Payload: &runnerv1.ServerMessage_CreatePod{
			CreatePod: result.DeferredCreateCommand,
		},
	})
	if err != nil {
		return workerplanner.WorkerPodLaunch{}, err
	}
	payload, err = launcher.sealer.SealPayload(payload)
	if err != nil {
		return workerplanner.WorkerPodLaunch{}, err
	}
	return workerplanner.WorkerPodLaunch{
		PodID: result.Pod.ID, PodKey: result.Pod.PodKey,
		RunnerID: result.Pod.RunnerID, CommandPayload: payload,
	}, nil
}

func optionalWorkerAlias(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func cloneWorkerPrompt(value *string) *string {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}

func int64Pointer(value int64) *int64 {
	return &value
}
