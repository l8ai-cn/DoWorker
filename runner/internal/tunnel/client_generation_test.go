package tunnel

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"

	"github.com/anthropics/agentsmesh/runner/internal/tunnelframe"
)

type blockingDispatcher struct {
	entered chan struct{}
	closed  chan struct{}
	once    sync.Once
}

func (d *blockingDispatcher) Dispatch(tunnelframe.Frame) {
	close(d.entered)
	<-d.closed
}

func (d *blockingDispatcher) SetSender(func(tunnelframe.Frame) error) {}
func (d *blockingDispatcher) Close() {
	d.once.Do(func() { close(d.closed) })
}

func TestClient_StopInterruptsBlockingDispatch(t *testing.T) {
	dispatcher := &blockingDispatcher{entered: make(chan struct{}), closed: make(chan struct{})}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgradeTunnel(w, r)
		if err != nil {
			return
		}
		defer conn.Close()
		readHello(t, conn)
		_ = conn.WriteMessage(websocket.BinaryMessage, tunnelframe.Encode(tunnelframe.Frame{Type: tunnelframe.TypeReqEnd}))
		<-r.Context().Done()
	}))
	defer srv.Close()

	client := NewClient(context.Background(), websocketURL(srv.URL), "tok", 7, 3, dispatcher)
	if err := client.Connect(); err != nil {
		t.Fatal(err)
	}
	client.Start()
	select {
	case <-dispatcher.entered:
	case <-time.After(time.Second):
		t.Fatal("dispatcher did not block")
	}

	stopped := make(chan struct{})
	go func() {
		client.Stop()
		close(stopped)
	}()
	select {
	case <-stopped:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Stop blocked behind Dispatch")
	}
}

type capturedSenderDispatcher struct {
	mu       sync.Mutex
	send     func(tunnelframe.Frame) error
	captured chan func(tunnelframe.Frame) error
}

func (d *capturedSenderDispatcher) Dispatch(tunnelframe.Frame) {
	d.mu.Lock()
	send := d.send
	d.mu.Unlock()
	d.captured <- send
}

func (d *capturedSenderDispatcher) SetSender(send func(tunnelframe.Frame) error) {
	d.mu.Lock()
	d.send = send
	d.mu.Unlock()
}

func (d *capturedSenderDispatcher) Close() {}

func TestClient_OldGenerationSenderCannotWriteToReplacement(t *testing.T) {
	dispatcher := &capturedSenderDispatcher{captured: make(chan func(tunnelframe.Frame) error, 1)}
	secondConnected := make(chan *websocket.Conn, 1)
	var connections int
	var mu sync.Mutex
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgradeTunnel(w, r)
		if err != nil {
			return
		}
		defer conn.Close()
		readHello(t, conn)
		mu.Lock()
		connections++
		number := connections
		mu.Unlock()
		if number == 1 {
			_ = conn.WriteMessage(websocket.BinaryMessage, tunnelframe.Encode(tunnelframe.Frame{Type: tunnelframe.TypeReqEnd}))
			return
		}
		secondConnected <- conn
		<-r.Context().Done()
	}))
	defer srv.Close()

	client := newTestClient(srv.URL)
	client.dispatcher = dispatcher
	dispatcher.SetSender(func(tunnelframe.Frame) error { return errors.New("not connected") })
	if err := client.Connect(); err != nil {
		t.Fatal(err)
	}
	client.Start()
	defer client.Stop()

	var oldSend func(tunnelframe.Frame) error
	select {
	case oldSend = <-dispatcher.captured:
	case <-time.After(time.Second):
		t.Fatal("old generation sender was not captured")
	}
	var replacement *websocket.Conn
	select {
	case replacement = <-secondConnected:
	case <-time.After(time.Second):
		t.Fatal("replacement connection was not established")
	}
	if err := oldSend(tunnelframe.Frame{Type: tunnelframe.TypeRespEnd, StreamID: 1}); err == nil {
		t.Fatal("old generation sender unexpectedly succeeded")
	}
	_ = replacement.SetReadDeadline(time.Now().Add(50 * time.Millisecond))
	if _, _, err := replacement.ReadMessage(); err == nil {
		t.Fatal("old generation frame reached replacement connection")
	}
}
