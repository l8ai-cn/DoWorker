package infra

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/agentpod"
	"github.com/l8ai-cn/agentcloud/backend/internal/domain/airesource"
	resource "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationresource"
	"github.com/l8ai-cn/agentcloud/backend/internal/domain/workerdependency"
	"github.com/l8ai-cn/agentcloud/backend/internal/domain/workerspec"
	"github.com/l8ai-cn/agentcloud/backend/pkg/slugkit"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPodRepositoryCreatesWorkerSpecSnapshotAtomically(t *testing.T) {
	db := workerSpecSnapshotDBForContract(t)
	repo := &podRepo{db: db}
	pod := workerSpecPodForRepoTest("7-standalone-aabbccdd")
	revision := workerSpecRevisionForRepoTest()
	artifactJSON, artifactDigest := workerSpecArtifactForRepoTest(t, 77)

	err := repo.CreateWithConfigAndWorkerSpec(
		context.Background(),
		pod,
		revision,
		workerSpecSnapshotForContract(t, 77),
		artifactJSON,
		artifactDigest,
	)

	require.NoError(t, err)
	require.NotNil(t, pod.WorkerSpecSnapshotID)
	assert.Positive(t, *pod.WorkerSpecSnapshotID)
	assert.Positive(t, pod.ID)
	assert.Positive(t, revision.ID)
	assert.Equal(t, pod.ID, revision.PodID)
	require.NotNil(t, pod.ActiveConfigRevisionID)
	assert.Equal(t, revision.ID, *pod.ActiveConfigRevisionID)

	var snapshotCount int64
	require.NoError(t, db.Table("worker_spec_snapshots").Count(&snapshotCount).Error)
	assert.Equal(t, int64(1), snapshotCount)
	var stored agentpod.Pod
	require.NoError(t, db.First(&stored, pod.ID).Error)
	assert.Equal(t, pod.WorkerSpecSnapshotID, stored.WorkerSpecSnapshotID)
	assert.Equal(t, int64(700), stored.ClusterID)
}

func TestPodRepositoryRollsBackSnapshotWhenConfigRevisionFails(t *testing.T) {
	db := workerSpecSnapshotDBForContract(t)
	require.NoError(t, db.Exec(`
CREATE TRIGGER reject_worker_spec_revision
BEFORE INSERT ON pod_config_revisions
BEGIN
	SELECT RAISE(ABORT, 'forced revision failure');
END
`).Error)
	repo := &podRepo{db: db}
	artifactJSON, artifactDigest := workerSpecArtifactForRepoTest(t, 77)

	err := repo.CreateWithConfigAndWorkerSpec(
		context.Background(),
		workerSpecPodForRepoTest("7-standalone-eeff0011"),
		workerSpecRevisionForRepoTest(),
		workerSpecSnapshotForContract(t, 77),
		artifactJSON,
		artifactDigest,
	)

	require.Error(t, err)
	for _, table := range []string{
		"worker_spec_snapshots",
		"worker_spec_dependency_artifacts",
		"pods",
		"pod_config_revisions",
	} {
		var count int64
		require.NoError(t, db.Table(table).Count(&count).Error)
		assert.Zero(t, count, table)
	}
}

func workerSpecPodForRepoTest(key string) *agentpod.Pod {
	return &agentpod.Pod{
		OrganizationID:  77,
		PodKey:          key,
		RunnerID:        1,
		AgentSlug:       "codex-cli",
		CreatedByID:     7,
		Status:          agentpod.StatusInitializing,
		AgentStatus:     agentpod.AgentStatusIdle,
		InteractionMode: agentpod.InteractionModeACP,
		AutomationLevel: agentpod.AutomationLevelAutonomous,
	}
}

func workerSpecRevisionForRepoTest() *agentpod.PodConfigRevision {
	return &agentpod.PodConfigRevision{
		Revision:       1,
		AgentfileLayer: "MODE acp\n",
		Status:         agentpod.ConfigRevisionStatusActive,
		ConfigSummary:  json.RawMessage(`{}`),
		CreatedByID:    7,
	}
}

func workerSpecArtifactForRepoTest(t *testing.T, organizationID int64) ([]byte, string) {
	t.Helper()
	spec := workerSpecForRepoContract()
	document := workerdependency.Document{
		Version:        workerdependency.VersionV1,
		OrganizationID: organizationID,
		Namespace:      slugkit.MustNewForTest("dev-org"),
		Worker: workerdependency.Worker{
			WorkerType:      spec.Runtime.WorkerType.Slug,
			AdapterID:       slugkit.MustNewForTest("codex-cli"),
			SpecVersion:     workerspec.VersionV1,
			SpecDigest:      workerdependency.TextDigest("workerspec"),
			DefinitionHash:  strings.Repeat("a", 64),
			AgentfileSource: "AGENT codex\nMODE pty\n",
		},
		Models: workerdependency.Models{Primary: &workerdependency.Model{
			Pin:                workerSpecArtifactPin(resource.KindModelBinding, "coding-model", 1001),
			ResourceRevision:   7,
			ConnectionID:       2001,
			ConnectionRevision: 9,
			ProviderKey:        slugkit.MustNewForTest("openai"),
			ProtocolAdapter:    slugkit.MustNewForTest("openai-compatible"),
			ModelID:            "gpt-5",
			BaseURL:            "https://api.example.com/v1",
			Modalities:         []airesource.Modality{airesource.ModalityChat},
			Capabilities:       []airesource.Capability{airesource.CapabilityTextGeneration},
		}},
		RuntimeBundles:   []workerdependency.RuntimeBundle{},
		SecretReferences: []workerdependency.SecretReference{},
		Placement: workerdependency.Placement{
			CatalogRevision: "catalog-v1",
			RuntimeImage: workerdependency.RuntimeImage{
				ID:        spec.Runtime.Image.ID,
				Reference: "registry.example.com/worker@" + spec.Runtime.Image.Digest,
				Digest:    spec.Runtime.Image.Digest,
			},
			ComputeTarget: workerSpecArtifactPin(resource.KindComputeTarget, "runner-pool", 52),
			ResourceProfile: &workerdependency.ResourcePin{
				Reference: workerSpecArtifactPin(
					resource.KindResourceProfile,
					"standard",
					63,
				).Reference,
				DomainID: 63,
			},
			Spec: spec.Placement,
		},
	}
	document.Worker.AgentfileSourceDigest = workerdependency.TextDigest(
		document.Worker.AgentfileSource,
	)
	normalized, err := workerdependency.NormalizeAndValidate(document)
	require.NoError(t, err)
	artifactJSON, digest, err := workerdependency.EncodeAndDigest(normalized)
	require.NoError(t, err)
	return artifactJSON, digest
}

func workerSpecArtifactPin(
	kind string,
	name string,
	domainID int64,
) workerdependency.ResourcePin {
	identity := kind + "\x00" + name
	return workerdependency.ResourcePin{
		DomainID: domainID,
		Reference: resource.Reference{
			APIVersion: resource.APIVersionV1Alpha1,
			Kind:       kind,
			Namespace:  slugkit.MustNewForTest("dev-org"),
			Name:       slugkit.MustNewForTest(name),
			UID:        uuid.NewSHA1(uuid.NameSpaceOID, []byte(identity)).String(),
			Revision:   1,
			Digest:     workerdependency.TextDigest(identity),
		},
	}
}
