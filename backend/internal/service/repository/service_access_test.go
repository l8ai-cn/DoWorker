package repository

import (
	"context"
	"testing"

	"github.com/anthropics/agentsmesh/backend/internal/domain/gitprovider"
	"github.com/stretchr/testify/require"
)

type accessibleRepositoryResolver func(
	context.Context,
	*Service,
	*gitprovider.Repository,
	int64,
	int64,
) (*gitprovider.Repository, error)

func TestGetAccessibleByID(t *testing.T) {
	testAccessibleRepositoryResolver(t, func(
		ctx context.Context,
		service *Service,
		repo *gitprovider.Repository,
		orgID int64,
		userID int64,
	) (*gitprovider.Repository, error) {
		return service.GetAccessibleByID(ctx, repo.ID, orgID, userID)
	})
}

func TestFindAccessibleByOrgSlug(t *testing.T) {
	testAccessibleRepositoryResolver(t, func(
		ctx context.Context,
		service *Service,
		repo *gitprovider.Repository,
		orgID int64,
		userID int64,
	) (*gitprovider.Repository, error) {
		return service.FindAccessibleByOrgSlug(ctx, orgID, userID, repo.Slug)
	})
}

func testAccessibleRepositoryResolver(t *testing.T, resolve accessibleRepositoryResolver) {
	t.Helper()

	ownerID := int64(41)
	otherUserID := int64(42)

	tests := []struct {
		name             string
		repositoryOrgID  int64
		lookupOrgID      int64
		visibility       string
		importedByUserID *int64
		persist          bool
		wantPermission   bool
	}{
		{
			name:            "organization repository in requested organization",
			repositoryOrgID: 7,
			lookupOrgID:     7,
			visibility:      "organization",
			persist:         true,
			wantPermission:  true,
		},
		{
			name:             "private repository imported by caller",
			repositoryOrgID:  7,
			lookupOrgID:      7,
			visibility:       "private",
			importedByUserID: &ownerID,
			persist:          true,
			wantPermission:   true,
		},
		{
			name:             "private repository imported by another user",
			repositoryOrgID:  7,
			lookupOrgID:      7,
			visibility:       "private",
			importedByUserID: &otherUserID,
			persist:          true,
		},
		{
			name:            "private repository without importer",
			repositoryOrgID: 7,
			lookupOrgID:     7,
			visibility:      "private",
			persist:         true,
		},
		{
			name:            "organization repository in another organization",
			repositoryOrgID: 8,
			lookupOrgID:     7,
			visibility:      "organization",
			persist:         true,
		},
		{
			name:             "private repository imported by caller in another organization",
			repositoryOrgID:  8,
			lookupOrgID:      7,
			visibility:       "private",
			importedByUserID: &ownerID,
			persist:          true,
		},
		{
			name:            "nonexistent repository",
			repositoryOrgID: 7,
			lookupOrgID:     7,
			visibility:      "organization",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, db := setupTestService(t)
			repo := &gitprovider.Repository{
				ID:               999999,
				OrganizationID:   tt.repositoryOrgID,
				ProviderType:     "github",
				ProviderBaseURL:  "https://github.com",
				ExternalID:       "access-test",
				Name:             "access-test",
				Slug:             "access/test",
				Visibility:       tt.visibility,
				ImportedByUserID: tt.importedByUserID,
				IsActive:         true,
			}
			if tt.persist {
				repo.ID = 0
				require.NoError(t, db.Create(repo).Error)
			}

			got, err := resolve(context.Background(), service, repo, tt.lookupOrgID, ownerID)
			if tt.wantPermission {
				require.NoError(t, err)
				require.NotNil(t, got)
				require.Equal(t, repo.ID, got.ID)
				return
			}

			require.ErrorIs(t, err, ErrNoPermission)
			require.Nil(t, got)
		})
	}

	t.Run("database error is propagated", func(t *testing.T) {
		service, db := setupTestService(t)
		sqlDB, err := db.DB()
		require.NoError(t, err)
		require.NoError(t, sqlDB.Close())

		got, err := resolve(context.Background(), service, &gitprovider.Repository{
			ID:   1,
			Slug: "access/test",
		}, 7, ownerID)
		require.Error(t, err)
		require.NotErrorIs(t, err, ErrNoPermission)
		require.Contains(t, err.Error(), "database is closed")
		require.Nil(t, got)
	})
}
