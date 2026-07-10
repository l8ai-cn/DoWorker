package virtualkey

import (
	"context"
	"errors"
	"testing"

	airesourcesvc "github.com/anthropics/agentsmesh/backend/internal/service/airesource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateValidatesSelectableModelResourceBeforeMinting(t *testing.T) {
	const (
		resourceID = int64(31)
		orgID      = int64(21)
		userID     = int64(11)
	)
	trace := &callTrace{}
	keyRepo := &fakeVirtualKeyRepository{trace: trace}
	resourceValidator := &fakeModelResourceValidator{trace: trace}
	service := NewService(keyRepo, resourceValidator)

	created, err := service.Create(context.Background(), CreateInput{
		OrgID: orgID, UserID: userID, ModelResourceID: resourceID, Name: "Worker key",
	})

	require.NoError(t, err)
	require.NotNil(t, created)
	assert.NotEmpty(t, created.Token)
	require.Len(t, keyRepo.created, 1)
	assert.Equal(t, resourceID, keyRepo.created[0].ModelResourceID)
	assert.Equal(t, []visibleModelCall{{id: resourceID, actor: airesourcesvc.Actor{UserID: userID}, orgID: orgID}}, resourceValidator.visibleCalls)
	assert.Equal(t, []string{"resource.ensure-selectable", "key.create"}, trace.calls)
}

func TestCreateRejectsUnselectableModelResource(t *testing.T) {
	const (
		resourceID = int64(31)
		orgID      = int64(21)
		userID     = int64(11)
	)
	resourceErr := airesourcesvc.ErrNotFound
	trace := &callTrace{}
	keyRepo := &fakeVirtualKeyRepository{trace: trace}
	resourceValidator := &fakeModelResourceValidator{err: resourceErr, trace: trace}
	service := NewService(keyRepo, resourceValidator)

	created, err := service.Create(context.Background(), CreateInput{
		OrgID: orgID, UserID: userID, ModelResourceID: resourceID, Name: "Worker key",
	})

	assert.Nil(t, created)
	assert.ErrorIs(t, err, resourceErr)
	assert.Empty(t, keyRepo.created)
	assert.Equal(t, []string{"resource.ensure-selectable"}, trace.calls)
}

func TestCreatePropagatesResourceAndRepositoryErrors(t *testing.T) {
	const (
		resourceID = int64(31)
		orgID      = int64(21)
		userID     = int64(11)
	)
	t.Run("resource query", func(t *testing.T) {
		resourceErr := errors.New("resource query failed")
		trace := &callTrace{}
		keyRepo := &fakeVirtualKeyRepository{trace: trace}
		resourceValidator := &fakeModelResourceValidator{err: resourceErr, trace: trace}
		service := NewService(keyRepo, resourceValidator)

		created, err := service.Create(context.Background(), CreateInput{
			OrgID: orgID, UserID: userID, ModelResourceID: resourceID,
		})

		assert.Nil(t, created)
		assert.ErrorIs(t, err, resourceErr)
		assert.Empty(t, keyRepo.created)
		assert.Equal(t, []string{"resource.ensure-selectable"}, trace.calls)
	})

	t.Run("key persistence", func(t *testing.T) {
		createErr := errors.New("key persistence failed")
		trace := &callTrace{}
		keyRepo := &fakeVirtualKeyRepository{createErr: createErr, trace: trace}
		resourceValidator := &fakeModelResourceValidator{trace: trace}
		service := NewService(keyRepo, resourceValidator)

		created, err := service.Create(context.Background(), CreateInput{
			OrgID: orgID, UserID: userID, ModelResourceID: resourceID,
		})

		assert.Nil(t, created)
		assert.ErrorIs(t, err, createErr)
		require.Len(t, keyRepo.created, 1)
		assert.Equal(t, []string{"resource.ensure-selectable", "key.create"}, trace.calls)
	})
}
