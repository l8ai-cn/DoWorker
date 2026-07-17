package backend

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRedeemPreviewBootstrap(t *testing.T) {
	for _, test := range []struct {
		name   string
		status int
		want   error
	}{
		{name: "redeemed", status: http.StatusNoContent},
		{name: "replay", status: http.StatusConflict, want: ErrPreviewBootstrapConsumed},
		{name: "unavailable", status: http.StatusServiceUnavailable, want: ErrPreviewBootstrapUnavailable},
	} {
		t.Run(test.name, func(t *testing.T) {
			var request PreviewBootstrapRedeemRequest
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/api/internal/relays/preview-bootstrap/redeem" {
					t.Fatalf("path = %q", r.URL.Path)
				}
				if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
					t.Fatal(err)
				}
				w.WriteHeader(test.status)
			}))
			defer server.Close()

			client := NewClient(server.URL, "internal-secret", "relay-1", "ws://relay", "test", 10)
			err := client.RedeemPreviewBootstrap(context.Background(), "bootstrap-1", PreviewSessionRegistration{
				ID: "session-1", PodKey: "pod-1", UserID: 42, OrgID: 3,
				ExpiresAt: time.Now().Add(time.Minute),
			})
			if !errors.Is(err, test.want) {
				t.Fatalf("error = %v, want %v", err, test.want)
			}
			if request.BootstrapID != "bootstrap-1" || request.SessionID != "session-1" {
				t.Fatalf("request = %+v", request)
			}
		})
	}
}

func TestAuthorizePreviewSession(t *testing.T) {
	for _, test := range []struct {
		status int
		want   error
	}{
		{status: http.StatusNoContent},
		{status: http.StatusUnauthorized, want: ErrPreviewSessionUnauthorized},
		{status: http.StatusServiceUnavailable, want: ErrPreviewSessionUnavailable},
	} {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/api/internal/relays/preview-sessions/authorize" {
				t.Fatalf("path = %q", r.URL.Path)
			}
			w.WriteHeader(test.status)
		}))
		client := NewClient(server.URL, "internal-secret", "relay-1", "ws://relay", "test", 10)
		err := client.AuthorizePreviewSession(context.Background(), PreviewSessionIdentity{
			ID: "session-1", PodKey: "pod-1", UserID: 42, OrgID: 3,
		})
		server.Close()
		if !errors.Is(err, test.want) {
			t.Fatalf("status %d error = %v, want %v", test.status, err, test.want)
		}
	}
}
