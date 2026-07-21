package runner

import (
	"log/slog"
	"sync/atomic"
	"testing"
	"time"

	"github.com/l8ai-cn/agentcloud/runner/internal/client"
	"github.com/l8ai-cn/agentcloud/runner/internal/config"
	"github.com/l8ai-cn/agentcloud/runner/internal/relay"
)

func TestOnSubscribePodCoalescesConcurrentSameRelay(t *testing.T) {
	store := NewInMemoryPodStore()
	pod := &Pod{PodKey: "pod-1", Status: PodStatusRunning}
	store.Put(pod.PodKey, pod)
	handler := NewRunnerMessageHandler(&Runner{cfg: &config.Config{}}, store, client.NewMockConnection())

	started := make(chan struct{})
	release := make(chan struct{})
	firstClient := &gatedRelayClient{
		MockClient: relay.NewMockClient("wss://relay.example"),
		started:    started,
		release:    release,
	}
	secondFactoryCall := make(chan struct{})
	var factoryCalls atomic.Int32
	handler.relayClientFactory = func(url, _, _ string, _ *slog.Logger) relay.RelayClient {
		if factoryCalls.Add(1) == 1 {
			return firstClient
		}
		close(secondFactoryCall)
		return relay.NewMockClient(url)
	}

	firstResult := make(chan error, 1)
	go func() {
		firstResult <- handler.OnSubscribePod(client.SubscribePodRequest{
			PodKey: pod.PodKey, RelayURL: "wss://relay.example", RunnerToken: "first-token",
		})
	}()
	<-started

	secondResult := make(chan error, 1)
	go func() {
		secondResult <- handler.OnSubscribePod(client.SubscribePodRequest{
			PodKey: pod.PodKey, RelayURL: "wss://relay.example", RunnerToken: "second-token",
		})
	}()

	select {
	case <-secondFactoryCall:
		t.Fatal("same relay subscription created a second client")
	case <-time.After(50 * time.Millisecond):
	}
	close(release)

	if err := <-firstResult; err != nil {
		t.Fatalf("first subscription failed: %v", err)
	}
	if err := <-secondResult; err != nil {
		t.Fatalf("second subscription failed: %v", err)
	}
	if factoryCalls.Load() != 1 {
		t.Fatalf("relay client factory calls = %d, want 1", factoryCalls.Load())
	}
	if len(firstClient.UpdateTokenCalls) != 2 {
		t.Fatalf("token refresh calls = %v, want both subscriptions", firstClient.UpdateTokenCalls)
	}
}
