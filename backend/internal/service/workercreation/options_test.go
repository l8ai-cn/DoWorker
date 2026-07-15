package workercreation

import (
	"context"
	"testing"

	agentdomain "github.com/anthropics/agentsmesh/backend/internal/domain/agent"
	runtimedomain "github.com/anthropics/agentsmesh/backend/internal/domain/workerruntime"
	specdomain "github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	specservice "github.com/anthropics/agentsmesh/backend/internal/service/workerspec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServiceListOptionsReturnsSelectableRuntimeAndBlockingReasons(t *testing.T) {
	codexSource := "AGENT codex\nEXECUTABLE codex\nMODE acp\n"
	unsupportedSource := "AGENT aider\nEXECUTABLE aider\nMODE pty\n"
	agents := &workerOptionsAgentProvider{agents: []*agentdomain.Agent{
		activeWorkerTypeAgentFor("codex-cli", "codex", codexSource),
		activeWorkerTypeAgentFor("aider", "aider", unsupportedSource),
	}}
	service := NewService(Deps{
		Catalog: runtimedomain.DefaultCatalog(),
		Definitions: staticWorkerDefinitions{
			"codex-cli": workerDefinition(
				"codex-cli", "codex", codexSource, "pty", "acp",
			),
			"aider": workerDefinition(
				"aider", "aider", unsupportedSource, "pty", "acp",
			),
		},
		Agents: agents,
	})
	targetID := int64(1)

	options, err := service.ListOptions(
		context.Background(),
		specservice.Scope{OrgID: 77, UserID: 7},
		OptionsFilter{
			ComputeTargetID: &targetID,
			DeploymentMode:  specdomain.DeploymentModeDedicated,
		},
	)

	require.NoError(t, err)
	assert.Equal(t, runtimedomain.DefaultCatalogRevision, options.Revision)
	require.Len(t, options.WorkerTypes, 2)
	assert.True(t, options.WorkerTypes[0].Selectable)
	assert.False(t, options.WorkerTypes[1].Selectable)
	assert.Contains(t, options.WorkerTypes[1].BlockingReason, "runtime image")
	require.Len(t, options.RuntimeImages, 3)
	assert.True(t, options.RuntimeImages[0].Selectable)
	require.Len(t, options.ComputeTargets, 2)
	assert.True(t, options.ComputeTargets[0].Selectable)
	assert.False(t, options.ComputeTargets[1].Selectable)
	assert.NotEmpty(t, options.ComputeTargets[1].BlockingReason)
	require.Len(t, options.DeploymentModes, 2)
	assert.False(t, options.DeploymentModes[1].Selectable)
	assert.Contains(t, options.DeploymentModes[1].BlockingReason, "compute target")
	require.Len(t, options.ResourceProfiles, 2)
	assert.True(t, options.ResourceProfiles[0].Selectable)
}

type workerOptionsAgentProvider struct {
	agents []*agentdomain.Agent
}

func (provider *workerOptionsAgentProvider) GetAgent(
	_ context.Context,
	slug string,
) (*agentdomain.Agent, error) {
	for _, agent := range provider.agents {
		if agent.Slug == slug {
			return agent, nil
		}
	}
	return nil, nil
}

func (provider *workerOptionsAgentProvider) ListBuiltinAgents(
	context.Context,
) ([]*agentdomain.Agent, error) {
	return provider.agents, nil
}
