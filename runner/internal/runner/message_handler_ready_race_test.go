package runner

import (
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/l8ai-cn/agentcloud/runner/internal/client"
	"github.com/l8ai-cn/agentcloud/runner/internal/config"
	"github.com/l8ai-cn/agentcloud/runner/internal/relay"
)

type gatedRelayClient struct {
	*relay.MockClient
	started chan struct{}
	release chan struct{}
}

func (c *gatedRelayClient) Connect() error {
	close(c.started)
	<-c.release
	return c.MockClient.Connect()
}

type gatedTunnelClient struct {
	fakeTunnel
	started chan struct{}
	release chan struct{}
}

func (c *gatedTunnelClient) Connect() error {
	close(c.started)
	<-c.release
	c.connected = true
	return nil
}

func TestRelayReadySerializesDifferentRelayChanges(t *testing.T) {
	store := NewInMemoryPodStore()
	pod := &Pod{PodKey: "pod-1", Status: PodStatusRunning}
	store.Put(pod.PodKey, pod)
	h := NewRunnerMessageHandler(
		&Runner{cfg: &config.Config{}},
		store,
		client.NewMockConnection(),
	)
	started := make(chan struct{})
	release := make(chan struct{})
	first := &gatedRelayClient{
		MockClient: relay.NewMockClient("wss://relay-one.example"),
		started:    started,
		release:    release,
	}
	second := relay.NewMockClient("wss://relay-two.example")
	h.relayClientFactory = func(
		url, _, _ string,
		_ *slog.Logger,
	) relay.RelayClient {
		if url == "wss://relay-one.example" {
			return first
		}
		return second
	}

	firstResult := make(chan error, 1)
	go func() {
		firstResult <- h.OnSubscribePod(client.SubscribePodRequest{
			PodKey:      pod.PodKey,
			RelayURL:    "wss://relay-one.example",
			RunnerToken: "first-token",
		})
	}()
	<-started

	secondResult := make(chan error, 1)
	go func() {
		secondResult <- h.OnSubscribePod(client.SubscribePodRequest{
			PodKey:      pod.PodKey,
			RelayURL:    "wss://relay-two.example",
			RunnerToken: "second-token",
		})
	}()
	select {
	case err := <-secondResult:
		t.Fatalf("second subscription completed before the first relay connection: %v", err)
	case <-time.After(50 * time.Millisecond):
	}
	close(release)

	if err := <-firstResult; err != nil {
		t.Fatalf("first subscription failed: %v", err)
	}
	if err := <-secondResult; err != nil {
		t.Fatalf("second subscription failed: %v", err)
	}
	if got := first.GetToken(); got != "first-token" {
		t.Fatalf("first relay token = %q, want first-token", got)
	}
	if got := pod.GetRelayClient(); got != second {
		t.Fatalf("active relay = %T %p, want relay two %p", got, got, second)
	}
	if got := second.GetToken(); got != "second-token" {
		t.Fatalf("second relay token = %q, want second-token", got)
	}
}

func TestTunnelReadyRejectsRequestSupersededByDifferentGateway(t *testing.T) {
	h := newTunnelTestHandler(t)
	started := make(chan struct{})
	release := make(chan struct{})
	first := &gatedTunnelClient{
		fakeTunnel: fakeTunnel{gatewayURL: "ws://gateway-one.example"},
		started:    started,
		release:    release,
	}
	h.tunnelClientFactory = func(url, _ string) tunnelRunner {
		if url == "ws://gateway-one.example" {
			return first
		}
		return &fakeTunnel{gatewayURL: url}
	}

	firstResult := make(chan error, 1)
	go func() {
		firstResult <- h.OnConnectTunnel(client.ConnectTunnelRequest{
			GatewayURL:  "ws://gateway-one.example",
			TunnelToken: "first-token",
		})
	}()
	<-started

	err := h.OnConnectTunnel(client.ConnectTunnelRequest{
		GatewayURL:  "ws://gateway-two.example",
		TunnelToken: "second-token",
	})
	if err != nil {
		t.Fatalf("second connection failed: %v", err)
	}
	close(release)

	err = <-firstResult
	if err == nil || !strings.Contains(err.Error(), "superseded") {
		t.Fatalf("first connection error = %v, want superseded", err)
	}
}
