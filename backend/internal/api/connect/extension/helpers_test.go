package extensionconnect

import (
	"context"
	"errors"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/require"

	"github.com/anthropics/agentsmesh/backend/internal/middleware"
)

type fakeOrg struct {
	id   int64
	slug string
}

func (f fakeOrg) GetID() int64    { return f.id }
func (f fakeOrg) GetSlug() string { return f.slug }
func (f fakeOrg) GetName() string { return f.slug }

type fakeOrgService struct {
	role string
}

func (f *fakeOrgService) GetBySlug(_ context.Context, slug string) (middleware.OrganizationGetter, error) {
	if slug == "missing" {
		return nil, errors.New("org not found")
	}
	return fakeOrg{id: 7, slug: slug}, nil
}

func (f *fakeOrgService) IsMember(context.Context, int64, int64) (bool, error) { return true, nil }

func (f *fakeOrgService) GetMemberRole(context.Context, int64, int64) (string, error) {
	return f.role, nil
}

// ctxAsUser populates the auth-interceptor stand-in: UserID matters because
// ResolveOrgScope rejects empty TenantContext as Unauthenticated.
func ctxAsUser(userID int64) context.Context {
	return middleware.SetTenant(context.Background(), &middleware.TenantContext{UserID: userID})
}

// connectCodeOf is the canonical accessor for the Connect error code,
// independent of the test framework's error helpers.
func connectCodeOf(t *testing.T, err error) connect.Code {
	t.Helper()
	var ce *connect.Error
	require.True(t, errors.As(err, &ce), "expected *connect.Error, got %v", err)
	return ce.Code()
}

func mustParseTime(t *testing.T, s string) time.Time {
	t.Helper()
	tt, err := time.Parse(time.RFC3339, s)
	require.NoError(t, err)
	return tt
}
