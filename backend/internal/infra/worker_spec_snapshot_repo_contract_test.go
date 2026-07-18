package infra

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	workerruntime "github.com/anthropics/agentsmesh/backend/internal/domain/workerruntime"
	"github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	workerspecservice "github.com/anthropics/agentsmesh/backend/internal/service/workerspec"
	"github.com/anthropics/agentsmesh/backend/internal/testkit"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestWorkerSpecSnapshotRepositoryPersistsCanonicalScopedDocuments(t *testing.T) {
	ctx := context.Background()
	db := workerSpecSnapshotDBForContract(t)
	repo := NewWorkerSpecSnapshotRepository(db)
	resolved := workerSpecSnapshotForContract(t, 77)

	created, err := repo.Create(ctx, resolved)
	require.NoError(t, err)
	require.Positive(t, created.ID)

	var storedSpec string
	require.NoError(t, db.Raw(
		"SELECT spec_json FROM worker_spec_snapshots WHERE id = ?",
		created.ID,
	).Scan(&storedSpec).Error)
	assert.JSONEq(t, string(resolved.SpecJSON()), storedSpec)

	loaded, err := repo.GetByID(ctx, 77, created.ID)
	require.NoError(t, err)
	assert.Equal(t, []int64{3, 9}, loaded.Spec.Workspace.SkillIDs)
	assert.Equal(t, created.Summary, loaded.Summary)

	_, err = repo.GetByID(ctx, 78, created.ID)
	assert.True(t, errors.Is(err, workerspec.ErrNotFound))

	other, err := repo.Create(ctx, workerSpecSnapshotForContract(t, 78))
	require.NoError(t, err)
	loadedBatch, err := repo.GetByIDs(ctx, 77, []int64{created.ID, other.ID})
	require.NoError(t, err)
	require.Len(t, loadedBatch, 1)
	assert.Equal(t, created.ID, loadedBatch[0].ID)
}

func TestWorkerSpecSnapshotRepositoryRejectsCorruptStoredDocuments(t *testing.T) {
	ctx := context.Background()
	db := workerSpecSnapshotDBForContract(t)
	repo := NewWorkerSpecSnapshotRepository(db)
	resolved := workerSpecSnapshotForContract(t, 79)
	snapshot, err := repo.Create(ctx, resolved)
	require.NoError(t, err)

	corrupt := snapshot.Summary
	corrupt.Alias = "other"
	encoded, err := workerspec.EncodeSummary(corrupt)
	require.NoError(t, err)
	require.NoError(t, db.Exec(
		"UPDATE worker_spec_snapshots SET summary_json = ? WHERE id = ?",
		encoded,
		snapshot.ID,
	).Error)

	_, err = repo.GetByID(ctx, 79, snapshot.ID)
	assert.True(t, errors.Is(err, workerspec.ErrSummaryMismatch))
}

func TestWorkerSpecSnapshotRepositoryReadsLegacyProtocolAdapterSnapshot(t *testing.T) {
	ctx := context.Background()
	db := workerSpecSnapshotDBForContract(t)
	repo := NewWorkerSpecSnapshotRepository(db)
	spec := workerSpecForRepoContract()
	summary, err := workerspec.Summarize(spec)
	require.NoError(t, err)
	spec.Runtime.ModelBinding.ProtocolAdapter = ""
	summary.ModelBinding.ProtocolAdapter = ""
	specJSON, err := json.Marshal(spec)
	require.NoError(t, err)
	summaryJSON, err := json.Marshal(summary)
	require.NoError(t, err)
	require.NoError(t, db.Exec(`
INSERT INTO worker_spec_snapshots (organization_id, version, spec_json, summary_json)
VALUES (?, ?, ?, ?)`,
		77, workerspec.VersionV1, specJSON, summaryJSON,
	).Error)

	snapshots, err := repo.ListByOrganization(ctx, 77)

	require.NoError(t, err)
	require.Len(t, snapshots, 1)
	assert.False(t, workerspec.HasResolvedProtocolAdapters(snapshots[0].Spec))
}

