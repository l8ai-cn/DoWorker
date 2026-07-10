package infra

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
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
	snapshot := workerSpecSnapshotForContract(t, 77)
	snapshot.Spec.Workspace.SkillIDs = []int64{9, 3}

	require.NoError(t, repo.Create(ctx, snapshot))
	require.Positive(t, snapshot.ID)

	var storedSpec string
	require.NoError(t, db.Raw(
		"SELECT spec_json FROM worker_spec_snapshots WHERE id = ?",
		snapshot.ID,
	).Scan(&storedSpec).Error)
	expected, err := workerspec.EncodeSpec(snapshot.Spec)
	require.NoError(t, err)
	assert.JSONEq(t, string(expected), storedSpec)

	loaded, err := repo.GetByID(ctx, 77, snapshot.ID)
	require.NoError(t, err)
	assert.Equal(t, []int64{3, 9}, loaded.Spec.Workspace.SkillIDs)
	assert.Equal(t, snapshot.Summary, loaded.Summary)

	_, err = repo.GetByID(ctx, 78, snapshot.ID)
	assert.True(t, errors.Is(err, workerspec.ErrNotFound))
}

func TestWorkerSpecSnapshotRepositoryRejectsCorruptStoredDocuments(t *testing.T) {
	ctx := context.Background()
	db := workerSpecSnapshotDBForContract(t)
	repo := NewWorkerSpecSnapshotRepository(db)
	snapshot := workerSpecSnapshotForContract(t, 79)
	require.NoError(t, repo.Create(ctx, snapshot))

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

func workerSpecSnapshotDBForContract(t *testing.T) *gorm.DB {
	t.Helper()
	db := testkit.SetupTestDB(t)
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
) *workerspec.Snapshot {
	t.Helper()
	snapshot, err := workerspec.NewSnapshot(organizationID, workerSpecForRepoContract())
	require.NoError(t, err)
	return snapshot
}

func workerSpecForRepoContract() workerspec.Spec {
	spec := workerspec.NewV1(
		workerspec.Runtime{
			ModelResourceID: 1001,
			WorkerType: workerspec.WorkerType{
				Slug:           slugkit.MustNewForTest("codex-cli"),
				DefinitionHash: strings.Repeat("a", 64),
			},
			Image: workerspec.RuntimeImage{
				ID:     41,
				Digest: "sha256:" + strings.Repeat("a", 64),
			},
		},
		workerspec.TypeConfig{
			SchemaVersion: 1,
			Values:        map[string]any{"mode": "careful"},
			SecretRefs:    map[string]workerspec.SecretReference{},
		},
		workerspec.Workspace{
			RepositoryID: int64PointerForRepoContract(22),
			Branch:       "main",
			SkillIDs:     []int64{3, 9},
		},
		workerspec.Lifecycle{},
		workerspec.Metadata{Alias: "worker"},
	)
	spec.Placement = workerspec.Placement{
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
	}
	return spec
}

func int64PointerForRepoContract(value int64) *int64 {
	return &value
}
