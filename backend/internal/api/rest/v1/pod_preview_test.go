package v1

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sort"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	relaysvc "github.com/anthropics/agentsmesh/backend/internal/service/relay"
)

type mockPreviewRelaySelector struct{ info *relaysvc.RelayInfo }

func (m *mockPreviewRelaySelector) SelectRelayForPodGeo(relaysvc.GeoSelectOptions) *relaysvc.RelayInfo {
	return m.info
}

type mockPreviewTokens struct {
	tokenTypes    []string
	previewPath   string
	previewOrigin string
}

func (m *mockPreviewTokens) GenerateTypedToken(podKey string, runnerID, userID, orgID int64, tokenType, previewTarget string, expiry time.Duration) (string, error) {
	m.tokenTypes = append(m.tokenTypes, tokenType)
	return "JWT-" + tokenType, nil
}

func (m *mockPreviewTokens) GeneratePreviewBootstrapToken(podKey string, runnerID, userID, orgID int64, previewTarget, previewPath, previewOrigin string, expiry time.Duration) (string, error) {
	m.tokenTypes = append(m.tokenTypes, "preview_bootstrap")
	m.previewPath = previewPath
	m.previewOrigin = previewOrigin
	return "JWT-preview-bootstrap", nil
}

type previewCommandSender struct {
	mockCommandSender
	sendConnectTunnelFn func(context.Context, int64, string, string) error
}

func (m *previewCommandSender) SendConnectTunnel(ctx context.Context, runnerID int64, tunnelURL, token string) error {
	if m.sendConnectTunnelFn != nil {
		return m.sendConnectTunnelFn(ctx, runnerID, tunnelURL, token)
	}
	return nil
}

func newPreviewHandler(pod *agentpod.Pod) *PodHandler {
	return &PodHandler{
		podService: &mockPodService{getPodFn: func(ctx context.Context, key string) (*agentpod.Pod, error) {
			return pod, nil
		}},
		commandSender:       &previewCommandSender{},
		relaySelector:       &mockPreviewRelaySelector{info: &relaysvc.RelayInfo{URL: "wss://relay.example.com/relay"}},
		relayTokens:         &mockPreviewTokens{},
		previewPublicOrigin: "https://preview.example.com",
	}
}

func performPreviewGET(h *PodHandler) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/pods/pod1/preview", nil)
	c.Params = gin.Params{{Key: "key", Value: "pod1"}}
	setPodTenantContext(c, 1, 10)
	h.GetPodPreview(c)
	return w
}

func TestGetPodPreview_ReturnsSessionURLWithoutRawToken(t *testing.T) {
	pod := &agentpod.Pod{PodKey: "pod1", RunnerID: 7, PreviewPort: 3000, PreviewPath: "/files/%25", Status: agentpod.StatusRunning, OrganizationID: 1, CreatedByID: 10}
	h := newPreviewHandler(pod)
	w := performPreviewGET(h)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
	}
	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, []string{"expires_at", "preview_base_url", "session_url"}, sortedKeys(resp))
	assert.Equal(t, "https://preview.example.com/preview/pod1/", resp["preview_base_url"])
	assert.Equal(t, "https://preview.example.com/preview/pod1/__session?token=JWT-preview", resp["session_url"])
	assert.NotEmpty(t, resp["expires_at"])
	assert.NotContains(t, resp, "token")
	assert.Equal(t, "/files/%25", h.relayTokens.(*mockPreviewTokens).previewPath)
	assert.Equal(t, "https://pod1.preview.example.com", h.relayTokens.(*mockPreviewTokens).previewOrigin)
}

func TestGetPodPreview_MissingPublicOriginReturns503(t *testing.T) {
	pod := &agentpod.Pod{PodKey: "pod1", RunnerID: 7, PreviewPort: 3000, Status: agentpod.StatusRunning, OrganizationID: 1, CreatedByID: 10}
	h := newPreviewHandler(pod)
	h.previewPublicOrigin = ""

	w := performPreviewGET(h)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	assert.JSONEq(t, `{"code":"preview_unavailable","error":"Preview is not available"}`, w.Body.String())
}

func TestGetPodPreview_MissingPublicOriginReturns503(t *testing.T) {
	pod := &agentpod.Pod{PodKey: "pod1", RunnerID: 7, PreviewPort: 3000, Status: agentpod.StatusRunning, OrganizationID: 1, CreatedByID: 10}
	h := newPreviewHandler(pod)
	h.previewPublicOrigin = ""

	w := performPreviewGET(h)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	assert.JSONEq(t, `{"code":"preview_unavailable","error":"Preview is not available"}`, w.Body.String())
}

func TestGetPodPreview_MissingCommandSenderReturns503(t *testing.T) {
	pod := &agentpod.Pod{PodKey: "pod1", RunnerID: 7, PreviewPort: 3000, Status: agentpod.StatusRunning, OrganizationID: 1, CreatedByID: 10}
	h := newPreviewHandler(pod)
	h.commandSender = nil

	w := performPreviewGET(h)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	assert.JSONEq(t, `{"code":"preview_unavailable","error":"Preview is not available"}`, w.Body.String())
}

func TestGetPodPreview_TunnelDispatchFailureReturns503WithoutPreviewToken(t *testing.T) {
	pod := &agentpod.Pod{PodKey: "pod1", RunnerID: 7, PreviewPort: 3000, Status: agentpod.StatusRunning, OrganizationID: 1, CreatedByID: 10}
	h := newPreviewHandler(pod)
	tokens := h.relayTokens.(*mockPreviewTokens)
	h.commandSender = &previewCommandSender{
		sendConnectTunnelFn: func(context.Context, int64, string, string) error {
			return errors.New("runner unavailable")
		},
	}

	w := performPreviewGET(h)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	assert.JSONEq(t, `{"code":"preview_unavailable","error":"Preview is not available"}`, w.Body.String())
	assert.Equal(t, []string{"tunnel"}, tokens.tokenTypes)
}

func TestGetPodPreview_DisabledReturns404(t *testing.T) {
	pod := &agentpod.Pod{PodKey: "pod1", RunnerID: 7, PreviewPort: 0, Status: agentpod.StatusRunning, OrganizationID: 1, CreatedByID: 10}
	w := performPreviewGET(newPreviewHandler(pod))
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestGetPodPreview_InactiveReturns409(t *testing.T) {
	pod := &agentpod.Pod{PodKey: "pod1", RunnerID: 7, PreviewPort: 3000, Status: agentpod.StatusCompleted, OrganizationID: 1, CreatedByID: 10}
	w := performPreviewGET(newPreviewHandler(pod))
	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", w.Code)
	}
}

func sortedKeys(values map[string]any) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
