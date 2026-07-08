package tunnel

import (
	"context"
	"testing"
	"time"
)

func TestRegistry_WaitForTunnel_Grace(t *testing.T) {
	r := NewRegistry()
	go func() {
		time.Sleep(80 * time.Millisecond)
		r.Register(newTunnelForTest2(5))
	}()
	tun := r.WaitForTunnel(context.Background(), 5, 500*time.Millisecond)
	if tun == nil {
		t.Fatal("expected tunnel within grace")
	}
}

func TestRegistry_WaitForTunnel_Timeout(t *testing.T) {
	r := NewRegistry()
	if tun := r.WaitForTunnel(context.Background(), 9, 50*time.Millisecond); tun != nil {
		t.Fatal("expected nil after grace timeout")
	}
}

func TestRegistry_ReconnectReplaces(t *testing.T) {
	r := NewRegistry()
	old := newTunnelForTest2(1)
	r.Register(old)
	newT := newTunnelForTest2(1)
	r.Register(newT)
	if r.Get(1) != newT {
		t.Fatal("new tunnel should replace old")
	}
	// old connection must have been closed on takeover.
	select {
	case <-old.Closed():
	default:
		t.Fatal("old tunnel should be closed on reconnect takeover")
	}
}

func newTunnelForTest2(runnerID int64) *Tunnel {
	return newTunnel(newMockConn(), runnerID, 0, 1<<20, nil)
}
