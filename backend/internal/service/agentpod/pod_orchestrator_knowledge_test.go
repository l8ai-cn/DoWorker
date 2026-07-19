package agentpod

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	kbDomain "github.com/anthropics/agentsmesh/backend/internal/domain/knowledgebase"
	kbservice "github.com/anthropics/agentsmesh/backend/internal/service/knowledgebase"
	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
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
			{
				KB:            kbFixture("team-docs"),
				Mode:          "rw",
				SSHCloneURL:   "ssh://git@gitea.local/am-kb/team-docs.git",
				GitKnownHosts: "gitea.local ssh-ed25519 host-key",
				GitPrivateKey: "rw-private-key",
			},
			{
				KB:            kbFixture("product-wiki"),
				Mode:          "ro",
				SSHCloneURL:   "ssh://git@gitea.local/am-kb/product-wiki.git",
				GitKnownHosts: "gitea.local ssh-ed25519 host-key",
				GitPrivateKey: "ro-private-key",
			},
		},
	}
	coord := &mockPodCoordinator{}
	orch, _, _ := setupOrchestrator(t, withCoordinator(coord), withKnowledgeResolver(resolver))

	_, err := createPodWithPlanSourceForTest(t, orch, context.Background(), &OrchestrateCreatePodRequest{
		OrganizationID:  1,
		UserID:          1,
		RunnerID:        1,
		AgentSlug:       "claude-code",
		ModelResourceID: testModelResourceID(),
		AgentfileLayer:  ptrStr("KNOWLEDGE team-docs [rw]"),
		KnowledgeMounts: []KnowledgeMountRequest{
			{Slug: "product-wiki"},
		},
	})
	require.NoError(t, err)

	require.Nil(t, resolver.lastRequested)

	require.NotNil(t, coord.lastCmd)
	require.NotNil(t, coord.lastCmd.SandboxConfig)
	mounts := coord.lastCmd.SandboxConfig.KnowledgeMounts
	require.Len(t, mounts, 2)
	teamDocs := knowledgeMountBySlugForTest(t, mounts, "team-docs")
	assert.Equal(t, "rw", teamDocs.Mode)
	assert.Equal(t, "kb/team-docs", teamDocs.MountPath)
	assert.Equal(t, "https://git.example.com/kb/team-docs.git", teamDocs.HttpCloneUrl)
	assert.Empty(t, teamDocs.SshCloneUrl)
	assert.Empty(t, teamDocs.GitKnownHosts)
	assert.Equal(t, "main", teamDocs.Branch)
	assert.Equal(t, strings.Repeat("e", 40), teamDocs.CommitSha)
	assert.Empty(t, teamDocs.GitPrivateKey)

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

func knowledgeMountBySlugForTest(
	t *testing.T,
	mounts []*runnerv1.KnowledgeMount,
	slug string,
) *runnerv1.KnowledgeMount {
	t.Helper()
	for _, mount := range mounts {
		if mount.Slug == slug {
			return mount
		}
	}
	require.Failf(t, "missing knowledge mount", slug)
	return nil
}

func TestCreatePod_KnowledgeDeclaredUsesArtifactPins(t *testing.T) {
	coord := &mockPodCoordinator{}
	orch, _, _ := setupOrchestrator(t, withCoordinator(coord))

	_, err := createPodWithPlanSourceForTest(t, orch, context.Background(), &OrchestrateCreatePodRequest{
		OrganizationID:  1,
		UserID:          1,
		RunnerID:        1,
		AgentSlug:       "claude-code",
		ModelResourceID: testModelResourceID(),
		AgentfileLayer:  ptrStr("KNOWLEDGE team-docs"),
	})
	require.NoError(t, err)
	require.NotNil(t, coord.lastCmd.SandboxConfig)
	assert.Len(t, coord.lastCmd.SandboxConfig.KnowledgeMounts, 1)
}

func TestCreatePod_NoKnowledgeMounts(t *testing.T) {
	resolver := &mockKnowledgeResolver{}
	coord := &mockPodCoordinator{}
	orch, _, _ := setupOrchestrator(t, withCoordinator(coord), withKnowledgeResolver(resolver))

	_, err := createPodWithPlanSourceForTest(t, orch, context.Background(), &OrchestrateCreatePodRequest{
		OrganizationID:  1,
		UserID:          1,
		RunnerID:        1,
		AgentSlug:       "claude-code",
		ModelResourceID: testModelResourceID(),
		AgentfileLayer:  ptrStr("CONFIG mcp_enabled = true"),
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
