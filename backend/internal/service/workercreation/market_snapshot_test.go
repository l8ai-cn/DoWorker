package workercreation

import (
	"context"
	"strings"
	"testing"

	resourcedomain "github.com/anthropics/agentsmesh/backend/internal/domain/airesource"
	"github.com/anthropics/agentsmesh/backend/internal/domain/skill"
	runtimedomain "github.com/anthropics/agentsmesh/backend/internal/domain/workerruntime"
	specdomain "github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	"github.com/anthropics/agentsmesh/backend/internal/service/airesource"
	"github.com/anthropics/agentsmesh/backend/internal/service/workerdefinition"
	specservice "github.com/anthropics/agentsmesh/backend/internal/service/workerspec"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrepareMarketSnapshotRebindsModelAndResolvesInTargetScope(t *testing.T) {
	fixture := newWorkerCreationServiceFixture()
	service := NewService(fixture.deps())
	source := portableMarketSource(t, service)
	require.Equal(t, int64(101), source.Runtime.ModelBinding.ResourceID)
	source.Metadata.SourceExpertID = int64PointerForMarketSnapshotTest(31)
	fixture.resources.resolved.Resource.ID = 202
	fixture.resources.resolved.Resource.ProviderConnectionID = 201
	fixture.resources.calls = 0
	targetScope := specservice.Scope{OrgID: 88, UserID: 9}

	snapshot, err := service.PrepareMarketSnapshot(
		context.Background(),
		targetScope,
		source,
		202,
		nil,
	)

	require.NoError(t, err)
	assert.Equal(t, int64(88), snapshot.OrganizationID())
	decoded, err := specdomain.DecodeSpec(snapshot.SpecJSON())
	require.NoError(t, err)
	assert.Equal(t, int64(202), decoded.Runtime.ModelBinding.ResourceID)
	assert.Equal(t, int64(7), decoded.Runtime.ModelBinding.ResourceRevision)
	assert.Equal(t, source.Runtime.WorkerType.Slug, decoded.Runtime.WorkerType.Slug)
	assert.Equal(t, source.Runtime.Image.ID, decoded.Runtime.Image.ID)
	assert.Equal(t, source.Placement.Policy, decoded.Placement.Policy)
	assert.Equal(t, source.Placement.ComputeTarget.ID, decoded.Placement.ComputeTarget.ID)
	assert.Equal(t, source.Placement.DeploymentMode, decoded.Placement.DeploymentMode)
	assert.Equal(t, source.Placement.ResourceProfile.ID, decoded.Placement.ResourceProfile.ID)
	assert.Equal(t, source.TypeConfig, decoded.TypeConfig)
	assert.Equal(t, source.Workspace.SkillIDs, decoded.Workspace.SkillIDs)
	assert.Equal(t, source.Workspace.Instructions, decoded.Workspace.Instructions)
	assert.Equal(t, source.Workspace.InitialTask, decoded.Workspace.InitialTask)
	assert.Equal(t, source.Lifecycle, decoded.Lifecycle)
	assert.Equal(t, "worker", decoded.Metadata.Alias)
	assert.Nil(t, decoded.Metadata.SourceExpertID)
	assert.Equal(t, int64(88), fixture.resources.orgID)
	assert.Equal(t, int64(9), fixture.resources.actor.UserID)
	assert.Equal(t, int64(202), fixture.resources.resourceID)
}

func TestPrepareMarketSnapshotRebindsRequiredToolModels(t *testing.T) {
	fixture := seedanceMarketSnapshotFixture()
	deps := fixture.deps()
	deps.Catalog = runtimedomain.DefaultCatalog()
	service := NewService(deps)
	draft := validWorkerCreationDraft()
	draft.WorkerSpec.WorkerTypeSlug = slugkit.MustNewForTest("video-studio")
	draft.WorkerSpec.Runtime.RuntimeImageID = 4
	draft.WorkerSpec.ToolModelResourceIDs = map[string]int64{
		"seedance-video": 102,
	}
	draft.WorkerSpec.TypeConfig.Values = map[string]any{}
	prepared, err := service.Prepare(
		context.Background(),
		specservice.Scope{OrgID: 77, UserID: 7},
		draft,
	)
	require.NoError(t, err)
	source := prepared.Spec
	source.Workspace.RepositoryID = nil
	source.Workspace.Branch = ""
	source.Workspace.KnowledgeMounts = nil
	source.Workspace.EnvBundleIDs = nil
	source.TypeConfig.SecretRefs = map[string]specdomain.SecretReference{}

	snapshot, err := service.PrepareMarketSnapshot(
		context.Background(),
		specservice.Scope{OrgID: 88, UserID: 9},
		source,
		202,
		map[string]int64{"seedance-video": 302},
	)

	require.NoError(t, err)
	decoded, err := specdomain.DecodeSpec(snapshot.SpecJSON())
	require.NoError(t, err)
	assert.Equal(t, int64(202), decoded.Runtime.ModelBinding.ResourceID)
	require.Len(t, decoded.Runtime.ToolModelBindings, 1)
	assert.Equal(t, "seedance-video", decoded.Runtime.ToolModelBindings[0].Role.String())
	assert.Equal(
		t,
		int64(302),
		decoded.Runtime.ToolModelBindings[0].ModelBinding.ResourceID,
	)
}

func TestPrepareMarketSnapshotPreservesCustomResources(t *testing.T) {
	fixture := newWorkerCreationServiceFixture()
	service := NewService(fixture.deps())
	source := portableMarketSource(t, service)
	source.Placement.ResourceProfile.ID = 0
	source.Placement.ResourceProfile.Custom = true
	fixture.resources.resolved.Resource.ID = 202

	snapshot, err := service.PrepareMarketSnapshot(
		context.Background(),
		specservice.Scope{OrgID: 88, UserID: 9},
		source,
		202,
		nil,
	)

	require.NoError(t, err)
	decoded, err := specdomain.DecodeSpec(snapshot.SpecJSON())
	require.NoError(t, err)
	assert.Zero(t, decoded.Placement.ResourceProfile.ID)
	assert.True(t, decoded.Placement.ResourceProfile.Custom)
	assert.Equal(
		t,
		source.Placement.ResourceProfile.Resources,
		decoded.Placement.ResourceProfile.Resources,
	)
}

func TestPrepareMarketSnapshotRejectsPrivateReferences(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*specdomain.Spec)
		match  string
	}{
		{
			name: "repository",
			mutate: func(source *specdomain.Spec) {
				id := int64(22)
				source.Workspace.RepositoryID = &id
				source.Workspace.Branch = "main"
			},
			match: "workspace.repository_id",
		},
		{
			name: "knowledge mount",
			mutate: func(source *specdomain.Spec) {
				source.Workspace.KnowledgeMounts = []specdomain.KnowledgeMount{
					{KnowledgeBaseID: 4, Mode: specdomain.KnowledgeMountReadOnly},
				}
			},
			match: "workspace.knowledge_mounts",
		},
		{
			name: "env bundle",
			mutate: func(source *specdomain.Spec) {
				source.Workspace.EnvBundleIDs = []specdomain.RuntimeEnvBundleID{5}
			},
			match: "workspace.env_bundle_ids",
		},
		{
			name: "config bundle",
			mutate: func(source *specdomain.Spec) {
				source.Workspace.ConfigBundleIDs = []int64{6}
			},
			match: "workspace.config_bundle_ids",
		},
		{
			name: "secret ref",
			mutate: func(source *specdomain.Spec) {
				source.TypeConfig.SecretRefs = map[string]specdomain.SecretReference{
					"SIGNING_KEY": {Kind: slugkit.MustNewForTest("env-bundle"), ID: 6},
				}
			},
			match: "type_config.secret_refs",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fixture := newWorkerCreationServiceFixture()
			service := NewService(fixture.deps())
			source := portableMarketSource(t, service)
			test.mutate(&source)
			fixture.resources.calls = 0

			_, err := service.PrepareMarketSnapshot(
				context.Background(),
				specservice.Scope{OrgID: 88, UserID: 9},
				source,
				202,
				nil,
			)

			require.Error(t, err)
			assert.ErrorIs(t, err, specservice.ErrInvalidDraft)
			assert.ErrorContains(t, err, test.match)
			assert.Zero(t, fixture.resources.calls)
		})
	}
}

