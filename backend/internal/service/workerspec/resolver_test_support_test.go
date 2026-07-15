package workerspec

import (
	"context"
	"errors"
	"strings"

	workerruntime "github.com/anthropics/agentsmesh/backend/internal/domain/workerruntime"
	domain "github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
)

var errCrossScopeForTest = errors.New("cross-scope resolution")

type resolverPortsForTest struct {
	calls             []string
	scopes            []Scope
	failAt            string
	failure           error
	workerType        WorkerTypeResolution
	runtime           workerruntime.Resolved
	modelBinding      domain.ModelBinding
	runtimeSelection  RuntimeSelection
	runtimeWorkerType slugkit.Slug
	modelResourceID   int64
}

func newResolverPortsForTest() *resolverPortsForTest {
	return &resolverPortsForTest{
		workerType: WorkerTypeResolution{
			WorkerType: domain.WorkerType{
				Slug:           mustSlugForTest("codex-cli"),
				DefinitionHash: strings.Repeat("a", 64),
			},
			SupportedInteractionModes: []domain.InteractionMode{
				domain.InteractionModePTY,
			},
			TypeSchema: domain.TypeSchema{
				Version: 7,
				Fields: map[string]domain.TypeFieldSchema{
					"mode":        {Kind: domain.TypeFieldSelect, Options: []string{"careful"}},
					"api-token":   {Kind: domain.TypeFieldSecret},
					"signing-key": {Kind: domain.TypeFieldSecret},
				},
			},
		},
		runtime: validResolvedRuntimeForTest(),
		modelBinding: domain.ModelBinding{
			ResourceID:         1001,
			ResourceRevision:   7,
			ConnectionID:       2001,
			ConnectionRevision: 9,
			ProviderKey:        mustSlugForTest("openai"),
			ModelID:            "gpt-5",
		},
		failure: errors.New("resolution failed"),
	}
}

func (ports *resolverPortsForTest) deps() ResolverDeps {
	return ResolverDeps{
		WorkerTypes: ports,
		Runtime:     ports,
		Models:      ports,
		Secrets:     ports,
		Workspaces:  ports,
	}
}

func (ports *resolverPortsForTest) ResolveWorkerType(
	_ context.Context,
	scope Scope,
	_ slugkit.Slug,
) (WorkerTypeResolution, error) {
	if err := ports.record("worker-type", scope); err != nil {
		return WorkerTypeResolution{}, err
	}
	return ports.workerType, nil
}

func (ports *resolverPortsForTest) ResolveRuntime(
	_ context.Context,
	scope Scope,
	workerType slugkit.Slug,
	selection RuntimeSelection,
) (workerruntime.Resolved, error) {
	ports.runtimeWorkerType = workerType
	ports.runtimeSelection = selection
	if err := ports.record("runtime", scope); err != nil {
		return workerruntime.Resolved{}, err
	}
	return ports.runtime, nil
}

func (ports *resolverPortsForTest) ResolveModel(
	_ context.Context,
	scope Scope,
	_ slugkit.Slug,
	resourceID int64,
) (domain.ModelBinding, error) {
	ports.modelResourceID = resourceID
	if err := ports.record("model", scope); err != nil {
		return domain.ModelBinding{}, err
	}
	return ports.modelBinding, nil
}

func (ports *resolverPortsForTest) ResolveSecretReference(
	_ context.Context,
	scope Scope,
	_ slugkit.Slug,
	field string,
	_ domain.SecretReference,
) error {
	return ports.record("secret:"+field, scope)
}

func (ports *resolverPortsForTest) ResolveWorkspace(
	_ context.Context,
	scope Scope,
	_ slugkit.Slug,
	workspace domain.Workspace,
) (domain.Workspace, error) {
	if err := ports.record("workspace", scope); err != nil {
		return domain.Workspace{}, err
	}
	return cloneWorkspaceForTest(workspace), nil
}

func (ports *resolverPortsForTest) record(call string, scope Scope) error {
	ports.calls = append(ports.calls, call)
	ports.scopes = append(ports.scopes, scope)
	if scope != validScopeForTest() {
		return errCrossScopeForTest
	}
	if ports.failAt == call || ports.failAt == "secret" && strings.HasPrefix(call, "secret:") {
		return ports.failure
	}
	return nil
}

