package virtualkey

import (
	"context"
	"errors"
	"testing"

	aimodeldomain "github.com/anthropics/agentsmesh/backend/internal/domain/aimodel"
	aimodelsvc "github.com/anthropics/agentsmesh/backend/internal/service/aimodel"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateValidatesVisibleModelBeforeMinting(t *testing.T) {
	const (
		modelID = int64(31)
		orgID   = int64(21)
		userID  = int64(11)
	)
	trace := &callTrace{}
	keyRepo := &fakeVirtualKeyRepository{trace: trace}
	modelRepo := &fakeAIModelRepository{
		trace: trace,
		models: map[int64]*aimodeldomain.AIModel{
			modelID: {ID: modelID, OrganizationID: int64Pointer(orgID), IsEnabled: true},
		},
	}
	service := NewService(keyRepo, aimodelsvc.NewService(modelRepo, nil))

	created, err := service.Create(context.Background(), CreateInput{
		OrgID: orgID, UserID: userID, AIModelID: modelID, Name: "Worker key",
	})

	require.NoError(t, err)
	require.NotNil(t, created)
	assert.NotEmpty(t, created.Token)
	require.Len(t, keyRepo.created, 1)
	assert.Equal(t, modelID, keyRepo.created[0].AIModelID)
	assert.Equal(t, []visibleModelCall{{id: modelID, userID: userID, orgID: orgID}}, modelRepo.visibleCalls)
	assert.Equal(t, []string{"model.get-visible", "key.create"}, trace.calls)
}

func TestCreateRejectsInvisibleModel(t *testing.T) {
	const (
		modelID = int64(31)
		orgID   = int64(21)
		userID  = int64(11)
	)
	tests := []struct {
		name  string
		model *aimodeldomain.AIModel
	}{
		{name: "missing"},
		{name: "foreign organization", model: &aimodeldomain.AIModel{
			ID: modelID, OrganizationID: int64Pointer(99), IsEnabled: true,
		}},
		{name: "foreign user", model: &aimodeldomain.AIModel{
			ID: modelID, UserID: int64Pointer(98), IsEnabled: true,
		}},
		{name: "disabled", model: &aimodeldomain.AIModel{
			ID: modelID, OrganizationID: int64Pointer(orgID), IsEnabled: false,
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			trace := &callTrace{}
			keyRepo := &fakeVirtualKeyRepository{trace: trace}
			models := map[int64]*aimodeldomain.AIModel{}
			if test.model != nil {
				models[modelID] = test.model
			}
			modelRepo := &fakeAIModelRepository{models: models, trace: trace}
			service := NewService(keyRepo, aimodelsvc.NewService(modelRepo, nil))

			created, err := service.Create(context.Background(), CreateInput{
				OrgID: orgID, UserID: userID, AIModelID: modelID, Name: "Worker key",
			})

			assert.Nil(t, created)
			assert.ErrorIs(t, err, aimodelsvc.ErrNotFound)
			assert.Empty(t, keyRepo.created)
			assert.Equal(t, []string{"model.get-visible"}, trace.calls)
		})
	}
}

func TestCreatePropagatesModelAndRepositoryErrors(t *testing.T) {
	const (
		modelID = int64(31)
		orgID   = int64(21)
		userID  = int64(11)
	)
	t.Run("model query", func(t *testing.T) {
		modelErr := errors.New("model query failed")
		trace := &callTrace{}
		keyRepo := &fakeVirtualKeyRepository{trace: trace}
		modelRepo := &fakeAIModelRepository{visibleErr: modelErr, trace: trace}
		service := NewService(keyRepo, aimodelsvc.NewService(modelRepo, nil))

		created, err := service.Create(context.Background(), CreateInput{
			OrgID: orgID, UserID: userID, AIModelID: modelID,
		})

		assert.Nil(t, created)
		assert.ErrorIs(t, err, modelErr)
		assert.Empty(t, keyRepo.created)
		assert.Equal(t, []string{"model.get-visible"}, trace.calls)
	})

	t.Run("key persistence", func(t *testing.T) {
		createErr := errors.New("key persistence failed")
		trace := &callTrace{}
		keyRepo := &fakeVirtualKeyRepository{createErr: createErr, trace: trace}
		modelRepo := &fakeAIModelRepository{
			trace: trace,
			models: map[int64]*aimodeldomain.AIModel{
				modelID: {ID: modelID, UserID: int64Pointer(userID), IsEnabled: true},
			},
		}
		service := NewService(keyRepo, aimodelsvc.NewService(modelRepo, nil))

		created, err := service.Create(context.Background(), CreateInput{
			OrgID: orgID, UserID: userID, AIModelID: modelID,
		})

		assert.Nil(t, created)
		assert.ErrorIs(t, err, createErr)
		require.Len(t, keyRepo.created, 1)
		assert.Equal(t, []string{"model.get-visible", "key.create"}, trace.calls)
	})
}
