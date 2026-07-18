package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/anthropics/agentsmesh/relay/internal/auth"
	relaybackend "github.com/anthropics/agentsmesh/relay/internal/backend"
	"github.com/anthropics/agentsmesh/relay/internal/config"
	"github.com/anthropics/agentsmesh/relay/internal/tunnel"
)

type previewSessionBackendStub struct {
	mu             sync.Mutex
	redeemErr      error
	authorizeErr   error
	redeemCalls    int
	authorizeCalls int
}

func (s *previewSessionBackendStub) RedeemPreviewBootstrap(
	_ context.Context,
	_ string,
	_ relaybackend.PreviewSessionRegistration,
) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.redeemCalls++
	return s.redeemErr
}

func (s *previewSessionBackendStub) AuthorizePreviewSession(
	_ context.Context,
	_ relaybackend.PreviewSessionIdentity,
) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.authorizeCalls++
	return s.authorizeErr
}

func (s *previewSessionBackendStub) setAuthorizeError(err error) {
	s.mu.Lock()
	s.authorizeErr = err
	s.mu.Unlock()
}

func (s *previewSessionBackendStub) authorizationCalls() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.authorizeCalls
}

func newTestPreviewHandler(t *testing.T) *PreviewHandler {
	return newTestPreviewHandlerWithBackend(t, &previewSessionBackendStub{})
}

func newTestPreviewHandlerWithBackend(t *testing.T, backend *previewSessionBackendStub) *PreviewHandler {
	t.Helper()
	v := auth.NewTokenValidator("s3cret", "iss")
	issuer := auth.NewPreviewSessionIssuer("s3cret", "iss")
	registry := tunnel.NewRegistry()
	limiter := tunnel.NewPodLimiter(32, 16, 5*time.Second)
	return NewPreviewHandler(v, issuer, backend, registry, limiter, PreviewConfig{
		ReconnectGrace:    50 * time.Millisecond,
		StreamTimeout:     5 * time.Second,
		StreamWindowBytes: 1 << 20,
		PublicHost:        "example.com",
	})
}

func TestPreview_RejectsUnexpectedHost(t *testing.T) {
	h := newTestPreviewHandler(t)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "https://app.example.com/preview/pod1/index.html", nil)

	h.HandlePreview(rec, req)

	if rec.Code != http.StatusMisdirectedRequest {
		t.Fatalf("expected 421, got %d", rec.Code)
	}
}

func TestPreview_UnauthorizedWithoutToken(t *testing.T) {
	h := newTestPreviewHandler(t)
	rec := httptest.NewRecorder()
	req := previewRequest("GET", "/preview/pod1/index.html", "pod1")
	h.HandlePreview(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestPreview_OfflineReturns502(t *testing.T) {
	h := newTestPreviewHandler(t)
	tok := mustPreviewSessionToken(t, "pod1", 7, "127.0.0.1:3000")
	rec := httptest.NewRecorder()
	req := previewRequest("GET", "/preview/pod1/index.html", "pod1")
	req.AddCookie(&http.Cookie{Name: previewCookieName, Value: tok})
	h.HandlePreview(rec, req) // registry has no runner 7
	if rec.Code != http.StatusBadGateway {
		t.Fatalf("expected 502, got %d", rec.Code)
	}
	if h.sessionBackend.(*previewSessionBackendStub).authorizationCalls() != 1 {
		t.Fatal("preview request must reauthorize the active session")
	}
}

func TestPreview_TokenPodKeyMismatchRejected(t *testing.T) {
	h := newTestPreviewHandler(t)
	tok := mustPreviewSessionToken(t, "other-pod", 7, "127.0.0.1:3000")
	rec := httptest.NewRecorder()
	req := previewRequest("GET", "/preview/pod1/index.html", "pod1")
	req.AddCookie(&http.Cookie{Name: previewCookieName, Value: tok})
	h.HandlePreview(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for pod_key mismatch, got %d", rec.Code)
	}
}

func TestPreview_NonPreviewTokenRejected(t *testing.T) {
	h := newTestPreviewHandler(t)
	tok, err := auth.GenerateTypedToken("s3cret", "iss", auth.TokenTypeBrowser, "", 7, 42, 3, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()
	req := previewRequest("GET", "/preview/pod1/index.html", "pod1")
	req.AddCookie(&http.Cookie{Name: previewCookieName, Value: tok})
	h.HandlePreview(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for non-preview token, got %d", rec.Code)
	}
}

func TestPreview_DoesNotAcceptBootstrapQueryToken(t *testing.T) {
	h := newTestPreviewHandler(t)
	token := mustPreviewBootstrapToken(t, "pod1", 7, "127.0.0.1:3000")
	rec := httptest.NewRecorder()
	req := previewRequest("GET", "/preview/pod1/index.html?token="+token, "pod1")

	h.HandlePreview(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
}

func TestJoinPreviewUpstreamPath(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name string
		base string
		rest string
		want string
	}{
		{name: "root mount root request", base: "/", rest: "", want: "/"},
		{name: "mounted root request", base: "/app", rest: "", want: "/app"},
		{name: "mounted asset", base: "/app", rest: "assets/app.js", want: "/app/assets/app.js"},
		{name: "preserve directory slash", base: "/app", rest: "docs/", want: "/app/docs/"},
		{name: "preserve escaped percent", base: "/files/%25", rest: "report%25.txt", want: "/files/%25/report%25.txt"},
		{name: "preserve double encoding", base: "/app", rest: "%252e%252e", want: "/app/%252e%252e"},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, err := joinPreviewUpstreamPath(tt.base, tt.rest)
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Fatalf("joinPreviewUpstreamPath(%q, %q) = %q, want %q", tt.base, tt.rest, got, tt.want)
			}
		})
	}
}

