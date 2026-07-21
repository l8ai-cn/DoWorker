package virtualkey

import (
	"context"
	"errors"
	"testing"

	virtualkeydomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/virtualkey"
	airesourcesvc "github.com/l8ai-cn/agentcloud/backend/internal/service/airesource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveResourceForScopeRejectsInvisibleAndRevokedKeys(t *testing.T) {
	const (
		keyID      = int64(41)
		resourceID = int64(31)
		orgID      = int64(21)
		userID     = int64(11)
	)
	tests := []struct {
		name   string
		lookup func(*Service) (int64, *int64, error)
		status string
	}{
		{
			name: "wrong organization",
			lookup: func(service *Service) (int64, *int64, error) {
				return service.ResolveResourceForScope(context.Background(), keyID, 22, userID)
			},
			status: virtualkeydomain.StatusActive,
		},
		{
			name: "wrong user",
			lookup: func(service *Service) (int64, *int64, error) {
				return service.ResolveResourceForScope(context.Background(), keyID, orgID, 12)
			},
			status: virtualkeydomain.StatusActive,
		},
		{
			name: "missing",
			lookup: func(service *Service) (int64, *int64, error) {
				return service.ResolveResourceForScope(context.Background(), 99, orgID, userID)
			},
			status: virtualkeydomain.StatusActive,
		},
		{
			name: "revoked",
			lookup: func(service *Service) (int64, *int64, error) {
				return service.ResolveResourceForScope(context.Background(), keyID, orgID, userID)
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
					keyID: {ID: keyID, OrganizationID: orgID, UserID: userID, ModelResourceID: resourceID, Status: test.status},
				},
			}
			resourceValidator := &fakeModelResourceValidator{trace: trace}
			service := NewService(keyRepo, resourceValidator)

			resolvedID, budget, err := test.lookup(service)

			assert.Zero(t, resolvedID)
			assert.Nil(t, budget)
			if test.status == virtualkeydomain.StatusRevoked {
				assert.ErrorIs(t, err, ErrRevoked)
			} else {
				assert.ErrorIs(t, err, ErrNotFound)
			}
			assert.Empty(t, resourceValidator.visibleCalls)
			assert.Empty(t, keyRepo.touchCalls)
			assert.Equal(t, []string{"key.get-scoped"}, trace.calls)
		})
	}
}

func TestResolveResourceForScopeValidatesResourceBeforeTouch(t *testing.T) {
	const (
		keyID      = int64(41)
		resourceID = int64(31)
		orgID      = int64(21)
		userID     = int64(11)
	)
	tests := []struct {
		name        string
		resourceErr error
	}{
		{name: "invisible resource", resourceErr: airesourcesvc.ErrNotFound},
		{name: "resource query failure", resourceErr: errors.New("resource query failed")},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			trace := &callTrace{}
			keyRepo := &fakeVirtualKeyRepository{
				trace: trace,
				keys: map[int64]*virtualkeydomain.VirtualAPIKey{
					keyID: {ID: keyID, OrganizationID: orgID, UserID: userID, ModelResourceID: resourceID, Status: virtualkeydomain.StatusActive},
				},
			}
			resourceValidator := &fakeModelResourceValidator{err: test.resourceErr, trace: trace}
			service := NewService(keyRepo, resourceValidator)

			resolvedID, budget, err := service.ResolveResourceForScope(context.Background(), keyID, orgID, userID)

			assert.Zero(t, resolvedID)
			assert.Nil(t, budget)
			assert.ErrorIs(t, err, test.resourceErr)
			assert.Empty(t, keyRepo.touchCalls)
			assert.Equal(t, []string{"key.get-scoped", "resource.ensure-selectable"}, trace.calls)
		})
	}
}

func TestResolveResourceForScopePropagatesScopedQueryAndTouchErrors(t *testing.T) {
	const (
		keyID  = int64(41)
		orgID  = int64(21)
		userID = int64(11)
	)
	t.Run("scoped query", func(t *testing.T) {
		queryErr := errors.New("scoped query failed")
		trace := &callTrace{}
		keyRepo := &fakeVirtualKeyRepository{scopedErr: queryErr, trace: trace}
		resourceValidator := &fakeModelResourceValidator{trace: trace}
		service := NewService(keyRepo, resourceValidator)

		resolvedID, budget, err := service.ResolveResourceForScope(context.Background(), keyID, orgID, userID)

		assert.Zero(t, resolvedID)
		assert.Nil(t, budget)
		assert.ErrorIs(t, err, queryErr)
		assert.Equal(t, []string{"key.get-scoped"}, trace.calls)
	})

	t.Run("touch", func(t *testing.T) {
		touchErr := errors.New("touch failed")
		trace := &callTrace{}
		budget := int64(500)
		keyRepo, resourceValidator := successfulResolveRepositories(trace, touchErr, &budget)
		service := NewService(keyRepo, resourceValidator)

		resolvedID, resolvedBudget, err := service.ResolveResourceForScope(context.Background(), keyID, orgID, userID)

		assert.Zero(t, resolvedID)
		assert.Nil(t, resolvedBudget)
		assert.ErrorIs(t, err, touchErr)
		assert.Equal(t, []int64{keyID}, keyRepo.touchCalls)
		assert.Equal(t, []string{"key.get-scoped", "resource.ensure-selectable", "key.touch"}, trace.calls)
	})
}

func TestResolveResourceForScopeReturnsResourceIDAndBudget(t *testing.T) {
	trace := &callTrace{}
	budget := int64(500)
	keyRepo, resourceValidator := successfulResolveRepositories(trace, nil, &budget)
	service := NewService(keyRepo, resourceValidator)

	resolvedID, resolvedBudget, err := service.ResolveResourceForScope(context.Background(), 41, 21, 11)

	require.NoError(t, err)
	assert.Equal(t, int64(31), resolvedID)
	require.NotNil(t, resolvedBudget)
	assert.Equal(t, budget, *resolvedBudget)
	assert.Equal(t, []scopedKeyCall{{id: 41, orgID: 21, userID: 11}}, keyRepo.scopedCalls)
	assert.Equal(t, []visibleModelCall{{id: 31, actor: airesourcesvc.Actor{UserID: 11}, orgID: 21}}, resourceValidator.visibleCalls)
	assert.Equal(t, []int64{41}, keyRepo.touchCalls)
	assert.Equal(t, []string{"key.get-scoped", "resource.ensure-selectable", "key.touch"}, trace.calls)
}

func successfulResolveRepositories(
	trace *callTrace, touchErr error, budget *int64,
) (*fakeVirtualKeyRepository, *fakeModelResourceValidator) {
	keyRepo := &fakeVirtualKeyRepository{
		trace:    trace,
		touchErr: touchErr,
		keys: map[int64]*virtualkeydomain.VirtualAPIKey{
			41: {
				ID: 41, OrganizationID: 21, UserID: 11, ModelResourceID: 31,
				Status: virtualkeydomain.StatusActive, TokenBudget: budget,
			},
		},
	}
	resourceValidator := &fakeModelResourceValidator{trace: trace}
	return keyRepo, resourceValidator
}
