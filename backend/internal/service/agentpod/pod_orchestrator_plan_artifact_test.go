package agentpod

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	resourcedomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/airesource"
	"github.com/l8ai-cn/agentcloud/backend/internal/domain/gitprovider"
	control "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationcontrol"
	resource "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationresource"
	"github.com/l8ai-cn/agentcloud/backend/internal/domain/user"
	"github.com/l8ai-cn/agentcloud/backend/internal/domain/workerdependency"
	specdomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/workerspec"
	resourcesvc "github.com/l8ai-cn/agentcloud/backend/internal/service/airesource"
	workercreation "github.com/l8ai-cn/agentcloud/backend/internal/service/workercreation"
	"github.com/l8ai-cn/agentcloud/backend/internal/service/workerdefinition"
	"github.com/l8ai-cn/agentcloud/backend/internal/service/workerdependencyartifact"
	"github.com/l8ai-cn/agentcloud/backend/pkg/slugkit"
	"github.com/stretchr/testify/require"
)

func planArtifactForTest(
	t *testing.T,
	ctx context.Context,
	spec *specdomain.Spec,
	layer string,
	model *resourcesvc.ResolvedResource,
	repository *gitprovider.Repository,
	userServices ...UserServiceForOrchestrator,
) (*workerdependencyartifact.Artifact, *workerdependency.Document) {
	t.Helper()
	definition := formalWorkerDefinitionForPlanTest(t, spec.Runtime.WorkerType.Slug.String())
	spec.Runtime.WorkerType.DefinitionHash = definition.DefinitionHash
	artifactLayer := planSourceArtifactLayer(*spec, layer)
	input := workerdependencyartifact.Input{
		Scope: control.Scope{
			OrganizationID:   ctxOrgID(ctx),
			OrganizationSlug: slugkit.MustNewForTest("test-org"),
			ActorID:          ctxUserID(ctx),
		},
		Definition:     definition,
		AgentfileLayer: artifactLayer,
		WorkerSpec:     *spec,
	}
	input.Dependencies = planResolvedDependenciesForTest(
		t,
		ctx,
		input.Scope,
		*spec,
		model,
		repository,
		planUserServiceForArtifactTest(userServices),
	)
	input.PlanReferences = planReferencesForTest(input.Dependencies)
	artifact, err := workerdependencyartifact.Build(input)
	require.NoError(t, err)
	document, err := workerdependency.Decode(artifact.JSON())
	require.NoError(t, err)
	return &artifact, &document
}

func preparedWorkerSpecForArtifactTest(
	t *testing.T,
	ctx context.Context,
	spec specdomain.Spec,
	layer string,
	repository *gitprovider.Repository,
) workercreation.Prepared {
	t.Helper()
	model := resolvedModelResourceFromSpecForArtifactTest(t, spec)
	artifact, dependencies := planArtifactForTest(
		t,
		ctx,
		&spec,
		layer,
		model,
		repository,
	)
	return workercreation.Prepared{
		Snapshot:       resolvedWorkerSpecFromSpecForPodServiceTest(t, ctxOrgID(ctx), spec),
		Spec:           spec,
		AgentfileLayer: planSourceArtifactLayer(spec, layer),
		Repository:     repository,
		Artifact:       artifact,
		Dependencies:   dependencies,
	}
}

func resolvedModelResourceFromSpecForArtifactTest(
	t *testing.T,
	spec specdomain.Spec,
) *resourcesvc.ResolvedResource {
	t.Helper()
	binding := spec.Runtime.ModelBinding
	if binding.IsEmpty() {
		return nil
	}
	provider, ok := resourcedomain.Provider(binding.ProviderKey.String())
	require.True(t, ok)
	return &resourcesvc.ResolvedResource{
		Provider: provider,
		Connection: resourcedomain.Connection{
			ID: binding.ConnectionID, ProviderKey: binding.ProviderKey,
			BaseURL: provider.DefaultBaseURL, Revision: binding.ConnectionRevision,
		},
		Resource: resourcedomain.ModelResource{
			ID: binding.ResourceID, ProviderConnectionID: binding.ConnectionID,
			ModelID: binding.ModelID, Revision: binding.ResourceRevision,
		},
		Credentials: map[string]string{"api_key": "sk-test"},
	}
}

