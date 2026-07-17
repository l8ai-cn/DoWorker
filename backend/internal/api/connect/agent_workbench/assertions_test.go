package agentworkbenchconnect

import (
	"context"
	"errors"
	"testing"

	"connectrpc.com/connect"
	"github.com/anthropics/agentsmesh/backend/internal/middleware"
	"github.com/anthropics/agentsmesh/backend/pkg/embedtoken"
	"github.com/stretchr/testify/require"
)

func ownerContext() context.Context {
	return middleware.SetTenant(context.Background(), &middleware.TenantContext{
		UserID: testUserID,
	})
}

func embedContext(capabilities ...string) context.Context {
	claims := &embedtoken.Claims{
		SessionID:        testSessionID,
		OrganizationID:   7,
		OrganizationSlug: testOrgSlug,
		UserID:           testUserID,
		Capabilities:     capabilities,
	}
	ctx := middleware.SetTenant(context.Background(), &middleware.TenantContext{
		OrganizationID:   claims.OrganizationID,
		OrganizationSlug: claims.OrganizationSlug,
		UserID:           claims.UserID,
		UserRole:         "embed",
	})
	return withEmbedAccess(ctx, claims)
}

func requireConnectCode(t *testing.T, err error, expected connect.Code) {
	t.Helper()
	require.Error(t, err)
	var connectError *connect.Error
	require.True(t, errors.As(err, &connectError))
	require.Equal(t, expected, connectError.Code())
}
