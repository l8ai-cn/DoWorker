package internal

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	previewsvc "github.com/l8ai-cn/agentcloud/backend/internal/service/preview"
	"github.com/gin-gonic/gin"
)

type previewSessionServiceStub struct {
	err       error
	bootstrap string
	record    previewsvc.SessionRecord
	identity  previewsvc.SessionIdentity
}

func (s *previewSessionServiceStub) Redeem(_ context.Context, bootstrap string, record previewsvc.SessionRecord, _ time.Duration) error {
	s.bootstrap = bootstrap
	s.record = record
	return s.err
}

func (s *previewSessionServiceStub) Authorize(_ context.Context, identity previewsvc.SessionIdentity) error {
	s.identity = identity
	return s.err
}

func TestRedeemPreviewBootstrap(t *testing.T) {
	gin.SetMode(gin.TestMode)
	expiresAt := time.Now().Add(15 * time.Minute).UTC().Format(time.RFC3339)
	for _, test := range []struct {
		name   string
		err    error
		status int
	}{
		{name: "redeemed", status: http.StatusNoContent},
		{name: "replay", err: previewsvc.ErrBootstrapConsumed, status: http.StatusConflict},
		{name: "unavailable", err: previewsvc.ErrStoreUnavailable, status: http.StatusServiceUnavailable},
	} {
		t.Run(test.name, func(t *testing.T) {
			service := &previewSessionServiceStub{err: test.err}
			router := gin.New()
			RegisterPreviewSessionRoutes(router.Group("/api/internal/relays"), service)
			body := `{"bootstrap_id":"bootstrap-1","session_id":"session-1","pod_key":"pod-1","user_id":42,"org_id":3,"expires_at":"` + expiresAt + `"}`
			request := httptest.NewRequest(http.MethodPost, "/api/internal/relays/preview-bootstrap/redeem", bytes.NewBufferString(body))
			request.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()

			router.ServeHTTP(response, request)

			if response.Code != test.status {
				t.Fatalf("status = %d, want %d; body=%s", response.Code, test.status, response.Body.String())
			}
			if test.status == http.StatusNoContent &&
				(service.bootstrap != "bootstrap-1" || service.record.ID != "session-1") {
				t.Fatalf("redeem args = %q %+v", service.bootstrap, service.record)
			}
		})
	}
}

func TestAuthorizePreviewSession(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, test := range []struct {
		name   string
		err    error
		status int
	}{
		{name: "authorized", status: http.StatusNoContent},
		{name: "denied", err: previewsvc.ErrSessionUnauthorized, status: http.StatusUnauthorized},
		{name: "unavailable", err: previewsvc.ErrAuthorizationUnavailable, status: http.StatusServiceUnavailable},
	} {
		t.Run(test.name, func(t *testing.T) {
			service := &previewSessionServiceStub{err: test.err}
			router := gin.New()
			RegisterPreviewSessionRoutes(router.Group("/api/internal/relays"), service)
			request := httptest.NewRequest(
				http.MethodPost,
				"/api/internal/relays/preview-sessions/authorize",
				bytes.NewBufferString(`{"session_id":"session-1","pod_key":"pod-1","user_id":42,"org_id":3}`),
			)
			request.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()

			router.ServeHTTP(response, request)

			if response.Code != test.status {
				t.Fatalf("status = %d, want %d; body=%s", response.Code, test.status, response.Body.String())
			}
			if test.status == http.StatusNoContent && service.identity.ID != "session-1" {
				t.Fatalf("authorize identity = %+v", service.identity)
			}
		})
	}
}