func formalWorkerDefinitionForPlanTest(t *testing.T, slug string) workerdefinition.Definition {
	t.Helper()
	root := filepath.Join("..", "..", "..", "..", "config", "worker-types", slug)
	source, err := os.ReadFile(filepath.Join(root, "definition.json"))
	require.NoError(t, err)
	agentfile, err := os.ReadFile(filepath.Join(root, "AgentFile"))
	require.NoError(t, err)
	definition, err := workerdefinition.ParseSnapshot(source, string(agentfile))
	require.NoError(t, err)
	return definition
}

func planResolvedDependenciesForTest(
	t *testing.T,
	ctx context.Context,
	scope control.Scope,
	spec specdomain.Spec,
	model *resourcesvc.ResolvedResource,
	repository *gitprovider.Repository,
	userService UserServiceForOrchestrator,
) workerdependencyartifact.ResolvedDependencies {
	t.Helper()
	return workerdependencyartifact.ResolvedDependencies{
		PrimaryModel:   planPrimaryModelForTest(t, scope, spec, model),
		ToolModels:     planToolModelsForTest(t, scope, spec),
		Repository:     planRepositoryForArtifactTest(t, ctx, scope, spec, repository, userService),
		Skills:         planSkillsForArtifactTest(t, scope, spec.Workspace.SkillIDs),
		KnowledgeBases: planKnowledgeForArtifactTest(t, scope, spec.Workspace.KnowledgeMounts),
		RuntimeBundles: planRuntimeBundlesForArtifactTest(t, scope, spec.Workspace),
		Placement:      planPlacementForArtifactTest(t, scope, spec),
	}
}

func planPrimaryModelForTest(
	t *testing.T,
	scope control.Scope,
	spec specdomain.Spec,
	model *resourcesvc.ResolvedResource,
) *workerdependencyartifact.ModelResolution {
	t.Helper()
	if spec.Runtime.ModelBinding.IsEmpty() {
		return nil
	}
	ref := planResolvedReferenceForTest(scope, resource.KindModelBinding, "primary-model", 2)
	resolution := modelResolutionForPlanTest(t, scope, ref, spec.Runtime.ModelBinding, model)
	return &resolution
}

func modelResolutionForPlanTest(
	t *testing.T,
	scope control.Scope,
	ref control.ResolvedReference,
	binding specdomain.ModelBinding,
	model *resourcesvc.ResolvedResource,
) workerdependencyartifact.ModelResolution {
	t.Helper()
	projection := planResourceProjectionForTest(t, scope, ref, binding.ResourceID)
	baseURL := ""
	modalities := []resourcedomain.Modality{resourcedomain.ModalityChat}
	capabilities := []resourcedomain.Capability{resourcedomain.CapabilityTextGeneration}
	if model != nil {
		baseURL = model.Connection.BaseURL
		if len(model.Resource.Modalities) > 0 {
			modalities = append([]resourcedomain.Modality{}, model.Resource.Modalities...)
		}
		if len(model.Resource.Capabilities) > 0 {
			capabilities = append([]resourcedomain.Capability{}, model.Resource.Capabilities...)
		}
	}
	if strings.TrimSpace(baseURL) == "" {
		provider, ok := resourcedomain.Provider(binding.ProviderKey.String())
		require.True(t, ok)
		baseURL = provider.DefaultBaseURL
	}
	return workerdependencyartifact.ModelResolution{
		ResourceResolution: projection,
		ResourceRevision:   binding.ResourceRevision,
		ConnectionID:       binding.ConnectionID,
		ConnectionRevision: binding.ConnectionRevision,
		ProviderKey:        binding.ProviderKey,
		ProtocolAdapter:    binding.ProtocolAdapter,
		ModelID:            binding.ModelID,
		BaseURL:            baseURL,
		Modalities:         modalities,
		Capabilities:       capabilities,
	}
}

