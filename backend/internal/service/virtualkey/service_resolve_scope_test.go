package virtualkey

import (
	"context"
	"errors"
	"testing"

	aimodeldomain "github.com/anthropics/agentsmesh/backend/internal/domain/aimodel"
	virtualkeydomain "github.com/anthropics/agentsmesh/backend/internal/domain/virtualkey"
	aimodelsvc "github.com/anthropics/agentsmesh/backend/internal/service/aimodel"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveModelForScopeRejectsInvisibleAndRevokedKeys(t *testing.T) {
	const (
		keyID   = int64(41)
		modelID = int64(31)
		orgID   = int64(21)
		userID  = int64(11)
	)
	tests := []struct {
		name   string
		lookup func(*Service) (*aimodelsvc.ResolvedModel, *int64, error)
		status string
	}{
		{
			name: "wrong organization",
			lookup: func(service *Service) (*aimodelsvc.ResolvedModel, *int64, error) {
				return service.ResolveModelForScope(context.Background(), keyID, 22, userID)
			},
			status: virtualkeydomain.StatusActive,
		},
		{
			name: "wrong user",
			lookup: func(service *Service) (*aimodelsvc.ResolvedModel, *int64, error) {
				return service.ResolveModelForScope(context.Background(), keyID, orgID, 12)
			},
			status: virtualkeydomain.StatusActive,
		},
		{
			name: "missing",
			lookup: func(service *Service) (*aimodelsvc.ResolvedModel, *int64, error) {
				return service.ResolveModelForScope(context.Background(), 99, orgID, userID)
			},
			status: virtualkeydomain.StatusActive,
		},
		{
			name: "revoked",
			lookup: func(service *Service) (*aimodelsvc.ResolvedModel, *int64, error) {
				return service.ResolveModelForScope(context.Background(), keyID, orgID, userID)
			},
			status: virtualkeydomain.StatusRevoked,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			trace := &callTrace{}
			keyRepo := &fakeVirtualKeyRepository{
				trace: trace,
				keys: map[int64]*virtualkeydomain.VirtualAPIKey{
					keyID: {ID: keyID, OrganizationID: orgID, UserID: userID, AIModelID: modelID, Status: test.status},
				},
			}
			modelRepo := &fakeAIModelRepository{trace: trace}
			service := NewService(keyRepo, aimodelsvc.NewService(modelRepo, nil))

			resolved, budget, err := test.lookup(service)

			assert.Nil(t, resolved)
			assert.Nil(t, budget)
			if test.status == virtualkeydomain.StatusRevoked {
				assert.ErrorIs(t, err, ErrRevoked)
			} else {
				assert.ErrorIs(t, err, ErrNotFound)
			}
			assert.Empty(t, modelRepo.visibleCalls)
			assert.Empty(t, keyRepo.touchCalls)
			assert.Equal(t, []string{"key.get-scoped"}, trace.calls)
		})
	}
}

func TestResolveModelForScopeValidatesModelBeforeTouch(t *testing.T) {
	const (
		keyID   = int64(41)
		modelID = int64(31)
		orgID   = int64(21)
		userID  = int64(11)
	)
	tests := []struct {
		name     string
		model    *aimodeldomain.AIModel
		modelErr error
		wantErr  error
	}{
		{name: "invisible model", wantErr: aimodelsvc.ErrNotFound},
		{name: "model query failure", modelErr: errors.New("model query failed")},
		{
			name: "credential resolution failure",
			model: &aimodeldomain.AIModel{
				ID: modelID, OrganizationID: int64Pointer(orgID), IsEnabled: true,
				EncryptedCredentials: "invalid-json",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			trace := &callTrace{}
			keyRepo := &fakeVirtualKeyRepository{
				trace: trace,
				keys: map[int64]*virtualkeydomain.VirtualAPIKey{
					keyID: {ID: keyID, OrganizationID: orgID, UserID: userID, AIModelID: modelID, Status: virtualkeydomain.StatusActive},
				},
			}
			models := map[int64]*aimodeldomain.AIModel{}
			if test.model != nil {
				models[modelID] = test.model
			}
			modelRepo := &fakeAIModelRepository{models: models, visibleErr: test.modelErr, trace: trace}
			service := NewService(keyRepo, aimodelsvc.NewService(modelRepo, nil))

			resolved, budget, err := service.ResolveModelForScope(context.Background(), keyID, orgID, userID)

			assert.Nil(t, resolved)
			assert.Nil(t, budget)
			require.Error(t, err)
			if test.wantErr != nil {
				assert.ErrorIs(t, err, test.wantErr)
			}
			if test.modelErr != nil {
				assert.ErrorIs(t, err, test.modelErr)
			}
			assert.Empty(t, keyRepo.touchCalls)
			assert.Equal(t, []string{"key.get-scoped", "model.get-visible"}, trace.calls)
		})
	}
}

