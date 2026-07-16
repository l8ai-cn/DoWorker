package runner

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/anthropics/agentsmesh/backend/internal/domain/executioncluster"
	runnerdomain "github.com/anthropics/agentsmesh/backend/internal/domain/runner"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegisterWithTokenUsesPersistedClusterAndPreservesLabels(t *testing.T) {
	db := setupTestDB(t)
	service := newTestService(db)
	ctx := context.Background()
	pkiService, tmpDir := setupTestPKI(t)
	defer os.RemoveAll(tmpDir)

	org := createTestOrg(t, db, "token-cluster-org")
	clusterID := createTestExecutionCluster(t, db, org.ID, "local")
	token := generateTestAuthKey()
	registrationToken := &runnerdomain.GRPCRegistrationToken{
		TokenHash:      hashToken(token),
		OrganizationID: org.ID,
		ClusterID:      clusterID,
		Labels:         runnerdomain.Labels{"cluster": "online", "region": "local", "runtime": "macos"},
		MaxUses:        1,
		ExpiresAt:      time.Now().Add(time.Hour),
	}
	require.NoError(t, db.Create(registrationToken).Error)

	response, err := service.RegisterWithToken(ctx, &RegisterWithTokenRequest{
		Token:  token,
		NodeID: "local-mac",
	}, pkiService)
	require.NoError(t, err)

	var registered runnerdomain.Runner
	require.NoError(t, db.First(&registered, response.RunnerID).Error)
	assert.Equal(t, clusterID, registered.ClusterID)
	assert.ElementsMatch(t, []string{"cluster=online", "region=local", "runtime=macos"}, []string(registered.Tags))
}

func TestGenerateGRPCRegistrationTokenRejectsClusterOutsideOrganization(t *testing.T) {
	db := setupTestDB(t)
	service := newTestService(db)
	service.SetExecutionClusterRepository(registrationClusterRepository{
		clusters: map[int64]*executioncluster.Cluster{
			92: {ID: 92, OrganizationID: 200, Slug: "local", Name: "本地集群", Kind: executioncluster.KindLocal, Status: executioncluster.StatusReady},
		},
	})
	org := createTestOrg(t, db, "token-owner-org")

	_, err := service.GenerateGRPCRegistrationToken(
		context.Background(),
		org.ID,
		1,
		&GenerateGRPCRegistrationTokenRequest{ClusterID: 92},
		"https://example.com",
	)

	require.ErrorIs(t, err, ErrExecutionClusterNotFound)
}

func TestAuthorizeRunnerBindsSelectedClusterWithOrganization(t *testing.T) {
	db := setupTestDB(t)
	service := newTestService(db)
	ctx := context.Background()
	org := createTestOrg(t, db, "interactive-cluster-org")
	service.SetExecutionClusterRepository(registrationClusterRepository{
		clusters: map[int64]*executioncluster.Cluster{
			93: {ID: 93, OrganizationID: org.ID, Slug: "local", Name: "本地集群", Kind: executioncluster.KindLocal, Status: executioncluster.StatusReady},
		},
	})
	pendingAuth := &runnerdomain.PendingAuth{
		AuthKey:    generateTestAuthKey(),
		MachineKey: "interactive-local-mac",
		Labels:     runnerdomain.Labels{"machine": "macos"},
		ExpiresAt:  time.Now().Add(15 * time.Minute),
	}
	require.NoError(t, db.Create(pendingAuth).Error)

	registered, err := service.AuthorizeRunner(ctx, pendingAuth.AuthKey, org.ID, 1, 93, "interactive-local")
	require.NoError(t, err)
	assert.Equal(t, int64(93), registered.ClusterID)
	assert.Equal(t, []string{"machine=macos"}, []string(registered.Tags))

	var claimed runnerdomain.PendingAuth
	require.NoError(t, db.First(&claimed, pendingAuth.ID).Error)
	require.NotNil(t, claimed.ClusterID)
	assert.Equal(t, int64(93), *claimed.ClusterID)
}

func TestAuthorizeRunnerRejectsClusterOutsideOrganization(t *testing.T) {
	db := setupTestDB(t)
	service := newTestService(db)
	ctx := context.Background()
	org := createTestOrg(t, db, "interactive-other-org")
	service.SetExecutionClusterRepository(registrationClusterRepository{
		clusters: map[int64]*executioncluster.Cluster{
			94: {ID: 94, OrganizationID: org.ID + 1, Slug: "local", Name: "其他组织集群", Kind: executioncluster.KindLocal, Status: executioncluster.StatusReady},
		},
	})
	pendingAuth := &runnerdomain.PendingAuth{
		AuthKey:    generateTestAuthKey(),
		MachineKey: "interactive-other-machine",
		ExpiresAt:  time.Now().Add(15 * time.Minute),
	}
	require.NoError(t, db.Create(pendingAuth).Error)

	_, err := service.AuthorizeRunner(ctx, pendingAuth.AuthKey, org.ID, 1, 94, "other-org-runner")

	require.ErrorIs(t, err, ErrExecutionClusterNotFound)
	var claimed runnerdomain.PendingAuth
	require.NoError(t, db.First(&claimed, pendingAuth.ID).Error)
	assert.False(t, claimed.Authorized)
	assert.Nil(t, claimed.ClusterID)
}

type registrationClusterRepository struct {
	clusters map[int64]*executioncluster.Cluster
}

func (r registrationClusterRepository) ListByOrganization(_ context.Context, organizationID int64) ([]*executioncluster.Cluster, error) {
	var clusters []*executioncluster.Cluster
	for _, cluster := range r.clusters {
		if cluster.OrganizationID == organizationID {
			clusters = append(clusters, cluster)
		}
	}
	return clusters, nil
}

func (r registrationClusterRepository) GetByIDAndOrganization(_ context.Context, id, organizationID int64) (*executioncluster.Cluster, error) {
	cluster := r.clusters[id]
	if cluster == nil || cluster.OrganizationID != organizationID {
		return nil, nil
	}
	return cluster, nil
}

func (r registrationClusterRepository) EnsureDefaults(ctx context.Context, organizationID int64) ([]*executioncluster.Cluster, error) {
	return r.ListByOrganization(ctx, organizationID)
}
