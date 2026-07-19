package workercreation

import (
	"context"
	"testing"

	runtimedomain "github.com/anthropics/agentsmesh/backend/internal/domain/workerruntime"
	specdomain "github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	specservice "github.com/anthropics/agentsmesh/backend/internal/service/workerspec"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServicePreflightResolvesAndCompilesCompleteWorkerSpec(t *testing.T) {
	fixture := newWorkerCreationServiceFixture()
	service := NewService(fixture.deps())

	result, err := service.Preflight(
		context.Background(),
		specservice.Scope{OrgID: 77, UserID: 7},
		validWorkerCreationDraft(),
	)

	require.NoError(t, err)
	require.Empty(t, result.BlockingErrors)
	require.Empty(t, result.Warnings)
	require.NotNil(t, result.Resolved)
	assert.Equal(t, runtimedomain.DefaultCatalogRevision(), result.OptionsRevision)
	assert.Equal(t, int64(101), result.Resolved.Spec.Runtime.ModelBinding.ResourceID)
	assert.Equal(t, int64(7), result.Resolved.Spec.Runtime.ModelBinding.ResourceRevision)
	assert.Equal(t, int64(201), result.Resolved.Spec.Runtime.ModelBinding.ConnectionID)
	assert.Equal(t, int64(9), result.Resolved.Spec.Runtime.ModelBinding.ConnectionRevision)
	assert.Equal(t, int64(1), result.Resolved.Spec.Runtime.Image.ID)
	assert.Equal(t, int64(1), result.Resolved.Spec.Placement.ComputeTarget.ID)
	assert.Equal(t, int64(1), result.Resolved.Spec.Placement.ResourceProfile.ID)
	assert.Equal(t, []int64{3}, result.Resolved.Spec.Workspace.SkillIDs)
	assert.Equal(t, []specdomain.SkillPackageBinding{{
		SkillID:     3,
		Slug:        "code-review",
		Version:     2,
		ContentSHA:  "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		StorageKey:  "skills/code-review.tar.gz",
		PackageSize: 128,
	}}, result.Resolved.Spec.Workspace.SkillPackages)
	assert.Equal(t, "worker", result.Resolved.Spec.Metadata.Alias)
	assert.Contains(t, result.Resolved.AgentfileLayer, `CONFIG approval_mode = "never"`)
	assert.Contains(t, result.Resolved.AgentfileLayer, `USE_ENV_BUNDLE "signing-secrets"`)
	assert.NotContains(t, result.Resolved.AgentfileLayer, "must-not-leak")
	require.Same(t, fixture.workspace.repositories.row, result.Resolved.Repository)
	assert.Equal(t, 1, fixture.workspace.repositories.calls)
	assert.Equal(t, []int64{3}, fixture.workspace.skills.ids)
	assert.Equal(t, []int64{4}, fixture.workspace.knowledge.ids)
	assert.Len(t, fixture.workspace.envBundles.ids, 2)
	assert.ElementsMatch(t, []int64{5, 6}, fixture.workspace.envBundles.ids)

	decoded, err := specdomain.DecodeSpec(result.Resolved.Snapshot.SpecJSON())
	require.NoError(t, err)
	assert.Equal(t, result.Resolved.Spec, decoded)
}

func TestServicePreflightReturnsBlockingIssueForStaleOptions(t *testing.T) {
	fixture := newWorkerCreationServiceFixture()
	service := NewService(fixture.deps())
	draft := validWorkerCreationDraft()
	draft.OptionsRevision = "runtime-catalog-stale"

	result, err := service.Preflight(
		context.Background(),
		specservice.Scope{OrgID: 77, UserID: 7},
		draft,
	)

	require.NoError(t, err)
	require.Len(t, result.BlockingErrors, 1)
	assert.Equal(t, "stale-options", result.BlockingErrors[0].Code)
	assert.Equal(t, "options_revision", result.BlockingErrors[0].Field)
	assert.Nil(t, result.Resolved)
	assert.Zero(t, fixture.resources.calls)
}

func TestServicePreflightReturnsBlockingIssueForInvalidDraft(t *testing.T) {
	service := NewService(newWorkerCreationServiceFixture().deps())
	draft := validWorkerCreationDraft()
	draft.WorkerSpec.TypeConfig.AutomationLevel = "unknown"

	result, err := service.Preflight(
		context.Background(),
		specservice.Scope{OrgID: 77, UserID: 7},
		draft,
	)

	require.NoError(t, err)
	require.Len(t, result.BlockingErrors, 1)
	assert.Equal(t, "invalid-draft", result.BlockingErrors[0].Code)
	assert.Nil(t, result.Resolved)
}

