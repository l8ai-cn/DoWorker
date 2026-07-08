package runner

import (
	"context"
	"errors"
	"fmt"
	"time"
)

type ChannelPromptQueuer struct {
	podStore PodStore
	queue    *PendingCommandQueue
}

func NewChannelPromptQueuer(podStore PodStore, queue *PendingCommandQueue) *ChannelPromptQueuer {
	return &ChannelPromptQueuer{podStore: podStore, queue: queue}
}

func (q *ChannelPromptQueuer) QueueChannelPrompt(ctx context.Context, podKey, prompt string, messageID int64) error {
	if q == nil || q.queue == nil || !q.queue.Enabled() {
		return ErrRunnerNotConnected
	}
	pod, err := q.podStore.GetByKey(ctx, podKey)
	if err != nil || pod == nil {
		return err
	}
	commandID := fmt.Sprintf("chmsg-%d", messageID)
	return q.queue.EnqueueSendPrompt(ctx, pod.OrganizationID, pod.RunnerID, podKey, commandID, prompt, 10*time.Minute)
}

func IsRunnerOffline(err error) bool {
	return errors.Is(err, ErrRunnerNotConnected)
}
