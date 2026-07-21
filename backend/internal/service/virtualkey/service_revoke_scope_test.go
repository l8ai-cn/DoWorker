package virtualkey

import (
	"context"
	"testing"

	virtualkeydomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/virtualkey"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRevokeUpdatesOnlyTheCallingUserScope(t *testing.T) {
	trace := &callTrace{}
	repo := &fakeVirtualKeyRepository{
		trace: trace,
		keys: map[int64]*virtualkeydomain.VirtualAPIKey{
			41: {
				ID: 41, OrganizationID: 21, UserID: 11,
				ModelResourceID: 31, Status: virtualkeydomain.StatusActive,
			},
		},
	}
	service := NewService(repo, &fakeModelResourceValidator{trace: trace})

	err := service.Revoke(context.Background(), 41, 21, 11)

	require.NoError(t, err)
	require.Len(t, repo.statusCalls, 1)
	assert.Equal(t, scopedStatusCall{
		id: 41, orgID: 21, userID: 11, status: virtualkeydomain.StatusRevoked,
	}, repo.statusCalls[0])
	assert.Equal(t, virtualkeydomain.StatusRevoked, repo.keys[41].Status)
}

func TestRevokeHidesKeysOutsideTheCallingUserScope(t *testing.T) {
	trace := &callTrace{}
	repo := &fakeVirtualKeyRepository{
		trace: trace,
		keys: map[int64]*virtualkeydomain.VirtualAPIKey{
			41: {
				ID: 41, OrganizationID: 21, UserID: 12,
				ModelResourceID: 31, Status: virtualkeydomain.StatusActive,
			},
		},
	}
	service := NewService(repo, &fakeModelResourceValidator{trace: trace})

	err := service.Revoke(context.Background(), 41, 21, 11)

	assert.ErrorIs(t, err, ErrNotFound)
	assert.Equal(t, virtualkeydomain.StatusActive, repo.keys[41].Status)
}
