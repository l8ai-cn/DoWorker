package v1

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/anthropics/agentsmesh/backend/internal/service/agentpod"
	"github.com/anthropics/agentsmesh/backend/internal/service/billing"
	"github.com/anthropics/agentsmesh/backend/pkg/apierr"
)

func TestMapOrchestratorErrorToHTTP(t *testing.T) {
	gin.SetMode(gin.TestMode)
	wrappedRepositoryError := fmt.Errorf("repository 17 org 9 permission denied: %w", agentpod.ErrCreateResourceUnavailable)

	tests := []struct {
		name           string
		err            error
		wantCode       int
		wantJSON       map[string]string // "code" field expected in response
		wantNotContain []string
	}{
		{
			name:     "ErrMissingRunnerID -> 400",
			err:      agentpod.ErrMissingRunnerID,
			wantCode: http.StatusBadRequest,
			wantJSON: map[string]string{"code": "MISSING_RUNNER_ID"},
		},
		{
			name:     "ErrMissingAgentSlug -> 400",
			err:      agentpod.ErrMissingAgentSlug,
			wantCode: http.StatusBadRequest,
			wantJSON: map[string]string{"code": "MISSING_AGENT_SLUG"},
		},
		{
			name:     "ErrSourcePodNotTerminated -> 400",
			err:      agentpod.ErrSourcePodNotTerminated,
			wantCode: http.StatusBadRequest,
			wantJSON: map[string]string{"code": "SOURCE_POD_NOT_TERMINATED"},
		},
		{
			name:     "ErrResumeRunnerMismatch -> 400",
			err:      agentpod.ErrResumeRunnerMismatch,
			wantCode: http.StatusBadRequest,
			wantJSON: map[string]string{"code": "RESUME_RUNNER_MISMATCH"},
		},
		{
			name:     "ErrCreateResourceUnavailable -> 400",
			err:      agentpod.ErrCreateResourceUnavailable,
			wantCode: http.StatusBadRequest,
			wantJSON: map[string]string{
				"code":  apierr.VALIDATION_FAILED,
				"error": "Selected repository is unavailable",
			},
		},
		{
			name:     "wrapped ErrCreateResourceUnavailable -> 400",
			err:      wrappedRepositoryError,
			wantCode: http.StatusBadRequest,
			wantJSON: map[string]string{
				"code":  apierr.VALIDATION_FAILED,
				"error": "Selected repository is unavailable",
			},
			wantNotContain: []string{"repository 17", "org 9", "permission denied"},
		},
		{
			name:     "joined repository and agentfile validation errors -> 400",
			err:      errors.Join(agentpod.ErrInvalidAgentfileLayer, wrappedRepositoryError),
			wantCode: http.StatusBadRequest,
			wantJSON: map[string]string{
				"code":  apierr.VALIDATION_FAILED,
				"error": "Selected repository is unavailable",
			},
			wantNotContain: []string{"repository 17", "org 9", "permission denied"},
		},
		{
			name:     "ErrQuotaExceeded -> 402",
			err:      billing.ErrQuotaExceeded,
			wantCode: http.StatusPaymentRequired,
			wantJSON: map[string]string{"code": "CONCURRENT_POD_QUOTA_EXCEEDED"},
		},
		{
			name:     "ErrSubscriptionFrozen -> 402",
			err:      billing.ErrSubscriptionFrozen,
			wantCode: http.StatusPaymentRequired,
			wantJSON: map[string]string{"code": "SUBSCRIPTION_FROZEN"},
		},
		{
			name:     "ErrSourcePodAccessDenied -> 403",
			err:      agentpod.ErrSourcePodAccessDenied,
			wantCode: http.StatusForbidden,
			wantJSON: map[string]string{"code": "SOURCE_POD_ACCESS_DENIED"},
		},
		{
			name:     "ErrSourcePodNotFound -> 404",
			err:      agentpod.ErrSourcePodNotFound,
			wantCode: http.StatusNotFound,
			wantJSON: map[string]string{"code": "SOURCE_POD_NOT_FOUND"},
		},
		{
			name:     "ErrSourcePodAlreadyResumed -> 409",
			err:      agentpod.ErrSourcePodAlreadyResumed,
			wantCode: http.StatusConflict,
			wantJSON: map[string]string{"code": "SOURCE_POD_ALREADY_RESUMED"},
		},
		{
			name:     "ErrSandboxAlreadyResumed -> 409",
			err:      agentpod.ErrSandboxAlreadyResumed,
			wantCode: http.StatusConflict,
			wantJSON: map[string]string{"code": "SANDBOX_ALREADY_RESUMED"},
		},
		{
			name:     "ErrConfigBuildFailed -> 500",
			err:      agentpod.ErrConfigBuildFailed,
			wantCode: http.StatusInternalServerError,
			wantJSON: map[string]string{"code": "POD_CONFIG_BUILD_FAILED"},
		},
		{
			name:     "ErrModelResourceResolverUnavailable -> 500",
			err:      agentpod.ErrModelResourceResolverUnavailable,
			wantCode: http.StatusInternalServerError,
			wantJSON: map[string]string{"code": "INTERNAL_ERROR"},
		},
		{
			name:     "ErrModelResourceCommandConflict -> 400",
			err:      agentpod.ErrModelResourceCommandConflict,
			wantCode: http.StatusBadRequest,
			wantJSON: map[string]string{"code": "VALIDATION_FAILED"},
		},
		{
			name:     "unknown error -> 500 fallback",
			err:      assert.AnError,
			wantCode: http.StatusInternalServerError,
			wantJSON: map[string]string{"error": "Failed to create pod", "code": "INTERNAL_ERROR"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			mapOrchestratorErrorToHTTP(c, tt.err)

			assert.Equal(t, tt.wantCode, w.Code)

			var resp map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &resp)
			require.NoError(t, err)

			for key, expectedVal := range tt.wantJSON {
				actual, ok := resp[key]
				require.True(t, ok, "expected key %q in response", key)
				assert.Equal(t, expectedVal, actual)
			}
			for _, sensitive := range tt.wantNotContain {
				assert.NotContains(t, w.Body.String(), sensitive)
			}
		})
	}
}

func TestLegacyPodCreateModelFieldsAreRejected(t *testing.T) {
	for _, field := range []string{
		"credential" + "_profile_id",
		"model" + "_config_id",
		"virtual_api" + "_key_id",
	} {
		t.Run(field, func(t *testing.T) {
			got, ok := legacyPodCreateModelField([]byte(`{"agent_slug":"codex-cli","` + field + `":99}`))

			require.True(t, ok)
			assert.Equal(t, field, got)
		})
	}
}
