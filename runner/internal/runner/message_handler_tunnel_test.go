package runner

import (
	"testing"

	"github.com/l8ai-cn/agentcloud/runner/internal/client"
	"github.com/l8ai-cn/agentcloud/runner/internal/config"
)

type fakeTunnel struct {
	gatewayURL string
	connected  bool
	started    bool
	stopped    bool
}

func (f *fakeTunnel) Connect() error       { f.connected = true; return nil }
func (f *fakeTunnel) Start()               { f.started = true }
func (f *fakeTunnel) Stop()                { f.stopped = true }
func (f *fakeTunnel) UpdateToken(string)   {}
func (f *fakeTunnel) GatewayURL() string   { return f.gatewayURL }
func (f *fakeTunnel) IsConnected() bool    { return f.connected }

type fakeTunnelFactory struct {
	created int
	last    *fakeTunnel
}

func (f *fakeTunnelFactory) New(gatewayURL, token string) tunnelRunner {
	f.created++
	f.last = &fakeTunnel{gatewayURL: gatewayURL}
	return f.last
}

func newTunnelTestHandler(t *testing.T) *RunnerMessageHandler {
	t.Helper()
	cfg := &config.Config{WorkspaceRoot: t.TempDir()}
	r := &Runner{cfg: cfg, podStore: NewInMemoryPodStore()}
	mockConn := client.NewMockConnection()
	return NewRunnerMessageHandler(r, r.podStore, mockConn)
}

func TestOnConnectTunnel_StartsClient(t *testing.T) {
	h := newTunnelTestHandler(t)
	factory := &fakeTunnelFactory{}
	h.tunnelClientFactory = factory.New

	err := h.OnConnectTunnel(client.ConnectTunnelRequest{
		GatewayURL:  "ws://127.0.0.1:1/relay",
		TunnelToken: "tok",
	})
	if err != nil {
		t.Fatal(err)
	}
	if factory.created != 1 {
		t.Fatalf("expected 1 client, got %d", factory.created)
	}
	if !factory.last.started {
		t.Fatal("expected client to be started")
	}
}

func TestOnConnectTunnel_ReuseSameGateway(t *testing.T) {
	h := newTunnelTestHandler(t)
	factory := &fakeTunnelFactory{}
	h.tunnelClientFactory = factory.New

	req := client.ConnectTunnelRequest{GatewayURL: "ws://127.0.0.1:1/relay", TunnelToken: "tok"}
	if err := h.OnConnectTunnel(req); err != nil {
		t.Fatal(err)
	}
	// Second call to the same gateway should reuse (update token), not recreate.
	if err := h.OnConnectTunnel(req); err != nil {
		t.Fatal(err)
	}
	if factory.created != 1 {
		t.Fatalf("expected reuse (1 client), got %d", factory.created)
	}
}
