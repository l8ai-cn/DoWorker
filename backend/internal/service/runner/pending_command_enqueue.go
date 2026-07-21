package runner

import (
	"context"
	"time"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/agentpod"
	runnerv1 "github.com/l8ai-cn/agentcloud/proto/gen/go/runner/v1"
)

func (q *PendingCommandQueue) EnqueueCreatePod(
	ctx context.Context,
	orgID, runnerID int64,
	podKey string,
	cmd *runnerv1.CreatePodCommand,
	ttl time.Duration,
) (time.Time, error) {
	payload, err := marshalServerMessage(&runnerv1.ServerMessage{
		Payload: &runnerv1.ServerMessage_CreatePod{CreatePod: cmd},
	})
	if err != nil {
		return time.Time{}, err
	}
	expiresAt, err := q.enqueue(
		ctx, orgID, runnerID, podKey,
		agentpod.CommandTypeCreatePod, podKey, payload, ttl,
	)
	if err != nil {
		return time.Time{}, err
	}
	q.maybeDrain(runnerID)
	return expiresAt, nil
}

func (q *PendingCommandQueue) EnqueueSendPrompt(
	ctx context.Context,
	orgID, runnerID int64,
	podKey, commandID, prompt string,
	ttl time.Duration,
) error {
	payload, err := marshalServerMessage(&runnerv1.ServerMessage{
		Payload: &runnerv1.ServerMessage_SendPrompt{
			SendPrompt: &runnerv1.SendPromptCommand{
				PodKey:    podKey,
				Prompt:    prompt,
				CommandId: commandID,
			},
		},
	})
	if err != nil {
		return err
	}
	_, err = q.enqueue(
		ctx, orgID, runnerID, podKey,
		agentpod.CommandTypeSendPrompt, commandID, payload, ttl,
	)
	if err != nil {
		return err
	}
	q.maybeDrain(runnerID)
	return nil
}
