package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/l8ai-cn/agentcloud/relay/internal/config"
)

func TestPreviewSessionSetsPartitionedCookieForExternalEmbed(t *testing.T) {
	h := newTestPreviewHandler(t)
	h.cfg.CookieMode = config.PreviewCookiePartitioned
	token := mustPreviewBootstrapToken(t, "pod1", 7, "127.0.0.1:3000")
	recorder := httptest.NewRecorder()
	request := previewRequest(
		http.MethodGet,
		"/preview/pod1/__session?token="+token,
		"pod1",
	)

	h.HandlePreviewSession(recorder, request)

	cookies := recorder.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("cookies = %d, want 1", len(cookies))
	}
	cookie := cookies[0]
	if !cookie.Partitioned {
		t.Fatal("external embed cookie must be Partitioned")
	}
	if cookie.SameSite != http.SameSiteNoneMode {
		t.Fatalf("SameSite = %v, want None", cookie.SameSite)
	}
	if !cookie.Secure {
		t.Fatal("partitioned cookie must be Secure")
	}
}
