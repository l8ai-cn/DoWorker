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

func TestFindAccessibleByOrgSlugSkipsInaccessibleMatches(t *testing.T) {
	ownerID := int64(41)

	for _, privateFirst := range []bool{true, false} {
		name := "organization repository inserted first"
		if privateFirst {
			name = "private repository inserted first"
		}
		t.Run(name, func(t *testing.T) {
			service, db := setupTestService(t)
			privateRepo := newOrgSlugAccessTestRepository("github", "private", int64Pointer(42))
			organizationRepo := newOrgSlugAccessTestRepository("gitlab", "organization", nil)
			repos := []*gitprovider.Repository{organizationRepo, privateRepo}
			if privateFirst {
				repos = []*gitprovider.Repository{privateRepo, organizationRepo}
			}
			for _, repo := range repos {
				require.NoError(t, db.Create(repo).Error)
			}

			got, err := service.FindAccessibleByOrgSlug(context.Background(), 7, ownerID, "access/test")
			require.NoError(t, err)
			require.Equal(t, organizationRepo.ID, got.ID)
		})
	}
}

func TestFindAccessibleByOrgSlugRejectsAmbiguousMatches(t *testing.T) {
	service, db := setupTestService(t)
	ownerID := int64(41)
	organizationRepo := newOrgSlugAccessTestRepository("github", "organization", nil)
	privateRepo := newOrgSlugAccessTestRepository("gitlab", "private", &ownerID)
	require.NoError(t, db.Create(organizationRepo).Error)
	require.NoError(t, db.Create(privateRepo).Error)

	got, err := service.FindAccessibleByOrgSlug(context.Background(), 7, ownerID, "access/test")
	require.ErrorIs(t, err, ErrAmbiguousRepositorySlug)
	require.EqualError(t, err, "repository slug is ambiguous")
	require.Nil(t, got)
}

func newOrgSlugAccessTestRepository(providerType, visibility string, importedByUserID *int64) *gitprovider.Repository {
	return &gitprovider.Repository{
		OrganizationID:   7,
		ProviderType:     providerType,
		ProviderBaseURL:  "https://" + providerType + ".com",
		ExternalID:       "access-test-" + providerType,
		Name:             "access-test-" + providerType,
		Slug:             "access/test",
		Visibility:       visibility,
		ImportedByUserID: importedByUserID,
		IsActive:         true,
	}
}

func int64Pointer(value int64) *int64 {
	return &value
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
