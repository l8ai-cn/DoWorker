package agentworkbench

import (
	"testing"

	agentworkbenchv2 "github.com/l8ai-cn/agentcloud/proto/gen/go/agent_workbench/v2"
	"github.com/stretchr/testify/require"
)

func TestDeltaHubDisconnectsLaggingSubscriber(t *testing.T) {
	hub := NewDeltaHub(1)
	subscription := hub.Subscribe("conv_1")
	t.Cleanup(subscription.Close)

	hub.Publish("conv_1", &agentworkbenchv2.SessionDeltaBatch{Revision: 1})
	hub.Publish("conv_1", &agentworkbenchv2.SessionDeltaBatch{Revision: 2})

	first := <-subscription.Deltas
	require.Equal(t, uint64(1), first.Revision)
	require.ErrorIs(t, <-subscription.Errors, ErrSubscriberLagged)
	_, open := <-subscription.Deltas
	require.False(t, open)
}

func TestDeltaHubPublishesIndependentClones(t *testing.T) {
	hub := NewDeltaHub(2)
	first := hub.Subscribe("conv_1")
	second := hub.Subscribe("conv_1")
	t.Cleanup(first.Close)
	t.Cleanup(second.Close)
	delta := &agentworkbenchv2.SessionDeltaBatch{Revision: 1, Digest: "digest"}

	hub.Publish("conv_1", delta)
	firstDelta := <-first.Deltas
	secondDelta := <-second.Deltas
	firstDelta.Digest = "changed"

	require.Equal(t, "digest", secondDelta.Digest)
}
