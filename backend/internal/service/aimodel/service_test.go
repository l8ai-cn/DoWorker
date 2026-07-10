package aimodel

import (
	"context"
	"errors"
	"testing"

	domain "github.com/anthropics/agentsmesh/backend/internal/domain/aimodel"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeRepository struct {
	models     map[int64]*domain.AIModel
	visibleErr error
}

func (r *fakeRepository) GetByID(_ context.Context, id int64) (*domain.AIModel, error) {
	return r.models[id], nil
}

func (r *fakeRepository) GetVisibleByID(_ context.Context, id, userID, orgID int64) (*domain.AIModel, error) {
	if r.visibleErr != nil {
		return nil, r.visibleErr
	}
	model := r.models[id]
	if model == nil || !model.IsEnabled {
		return nil, nil
	}
	if model.OrganizationID != nil && *model.OrganizationID == orgID {
		return model, nil
	}
	if model.UserID != nil && *model.UserID == userID {
		return model, nil
	}
	return nil, nil
}

func (r *fakeRepository) Create(context.Context, *domain.AIModel) error { return nil }
func (r *fakeRepository) Save(context.Context, *domain.AIModel) error   { return nil }
func (r *fakeRepository) Delete(context.Context, int64) error           { return nil }
func (r *fakeRepository) ListVisible(context.Context, int64, int64) ([]*domain.AIModel, error) {
	return nil, nil
}
func (r *fakeRepository) DefaultVisible(context.Context, int64, int64) (*domain.AIModel, error) {
	return nil, nil
}
func (r *fakeRepository) ClearDefaults(context.Context, int64, int64) error { return nil }
func (r *fakeRepository) CountOrg(context.Context, int64) (int64, error)    { return 0, nil }
func (r *fakeRepository) FirstVisibleByProvider(context.Context, int64, int64, string) (*domain.AIModel, error) {
	return nil, nil
}

func TestResolveVisible(t *testing.T) {
	const (
		userID      = int64(11)
		orgID       = int64(21)
		otherUserID = int64(12)
		otherOrgID  = int64(22)
	)
	models := map[int64]*domain.AIModel{
		1: {ID: 1, OrganizationID: pointer(orgID), IsEnabled: true, EncryptedCredentials: `{"api_key":"org"}`},
		2: {ID: 2, UserID: pointer(userID), IsEnabled: true, EncryptedCredentials: `{"api_key":"user"}`},
		3: {ID: 3, OrganizationID: pointer(otherOrgID), IsEnabled: true, EncryptedCredentials: "invalid-json"},
		4: {ID: 4, UserID: pointer(otherUserID), IsEnabled: true, EncryptedCredentials: "invalid-json"},
		5: {ID: 5, OrganizationID: pointer(orgID), IsEnabled: false, EncryptedCredentials: "invalid-json"},
		6: {ID: 6, OrganizationID: pointer(orgID), IsEnabled: true, EncryptedCredentials: "invalid-json"},
	}
	service := NewService(&fakeRepository{models: models}, nil)

	tests := []struct {
		name string
		id   int64
		want string
	}{
		{name: "same-org shared", id: 1, want: "org"},
		{name: "current-user private", id: 2, want: "user"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resolved, err := service.ResolveVisible(context.Background(), test.id, userID, orgID)
			require.NoError(t, err)
			assert.Equal(t, test.id, resolved.Model.ID)
			assert.Equal(t, test.want, resolved.Credentials["api_key"])
		})
	}

	for _, test := range []struct {
		name string
		id   int64
	}{
		{name: "other-org", id: 3},
		{name: "other-user", id: 4},
		{name: "disabled", id: 5},
		{name: "missing", id: 99},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, err := service.ResolveVisible(context.Background(), test.id, userID, orgID)
			assert.ErrorIs(t, err, ErrNotFound)
		})
	}

	t.Run("visible invalid credentials", func(t *testing.T) {
		_, err := service.ResolveVisible(context.Background(), 6, userID, orgID)
		require.Error(t, err)
		assert.NotErrorIs(t, err, ErrNotFound)
	})

	t.Run("repository error", func(t *testing.T) {
		repoErr := errors.New("query failed")
		service := NewService(&fakeRepository{visibleErr: repoErr}, nil)
		_, err := service.ResolveVisible(context.Background(), 1, userID, orgID)
		assert.ErrorIs(t, err, repoErr)
	})
}

func TestGetVisibleDoesNotDecrypt(t *testing.T) {
	const (
		userID = int64(11)
		orgID  = int64(21)
	)
	repo := &fakeRepository{models: map[int64]*domain.AIModel{
		1: {ID: 1, OrganizationID: pointer(orgID), IsEnabled: true, EncryptedCredentials: "invalid-json"},
	}}
	service := NewService(repo, nil)

	model, err := service.GetVisible(context.Background(), 1, userID, orgID)

	require.NoError(t, err)
	assert.Equal(t, int64(1), model.ID)
}

func pointer(value int64) *int64 { return &value }
