package proxy

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/l8ai-cn/agentcloud/relay/internal/protocol/tunnelframe"
	"github.com/l8ai-cn/agentcloud/relay/internal/tunnel"
)

func TestPreviewWebSocketOriginCheckIsExact(t *testing.T) {
	for _, test := range []struct {
		name   string
		origin string
		want   bool
	}{
		{name: "exact", origin: "https://pod1.preview.example.com", want: true},
		{name: "sibling pod", origin: "https://pod2.preview.example.com"},
		{name: "application", origin: "https://app.example.com"},
		{name: "missing"},
	} {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "https://pod1.preview.example.com/preview/pod1/socket", nil)
			if test.origin != "" {
				request.Header.Set("Origin", test.origin)
			}
			if got := previewWebSocketOriginAllowed(request, "https://pod1.preview.example.com"); got != test.want {
				t.Fatalf("allowed = %v, want %v", got, test.want)
			}
		})
	}
}

func TestProxyWebSocketRejectsWrongOriginBeforeUpgrade(t *testing.T) {
	fake := newFakeTunnel()
	fake.onReqStart = func(stream *tunnel.Stream, _ tunnelframe.ReqStartPayload) {
		fake.inject(stream.ID, tunnelframe.TypeRespStart, mustJSON(tunnelframe.RespStartPayload{
			Status: http.StatusSwitchingProtocols,
		}))
	}
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "https://pod1.preview.example.com/preview/pod1/socket", nil)
	request.Header.Set("Connection", "Upgrade")
	request.Header.Set("Upgrade", "websocket")
	request.Header.Set("Sec-WebSocket-Version", "13")
	request.Header.Set("Sec-WebSocket-Key", "dGhlIHNhbXBsZSBub25jZQ==")
	request.Header.Set("Origin", "https://pod2.preview.example.com")

	err := ProxyWebSocket(context.Background(), fake, recorder, request, ProxyParams{
		PodKey:         "pod1",
		Target:         "127.0.0.1:3000",
		Path:           "/socket",
		ExpectedOrigin: "https://pod1.preview.example.com",
	})

	if err == nil {
		t.Fatal("expected wrong origin to reject the websocket upgrade")
	}
	if recorder.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", recorder.Code)
	}
}