func TestWorkerSpecSnapshotRepositoryRejectsZeroAggregate(t *testing.T) {
	ctx := context.Background()
	db := workerSpecSnapshotDBForContract(t)
	repo := NewWorkerSpecSnapshotRepository(db)

	snapshot, err := repo.Create(ctx, workerspecservice.ResolvedSnapshot{})

	assert.Equal(t, workerspec.Snapshot{}, snapshot)
	assert.ErrorIs(t, err, workerspecservice.ErrInvalidResolvedSnapshot)
	var count int64
	require.NoError(t, db.Table("worker_spec_snapshots").Count(&count).Error)
	assert.Zero(t, count)
}

func workerSpecSnapshotDBForContract(t *testing.T) *gorm.DB {
	t.Helper()
	db := testkit.SetupTestDB(t)
	require.NoError(t, db.Exec(`
INSERT INTO runners (id, organization_id, cluster_id, node_id)
VALUES (1, 77, 700, 'workerspec-runner')
`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE worker_spec_snapshots (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		organization_id INTEGER NOT NULL,
		version INTEGER NOT NULL,
		spec_json BLOB NOT NULL,
		summary_json BLOB NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`).Error)
	return db
}

func workerSpecSnapshotForContract(
	t *testing.T,
	organizationID int64,
) workerspecservice.ResolvedSnapshot {
	t.Helper()
	spec := workerSpecForRepoContract()
	ports := &workerSpecResolutionPorts{spec: spec}
	resolver := workerspecservice.NewResolver(workerspecservice.ResolverDeps{
		WorkerTypes: ports,
		Runtime:     ports,
		Models:      ports,
		ToolModels:  ports,
		Secrets:     ports,
		Workspaces:  ports,
	})
	snapshot, err := resolver.Resolve(
		context.Background(),
		workerspecservice.Scope{OrgID: organizationID, UserID: 7},
		workerspecservice.Draft{
			ModelResourceID: spec.Runtime.ModelBinding.ResourceID,
			ToolModelResourceIDs: map[string]int64{
				"video-generator": 3001,
			},
			WorkerTypeSlug: spec.Runtime.WorkerType.Slug,
			Runtime: workerspecservice.RuntimeSelection{
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
	return snapshot
}

type workerSpecResolutionPorts struct {
	spec workerspec.Spec
}

func (ports *workerSpecResolutionPorts) ResolveWorkerType(
	context.Context,
	workerspecservice.Scope,
	slugkit.Slug,
) (workerspecservice.WorkerTypeResolution, error) {
	return workerspecservice.WorkerTypeResolution{
		WorkerType: ports.spec.Runtime.WorkerType,
		SupportedInteractionModes: []workerspec.InteractionMode{
			ports.spec.TypeConfig.InteractionMode,
		},
		TypeSchema: workerspec.TypeSchema{
			Version: ports.spec.TypeConfig.SchemaVersion,
			Fields: map[string]workerspec.TypeFieldSchema{
				"mode": {
					Kind:    workerspec.TypeFieldSelect,
					Options: []string{"careful"},
				},
			},
		},
		ModelRequirement: workerspec.ModelRequirement{
			Required: true,
			ProtocolAdapters: []slugkit.Slug{
				ports.spec.Runtime.ModelBinding.ProtocolAdapter,
			},
		},
		ToolModelRequirements: []workerspec.ToolModelRequirement{{
			Role: ports.spec.Runtime.ToolModelBindings[0].Role,
			ProviderKeys: []slugkit.Slug{
				ports.spec.Runtime.ToolModelBindings[0].ModelBinding.ProviderKey,
			},
			ProtocolAdapters: []slugkit.Slug{
				ports.spec.Runtime.ToolModelBindings[0].ModelBinding.ProtocolAdapter,
			},
			Modality:    ports.spec.Runtime.ToolModelBindings[0].Modality,
			Capability:  ports.spec.Runtime.ToolModelBindings[0].Capability,
			Environment: ports.spec.Runtime.ToolModelBindings[0].Environment,
		}},
	}, nil
}

func (ports *workerSpecResolutionPorts) ResolveRuntime(
	context.Context,
	workerspecservice.Scope,
	slugkit.Slug,
	workerspecservice.RuntimeSelection,
) (workerruntime.Resolved, error) {
	return workerruntime.Resolved{
		RuntimeImage: ports.spec.Runtime.Image,
		Placement:    ports.spec.Placement,
	}, nil
}

func (ports *workerSpecResolutionPorts) ResolveModel(
	context.Context,
	workerspecservice.Scope,
	workerspec.ModelRequirement,
	int64,
) (workerspec.ModelBinding, error) {
	return ports.spec.Runtime.ModelBinding, nil
}

func (ports *workerSpecResolutionPorts) ResolveToolModel(
	_ context.Context,
	_ workerspecservice.Scope,
	requirement workerspec.ToolModelRequirement,
	resourceID int64,
) (workerspec.ToolModelBinding, error) {
	for _, binding := range ports.spec.Runtime.ToolModelBindings {
		if binding.Role == requirement.Role &&
			binding.ModelBinding.ResourceID == resourceID {
			return binding, nil
		}
	}
	return workerspec.ToolModelBinding{}, workerspec.ErrNotFound
}

func (*workerSpecResolutionPorts) ResolveSecretReference(
	context.Context,
	workerspecservice.Scope,
	slugkit.Slug,
	string,
	workerspec.SecretReference,
) error {
	return nil
}

func (*workerSpecResolutionPorts) ResolveWorkspace(
	_ context.Context,
	_ workerspecservice.Scope,
	_ slugkit.Slug,
	workspace workerspec.Workspace,
) (workerspec.Workspace, error) {
	return workspace, nil
}

func workerSpecForRepoContract() workerspec.Spec {
	spec := workerspec.NewV1(
		workerspec.Runtime{
			ModelBinding: workerspec.ModelBinding{
				ResourceID:         1001,
				ResourceRevision:   7,
				ConnectionID:       2001,
				ConnectionRevision: 9,
				ProviderKey:        slugkit.MustNewForTest("openai"),
				ProtocolAdapter:    slugkit.MustNewForTest("openai-compatible"),
				ModelID:            "gpt-5",
			},
			ToolModelBindings: []workerspec.ToolModelBinding{{
				Role: slugkit.MustNewForTest("video-generator"),
				ModelBinding: workerspec.ModelBinding{
					ResourceID: 3001, ResourceRevision: 4,
					ConnectionID: 4001, ConnectionRevision: 5,
					ProviderKey:     slugkit.MustNewForTest("volcengine"),
					ProtocolAdapter: slugkit.MustNewForTest("openai-compatible"),
					ModelID:         "video-1",
				},
				Modality: "video", Capability: "video-generation",
				Environment: workerspec.ToolModelEnvironment{
					APIKey: "VIDEO_API_KEY", BaseURL: "VIDEO_BASE_URL", ModelID: "VIDEO_MODEL_ID",
				},
			}},
			WorkerType: workerspec.WorkerType{
				Slug:           slugkit.MustNewForTest("codex-cli"),
				DefinitionHash: strings.Repeat("a", 64),
			},
			Image: workerspec.RuntimeImage{
				ID:     41,
				Digest: "sha256:" + strings.Repeat("a", 64),
			},
		},
		workerspec.Placement{
			Policy: workerspec.PlacementPolicyExplicit,
			ComputeTarget: workerspec.ComputeTarget{
				ID:   52,
				Kind: workerspec.ComputeTargetKindKubernetes,
			},
			DeploymentMode: workerspec.DeploymentModePooled,
			ResourceProfile: workerspec.ResourceProfile{
				ID: 63,
				Resources: workerspec.ResourceRequestsLimits{
					CPURequestMilliCPU: 250,
					CPULimitMilliCPU:   500,
					MemoryRequestBytes: 256 << 20,
					MemoryLimitBytes:   512 << 20,
				},
			},
		},
		workerspec.TypeConfig{
			SchemaVersion:   1,
			Values:          map[string]any{"mode": "careful"},
			SecretRefs:      map[string]workerspec.SecretReference{},
			InteractionMode: workerspec.InteractionModePTY,
			AutomationLevel: workerspec.AutomationLevelAutonomous,
		},
		workerspec.Workspace{
			RepositoryID: int64PointerForRepoContract(22),
			Branch:       "main",
			SkillIDs:     []int64{3, 9},
		},
		workerspec.Lifecycle{
			TerminationPolicy: workerspec.TerminationPolicyManual,
		},
		workerspec.Metadata{Alias: "worker"},
	)
	return spec
}

func int64PointerForRepoContract(value int64) *int64 {
	return &value
}
