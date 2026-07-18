package workercreation

import (
	"context"
	"errors"
	"testing"

	agentdomain "github.com/anthropics/agentsmesh/backend/internal/domain/agent"
	runtimedomain "github.com/anthropics/agentsmesh/backend/internal/domain/workerruntime"
	specdomain "github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	"github.com/anthropics/agentsmesh/backend/internal/service/workerdefinition"
	specservice "github.com/anthropics/agentsmesh/backend/internal/service/workerspec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServiceListOptionsReturnsSelectableRuntimeAndBlockingReasons(t *testing.T) {
	codexSource := "AGENT codex\nEXECUTABLE codex\nMODE acp\n"
	unsupportedSource := "AGENT aider\nEXECUTABLE aider\nMODE pty\n"
	aider := activeWorkerTypeAgentFor("aider", "aider", unsupportedSource)
	aider.SupportedModes = "pty"
	agents := &workerOptionsAgentProvider{agents: []*agentdomain.Agent{
		activeWorkerTypeAgentFor("codex-cli", "codex", codexSource),
		aider,
	}}
	service := NewService(Deps{
		Catalog: enabledCodexRuntimeCatalog(),
		Definitions: staticWorkerDefinitions{
			"codex-cli": workerDefinition("codex-cli", "codex", codexSource, "pty", "acp"),
			"aider":     workerDefinition("aider", "aider", unsupportedSource, "pty"),
		},
		Agents: agents,
		Runners: workerOptionsRunnerAvailability{
			available: true,
		},
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
	assert.Equal(t, runtimedomain.DefaultCatalogRevision(), options.Revision)
	require.Len(t, options.WorkerTypes, 2)
	assert.True(t, options.WorkerTypes[0].Selectable)
	assert.Equal(
		t,
		[]string{"openai-compatible"},
		options.WorkerTypes[0].ModelProtocolAdapters,
	)
	assert.False(t, options.WorkerTypes[1].Selectable)
	assert.Contains(t, options.WorkerTypes[1].BlockingReason, "runtime image")
	require.Len(t, options.RuntimeImages, 1)
	assert.True(t, options.RuntimeImages[0].Selectable)
	assert.Contains(t, options.RuntimeImages, RuntimeImageOption{
		Image:      enabledCodexRuntimeCatalog().ImagesFor("codex-cli")[0],
		Selectable: true,
	})
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

func TestServiceListOptionsExposesModelRequirement(t *testing.T) {
	source := "AGENT cursor\nEXECUTABLE agent\nMODE acp\n"
	definition := workerDefinition("cursor-cli", "agent", source, "acp")
	definition.ModelRequirement = workerdefinition.ModelRequirement{}
	agents := &workerOptionsAgentProvider{agents: []*agentdomain.Agent{
		activeWorkerTypeAgentFor("cursor-cli", "agent", source),
	}}
	service := NewService(Deps{
		Catalog: runtimedomain.DefaultCatalog(),
		Definitions: staticWorkerDefinitions{
			"cursor-cli": definition,
		},
		Agents: agents,
		Runners: workerOptionsRunnerAvailability{
			available: true,
		},
	})

	options, err := service.ListOptions(
		context.Background(),
		specservice.Scope{OrgID: 77, UserID: 7},
		OptionsFilter{},
	)

	require.NoError(t, err)
	require.Len(t, options.WorkerTypes, 1)
	assert.False(t, options.WorkerTypes[0].RequiresModelResource)
}

func TestServiceListOptionsRequiresRunnerAvailabilityResolver(t *testing.T) {
	source := "AGENT codex\nEXECUTABLE codex\nMODE acp\n"
	service := NewService(Deps{
		Catalog: enabledCodexRuntimeCatalog(),
		Definitions: staticWorkerDefinitions{
			"codex-cli": workerDefinition("codex-cli", "codex", source, "pty", "acp"),
		},
		Agents: &workerOptionsAgentProvider{agents: []*agentdomain.Agent{
			activeWorkerTypeAgentFor("codex-cli", "codex", source),
		}},
	})

	_, err := service.ListOptions(
		context.Background(),
		specservice.Scope{OrgID: 77, UserID: 7},
		OptionsFilter{},
	)

	assert.ErrorIs(t, err, specservice.ErrResolverUnavailable)
}

func TestServiceListOptionsBlocksEnabledImageWithoutOnlineRunner(t *testing.T) {
	source := "AGENT codex\nEXECUTABLE codex\nMODE acp\n"
	service := NewService(Deps{
		Catalog: enabledCodexRuntimeCatalog(),
		Definitions: staticWorkerDefinitions{
			"codex-cli": workerDefinition("codex-cli", "codex", source, "pty", "acp"),
		},
		Agents: &workerOptionsAgentProvider{agents: []*agentdomain.Agent{
			activeWorkerTypeAgentFor("codex-cli", "codex", source),
		}},
		Runners: workerOptionsRunnerAvailability{},
	})

	options, err := service.ListOptions(
		context.Background(),
		specservice.Scope{OrgID: 77, UserID: 7},
		OptionsFilter{},
	)

	require.NoError(t, err)
	require.Len(t, options.WorkerTypes, 1)
	assert.False(t, options.WorkerTypes[0].Selectable)
	assert.Equal(
		t,
		"No online Runner currently supports this worker type",
		options.WorkerTypes[0].BlockingReason,
	)
}

func TestServiceListOptionsReturnsRunnerAvailabilityErrors(t *testing.T) {
	source := "AGENT codex\nEXECUTABLE codex\nMODE acp\n"
	runnerFailure := errors.New("runner query failed")
	service := NewService(Deps{
		Catalog: enabledCodexRuntimeCatalog(),
		Definitions: staticWorkerDefinitions{
			"codex-cli": workerDefinition("codex-cli", "codex", source, "pty", "acp"),
		},
		Agents: &workerOptionsAgentProvider{agents: []*agentdomain.Agent{
			activeWorkerTypeAgentFor("codex-cli", "codex", source),
		}},
		Runners: workerOptionsRunnerAvailability{err: runnerFailure},
	})

	_, err := service.ListOptions(
		context.Background(),
		specservice.Scope{OrgID: 77, UserID: 7},
		OptionsFilter{},
	)

	assert.ErrorIs(t, err, runnerFailure)
}

type workerOptionsAgentProvider struct {
	agents []*agentdomain.Agent
}

type workerOptionsRunnerAvailability struct {
	available bool
	err       error
}

func (resolver workerOptionsRunnerAvailability) HasAvailableRunnerForAgent(
	context.Context,
	int64,
	int64,
	string,
) (bool, error) {
	return resolver.available, resolver.err
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
