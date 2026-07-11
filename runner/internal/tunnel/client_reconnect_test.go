package tunnel

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gorilla/websocket"

	"github.com/anthropics/agentsmesh/runner/internal/tunnelframe"
)

func TestClient_ReconnectsAndSendsHelloAfterServerDisconnect(t *testing.T) {
	var connections atomic.Int32
	secondHello := make(chan tunnelframe.HelloPayload, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := (&websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}).Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		hello := readHello(t, conn)
		if connections.Add(1) == 1 {
			return
		}
		secondHello <- hello
	}))
	defer srv.Close()

	cl := NewClient(context.Background(), websocketURL(srv.URL), "tok", 7, 3, nil)
	if err := cl.Connect(); err != nil {
		t.Fatal(err)
	}
	cl.Start()
	defer cl.Stop()

	select {
	case hello := <-secondHello:
		if hello.RunnerID != "7" || hello.OrgID != "3" {
			t.Fatalf("unexpected hello: %+v", hello)
		}
	case <-time.After(time.Second):
		t.Fatal("client did not reconnect and send a second HELLO")
	}
}

func TestClient_StartLaunchesOneReconnectLoop(t *testing.T) {
	var connections atomic.Int32
	secondHello := make(chan struct{}, 1)
	release := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgradeTunnel(w, r)
		if err != nil {
			return
		}
		defer conn.Close()
		readHello(t, conn)
		if connections.Add(1) == 1 {
			return
		}
		select {
		case secondHello <- struct{}{}:
		default:
		}
		<-release
	}))
	defer srv.Close()

	cl := newTestClient(srv.URL)
	if err := cl.Connect(); err != nil {
		t.Fatal(err)
	}
	for range 10 {
		cl.Start()
	}
	select {
	case <-secondHello:
	case <-time.After(time.Second):
		t.Fatal("second connection not established")
	}
	time.Sleep(30 * time.Millisecond)
	if got := connections.Load(); got != 2 {
		t.Fatalf("connections = %d, want 2", got)
	}
	close(release)
	cl.Stop()
}

func TestClient_ReconnectUsesExponentialBackoff(t *testing.T) {
	var connections atomic.Int32
	thirdHello := make(chan struct{}, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgradeTunnel(w, r)
		if err != nil {
			return
		}
		defer conn.Close()
		readHello(t, conn)
		if connections.Add(1) == 3 {
			thirdHello <- struct{}{}
		}
	}))
	defer srv.Close()

	var mu sync.Mutex
	var delays []time.Duration
	cl := newTestClient(srv.URL)
	cl.reconnect.jitter = func(delay time.Duration) time.Duration {
		mu.Lock()
		delays = append(delays, delay)
		mu.Unlock()
		return time.Millisecond
	}
	if err := cl.Connect(); err != nil {
		t.Fatal(err)
	}
	cl.Start()
	select {
	case <-thirdHello:
	case <-time.After(time.Second):
		t.Fatal("third connection not established")
	}
	cl.Stop()

	mu.Lock()
	defer mu.Unlock()
	if len(delays) < 2 || delays[0] != 10*time.Millisecond || delays[1] != 20*time.Millisecond {
		t.Fatalf("backoff inputs = %v, want prefix [10ms 20ms]", delays)
	}
}

func TestClient_ReconnectsAfterDialFailure(t *testing.T) {
	reserved, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	address := reserved.Addr().String()
	if err := reserved.Close(); err != nil {
		t.Fatal(err)
	}

	retryScheduled := make(chan struct{}, 1)
	helloReceived := make(chan struct{}, 1)
	cl := NewClient(context.Background(), "ws://"+address, "tok", 7, 3, nil)
	cl.reconnect = reconnectPolicy{
		initial: 20 * time.Millisecond, maximum: 40 * time.Millisecond,
		maxPermanent: maxPermanentConnectFailures,
		jitter: func(delay time.Duration) time.Duration {
			select {
			case retryScheduled <- struct{}{}:
			default:
			}
			return delay
		},
	}
	cl.Start()
	select {
	case <-retryScheduled:
	case <-time.After(time.Second):
		t.Fatal("dial failure did not schedule a retry")
	}

	listener, err := net.Listen("tcp", address)
	if err != nil {
		t.Fatal(err)
	}
	server := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, upgradeErr := upgradeTunnel(w, r)
		if upgradeErr != nil {
			return
		}
		defer conn.Close()
		readHello(t, conn)
		helloReceived <- struct{}{}
		<-cl.ctx.Done()
	})}
	go server.Serve(listener)
	defer server.Shutdown(context.Background())
	defer cl.Stop()

	select {
	case <-helloReceived:
	case <-time.After(time.Second):
		t.Fatal("client did not reconnect after dial failure")
	}
}

