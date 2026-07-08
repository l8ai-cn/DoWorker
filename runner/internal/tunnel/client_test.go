package tunnel

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"

	"github.com/anthropics/agentsmesh/runner/internal/tunnelframe"
)

func TestClient_ConnectAndHello(t *testing.T) {
	got := make(chan tunnelframe.HelloPayload, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := (&websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}).Upgrade(w, r, nil)
		if err != nil {
			return
		}
		_, data, err := c.ReadMessage()
		if err != nil {
			return
		}
		f, _ := tunnelframe.Decode(data)
		var hp tunnelframe.HelloPayload
		_ = json.Unmarshal(f.Payload, &hp)
		got <- hp
	}))
	defer srv.Close()

	url := "ws" + strings.TrimPrefix(srv.URL, "http")
	cl := NewClient(context.Background(), url, "tok", 7, 3, nil)
	if err := cl.Connect(); err != nil {
		t.Fatal(err)
	}
	defer cl.Stop()
	select {
	case hp := <-got:
		if hp.RunnerID != "7" {
			t.Fatalf("bad runner id %q", hp.RunnerID)
		}
	case <-time.After(time.Second):
		t.Fatal("no hello received")
	}
}
