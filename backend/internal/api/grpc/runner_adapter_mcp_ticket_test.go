package grpc

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/anthropics/agentsmesh/backend/internal/service/agentpod"
)

func TestMapOrchestratorErrorToMCP(t *testing.T) {
	wrappedRepositoryError := fmt.Errorf("repository 17 org 9 permission denied: %w", agentpod.ErrCreateResourceUnavailable)
	tests := []struct {
		name           string
		err            error
		wantCode       int32
		wantMessage    string
		wantNotContain []string
	}{
		{
			name:        "ErrMissingRunnerID -> 400",
			err:         agentpod.ErrMissingRunnerID,
			wantCode:    400,
			wantMessage: "runner_id is required",
		},
		{
			name:        "ErrMissingAgentSlug -> 400",
			err:         agentpod.ErrMissingAgentSlug,
			wantCode:    400,
			wantMessage: "agent_slug is required",
		},
		{
			name:        "ErrSourcePodNotTerminated -> 400",
			err:         agentpod.ErrSourcePodNotTerminated,
			wantCode:    400,
			wantMessage: "source pod is not terminated",
		},
		{
			name:        "ErrResumeRunnerMismatch -> 400",
			err:         agentpod.ErrResumeRunnerMismatch,
			wantCode:    400,
			wantMessage: "resume requires same runner",
		},
		{
			name:        "ErrCreateResourceUnavailable -> 400",
			err:         agentpod.ErrCreateResourceUnavailable,
			wantCode:    400,
			wantMessage: "Selected repository is unavailable",
		},
		{
			name:           "wrapped ErrCreateResourceUnavailable -> 400",
			err:            wrappedRepositoryError,
			wantCode:       400,
			wantMessage:    "Selected repository is unavailable",
			wantNotContain: []string{"repository 17", "org 9", "permission denied"},
		},
		{
			name:           "joined repository and agentfile validation errors -> 400",
			err:            errors.Join(agentpod.ErrInvalidAgentfileLayer, wrappedRepositoryError),
			wantCode:       400,
			wantMessage:    "Selected repository is unavailable",
			wantNotContain: []string{"repository 17", "org 9", "permission denied"},
		},
		{
			name:        "ErrSourcePodAccessDenied -> 403",
			err:         agentpod.ErrSourcePodAccessDenied,
			wantCode:    403,
			wantMessage: "source pod access denied",
		},
		{
			name:        "ErrSourcePodNotFound -> 404",
			err:         agentpod.ErrSourcePodNotFound,
			wantCode:    404,
			wantMessage: "source pod not found",
		},
		{
			name:        "ErrSourcePodAlreadyResumed -> 409",
			err:         agentpod.ErrSourcePodAlreadyResumed,
			wantCode:    409,
			wantMessage: "source pod already resumed",
		},
		{
			name:        "ErrSandboxAlreadyResumed -> 409",
			err:         agentpod.ErrSandboxAlreadyResumed,
			wantCode:    409,
			wantMessage: "sandbox already resumed",
		},
		{
			name:        "ErrConfigBuildFailed -> 500",
			err:         agentpod.ErrConfigBuildFailed,
			wantCode:    500,
			wantMessage: "failed to build pod configuration",
		},
		{
			name:        "unknown error -> 500 with details",
			err:         assert.AnError,
			wantCode:    500,
			wantMessage: "failed to create pod: assert.AnError general error for testing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapOrchestratorErrorToMCP(tt.err)
			require.NotNil(t, result)
			assert.Equal(t, tt.wantCode, result.code)
			assert.Equal(t, tt.wantMessage, result.message)
			for _, sensitive := range tt.wantNotContain {
				assert.NotContains(t, result.message, sensitive)
			}
		})
	}
}
