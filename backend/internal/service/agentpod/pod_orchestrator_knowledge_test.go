package agentpod

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	kbDomain "github.com/anthropics/agentsmesh/backend/internal/domain/knowledgebase"
	kbservice "github.com/anthropics/agentsmesh/backend/internal/service/knowledgebase"
)

type mockKnowledgeResolver struct {
	lastRequested []kbservice.MountRequest
	mounts        []*kbservice.ResolvedMount
	err           error
}

func (m *mockKnowledgeResolver) ResolveMountsForPod(_ context.Context, _ int64, _ string, requested []kbservice.MountRequest) ([]*kbservice.ResolvedMount, error) {
	m.lastRequested = requested
	return m.mounts, m.err
}

func (m *mockKnowledgeResolver) CloneToken() string { return "kb-token" }

func withKnowledgeResolver(r KnowledgeBaseResolverForOrchestrator) func(*PodOrchestratorDeps) {
	return func(d *PodOrchestratorDeps) { d.KnowledgeBases = r }
}

func kbFixture(slug string) *kbDomain.KnowledgeBase {
	return &kbDomain.KnowledgeBase{
		Slug:          slug,
		HTTPCloneURL:  "http://gitea.local/am-kb/" + slug + ".git",
		DefaultBranch: "main",
	}
}

func TestCreatePod_KnowledgeMountsFromLayerAndRequest(t *testing.T) {
	resolver := &mockKnowledgeResolver{
		mounts: []*kbservice.ResolvedMount{
			{KB: kbFixture("team-docs"), Mode: "rw"},
			{KB: kbFixture("product-wiki"), Mode: "ro"},
		},
	}
	coord := &mockPodCoordinator{}
	orch, _, _ := setupOrchestrator(t, withCoordinator(coord), withKnowledgeResolver(resolver))

	_, err := orch.CreatePod(context.Background(), &OrchestrateCreatePodRequest{
		OrganizationID: 1,
		UserID:         1,
		RunnerID:       1,
		AgentSlug:      "claude-code",
		AgentfileLayer: ptrStr("KNOWLEDGE team-docs [rw]"),
		KnowledgeMounts: []KnowledgeMountRequest{
			{Slug: "product-wiki"},
		},
	})
	require.NoError(t, err)

	require.Equal(t, []kbservice.MountRequest{
		{KBSlug: "team-docs", Mode: "rw"},
		{KBSlug: "product-wiki", Mode: ""},
	}, resolver.lastRequested, "Agentfile declarations first, request selections last")

	require.NotNil(t, coord.lastCmd)
	require.NotNil(t, coord.lastCmd.SandboxConfig)
	mounts := coord.lastCmd.SandboxConfig.KnowledgeMounts
	require.Len(t, mounts, 2)
	assert.Equal(t, "team-docs", mounts[0].Slug)
	assert.Equal(t, "rw", mounts[0].Mode)
	assert.Equal(t, "kb/team-docs", mounts[0].MountPath)
	assert.Equal(t, "http://gitea.local/am-kb/team-docs.git", mounts[0].HttpCloneUrl)
	assert.Equal(t, "main", mounts[0].Branch)
	assert.Equal(t, "kb-token", mounts[0].GitToken)

	var readme string
	for _, f := range coord.lastCmd.FilesToCreate {
		if strings.HasSuffix(f.Path, "kb/README.md") {
			readme = f.Content
		}
	}
	require.NotEmpty(t, readme, "kb/README.md context file must be injected")
	assert.Contains(t, readme, "team-docs")
	assert.Contains(t, readme, "read-write")
	assert.Contains(t, readme, "llms.txt")
}

func TestCreatePod_KnowledgeDeclaredButFeatureDisabled(t *testing.T) {
	coord := &mockPodCoordinator{}
	orch, _, _ := setupOrchestrator(t, withCoordinator(coord))

	_, err := orch.CreatePod(context.Background(), &OrchestrateCreatePodRequest{
		OrganizationID: 1,
		UserID:         1,
		RunnerID:       1,
		AgentSlug:      "claude-code",
		AgentfileLayer: ptrStr("KNOWLEDGE team-docs"),
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrConfigBuildFailed)
}

func TestCreatePod_NoKnowledgeMounts(t *testing.T) {
	resolver := &mockKnowledgeResolver{}
	coord := &mockPodCoordinator{}
	orch, _, _ := setupOrchestrator(t, withCoordinator(coord), withKnowledgeResolver(resolver))

	_, err := orch.CreatePod(context.Background(), &OrchestrateCreatePodRequest{
		OrganizationID: 1,
		UserID:         1,
		RunnerID:       1,
		AgentSlug:      "claude-code",
		AgentfileLayer: ptrStr("CONFIG mcp_enabled = true"),
	})
	require.NoError(t, err)
	require.NotNil(t, coord.lastCmd)
	if coord.lastCmd.SandboxConfig != nil {
		assert.Empty(t, coord.lastCmd.SandboxConfig.KnowledgeMounts)
	}
	for _, f := range coord.lastCmd.FilesToCreate {
		assert.False(t, strings.HasSuffix(f.Path, "kb/README.md"))
	}
}
