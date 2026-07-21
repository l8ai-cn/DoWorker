package podconnect

import (
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/l8ai-cn/agentcloud/backend/internal/middleware"
	podv1 "github.com/l8ai-cn/agentcloud/proto/gen/go/pod/v1"
)

func TestBuildCreatePodRequestRequiresWorkerSpecForFreshCreate(t *testing.T) {
	_, err := buildCreatePodRequest(
		&podv1.CreatePodRequest{AgentSlug: "codex-cli"},
		&middleware.TenantContext{
			OrganizationID: 7, OrganizationSlug: "acme", UserID: 42,
		},
	)

	require.Error(t, err)
	assert.Equal(t, connect.CodeFailedPrecondition, connectCodeOf(t, err))
}

func TestBuildCreatePodRequestUsesStructuredWorkerDraft(t *testing.T) {
	request, err := buildCreatePodRequest(
		&podv1.CreatePodRequest{
			WorkerSpec: completeWorkerDraftProto(),
			Cols:       120,
			Rows:       40,
		},
		&middleware.TenantContext{
			OrganizationID: 7, OrganizationSlug: "acme", UserID: 42,
		},
	)

	require.NoError(t, err)
	require.NotNil(t, request.WorkerSpecDraft)
	assert.Equal(t, int64(7), request.OrganizationID)
	assert.Equal(t, int64(42), request.UserID)
	assert.Empty(t, request.AgentSlug)
	assert.Zero(t, request.RunnerID)
	assert.Nil(t, request.AgentfileLayer)
	assert.Nil(t, request.ModelResourceID)
	assert.Equal(t, "acme", request.WorkerSpecDraft.OrganizationSlug.String())
}

func TestBuildCreatePodRequestUsesSourceLineage(t *testing.T) {
	resume := true
	request, err := buildCreatePodRequest(
		&podv1.CreatePodRequest{
			SourcePodKey:       stringPointer("source-pod"),
			ResumeAgentSession: &resume,
			Cols:               120,
			Rows:               40,
		},
		&middleware.TenantContext{OrganizationID: 7, UserID: 42},
	)

	require.NoError(t, err)
	assert.Equal(t, "source-pod", request.SourcePodKey)
	assert.Equal(t, &resume, request.ResumeAgentSession)
	assert.Nil(t, request.WorkerSpecDraft)
	assert.Empty(t, request.AgentSlug)
	assert.Zero(t, request.RunnerID)
}

func TestBuildCreatePodRequestRejectsResumeRuntimeOverrides(t *testing.T) {
	_, err := buildCreatePodRequest(
		&podv1.CreatePodRequest{
			SourcePodKey: stringPointer("source-pod"),
			AgentSlug:    "codex-cli",
		},
		&middleware.TenantContext{OrganizationID: 7, UserID: 42},
	)

	require.Error(t, err)
	assert.Equal(t, connect.CodeFailedPrecondition, connectCodeOf(t, err))
}

func stringPointer(value string) *string {
	return &value
}