func TestServicePreflightRejectsUnsupportedInteractionMode(t *testing.T) {
	fixture := newWorkerCreationServiceFixture()
	fixture.agents.agent.SupportedModes = "pty"
	definition := fixture.definitions["codex-cli"]
	definition.Modes = []string{"pty"}
	fixture.definitions["codex-cli"] = definition
	service := NewService(fixture.deps())

	result, err := service.Preflight(
		context.Background(),
		specservice.Scope{OrgID: 77, UserID: 7},
		validWorkerCreationDraft(),
	)

	require.NoError(t, err)
	require.Len(t, result.BlockingErrors, 1)
	assert.Equal(t, "invalid-draft", result.BlockingErrors[0].Code)
	assert.Equal(t, "worker_spec.type_config.interaction_mode", result.BlockingErrors[0].Field)
	assert.Nil(t, result.Resolved)
}

func TestServicePreflightDoesNotDisguiseInfrastructureErrors(t *testing.T) {
	fixture := newWorkerCreationServiceFixture()
	fixture.resources.err = assert.AnError
	service := NewService(fixture.deps())

	result, err := service.Preflight(
		context.Background(),
		specservice.Scope{OrgID: 77, UserID: 7},
		validWorkerCreationDraft(),
	)

	assert.ErrorIs(t, err, assert.AnError)
	assert.Empty(t, result.BlockingErrors)
	assert.Nil(t, result.Resolved)
}

func TestServiceValidateWorkerTypeSnapshotRejectsDefinitionDrift(t *testing.T) {
	fixture := newWorkerCreationServiceFixture()
	service := NewService(fixture.deps())
	prepared, err := service.Prepare(
		context.Background(),
		specservice.Scope{OrgID: 77, UserID: 7},
		validWorkerCreationDraft(),
	)
	require.NoError(t, err)
	source := *fixture.agents.agent.AgentfileSource + "MCP ON\n"
	fixture.agents.agent.AgentfileSource = &source
	fixture.definitions["codex-cli"] = workerDefinition("codex-cli", "codex", source, "pty", "acp")

	err = service.ValidateWorkerTypeSnapshot(
		context.Background(),
		specservice.Scope{OrgID: 77, UserID: 7},
		prepared.Spec.Runtime.WorkerType,
	)

	assert.ErrorIs(t, err, ErrWorkerTypeDefinitionChanged)
}

type workerCreationServiceFixture struct {
	agents      *workerTypeAgentProvider
	definitions staticWorkerDefinitions
	resources   *modelResourceService
	workspace   *workspaceFixture
}

func newWorkerCreationServiceFixture() *workerCreationServiceFixture {
	source := `AGENT codex
EXECUTABLE codex
CONFIG approval_mode SELECT("untrusted", "on-request", "never") = "on-request"
ENV SIGNING_KEY SECRET OPTIONAL
`
	return &workerCreationServiceFixture{
		agents: &workerTypeAgentProvider{agent: activeWorkerTypeAgent(source)},
		definitions: staticWorkerDefinitions{
			"codex-cli": workerDefinition("codex-cli", "codex", source, "pty", "acp"),
		},
		resources: validModelResourceService(),
		workspace: newWorkspaceFixture(),
	}
}

func (fixture *workerCreationServiceFixture) deps() Deps {
	return Deps{
		Catalog:      enabledCodexRuntimeCatalog(),
		Definitions:  fixture.definitions,
		Agents:       fixture.agents,
		Models:       fixture.resources,
		Repositories: fixture.workspace.repositories,
		Skills:       fixture.workspace.skills,
		Knowledge:    fixture.workspace.knowledge,
		EnvBundles:   fixture.workspace.envBundles,
		Commits:      fixture.workspace.commits,
	}
}

func validWorkerCreationDraft() Draft {
	return Draft{
		OptionsRevision:  runtimedomain.DefaultCatalogRevision(),
		OrganizationSlug: slugkit.MustNewForTest("dev-org"),
		WorkerSpec: specservice.Draft{
			ModelResourceID: 101,
			WorkerTypeSlug:  slugkit.MustNewForTest("codex-cli"),
			Runtime: specservice.RuntimeSelection{
				RuntimeImageID:    1,
				PlacementPolicy:   specdomain.PlacementPolicyExplicit,
				ComputeTargetID:   1,
				DeploymentMode:    specdomain.DeploymentModePooled,
				ResourceProfileID: 1,
			},
			TypeConfig: specdomain.TypeConfig{
				SchemaVersion: 1,
				Values: map[string]any{
					"approval_mode": "never",
				},
				SecretRefs: map[string]specdomain.SecretReference{
					"SIGNING_KEY": {
						Kind: slugkit.MustNewForTest("env-bundle"),
						ID:   6,
					},
				},
				InteractionMode: specdomain.InteractionModeACP,
				AutomationLevel: specdomain.AutomationLevelAutonomous,
			},
			Workspace: validWorkspaceDraft(),
			Lifecycle: specdomain.Lifecycle{
				TerminationPolicy: specdomain.TerminationPolicyManual,
			},
			Metadata: specdomain.Metadata{Alias: "  worker  "},
		},
	}
}
