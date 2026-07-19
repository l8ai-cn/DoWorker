package coordinator

import (
	"context"
	"errors"
	"testing"

	coordinatordom "github.com/anthropics/agentsmesh/backend/internal/domain/coordinator"
	"github.com/anthropics/agentsmesh/backend/internal/domain/workerdependency"
	workerspecdom "github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateProjectDerivesWorkerTypeFromSnapshotArtifact(t *testing.T) {
	snapshotID := int64(101)
	store := newFakeStore()
	service := NewService(Deps{
		Store:     store,
		Snapshots: projectSnapshots{snapshotID: projectSnapshot(1, snapshotID, "codex-cli")},
		Artifacts: projectArtifacts{snapshotID: projectArtifact("codex-cli")},
	})

	project, err := service.CreateProject(context.Background(), &CreateProjectRequest{
		OrganizationID: 1, RepositoryID: 9, Name: "Release queue",
		WorkerSpecSnapshotID: &snapshotID, CreatedByID: 7,
	})

	require.NoError(t, err)
	assert.Equal(t, "codex-cli", project.AgentSlug)
	require.Len(t, store.projects, 1)
	assert.Equal(t, "codex-cli", store.projects[0].AgentSlug)
}

func TestCreateProjectBlocksMissingOrMismatchedArtifact(t *testing.T) {
	snapshotID := int64(101)
	snapshot := projectSnapshot(1, snapshotID, "codex-cli")
	cases := []struct {
		name      string
		snapshots projectSnapshots
		artifacts projectArtifacts
		want      error
	}{
		{
			name:      "missing artifact",
			snapshots: projectSnapshots{snapshotID: snapshot},
			artifacts: projectArtifacts{},
			want:      ErrCoordinatorWorkerSpecArtifactRequired,
		},
		{
			name:      "cross organization snapshot",
			snapshots: projectSnapshots{snapshotID: projectSnapshot(2, snapshotID, "codex-cli")},
			artifacts: projectArtifacts{snapshotID: projectArtifact("codex-cli")},
			want:      ErrCoordinatorWorkerSpecSnapshotMismatch,
		},
		{
			name:      "artifact worker mismatch",
			snapshots: projectSnapshots{snapshotID: snapshot},
			artifacts: projectArtifacts{snapshotID: projectArtifact("gemini-cli")},
			want:      ErrCoordinatorWorkerSpecSnapshotMismatch,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			service := NewService(Deps{
				Store: newFakeStore(), Snapshots: tc.snapshots, Artifacts: tc.artifacts,
			})
			_, err := service.CreateProject(context.Background(), &CreateProjectRequest{
				OrganizationID: 1, RepositoryID: 9, Name: "Release queue",
				WorkerSpecSnapshotID: &snapshotID, CreatedByID: 7,
			})
			require.ErrorIs(t, err, tc.want)
		})
	}
}

func TestUpdateProjectRebindsSnapshotAndRejectsIndependentAgent(t *testing.T) {
	oldSnapshotID := int64(101)
	newSnapshotID := int64(102)
	store := newFakeStore()
	store.projects = []*coordinatordom.Project{{
		ID: 7, OrganizationID: 1, AgentSlug: "codex-cli", WorkerSpecSnapshotID: &oldSnapshotID,
	}}
	service := NewService(Deps{
		Store: store,
		Snapshots: projectSnapshots{
			oldSnapshotID: projectSnapshot(1, oldSnapshotID, "codex-cli"),
			newSnapshotID: projectSnapshot(1, newSnapshotID, "gemini-cli"),
		},
		Artifacts: projectArtifacts{
			oldSnapshotID: projectArtifact("codex-cli"),
			newSnapshotID: projectArtifact("gemini-cli"),
		},
	})

	require.NoError(t, service.UpdateProject(
		context.Background(),
		1,
		7,
		map[string]any{"worker_spec_snapshot_id": newSnapshotID},
	))
	assert.Equal(t, "gemini-cli", store.projects[0].AgentSlug)
	require.NotNil(t, store.projects[0].WorkerSpecSnapshotID)
	assert.Equal(t, newSnapshotID, *store.projects[0].WorkerSpecSnapshotID)

	err := service.UpdateProject(
		context.Background(),
		1,
		7,
		map[string]any{"agent_slug": "do-agent"},
	)
	require.ErrorIs(t, err, ErrCoordinatorAgentSlugDerived)
}

type projectSnapshots map[int64]workerspecdom.Snapshot

func (snapshots projectSnapshots) GetByID(
	_ context.Context,
	_ int64,
	id int64,
) (workerspecdom.Snapshot, error) {
	snapshot, found := snapshots[id]
	if !found {
		return workerspecdom.Snapshot{}, workerspecdom.ErrNotFound
	}
	return snapshot, nil
}

type projectArtifacts map[int64]workerdependency.Document

func (artifacts projectArtifacts) GetBySnapshotID(
	_ context.Context,
	_ int64,
	id int64,
) (workerdependency.Document, error) {
	artifact, found := artifacts[id]
	if !found {
		return workerdependency.Document{}, errors.New("artifact not found")
	}
	return artifact, nil
}

func projectSnapshot(orgID, id int64, workerType string) workerspecdom.Snapshot {
	return workerspecdom.Snapshot{
		ID: id, OrganizationID: orgID,
		Spec: workerspecdom.Spec{
			Runtime: workerspecdom.Runtime{
				WorkerType: workerspecdom.WorkerType{
					Slug: slugkit.MustNewForTest(workerType),
				},
			},
		},
	}
}

func projectArtifact(workerType string) workerdependency.Document {
	return workerdependency.Document{
		Worker: workerdependency.Worker{
			WorkerType: slugkit.MustNewForTest(workerType),
		},
	}
}