func TestPrepareMarketSnapshotRejectsInvalidModelResourceID(t *testing.T) {
	fixture := newWorkerCreationServiceFixture()
	service := NewService(fixture.deps())
	source := portableMarketSource(t, service)
	fixture.resources.calls = 0

	_, err := service.PrepareMarketSnapshot(
		context.Background(),
		specservice.Scope{OrgID: 88, UserID: 9},
		source,
		0,
		nil,
	)

	require.Error(t, err)
	assert.ErrorIs(t, err, specservice.ErrInvalidDraft)
	assert.ErrorContains(t, err, "model_resource_id")
	assert.Zero(t, fixture.resources.calls)
}

func TestPrepareMarketSnapshotRejectsNonPlatformOrUnavailableSkill(t *testing.T) {
	tests := map[string]func(*skill.Skill){
		"organization scoped": func(row *skill.Skill) {
			organizationID := int64(77)
			row.OrganizationID = &organizationID
		},
		"inactive": func(row *skill.Skill) {
			row.IsActive = false
		},
		"package missing": func(row *skill.Skill) {
			row.StorageKey = ""
		},
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			fixture := newWorkerCreationServiceFixture()
			service := NewService(fixture.deps())
			source := portableMarketSource(t, service)
			source.Workspace.SkillPackages = nil
			mutate(fixture.workspace.skills.rows[3])
			fixture.resources.resolved.Resource.ID = 202

			_, err := service.PrepareMarketSnapshot(
				context.Background(),
				specservice.Scope{OrgID: 88, UserID: 9},
				source,
				202,
				nil,
			)
			require.Error(t, err)
			assert.ErrorIs(t, err, specservice.ErrInvalidDraft)
			assert.ErrorContains(t, err, "workspace.skill_ids")
		})
	}
}