type snapshotRepositoryForTest struct {
	createCalls int
	captured    ResolvedSnapshot
}

func (repository *snapshotRepositoryForTest) Create(
	_ context.Context,
	resolved ResolvedSnapshot,
) (domain.Snapshot, error) {
	repository.createCalls++
	repository.captured = resolved
	spec, err := domain.DecodeSpec(resolved.SpecJSON())
	if err != nil {
		return domain.Snapshot{}, err
	}
	summary, err := domain.DecodeSummary(resolved.SummaryJSON())
	if err != nil {
		return domain.Snapshot{}, err
	}
	return domain.Snapshot{
		ID:             901,
		OrganizationID: resolved.OrganizationID(),
		Spec:           spec,
		Summary:        summary,
	}, nil
}

func (*snapshotRepositoryForTest) GetByID(
	context.Context,
	int64,
	int64,
) (domain.Snapshot, error) {
	return domain.Snapshot{}, domain.ErrNotFound
}

func (*snapshotRepositoryForTest) ListByOrganization(
	context.Context,
	int64,
) ([]domain.Snapshot, error) {
	return nil, nil
}

func validScopeForTest() Scope {
	return Scope{OrgID: 77, UserID: 7}
}

func validDraftForTest() Draft {
	repositoryID := int64(22)
	return Draft{
		ModelResourceID: 1001,
		WorkerTypeSlug:  mustSlugForTest("codex-cli"),
		Runtime: RuntimeSelection{
			RuntimeImageID:    41,
			PlacementPolicy:   domain.PlacementPolicyExplicit,
			ComputeTargetID:   52,
			DeploymentMode:    domain.DeploymentModeDedicated,
			ResourceProfileID: 63,
		},
		TypeConfig: domain.TypeConfig{
			SchemaVersion: 7,
			Values: map[string]any{
				"mode": "careful",
			},
			SecretRefs: map[string]domain.SecretReference{
				"signing-key": {Kind: mustSlugForTest("env-bundle"), ID: 82},
				"api-token":   {Kind: mustSlugForTest("env-bundle"), ID: 81},
			},
			InteractionMode: domain.InteractionModePTY,
			AutomationLevel: domain.AutomationLevelAutonomous,
		},
		Workspace: domain.Workspace{
			RepositoryID:    &repositoryID,
			Branch:          "main",
			SkillIDs:        []int64{9, 3},
			KnowledgeMounts: []domain.KnowledgeMount{},
			EnvBundleIDs:    []domain.RuntimeEnvBundleID{},
			Instructions:    "Keep reviews strict.",
			InitialTask:     "Review the change.",
		},
		Lifecycle: domain.Lifecycle{
			TerminationPolicy: domain.TerminationPolicyManual,
		},
		Metadata: domain.Metadata{Alias: "  worker  "},
	}
}

func validPlacementForTest() domain.Placement {
	return domain.Placement{
		Policy: domain.PlacementPolicyExplicit,
		ComputeTarget: domain.ComputeTarget{
			ID:   52,
			Kind: domain.ComputeTargetKindKubernetes,
		},
		DeploymentMode: domain.DeploymentModeDedicated,
		ResourceProfile: domain.ResourceProfile{
			ID: 63,
			Resources: domain.ResourceRequestsLimits{
				CPURequestMilliCPU: 500,
				CPULimitMilliCPU:   1000,
				MemoryRequestBytes: 512 << 20,
				MemoryLimitBytes:   1024 << 20,
			},
		},
	}
}

func cloneWorkspaceForTest(workspace domain.Workspace) domain.Workspace {
	cloned := workspace
	if workspace.RepositoryID != nil {
		repositoryID := *workspace.RepositoryID
		cloned.RepositoryID = &repositoryID
	}
	cloned.SkillIDs = append([]int64{}, workspace.SkillIDs...)
	cloned.KnowledgeMounts = append([]domain.KnowledgeMount{}, workspace.KnowledgeMounts...)
	cloned.EnvBundleIDs = append([]domain.RuntimeEnvBundleID{}, workspace.EnvBundleIDs...)
	return cloned
}

func mustSlugForTest(value string) slugkit.Slug {
	return slugkit.MustNewForTest(value)
}