func TestResolveModelForScopePropagatesScopedQueryAndTouchErrors(t *testing.T) {
	const (
		keyID   = int64(41)
		modelID = int64(31)
		orgID   = int64(21)
		userID  = int64(11)
	)
	t.Run("scoped query", func(t *testing.T) {
		queryErr := errors.New("scoped query failed")
		trace := &callTrace{}
		keyRepo := &fakeVirtualKeyRepository{scopedErr: queryErr, trace: trace}
		modelRepo := &fakeAIModelRepository{trace: trace}
		service := NewService(keyRepo, aimodelsvc.NewService(modelRepo, nil))

		resolved, budget, err := service.ResolveModelForScope(context.Background(), keyID, orgID, userID)

		assert.Nil(t, resolved)
		assert.Nil(t, budget)
		assert.ErrorIs(t, err, queryErr)
		assert.Equal(t, []string{"key.get-scoped"}, trace.calls)
	})

	t.Run("touch", func(t *testing.T) {
		touchErr := errors.New("touch failed")
		trace := &callTrace{}
		budget := int64(500)
		keyRepo, modelRepo := successfulResolveRepositories(trace, touchErr, &budget)
		service := NewService(keyRepo, aimodelsvc.NewService(modelRepo, nil))

		resolved, resolvedBudget, err := service.ResolveModelForScope(context.Background(), keyID, orgID, userID)

		assert.Nil(t, resolved)
		assert.Nil(t, resolvedBudget)
		assert.ErrorIs(t, err, touchErr)
		assert.Equal(t, []int64{keyID}, keyRepo.touchCalls)
		assert.Equal(t, []string{"key.get-scoped", "model.get-visible", "key.touch"}, trace.calls)
	})
}

func TestResolveModelForScopeReturnsResolvedModelAndBudget(t *testing.T) {
	trace := &callTrace{}
	budget := int64(500)
	keyRepo, modelRepo := successfulResolveRepositories(trace, nil, &budget)
	service := NewService(keyRepo, aimodelsvc.NewService(modelRepo, nil))

	resolved, resolvedBudget, err := service.ResolveModelForScope(context.Background(), 41, 21, 11)

	require.NoError(t, err)
	require.NotNil(t, resolved)
	assert.Equal(t, int64(31), resolved.Model.ID)
	assert.Equal(t, "secret", resolved.Credentials["api_key"])
	require.NotNil(t, resolvedBudget)
	assert.Equal(t, budget, *resolvedBudget)
	assert.Equal(t, []scopedKeyCall{{id: 41, orgID: 21, userID: 11}}, keyRepo.scopedCalls)
	assert.Equal(t, []visibleModelCall{{id: 31, userID: 11, orgID: 21}}, modelRepo.visibleCalls)
	assert.Equal(t, []int64{41}, keyRepo.touchCalls)
	assert.Equal(t, []string{"key.get-scoped", "model.get-visible", "key.touch"}, trace.calls)
}

func successfulResolveRepositories(
	trace *callTrace, touchErr error, budget *int64,
) (*fakeVirtualKeyRepository, *fakeAIModelRepository) {
	keyRepo := &fakeVirtualKeyRepository{
		trace:    trace,
		touchErr: touchErr,
		keys: map[int64]*virtualkeydomain.VirtualAPIKey{
			41: {
				ID: 41, OrganizationID: 21, UserID: 11, AIModelID: 31,
				Status: virtualkeydomain.StatusActive, TokenBudget: budget,
			},
		},
	}
	modelRepo := &fakeAIModelRepository{
		trace: trace,
		models: map[int64]*aimodeldomain.AIModel{
			31: {
				ID: 31, OrganizationID: int64Pointer(21), IsEnabled: true,
				EncryptedCredentials: `{"api_key":"secret"}`,
			},
		},
	}
	return keyRepo, modelRepo
}
