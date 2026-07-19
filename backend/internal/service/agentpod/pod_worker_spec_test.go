package agentpod

import (
	"context"
	"strings"
	"testing"

	runtimedomain "github.com/anthropics/agentsmesh/backend/internal/domain/workerruntime"
	specdomain "github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	"github.com/anthropics/agentsmesh/backend/internal/infra"
	specservice "github.com/anthropics/agentsmesh/backend/internal/service/workerspec"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPodServicePersistsResolvedWorkerSpecWithPod(t *testing.T) {
	db := setupTestDB(t)
	ensureWorkerSpecSnapshotTable(t, db)
	service := NewPodService(infra.NewPodRepository(db))
	prepared := preparedWorkerSpecForArtifactTest(
		t,
		context.WithValue(
			context.WithValue(context.Background(), ctxKeyOrgID, int64(1)),
			ctxKeyUserID,
			int64(1),
		),
		podServiceWorkerSpec(),
		"MODE acp\n",
		nil,
	)
	resolved := prepared.Snapshot

	pod, err := service.CreatePod(context.Background(), &CreatePodRequest{
		OrganizationID:                 1,
		RunnerID:                       1,
		AgentSlug:                      "codex-cli",
		CreatedByID:                    1,
		InteractionMode:                "acp",
		AutomationLevel:                "autonomous",
		AgentfileLayer:                 "MODE acp\n",
		ResolvedWorkerSpec:             &resolved,
		WorkerDependencyArtifactJSON:   prepared.Artifact.JSON(),
		WorkerDependencyArtifactDigest: prepared.Artifact.Digest(),
	})

	require.NoError(t, err)
	require.NotNil(t, pod.WorkerSpecSnapshotID)
	var snapshotCount int64
	require.NoError(t, db.Table("worker_spec_snapshots").Count(&snapshotCount).Error)
	assert.Equal(t, int64(1), snapshotCount)
}

func TestPodServicePersistsWorkerSpecWithoutMainModelAsNull(t *testing.T) {
	db := setupTestDB(t)
	ensureWorkerSpecSnapshotTable(t, db)
	service := NewPodService(infra.NewPodRepository(db))
	spec := podServiceWorkerSpec()
	spec.Runtime.ModelBinding = specdomain.ModelBinding{}
	spec.Runtime.WorkerType.Slug = slugkit.MustNewForTest("cursor-cli")
	prepared := preparedWorkerSpecForArtifactTest(
		t,
		context.WithValue(
			context.WithValue(context.Background(), ctxKeyOrgID, int64(1)),
			ctxKeyUserID,
			int64(1),
		),
		spec,
		"MODE acp\n",
		nil,
	)
	spec = prepared.Spec
	resolved := prepared.Snapshot
	orchestrate := &OrchestrateCreatePodRequest{}
	projectPreparedWorkerSpec(orchestrate, prepared)

	require.Nil(t, orchestrate.ModelResourceID)
	pod, err := service.CreatePod(context.Background(), &CreatePodRequest{
		OrganizationID: 1, RunnerID: 1, CreatedByID: 1,
		AgentSlug: orchestrate.AgentSlug, ModelResourceID: orchestrate.ModelResourceID,
		InteractionMode: string(spec.TypeConfig.InteractionMode),
		AutomationLevel: string(spec.TypeConfig.AutomationLevel),
		AgentfileLayer:  "MODE acp\n", ResolvedWorkerSpec: &resolved,
		WorkerDependencyArtifactJSON:   prepared.Artifact.JSON(),
		WorkerDependencyArtifactDigest: prepared.Artifact.Digest(),
	})

	require.NoError(t, err)
	assert.Nil(t, pod.ModelResourceID)
	var persisted specdomainSnapshotPod
	require.NoError(t, db.Table("pods AS p").
		Select("p.model_resource_id, r.model_resource_id AS revision_model_resource_id").
		Joins("JOIN pod_config_revisions r ON r.pod_id = p.id").
		Where("p.id = ?", pod.ID).
		Scan(&persisted).Error)
	assert.Nil(t, persisted.ModelResourceID)
	assert.Nil(t, persisted.RevisionModelResourceID)
}

type specdomainSnapshotPod struct {
	ModelResourceID         *int64
	RevisionModelResourceID *int64
}

func TestPodServiceResumeInheritsExistingWorkerSpecSnapshot(t *testing.T) {
	db := setupTestDB(t)
	service := NewPodService(infra.NewPodRepository(db))
	snapshotID := int64(91)

	pod, err := service.CreatePod(context.Background(), &CreatePodRequest{
		OrganizationID:       1,
		RunnerID:             1,
		AgentSlug:            "codex-cli",
		CreatedByID:          1,
		WorkerSpecSnapshotID: &snapshotID,
	})

	require.NoError(t, err)
	require.NotNil(t, pod.WorkerSpecSnapshotID)
	assert.Equal(t, snapshotID, *pod.WorkerSpecSnapshotID)
}

func TestPodServiceRejectsNewAndInheritedWorkerSpecTogether(t *testing.T) {
	db := setupTestDB(t)
	service := NewPodService(infra.NewPodRepository(db))
	resolved := resolvedWorkerSpecForPodServiceTest(t, 1)
	snapshotID := int64(91)

	pod, err := service.CreatePod(context.Background(), &CreatePodRequest{
		OrganizationID:       1,
		RunnerID:             1,
		AgentSlug:            "codex-cli",
		CreatedByID:          1,
		ResolvedWorkerSpec:   &resolved,
		WorkerSpecSnapshotID: &snapshotID,
	})

	require.ErrorIs(t, err, ErrConflictingWorkerSpecPersistence)
	assert.Nil(t, pod)
}

func TestProjectWorkerSpecOmitsModelResourceForCredentialManagedWorker(t *testing.T) {
	spec := podServiceWorkerSpec()
	spec.Runtime.ModelBinding = specdomain.ModelBinding{}
	request := &OrchestrateCreatePodRequest{}

	projectWorkerSpec(request, spec, "MODE acp\n", nil)

	assert.Nil(t, request.ModelResourceID)
}

func TestProjectWorkerSpecAppliesLifecyclePolicy(t *testing.T) {
	tests := []struct {
		policy    specdomain.TerminationPolicy
		perpetual bool
	}{
		{policy: specdomain.TerminationPolicyManual, perpetual: true},
		{policy: specdomain.TerminationPolicyOnIdle, perpetual: false},
		{policy: specdomain.TerminationPolicyOnCompleted, perpetual: false},
	}
	for _, test := range tests {
		t.Run(string(test.policy), func(t *testing.T) {
			spec := podServiceWorkerSpec()
			spec.Lifecycle.TerminationPolicy = test.policy
			request := &OrchestrateCreatePodRequest{}

			projectWorkerSpec(request, spec, "MODE acp\n", nil)

			assert.Equal(t, test.perpetual, request.Perpetual)
		})
	}
}

func resolvedWorkerSpecForPodServiceTest(
	t *testing.T,
	organizationID int64,
) specservice.ResolvedSnapshot {
	t.Helper()
	return resolvedWorkerSpecFromSpecForPodServiceTest(t, organizationID, podServiceWorkerSpec())
}

func resolvedWorkerSpecFromSpecForPodServiceTest(
	t *testing.T,
	organizationID int64,
	spec specdomain.Spec,
) specservice.ResolvedSnapshot {
	t.Helper()
	ports := &podServiceWorkerSpecPorts{spec: spec}
	resolver := specservice.NewResolver(specservice.ResolverDeps{
		WorkerTypes: ports,
		Runtime:     ports,
		Models:      ports,
		ToolModels:  ports,
		Secrets:     ports,
		Workspaces:  ports,
	})
	resolved, err := resolver.Resolve(
		context.Background(),
		specservice.Scope{OrgID: organizationID, UserID: 1},
		specservice.Draft{
			ModelResourceID: spec.Runtime.ModelBinding.ResourceID,
			WorkerTypeSlug:  spec.Runtime.WorkerType.Slug,
			Runtime: specservice.RuntimeSelection{
				RuntimeImageID:    spec.Runtime.Image.ID,
				PlacementPolicy:   spec.Placement.Policy,
				ComputeTargetID:   spec.Placement.ComputeTarget.ID,
				DeploymentMode:    spec.Placement.DeploymentMode,
				ResourceProfileID: spec.Placement.ResourceProfile.ID,
			},
			TypeConfig: spec.TypeConfig,
			Workspace:  spec.Workspace,
			Lifecycle:  spec.Lifecycle,
			Metadata:   spec.Metadata,
		},
	)
	require.NoError(t, err)
	return resolved
}

type podServiceWorkerSpecPorts struct {
	spec specdomain.Spec
}

func (ports *podServiceWorkerSpecPorts) ResolveWorkerType(
	context.Context,
	specservice.Scope,
	slugkit.Slug,
) (specservice.WorkerTypeResolution, error) {
	modelRequirement := specdomain.ModelRequirement{
		Required: !ports.spec.Runtime.ModelBinding.IsEmpty(),
	}
	if modelRequirement.Required {
		modelRequirement.ProtocolAdapters = []slugkit.Slug{
			ports.spec.Runtime.ModelBinding.ProtocolAdapter,
		}
	}
	return specservice.WorkerTypeResolution{
		WorkerType: ports.spec.Runtime.WorkerType,
		SupportedInteractionModes: []specdomain.InteractionMode{
			ports.spec.TypeConfig.InteractionMode,
		},
		TypeSchema: specdomain.TypeSchema{
			Version: 1,
			Fields:  map[string]specdomain.TypeFieldSchema{},
		},
		ModelRequirement: modelRequirement,
	}, nil
}

func (ports *podServiceWorkerSpecPorts) ResolveRuntime(
	context.Context,
	specservice.Scope,
	slugkit.Slug,
	specservice.RuntimeSelection,
) (runtimedomain.Resolved, error) {
	return runtimedomain.Resolved{
		RuntimeImage: ports.spec.Runtime.Image,
		Placement:    ports.spec.Placement,
	}, nil
}

func (ports *podServiceWorkerSpecPorts) ResolveModel(
	context.Context,
	specservice.Scope,
	specdomain.ModelRequirement,
	int64,
) (specdomain.ModelBinding, error) {
	return ports.spec.Runtime.ModelBinding, nil
}

func (ports *podServiceWorkerSpecPorts) ResolveToolModel(
	_ context.Context,
	_ specservice.Scope,
	requirement specdomain.ToolModelRequirement,
	resourceID int64,
) (specdomain.ToolModelBinding, error) {
	for _, binding := range ports.spec.Runtime.ToolModelBindings {
		if binding.Role == requirement.Role &&
			binding.ModelBinding.ResourceID == resourceID {
			return binding, nil
		}
	}
	return specdomain.ToolModelBinding{}, specdomain.ErrNotFound
}

func (*podServiceWorkerSpecPorts) ResolveSecretReference(
	context.Context,
	specservice.Scope,
	slugkit.Slug,
	string,
	specdomain.SecretReference,
) error {
	return nil
}

func (*podServiceWorkerSpecPorts) ResolveWorkspace(
	_ context.Context,
	_ specservice.Scope,
	_ slugkit.Slug,
	workspace specdomain.Workspace,
) (specdomain.Workspace, error) {
	return workspace, nil
}

func podServiceWorkerSpec() specdomain.Spec {
	return specdomain.NewV1(
		specdomain.Runtime{
			ModelBinding: specdomain.ModelBinding{
				ResourceID:         101,
				ResourceRevision:   7,
				ConnectionID:       201,
				ConnectionRevision: 9,
				ProviderKey:        slugkit.MustNewForTest("openai"),
				ProtocolAdapter:    slugkit.MustNewForTest("openai-compatible"),
				ModelID:            "gpt-5",
			},
			WorkerType: specdomain.WorkerType{
				Slug:           slugkit.MustNewForTest("codex-cli"),
				DefinitionHash: strings.Repeat("a", 64),
			},
			Image: specdomain.RuntimeImage{
				ID:     1,
				Digest: "sha256:" + strings.Repeat("b", 64),
			},
		},
		specdomain.Placement{
			Policy: specdomain.PlacementPolicyExplicit,
			ComputeTarget: specdomain.ComputeTarget{
				ID:   1,
				Kind: specdomain.ComputeTargetKindRunnerPool,
			},
			DeploymentMode: specdomain.DeploymentModePooled,
			ResourceProfile: specdomain.ResourceProfile{
				ID: 1,
				Resources: specdomain.ResourceRequestsLimits{
					CPURequestMilliCPU: 200,
					CPULimitMilliCPU:   1000,
					MemoryRequestBytes: 256 << 20,
					MemoryLimitBytes:   1 << 30,
				},
			},
		},
		specdomain.TypeConfig{
			SchemaVersion:   1,
			Values:          map[string]any{},
			SecretRefs:      map[string]specdomain.SecretReference{},
			InteractionMode: specdomain.InteractionModeACP,
			AutomationLevel: specdomain.AutomationLevelAutonomous,
		},
		specdomain.Workspace{
			SkillIDs:               []int64{},
			KnowledgeMounts:        []specdomain.KnowledgeMount{},
			EnvBundleIDs:           []specdomain.RuntimeEnvBundleID{},
			ConfigDocumentBindings: []specdomain.ConfigDocumentBinding{},
			InitialTask:            "Run checks.",
		},
		specdomain.Lifecycle{TerminationPolicy: specdomain.TerminationPolicyManual},
		specdomain.Metadata{Alias: "worker"},
	)
}