func TestJoinPreviewUpstreamPathRejectsTraversal(t *testing.T) {
	for _, rest := range []string{"../admin", "%2e%2e/admin"} {
		if _, err := joinPreviewUpstreamPath("/app", rest); err == nil {
			t.Fatalf("expected traversal %q to be rejected", rest)
		}
	}
}

func TestPreviewSession_SetsCookieAndRedirects(t *testing.T) {
	backend := &previewSessionBackendStub{}
	h := newTestPreviewHandlerWithBackend(t, backend)
	tok := mustPreviewBootstrapToken(t, "pod1", 7, "127.0.0.1:3000")
	rec := httptest.NewRecorder()
	req := previewRequest("GET", "/preview/pod1/__session?token="+tok, "pod1")
	h.HandlePreviewSession(rec, req)
	if rec.Code != http.StatusFound {
		t.Fatalf("expected 302, got %d", rec.Code)
	}
	resp := rec.Result()
	found := false
	for _, c := range resp.Cookies() {
		if c.Name == previewCookieName {
			found = true
			if c.Value == tok {
				t.Fatal("cookie must contain a distinct preview_session token")
			}
			if !c.HttpOnly {
				t.Fatalf("cookie must be HttpOnly")
			}
			if !c.Secure {
				t.Fatal("cookie must be Secure")
			}
			if c.SameSite != http.SameSiteStrictMode {
				t.Fatalf("SameSite = %v, want Strict", c.SameSite)
			}
			if c.MaxAge != int(previewSessionTTL.Seconds()) {
				t.Fatalf("MaxAge = %d", c.MaxAge)
			}
			if c.Path != "/preview/pod1" {
				t.Fatalf("unexpected cookie path %q", c.Path)
			}
			if c.Domain != "" {
				t.Fatalf("cookie must remain host-only, domain=%q", c.Domain)
			}
			claims, err := h.validator.ValidatePreviewToken(c.Value, auth.TokenTypePreviewSession, "https://pod1.preview.example.com")
			if err != nil || claims.PodKey != "pod1" {
				t.Fatalf("session token invalid: claims=%+v err=%v", claims, err)
			}
		}
	}
	if !found {
		t.Fatalf("expected gw_preview cookie to be set")
	}
	if loc := resp.Header.Get("Location"); loc != "/preview/pod1/" {
		t.Fatalf("unexpected redirect location %q", loc)
	}
	if resp.Header.Get("Cache-Control") != "no-store" {
		t.Fatal("session response must not be cached")
	}
	if resp.Header.Get("Referrer-Policy") != "no-referrer" {
		t.Fatal("session response must suppress the bootstrap URL referrer")
	}
	if backend.redeemCalls != 1 {
		t.Fatalf("redeem calls = %d, want 1", backend.redeemCalls)
	}
}

