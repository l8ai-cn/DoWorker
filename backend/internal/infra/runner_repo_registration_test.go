package infra

import (
	"context"
	"testing"
	"time"

	"github.com/anthropics/agentsmesh/backend/internal/domain/runner"
	"github.com/anthropics/agentsmesh/backend/internal/testkit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunnerRegistrationRepositoryAuthorizesPendingAuthWithRunner(t *testing.T) {
	db := testkit.SetupTestDB(t)
	repository := NewRunnerRepository(db)
	pendingAuth := &runner.PendingAuth{
		AuthKey:    "claim-pending-auth",
		MachineKey: "local-mac",
		ExpiresAt:  time.Now().Add(time.Minute),
	}
	require.NoError(t, db.Create(pendingAuth).Error)

	createdRunner := &runner.Runner{
		OrganizationID:    17,
		ClusterID:         29,
		NodeID:            "local-mac",
		Status:            runner.RunnerStatusOffline,
		MaxConcurrentPods: 5,
		Visibility:        runner.VisibilityOrganization,
	}
	rowsAffected, err := repository.AuthorizePendingAuthAtomic(
		context.Background(),
		pendingAuth.ID,
		17,
		29,
		createdRunner,
	)
	require.NoError(t, err)
	assert.Equal(t, int64(1), rowsAffected)
	assert.NotZero(t, createdRunner.ID)

	var claimed runner.PendingAuth
	require.NoError(t, db.First(&claimed, pendingAuth.ID).Error)
	assert.True(t, claimed.Authorized)
	require.NotNil(t, claimed.OrganizationID)
	assert.Equal(t, int64(17), *claimed.OrganizationID)
	require.NotNil(t, claimed.ClusterID)
	assert.Equal(t, int64(29), *claimed.ClusterID)
	require.NotNil(t, claimed.RunnerID)
	assert.Equal(t, createdRunner.ID, *claimed.RunnerID)
}

func TestRunnerRegistrationRepositoryRollsBackPendingAuthWhenRunnerCreateFails(t *testing.T) {
	db := testkit.SetupTestDB(t)
	repository := NewRunnerRepository(db)
	require.NoError(t, db.Exec(`
		CREATE UNIQUE INDEX runners_organization_node_unique
		ON runners (organization_id, node_id)
	`).Error)
	require.NoError(t, db.Create(&runner.Runner{
		OrganizationID:    17,
		ClusterID:         29,
		NodeID:            "occupied-node",
		Status:            runner.RunnerStatusOffline,
		MaxConcurrentPods: 5,
		Visibility:        runner.VisibilityOrganization,
	}).Error)
	pendingAuth := &runner.PendingAuth{
		AuthKey:    "rollback-pending-auth",
		MachineKey: "local-mac",
		ExpiresAt:  time.Now().Add(time.Minute),
	}
	require.NoError(t, db.Create(pendingAuth).Error)

	rowsAffected, err := repository.AuthorizePendingAuthAtomic(
		context.Background(),
		pendingAuth.ID,
		17,
		29,
		&runner.Runner{
			OrganizationID:    17,
			ClusterID:         29,
			NodeID:            "occupied-node",
			Status:            runner.RunnerStatusOffline,
			MaxConcurrentPods: 5,
			Visibility:        runner.VisibilityOrganization,
		},
	)

	require.Error(t, err)
	assert.Equal(t, int64(0), rowsAffected)
	var stored runner.PendingAuth
	require.NoError(t, db.First(&stored, pendingAuth.ID).Error)
	assert.False(t, stored.Authorized)
	assert.Nil(t, stored.OrganizationID)
	assert.Nil(t, stored.ClusterID)
	assert.Nil(t, stored.RunnerID)
}
