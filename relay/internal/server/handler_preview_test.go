package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/anthropics/agentsmesh/relay/internal/auth"
	"github.com/anthropics/agentsmesh/relay/internal/tunnel"
)

func newTestPreviewHandler(t *testing.T) *PreviewHandler {
	t.Helper()
	v := auth.NewTokenValidator("s3cret", "iss")
	registry := tunnel.NewRegistry()
	limiter := tunnel.NewPodLimiter(32, 16, 5*time.Second)
	return NewPreviewHandler(v, registry, limiter, PreviewConfig{
		ReconnectGrace:    50 * time.Millisecond,
		StreamTimeout:     5 * time.Second,
		StreamWindowBytes: 1 << 20,
	})
}

func TestPreview_UnauthorizedWithoutToken(t *testing.T) {
	h := newTestPreviewHandler(t)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/preview/pod1/index.html", nil)
	h.HandlePreview(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestPreview_OfflineReturns502(t *testing.T) {
	h := newTestPreviewHandler(t)
	tok := mustPreviewToken(t, "pod1", 7, "127.0.0.1:3000")
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/preview/pod1/index.html?token="+tok, nil)
	h.HandlePreview(rec, req) // registry has no runner 7
	if rec.Code != http.StatusBadGateway {
		t.Fatalf("expected 502, got %d", rec.Code)
	}
}

func TestPreview_TokenPodKeyMismatchRejected(t *testing.T) {
	h := newTestPreviewHandler(t)
	tok := mustPreviewToken(t, "other-pod", 7, "127.0.0.1:3000")
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/preview/pod1/index.html?token="+tok, nil)
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
	req := httptest.NewRequest("GET", "/preview/pod1/index.html?token="+tok, nil)
	h.HandlePreview(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for non-preview token, got %d", rec.Code)
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
	if _, err := joinPreviewUpstreamPath("/app", "../admin"); err == nil {
		t.Fatal("expected traversal to be rejected")
	}
}

func TestPreviewSession_SetsCookieAndRedirects(t *testing.T) {
	h := newTestPreviewHandler(t)
	tok := mustPreviewToken(t, "pod1", 7, "127.0.0.1:3000")
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/preview/pod1/__session?token="+tok, nil)
	h.HandlePreviewSession(rec, req)
	if rec.Code != http.StatusFound {
		t.Fatalf("expected 302, got %d", rec.Code)
	}
	resp := rec.Result()
	found := false
	for _, c := range resp.Cookies() {
		if c.Name == previewCookieName {
			found = true
			if c.Value != tok {
				t.Fatalf("cookie value mismatch")
			}
			if !c.HttpOnly {
				t.Fatalf("cookie must be HttpOnly")
			}
			if c.Path != "/preview/pod1" {
				t.Fatalf("unexpected cookie path %q", c.Path)
			}
		}
	}
	if !found {
		t.Fatalf("expected gw_preview cookie to be set")
	}
	if loc := resp.Header.Get("Location"); loc != "/preview/pod1/" {
		t.Fatalf("unexpected redirect location %q", loc)
	}
}

// mustPreviewToken mints a preview token bound to a pod_key. Production
// preview tokens are signed by the backend (which does set pod_key); this
// builds an equivalent claim set directly since auth.GenerateTypedToken
// (a gateway-side test helper) doesn't take a pod_key parameter.
func mustPreviewToken(t *testing.T, podKey string, runnerID int64, target string) string {
	return mustPreviewTokenWithPath(t, podKey, runnerID, target, "/")
}

func mustPreviewTokenWithPath(t *testing.T, podKey string, runnerID int64, target, previewPath string) string {
	t.Helper()
	now := time.Now()
	claims := &auth.RelayClaims{
		PodKey:        podKey,
		RunnerID:      runnerID,
		UserID:        42,
		OrgID:         3,
		TokenType:     auth.TokenTypePreview,
		PreviewTarget: target,
		PreviewPath:   previewPath,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "iss",
			Subject:   podKey,
		},
	}
	tok, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte("s3cret"))
	if err != nil {
		t.Fatal(err)
	}
	return tok
}
