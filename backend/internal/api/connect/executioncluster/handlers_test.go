package executionclusterconnect

import (
	"context"
	"errors"
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/require"

	"github.com/l8ai-cn/agentcloud/backend/internal/middleware"
	executioncluster "github.com/l8ai-cn/agentcloud/backend/internal/service/executioncluster"
	executionclusterv1 "github.com/l8ai-cn/agentcloud/proto/gen/go/execution_cluster/v1"
)

func TestCreateRegistrationCommandRejectsOtherOrganizationCluster(t *testing.T) {
	server := NewServer(
		&fakeExecutionClusterService{createError: executioncluster.ErrClusterNotFound},
		&fakeOrganizationService{role: "admin"},
	)

	_, err := server.CreateRegistrationCommand(
		asUser(42),
		connect.NewRequest(&executionclusterv1.CreateRegistrationCommandRequest{
			OrgSlug:   "org-one",
			ClusterId: 99,
			NodeName:  "local-mac",
		}),
	)

	require.Error(t, err)
	require.Equal(t, connect.CodeNotFound, connect.CodeOf(err))
}

type fakeExecutionClusterService struct {
	createError error
}

func (s *fakeExecutionClusterService) List(context.Context, int64) ([]executioncluster.View, error) {
	return nil, nil
}

func (s *fakeExecutionClusterService) CreateRegistrationCommand(
	context.Context,
	int64,
	int64,
	int64,
	string,
) (executioncluster.RegistrationCommand, error) {
	return executioncluster.RegistrationCommand{}, s.createError
}

type fakeOrganizationService struct {
	role string
}

func (s *fakeOrganizationService) GetBySlug(_ context.Context, slug string) (middleware.OrganizationGetter, error) {
	if slug == "missing" {
		return nil, errors.New("organization not found")
	}
	return fakeOrganization{id: 7, slug: slug}, nil
}

func (s *fakeOrganizationService) IsMember(context.Context, int64, int64) (bool, error) {
	return true, nil
}

func (s *fakeOrganizationService) GetMemberRole(context.Context, int64, int64) (string, error) {
	return s.role, nil
}

type fakeOrganization struct {
	id   int64
	slug string
}

func (o fakeOrganization) GetID() int64 {
	return o.id
}

func (o fakeOrganization) GetSlug() string {
	return o.slug
}

func (o fakeOrganization) GetName() string {
	return o.slug
}

func asUser(userID int64) context.Context {
	return middleware.SetTenant(context.Background(), &middleware.TenantContext{UserID: userID})
}
