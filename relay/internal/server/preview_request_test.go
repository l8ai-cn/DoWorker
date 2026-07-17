package server

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/websocket"
)

func doPreviewGET(t *testing.T, target, podKey, token string) *http.Response {
	t.Helper()
	request, err := http.NewRequest(http.MethodGet, target, nil)
	if err != nil {
		t.Fatal(err)
	}
	request.Host = podKey + ".preview.example.com"
	request.AddCookie(&http.Cookie{Name: previewCookieName, Value: token})
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	return response
}

func dialPreviewWebSocket(
	t *testing.T,
	gateway *httptest.Server,
	podKey string,
	path string,
	token string,
) (*websocket.Conn, *http.Response, error) {
	t.Helper()
	dialer := *websocket.DefaultDialer
	dialer.NetDialContext = func(ctx context.Context, network, _ string) (net.Conn, error) {
		return (&net.Dialer{}).DialContext(ctx, network, gateway.Listener.Addr().String())
	}
	target := "ws://" + podKey + ".preview.example.com" + path
	header := http.Header{
		"Cookie": {previewCookieName + "=" + token},
		"Origin": {"https://" + podKey + ".preview.example.com"},
	}
	return dialer.Dial(target, header)
}
