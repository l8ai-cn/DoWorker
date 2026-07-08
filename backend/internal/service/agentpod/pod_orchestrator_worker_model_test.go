package agentpod

import (
	"context"
	"strings"
	"testing"

	agentDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agent"
	"github.com/anthropics/agentsmesh/backend/internal/domain/aimodel"
	aimodelsvc "github.com/anthropics/agentsmesh/backend/internal/service/aimodel"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeAIPool struct{ rm *aimodelsvc.ResolvedModel }

func (f fakeAIPool) Resolve(_ context.Context, _ int64) (*aimodelsvc.ResolvedModel, error) {
	return f.rm, nil
}

func resolvedFixture() *aimodelsvc.ResolvedModel {
	return &aimodelsvc.ResolvedModel{
		Model:       &aimodel.AIModel{ProviderType: "anthropic", Model: "claude-sonnet"},
		Credentials: map[string]string{"api_key": "sk-test"},
	}
}

func TestApplyWorkerModel_ConfigBundleForDoAgent(t *testing.T) {
	o := NewPodOrchestrator(&PodOrchestratorDeps{AIModelPool: fakeAIPool{rm: resolvedFixture()}})
	id := int64(5)
	budget := int64(1000)
	req := &OrchestrateCreatePodRequest{AgentSlug: "do-agent", ModelConfigID: &id, TokenBudget: &budget}

	require.NoError(t, o.applyWorkerModel(context.Background(), req, &agentDomain.Agent{Executable: "do-agent"}))

	require.NotNil(t, req.AgentfileLayer)
	layer := *req.AgentfileLayer
	assert.Contains(t, layer, `USE_CONFIG_BUNDLE "worker-model"`)
	assert.Contains(t, layer, `token_budget = "1000"`)
	assert.NotNil(t, req.SessionConfigBundles["worker-model"])
}

func TestApplyWorkerModel_NoBindingIsNoop(t *testing.T) {
	o := NewPodOrchestrator(&PodOrchestratorDeps{})
	req := &OrchestrateCreatePodRequest{AgentSlug: "do-agent"}

	require.NoError(t, o.applyWorkerModel(context.Background(), req, nil))
	assert.Nil(t, req.AgentfileLayer)
	assert.Empty(t, req.SessionConfigBundles)
}

func TestAppendAgentfileLayerJoinsLines(t *testing.T) {
	var layer *string
	appendAgentfileLayer(&layer, `USE_ENV_BUNDLE "a"`)
	appendAgentfileLayer(&layer, `CONFIG token_budget = "5"`)
	require.NotNil(t, layer)
	lines := strings.Split(*layer, "\n")
	assert.Equal(t, 2, len(lines))
}