func TestPreviewSessionRejectsReplayedBootstrap(t *testing.T) {
	backend := &previewSessionBackendStub{redeemErr: relaybackend.ErrPreviewBootstrapConsumed}
	h := newTestPreviewHandlerWithBackend(t, backend)
	token := mustPreviewBootstrapToken(t, "pod1", 7, "127.0.0.1:3000")
	rec := httptest.NewRecorder()
	req := previewRequest("GET", "/preview/pod1/__session?token="+token, "pod1")

	h.HandlePreviewSession(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
	if len(rec.Result().Cookies()) != 0 {
		t.Fatal("replayed bootstrap must not set a cookie")
	}
}

func TestPreviewSessionFailsClosedWhenBackendUnavailable(t *testing.T) {
	backend := &previewSessionBackendStub{redeemErr: relaybackend.ErrPreviewBootstrapUnavailable}
	h := newTestPreviewHandlerWithBackend(t, backend)
	token := mustPreviewBootstrapToken(t, "pod1", 7, "127.0.0.1:3000")
	rec := httptest.NewRecorder()
	req := previewRequest("GET", "/preview/pod1/__session?token="+token, "pod1")

	h.HandlePreviewSession(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", rec.Code)
	}
}

func TestPreviewRejectsRevokedSession(t *testing.T) {
	backend := &previewSessionBackendStub{authorizeErr: relaybackend.ErrPreviewSessionUnauthorized}
	h := newTestPreviewHandlerWithBackend(t, backend)
	token := mustPreviewSessionToken(t, "pod1", 7, "127.0.0.1:3000")
	recorder := httptest.NewRecorder()
	request := previewRequest("GET", "/preview/pod1/index.html", "pod1")
	request.AddCookie(&http.Cookie{Name: previewCookieName, Value: token})

	h.HandlePreview(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", recorder.Code)
	}
}

func TestPreviewSessionDoesNotRedeemWhenIssuerUnavailable(t *testing.T) {
	backend := &previewSessionBackendStub{}
	h := newTestPreviewHandlerWithBackend(t, backend)
	h.sessionIssuer = nil
	token := mustPreviewBootstrapToken(t, "pod1", 7, "127.0.0.1:3000")
	recorder := httptest.NewRecorder()
	request := previewRequest("GET", "/preview/pod1/__session?token="+token, "pod1")

	h.HandlePreviewSession(recorder, request)

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", recorder.Code)
	}
	backend.mu.Lock()
	redeemCalls := backend.redeemCalls
	backend.mu.Unlock()
	if redeemCalls != 0 {
		t.Fatal("bootstrap must remain redeemable when session signing is unavailable")
	}
}

func mustPreviewToken(t *testing.T, podKey string, runnerID int64, target string) string {
	return mustPreviewSessionToken(t, podKey, runnerID, target)
}

func mustPreviewTokenWithPath(t *testing.T, podKey string, runnerID int64, target, previewPath string) string {
	return mustTypedPreviewToken(t, auth.TokenTypePreviewSession, podKey, runnerID, target, previewPath)
}

func mustPreviewSessionToken(t *testing.T, podKey string, runnerID int64, target string) string {
	return mustTypedPreviewToken(t, auth.TokenTypePreviewSession, podKey, runnerID, target, "/")
}

func mustPreviewBootstrapToken(t *testing.T, podKey string, runnerID int64, target string) string {
	return mustTypedPreviewToken(t, auth.TokenTypePreviewBootstrap, podKey, runnerID, target, "/")
}

func mustTypedPreviewToken(t *testing.T, tokenType auth.TokenType, podKey string, runnerID int64, target, previewPath string) string {
	t.Helper()
	now := time.Now()
	claims := &auth.RelayClaims{
		PodKey:        podKey,
		RunnerID:      runnerID,
		UserID:        42,
		OrgID:         3,
		TokenType:     tokenType,
		PreviewTarget: target,
		PreviewPath:   previewPath,
		PreviewOrigin: "https://" + podKey + ".preview.example.com",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "iss",
			Subject:   podKey,
			ID:        "jti-1",
			Audience:  jwt.ClaimStrings{"https://" + podKey + ".preview.example.com"},
		},
	}
	tok, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte("s3cret"))
	if err != nil {
		t.Fatal(err)
	}
	return tok
}