func planToolModelsForTest(
	t *testing.T,
	scope control.Scope,
	spec specdomain.Spec,
) []workerdependencyartifact.ToolModelResolution {
	t.Helper()
	out := make([]workerdependencyartifact.ToolModelResolution, 0, len(spec.Runtime.ToolModelBindings))
	for _, binding := range spec.Runtime.ToolModelBindings {
		tool := planResolvedReferenceForTest(scope, resource.KindToolBinding, binding.Role.String(), 3)
		modelRef := planResolvedReferenceForTest(scope, resource.KindModelBinding, binding.Role.String()+"-model", 4)
		out = append(out, workerdependencyartifact.ToolModelResolution{
			Binding:  tool,
			Role:     binding.Role,
			Model:    modelResolutionForPlanTest(t, scope, modelRef, binding.ModelBinding, nil),
			Modality: binding.Modality, Capability: binding.Capability,
			Environment: workerdependencyartifact.ToolModelEnvironmentResolution{
				APIKeyTarget:  binding.Environment.APIKey,
				BaseURLTarget: binding.Environment.BaseURL,
				ModelIDTarget: binding.Environment.ModelID,
			},
		})
	}
	return out
}

func planRepositoryForArtifactTest(
	t *testing.T,
	ctx context.Context,
	scope control.Scope,
	spec specdomain.Spec,
	repository *gitprovider.Repository,
	userService UserServiceForOrchestrator,
) *workerdependencyartifact.RepositoryResolution {
	t.Helper()
	if spec.Workspace.RepositoryID == nil || repository == nil {
		return nil
	}
	ref := planResolvedReferenceForTest(scope, resource.KindRepository, "repository", 5)
	credential := planRepositoryCredentialForArtifactTest(t, ctx, scope, userService)
	preparationScript := stringPtrValue(repository.PreparationScript)
	return &workerdependencyartifact.RepositoryResolution{
		ResourceResolution:        planResourceProjectionForTest(t, scope, ref, *spec.Workspace.RepositoryID),
		HTTPCloneURL:              repository.HttpCloneURL,
		SSHCloneURL:               repository.SshCloneURL,
		Branch:                    spec.Workspace.Branch,
		CommitSHA:                 strings.Repeat("d", 40),
		CredentialType:            credential.Type,
		CredentialID:              credential.CredentialID,
		CredentialOwnerUserID:     credential.OwnerUserID,
		PreparationScript:         preparationScript,
		PreparationScriptDigest:   planPreparationScriptDigestForTest(preparationScript),
		PreparationTimeoutSeconds: uint32(intPtrValue(repository.PreparationTimeout)),
	}
}

func planUserServiceForArtifactTest(
	services []UserServiceForOrchestrator,
) UserServiceForOrchestrator {
	if len(services) == 0 {
		return nil
	}
	return services[0]
}

func planRepositoryCredentialForArtifactTest(
	t *testing.T,
	ctx context.Context,
	scope control.Scope,
	userService UserServiceForOrchestrator,
) workerdependency.RepositoryCredential {
	t.Helper()
	if userService == nil {
		return workerdependency.RepositoryCredential{Type: workerdependency.RepositoryCredentialTypeNone}
	}
	credential, err := userService.GetDefaultGitCredential(ctx, scope.ActorID)
	if err != nil || credential == nil {
		return workerdependency.RepositoryCredential{Type: workerdependency.RepositoryCredentialTypeNone}
	}
	switch credential.CredentialType {
	case user.CredentialTypeRunnerLocal:
		return workerdependency.RepositoryCredential{Type: user.CredentialTypeRunnerLocal}
	case user.CredentialTypeOAuth, user.CredentialTypePAT, user.CredentialTypeSSHKey:
		return workerdependency.RepositoryCredential{
			Type:         credential.CredentialType,
			CredentialID: &credential.ID,
			OwnerUserID:  scope.ActorID,
		}
	default:
		require.Failf(t, "unsupported credential type", credential.CredentialType)
		return workerdependency.RepositoryCredential{}
	}
}

func planPreparationScriptDigestForTest(script string) string {
	if strings.TrimSpace(script) == "" {
		return ""
	}
	return workerdependency.TextDigest(script)
}
