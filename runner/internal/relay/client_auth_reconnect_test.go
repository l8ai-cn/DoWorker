package relay

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestAuthFailureCircuitBreaker(t *testing.T) {
	var connectionAttempts atomic.Int32
	var closeCalled atomic.Bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		connectionAttempts.Add(1)
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	c := NewClient(context.TODO(), "ws"+strings.TrimPrefix(srv.URL, "http"), "pod-1", "test-token", nil)
	c.SetCloseHandler(func() { closeCalled.Store(true) })
	c.SetTokenExpiredHandler(func() string { return "" })
	close(c.writeExitCh)
	c.wg.Add(1)
	c.reconnecting.Store(true)
	go c.reconnectLoop()

	deadline := time.After(30 * time.Second)
	for {
		select {
		case <-deadline:
			t.Fatal("timeout: reconnect loop did not stop after auth failures")
		default:
			if closeCalled.Load() {
				if attempts := connectionAttempts.Load(); attempts > int32(maxConsecutiveAuthFailures)+1 {
					t.Errorf("too many attempts: got %d, want <= %d", attempts, maxConsecutiveAuthFailures+1)
				}
				return
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func TestAuthFailureResetsOnTransientError(t *testing.T) {
	var connectionAttempts atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempt := connectionAttempts.Add(1)
		if attempt <= 3 {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if attempt == 4 {
			w.Header().Set("Connection", "close")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		conn, err := upgradeReadyPublisher(w, r)
		if err != nil {
			return
		}
		defer conn.Close()
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	}))
	defer srv.Close()

	c := NewClient(context.TODO(), "ws"+strings.TrimPrefix(srv.URL, "http"), "pod-1", "test-token", nil)
	c.SetTokenExpiredHandler(func() string { return "" })
	reconnected := make(chan struct{})
	c.SetReconnectHandler(func() { close(reconnected) })
	close(c.writeExitCh)
	c.wg.Add(1)
	c.reconnecting.Store(true)
	go c.reconnectLoop()
	defer c.Stop()

	select {
	case <-reconnected:
		if connectionAttempts.Load() < 5 {
			t.Errorf("expected at least 5 attempts, got %d", connectionAttempts.Load())
		}
	case <-time.After(30 * time.Second):
		t.Fatal("timeout: should have reconnected after transient error reset counter")
	}
}

func TestAuthFailure_TokenRefreshSuccessThenFailAgain(t *testing.T) {
	var connectionAttempts atomic.Int32
	var closeCalled atomic.Bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		connectionAttempts.Add(1)
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	c := NewClient(context.TODO(), "ws"+strings.TrimPrefix(srv.URL, "http"), "pod-1", "test-token", nil)
	c.SetCloseHandler(func() { closeCalled.Store(true) })
	c.SetTokenExpiredHandler(func() string { return "refreshed-token" })
	close(c.writeExitCh)
	c.wg.Add(1)
	c.reconnecting.Store(true)
	go c.reconnectLoop()

	deadline := time.After(30 * time.Second)
	for {
		select {
		case <-deadline:
			t.Fatal("timeout: circuit breaker should have triggered")
		default:
			if closeCalled.Load() {
				if attempts := connectionAttempts.Load(); attempts > int32(maxConsecutiveAuthFailures)+2 {
					t.Errorf("too many attempts after refresh: got %d", attempts)
				}
				return
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func TestStopIdempotent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgradeReadyPublisher(w, r)
		if err != nil {
			return
		}
		defer conn.Close()
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	}))
	defer srv.Close()

	c := NewClient(context.TODO(), "ws"+strings.TrimPrefix(srv.URL, "http"), "pod-1", "test-token", nil)
	if err := c.Connect(); err != nil {
		t.Fatalf("Connect: %v", err)
	}
	c.Start()

	done := make(chan struct{})
	for range 5 {
		go c.Stop()
	}
	go func() {
		time.Sleep(100 * time.Millisecond)
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(6 * time.Second):
		t.Error("Multiple Stop() calls hung")
	}
}