func portableMarketSource(t *testing.T, service *Service) specdomain.Spec {
	t.Helper()
	prepared, err := service.Prepare(
		context.Background(),
		specservice.Scope{OrgID: 77, UserID: 7},
		validWorkerCreationDraft(),
	)
	require.NoError(t, err)
	source := prepared.Spec
	source.Workspace.RepositoryID = nil
	source.Workspace.Branch = ""
	source.Workspace.KnowledgeMounts = []specdomain.KnowledgeMount{}
	source.Workspace.EnvBundleIDs = []specdomain.RuntimeEnvBundleID{}
	source.TypeConfig.SecretRefs = map[string]specdomain.SecretReference{}
	return source
}

func int64PointerForMarketSnapshotTest(value int64) *int64 {
	return &value
}

func seedanceMarketSnapshotFixture() *workerCreationServiceFixture {
	source := `AGENT video-studio
EXECUTABLE video-studio-codex
MODE pty
MODE acp
ENV SIGNING_KEY SECRET OPTIONAL
`
	definition := workerDefinition(
		"video-studio",
		"video-studio-codex",
		source,
		"pty",
		"acp",
	)
	definition.DefinitionHash = strings.Repeat("b", 64)
	definition.Image = workerdefinition.Image{
		Runtime: "video-studio", VersionProbe: []string{"video-studio-codex", "--version"},
	}
	definition.ToolModelRequirements = []workerdefinition.ToolModelRequirement{
		{
			ID: "seedance-video", ProviderKeys: []string{"doubao"},
			ProtocolAdapters: []string{"openai-compatible"},
			Modality:         "video", Capability: "video-generation",
			Environment: workerdefinition.ToolModelEnvironment{
				APIKey: "SEEDANCE_API_KEY", BaseURL: "SEEDANCE_BASE_URL",
				ModelID: "SEEDANCE_MODEL",
			},
		},
	}
	resources := validModelResourceService()
	resources.resolvedByID = map[int64]*airesource.ResolvedResource{
		101: marketResolvedResource(101, "openai", "gpt-5"),
		102: marketResolvedResource(102, "doubao", "doubao-seedance-2-0-260128"),
		202: marketResolvedResource(202, "openai", "gpt-5.5"),
		302: marketResolvedResource(302, "doubao", "doubao-seedance-2-0-260128"),
	}
	return &workerCreationServiceFixture{
		agents: &workerTypeAgentProvider{
			agent: activeWorkerTypeAgentFor(
				"video-studio",
				"video-studio-codex",
				source,
			),
		},
		definitions: staticWorkerDefinitions{"video-studio": definition},
		resources:   resources,
		workspace:   newWorkspaceFixture(),
	}
}

func marketResolvedResource(
	id int64,
	provider, modelID string,
) *airesource.ResolvedResource {
	providerSlug := slugkit.MustNewForTest(provider)
	return &airesource.ResolvedResource{
		Provider: resourcedomain.ProviderDefinition{
			Key: providerSlug, ProtocolAdapter: "openai-compatible",
		},
		Connection: resourcedomain.Connection{
			ID: id + 1000, ProviderKey: providerSlug, Revision: 1,
		},
		Resource: resourcedomain.ModelResource{
			ID: id, ProviderConnectionID: id + 1000, ModelID: modelID, Revision: 1,
		},
	}
}
