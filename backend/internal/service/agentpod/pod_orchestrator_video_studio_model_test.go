package agentpod

import (
	"context"
	"testing"

	agentDomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/agent"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplyWorkerModelUsesCodexHarnessForVideoStudio(t *testing.T) {
	resolver := &recordingModelResourceResolver{resource: resolvedOpenAIResource()}
	orchestrator := NewPodOrchestrator(&PodOrchestratorDeps{
		ModelResources: resolver,
	})
	resourceID := int64(9)
	req := &OrchestrateCreatePodRequest{
		AgentSlug:       "video-studio",
		UserID:          7,
		OrganizationID:  11,
		ModelResourceID: &resourceID,
	}

	require.NoError(t, orchestrator.applyWorkerModel(
		context.Background(),
		req,
		&agentDomain.Agent{Executable: "video-studio-codex"},
	))

	assert.Equal(t, []string{"openai-compatible"}, resolver.requirements.AllowedProtocolAdapters)
	assert.Equal(t, "sk-test", req.ModelResourceEnv["OPENAI_API_KEY"])
	assert.Equal(t, "gpt-5.1", req.ModelResourceEnv["OPENAI_MODEL"])
}
