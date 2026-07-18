package sessionapi

import (
	"errors"
	"time"

	podDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	"github.com/anthropics/agentsmesh/backend/internal/service/agentpod"
	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
	"google.golang.org/protobuf/proto"
)

var errDeferredSessionCommandInvalid = errors.New("deferred session command is invalid")

func pendingSessionCreateCommand(
	result *agentpod.OrchestrateCreatePodResult,
	ttl time.Duration,
) (*podDomain.PendingCommand, error) {
	if result == nil || result.Pod == nil ||
		result.DeferredCreateCommand == nil ||
		result.Pod.PodKey == "" ||
		result.Pod.PodKey != result.DeferredCreateCommand.GetPodKey() ||
		result.Pod.RunnerID <= 0 ||
		ttl <= 0 {
		return nil, errDeferredSessionCommandInvalid
	}
	payload, err := proto.Marshal(&runnerv1.ServerMessage{
		Payload: &runnerv1.ServerMessage_CreatePod{
			CreatePod: result.DeferredCreateCommand,
		},
	})
	if err != nil {
		return nil, err
	}
	now := time.Now()
	return &podDomain.PendingCommand{
		OrganizationID: result.Pod.OrganizationID,
		RunnerID:       result.Pod.RunnerID,
		PodKey:         result.Pod.PodKey,
		CommandType:    podDomain.CommandTypeCreatePod,
		CommandID:      result.Pod.PodKey,
		Payload:        payload,
		ExpiresAt:      now.Add(ttl),
		CreatedAt:      now,
	}, nil
}
