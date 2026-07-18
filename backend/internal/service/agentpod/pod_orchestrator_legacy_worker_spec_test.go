package agentpod

import (
	"context"
	"encoding/json"
	"testing"

	control "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
	specdomain "github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	"github.com/anthropics/agentsmesh/backend/internal/infra"
	workercreation "github.com/anthropics/agentsmesh/backend/internal/service/workercreation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrepareSnapshotWorkerCreateRejectsPersistedLegacyProtocolAdapter(t *testing.T) {
	db := setupTestDB(t)
	require.NoError(t, db.Exec(`CREATE TABLE worker_spec_snapshots (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		organization_id INTEGER NOT NULL,
		version INTEGER NOT NULL,
		spec_json BLOB NOT NULL,
		summary_json BLOB NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`).Error)
	spec := podServiceWorkerSpec()
	summary, err := specdomain.Summarize(spec)
	require.NoError(t, err)
	spec.Runtime.ModelBinding.ProtocolAdapter = ""
	summary.ModelBinding.ProtocolAdapter = ""
	specJSON, err := json.Marshal(spec)
	require.NoError(t, err)
	canonicalSpec, err := control.CanonicalJSONObject(specJSON)
	require.NoError(t, err)
	specDigest, err := control.DigestCanonicalJSON(canonicalSpec)
	require.NoError(t, err)
	summaryJSON, err := json.Marshal(summary)
	require.NoError(t, err)
	require.NoError(t, db.Exec(`
INSERT INTO worker_spec_snapshots
	(id, organization_id, version, spec_json, summary_json)
VALUES (?, ?, ?, ?, ?)`,
		91, 7, specdomain.VersionV1, specJSON, summaryJSON,
	).Error)
	snapshotID := int64(91)
	preparer := &snapshotWorkerCreationPreparer{
		prepared: workercreation.PreparedSnapshot{
			Spec:           podServiceWorkerSpec(),
			AgentfileLayer: "MODE acp\n",
		},
	}
	orchestrator := NewPodOrchestrator(&PodOrchestratorDeps{
		WorkerCreation:     preparer,
		WorkerSpecs:        infra.NewWorkerSpecSnapshotRepository(db),
		WorkerDependencies: snapshotDependencyLoaderWithDigest(7, specDigest),
	})

	err = orchestrator.prepareSnapshotWorkerCreate(
		context.Background(),
		&OrchestrateCreatePodRequest{
			OrganizationID:       7,
			UserID:               5,
			WorkerSpecSnapshotID: &snapshotID,
		},
	)

	require.ErrorIs(t, err, ErrWorkerSpecSnapshotMismatch)
	assert.Zero(t, preparer.snapshotCalls)
}
