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

type fakeAIPool struct {
	rm        *aimodelsvc.ResolvedModel
	defaultRM *aimodelsvc.ResolvedModel
}

func (f fakeAIPool) Resolve(_ context.Context, _ int64) (*aimodelsvc.ResolvedModel, error) {
	return f.rm, nil
}

func (f fakeAIPool) ResolveDefaultForAgent(_ context.Context, _, _ int64, _ string) (*aimodelsvc.ResolvedModel, error) {
	return f.defaultRM, nil
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

func TestApplyWorkerModel_OrgDefaultForDoAgent(t *testing.T) {
	rm := &aimodelsvc.ResolvedModel{
		Model:       &aimodel.AIModel{ProviderType: "minimax", Model: "MiniMax-M3", BaseURL: "https://api.minimax.chat/anthropic"},
		Credentials: map[string]string{"api_key": "sk-test"},
	}
	o := NewPodOrchestrator(&PodOrchestratorDeps{AIModelPool: fakeAIPool{defaultRM: rm}})
	req := &OrchestrateCreatePodRequest{AgentSlug: "do-agent", OrganizationID: 1, UserID: 1}

	require.NoError(t, o.applyWorkerModel(context.Background(), req, &agentDomain.Agent{Executable: "do-agent"}))
	require.NotNil(t, req.AgentfileLayer)
	assert.Contains(t, *req.AgentfileLayer, `USE_CONFIG_BUNDLE "worker-model"`)
}

func TestAppendAgentfileLayerJoinsLines(t *testing.T) {
	var layer *string
	appendAgentfileLayer(&layer, `USE_ENV_BUNDLE "a"`)
	appendAgentfileLayer(&layer, `CONFIG token_budget = "5"`)
	require.NotNil(t, layer)
	lines := strings.Split(*layer, "\n")
	assert.Equal(t, 2, len(lines))
}
