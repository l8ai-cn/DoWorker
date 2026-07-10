package agentpod

import (
	"context"
	"errors"
	"strings"
	"testing"

	agentDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agent"
	"github.com/anthropics/agentsmesh/backend/internal/domain/aimodel"
	aimodelsvc "github.com/anthropics/agentsmesh/backend/internal/service/aimodel"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeAIPool struct {
	rm             *aimodelsvc.ResolvedModel
	defaultRM      *aimodelsvc.ResolvedModel
	resolveErr     error
	visibleCalls   int
	defaultCalls   int
	modelID        int64
	userID         int64
	organizationID int64
}

func (f *fakeAIPool) ResolveVisible(_ context.Context, id, userID, organizationID int64) (*aimodelsvc.ResolvedModel, error) {
	f.visibleCalls++
	f.modelID = id
	f.userID = userID
	f.organizationID = organizationID
	return f.rm, f.resolveErr
}

func (f *fakeAIPool) ResolveDefaultForAgent(_ context.Context, _, _ int64, _ string) (*aimodelsvc.ResolvedModel, error) {
	f.defaultCalls++
	return f.defaultRM, nil
}

func resolvedFixture() *aimodelsvc.ResolvedModel {
	return &aimodelsvc.ResolvedModel{
		Model:       &aimodel.AIModel{ProviderType: "anthropic", Model: "claude-sonnet"},
		Credentials: map[string]string{"api_key": "sk-test"},
	}
}

func TestApplyWorkerModel_ConfigBundleForDoAgent(t *testing.T) {
	o := NewPodOrchestrator(&PodOrchestratorDeps{AIModelPool: &fakeAIPool{rm: resolvedFixture()}})
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
	o := NewPodOrchestrator(&PodOrchestratorDeps{AIModelPool: &fakeAIPool{defaultRM: rm}})
	req := &OrchestrateCreatePodRequest{AgentSlug: "do-agent", OrganizationID: 1, UserID: 1}

	require.NoError(t, o.applyWorkerModel(context.Background(), req, &agentDomain.Agent{Executable: "do-agent"}))
	require.NotNil(t, req.AgentfileLayer)
	assert.Contains(t, *req.AgentfileLayer, `USE_CONFIG_BUNDLE "worker-model"`)
}

func TestResolvePoolModel_ExplicitModelUsesCallerScope(t *testing.T) {
	pool := &fakeAIPool{rm: resolvedFixture()}
	o := NewPodOrchestrator(&PodOrchestratorDeps{AIModelPool: pool})
	modelID := int64(5)
	req := &OrchestrateCreatePodRequest{
		ModelConfigID:  modelIDPointer(modelID),
		UserID:         11,
		OrganizationID: 21,
	}

	resolved, budget, err := o.resolvePoolModel(context.Background(), req, nil)

	require.NoError(t, err)
	assert.Same(t, pool.rm, resolved)
	assert.Nil(t, budget)
	assert.Equal(t, 1, pool.visibleCalls)
	assert.Equal(t, modelID, pool.modelID)
	assert.Equal(t, int64(11), pool.userID)
	assert.Equal(t, int64(21), pool.organizationID)
	assert.Zero(t, pool.defaultCalls)
}

func TestResolvePoolModel_ExplicitModelErrorDoesNotUseDefault(t *testing.T) {
	resolveErr := errors.New("resolve visible model")
	pool := &fakeAIPool{resolveErr: resolveErr, defaultRM: resolvedFixture()}
	o := NewPodOrchestrator(&PodOrchestratorDeps{AIModelPool: pool})
	req := &OrchestrateCreatePodRequest{
		ModelConfigID:  modelIDPointer(5),
		UserID:         11,
		OrganizationID: 21,
	}

	resolved, budget, err := o.resolvePoolModel(context.Background(), req, nil)

	assert.ErrorIs(t, err, resolveErr)
	assert.Nil(t, resolved)
	assert.Nil(t, budget)
	assert.Zero(t, pool.defaultCalls)
}

func TestAppendAgentfileLayerJoinsLines(t *testing.T) {
	var layer *string
	appendAgentfileLayer(&layer, `USE_ENV_BUNDLE "a"`)
	appendAgentfileLayer(&layer, `CONFIG token_budget = "5"`)
	require.NotNil(t, layer)
	lines := strings.Split(*layer, "\n")
	assert.Equal(t, 2, len(lines))
}

func modelIDPointer(value int64) *int64 { return &value }
