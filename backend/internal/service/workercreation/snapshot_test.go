package workercreation

import (
	"context"
	"testing"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/workerdependency"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	specdomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/workerspec"
	specservice "github.com/l8ai-cn/agentcloud/backend/internal/service/workerspec"
)

func TestPrepareSnapshotCompilesThePersistedSpec(t *testing.T) {
	fixture := newWorkerCreationServiceFixture()
	service := NewService(fixture.deps())
	scope := specservice.Scope{OrgID: 77, UserID: 7}
	source, err := service.Prepare(
		context.Background(),
		scope,
		validWorkerCreationDraft(),
	)
	require.NoError(t, err)

	replayed, err := service.PrepareSnapshot(
		context.Background(),
		scope,
		specdomain.Snapshot{
			ID:             91,
			OrganizationID: scope.OrgID,
			Spec:           source.Spec,
		},
	)

	require.NoError(t, err)
	assert.Equal(t, source.Spec, replayed.Spec)
	assert.Equal(t, source.AgentfileLayer, replayed.AgentfileLayer)
	require.NotNil(t, replayed.Repository)
	assert.Equal(t, *source.Spec.Workspace.RepositoryID, replayed.Repository.ID)
}

func TestPrepareSnapshotWithDependenciesReplaysFreshArtifact(t *testing.T) {
	fixture := newWorkerCreationServiceFixture()
	service := NewService(fixture.deps())
	scope := specservice.Scope{OrgID: 77, UserID: 7}
	source, err := service.Prepare(
		context.Background(),
		scope,
		validWorkerCreationDraft(),
	)
	require.NoError(t, err)
	require.NotNil(t, source.Artifact)

	document, err := workerdependency.Decode(source.Artifact.JSON())
	require.NoError(t, err)
	replayed, err := service.PrepareSnapshotWithDependencies(
		context.Background(),
		scope,
		specdomain.Snapshot{
			ID:             91,
			OrganizationID: scope.OrgID,
			Spec:           source.Spec,
		},
		document,
	)

	require.NoError(t, err)
	assert.Equal(t, source.Spec, replayed.Spec)
	assert.Contains(t, replayed.AgentfileLayer, `REPO "repository-22"`)
	assert.NotContains(t, replayed.AgentfileLayer, `REPO "org/repo"`)
	require.NotNil(t, replayed.Dependencies)
	assert.Equal(t, source.Artifact.Digest(), workerdependencyMustDigest(t, *replayed.Dependencies))
}

func TestPrepareSnapshotUsesPinnedSkillPackageAfterCatalogChanges(t *testing.T) {
	fixture := newWorkerCreationServiceFixture()
	service := NewService(fixture.deps())
	scope := specservice.Scope{OrgID: 77, UserID: 7}
	source, err := service.Prepare(
		context.Background(),
		scope,
		validWorkerCreationDraft(),
	)
	require.NoError(t, err)
	skill := fixture.workspace.skills.rows[source.Spec.Workspace.SkillIDs[0]]
	skill.IsActive = false
	skill.ContentSha = "changed-sha"
	skill.StorageKey = "skills/changed.tar.gz"

	replayed, err := service.PrepareSnapshot(
		context.Background(),
		scope,
		specdomain.Snapshot{
			ID:             91,
			OrganizationID: scope.OrgID,
			Spec:           source.Spec,
		},
	)

	require.NoError(t, err)
	assert.Equal(t, source.Spec.Workspace.SkillPackages, replayed.Spec.Workspace.SkillPackages)
}

func workerdependencyMustDigest(
	t *testing.T,
	document workerdependency.Document,
) string {
	t.Helper()
	digest, err := workerdependency.Digest(document)
	require.NoError(t, err)
	return digest
}

func TestPrepareLegacySnapshotFailsWhenAReferencedSkillIsUnavailable(t *testing.T) {
	fixture := newWorkerCreationServiceFixture()
	service := NewService(fixture.deps())
	scope := specservice.Scope{OrgID: 77, UserID: 7}
	source, err := service.Prepare(
		context.Background(),
		scope,
		validWorkerCreationDraft(),
	)
	require.NoError(t, err)
	source.Spec.Workspace.SkillPackages = nil
	fixture.workspace.skills.rows[source.Spec.Workspace.SkillIDs[0]].IsActive = false

	_, err = service.PrepareSnapshot(
		context.Background(),
		scope,
		specdomain.Snapshot{
			ID:             91,
			OrganizationID: scope.OrgID,
			Spec:           source.Spec,
		},
	)

	require.Error(t, err)
	assert.ErrorIs(t, err, specservice.ErrInvalidDraft)
}