func TestClient_StopInterruptsReconnectBackoffAndIsIdempotent(t *testing.T) {
	backoffStarted := make(chan struct{}, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgradeTunnel(w, r)
		if err != nil {
			return
		}
		defer conn.Close()
		readHello(t, conn)
	}))
	defer srv.Close()

	cl := newTestClient(srv.URL)
	cl.reconnect.initial = time.Hour
	cl.reconnect.maximum = time.Hour
	cl.reconnect.jitter = func(delay time.Duration) time.Duration {
		backoffStarted <- struct{}{}
		return delay
	}
	if err := cl.Connect(); err != nil {
		t.Fatal(err)
	}
	cl.Start()
	select {
	case <-backoffStarted:
	case <-time.After(time.Second):
		t.Fatal("reconnect backoff did not start")
	}

	started := time.Now()
	cl.Stop()
	cl.Stop()
	if elapsed := time.Since(started); elapsed > 100*time.Millisecond {
		t.Fatalf("Stop took %v while interrupting backoff", elapsed)
	}
}

func TestClient_StopInterruptsHandshake(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	accepted := make(chan net.Conn, 1)
	go func() {
		conn, acceptErr := listener.Accept()
		if acceptErr == nil {
			accepted <- conn
		}
	}()

	cl := NewClient(context.Background(), "ws://"+listener.Addr().String(), "tok", 7, 3, nil)
	cl.Start()
	var serverConn net.Conn
	select {
	case serverConn = <-accepted:
		defer serverConn.Close()
	case <-time.After(time.Second):
		t.Fatal("handshake connection not accepted")
	}

	started := time.Now()
	cl.Stop()
	if elapsed := time.Since(started); elapsed > 100*time.Millisecond {
		t.Fatalf("Stop took %v while interrupting handshake", elapsed)
	}
}

func TestClient_PermanentHandshakeFailuresAreLimited(t *testing.T) {
	var requests atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	}))
	defer srv.Close()

	cl := newTestClient(srv.URL)
	cl.reconnect.maxPermanent = 3
	cl.reconnect.jitter = func(time.Duration) time.Duration { return 0 }
	cl.Start()
	waitForCount(t, &requests, 3)
	time.Sleep(30 * time.Millisecond)
	cl.Stop()
	if got := requests.Load(); got != 3 {
		t.Fatalf("requests = %d, want 3", got)
	}
}

func TestClient_OldGenerationCannotDisconnectReplacement(t *testing.T) {
	var connections atomic.Int32
	secondHello := make(chan struct{}, 1)
	release := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgradeTunnel(w, r)
		if err != nil {
			return
		}
		defer conn.Close()
		readHello(t, conn)
		if connections.Add(1) == 1 {
			<-release
			return
		}
		select {
		case secondHello <- struct{}{}:
		default:
		}
		<-release
	}))
	defer srv.Close()

	cl := newTestClient(srv.URL)
	if err := cl.Connect(); err != nil {
		t.Fatal(err)
	}
	cl.Start()
	if err := cl.Connect(); err != nil {
		t.Fatal(err)
	}
	select {
	case <-secondHello:
	case <-time.After(time.Second):
		t.Fatal("replacement connection not established")
	}
	time.Sleep(50 * time.Millisecond)
	if !cl.IsConnected() {
		t.Fatal("old generation marked replacement disconnected")
	}
	if got := connections.Load(); got != 2 {
		t.Fatalf("connections = %d, want 2", got)
	}
	close(release)
	cl.Stop()
}

func newTestClient(serverURL string) *Client {
	cl := NewClient(context.Background(), websocketURL(serverURL), "tok", 7, 3, nil)
	cl.reconnect = reconnectPolicy{
		initial:      10 * time.Millisecond,
		maximum:      40 * time.Millisecond,
		maxPermanent: maxPermanentConnectFailures,
		jitter:       func(delay time.Duration) time.Duration { return delay },
	}
	return cl
}

func upgradeTunnel(w http.ResponseWriter, r *http.Request) (*websocket.Conn, error) {
	return (&websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}).Upgrade(w, r, nil)
}

func waitForCount(t *testing.T, count *atomic.Int32, want int32) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for count.Load() < want {
		if time.Now().After(deadline) {
			t.Fatalf("count = %d, want %d", count.Load(), want)
		}
		time.Sleep(time.Millisecond)
	}
}

func readHello(t *testing.T, conn *websocket.Conn) tunnelframe.HelloPayload {
	t.Helper()
	hello := readHelloOnly(t, conn)
	if err := conn.WriteMessage(websocket.BinaryMessage, tunnelframe.Encode(tunnelframe.Frame{Type: tunnelframe.TypeHelloAck})); err != nil {
		t.Errorf("write HELLO_ACK: %v", err)
	}
	return hello
}

func readHelloOnly(t *testing.T, conn *websocket.Conn) tunnelframe.HelloPayload {
	t.Helper()
	_, data, err := conn.ReadMessage()
	if err != nil {
		t.Errorf("read HELLO: %v", err)
		return tunnelframe.HelloPayload{}
	}
	frame, err := tunnelframe.Decode(data)
	if err != nil {
		t.Errorf("decode HELLO: %v", err)
		return tunnelframe.HelloPayload{}
	}
	if frame.Type != tunnelframe.TypeHello {
		t.Errorf("frame type = %v, want HELLO", frame.Type)
	}
	var hello tunnelframe.HelloPayload
	if err := json.Unmarshal(frame.Payload, &hello); err != nil {
		t.Errorf("decode HELLO payload: %v", err)
	}
	return hello
}

func websocketURL(serverURL string) string {
	return "ws" + strings.TrimPrefix(serverURL, "http")
}
