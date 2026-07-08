package v1

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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

type mockPreviewTokens struct{}

func (m *mockPreviewTokens) GenerateTypedToken(podKey string, runnerID, userID, orgID int64, tokenType, previewTarget string, expiry time.Duration) (string, error) {
	return "JWT-" + tokenType, nil
}

func newPreviewHandler(pod *agentpod.Pod) *PodHandler {
	return &PodHandler{
		podService: &mockPodService{getPodFn: func(ctx context.Context, key string) (*agentpod.Pod, error) {
			return pod, nil
		}},
		commandSender: &mockCommandSender{},
		relaySelector: &mockPreviewRelaySelector{info: &relaysvc.RelayInfo{URL: "wss://example.com/relay"}},
		relayTokens:   &mockPreviewTokens{},
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

func TestGetPodPreview_ReturnsTokenAndURL(t *testing.T) {
	pod := &agentpod.Pod{PodKey: "pod1", RunnerID: 7, PreviewPort: 3000, Status: agentpod.StatusRunning, OrganizationID: 1, CreatedByID: 10}
	w := performPreviewGET(newPreviewHandler(pod))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
	}
	var resp struct {
		PreviewBaseURL string `json:"preview_base_url"`
		SessionURL     string `json:"session_url"`
		Token          string `json:"token"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.NotEmpty(t, resp.Token)
	assert.Contains(t, resp.PreviewBaseURL, "/preview/pod1/")
	assert.Contains(t, resp.SessionURL, "__session")
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
