package workercreation

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	specdomain "github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	specservice "github.com/anthropics/agentsmesh/backend/internal/service/workerspec"
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

func TestPrepareSnapshotFailsWhenAReferencedResourceIsUnavailable(t *testing.T) {
	fixture := newWorkerCreationServiceFixture()
	service := NewService(fixture.deps())
	scope := specservice.Scope{OrgID: 77, UserID: 7}
	source, err := service.Prepare(
		context.Background(),
		scope,
		validWorkerCreationDraft(),
	)
	require.NoError(t, err)
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
