package orchestrationcontrol

import (
	"context"
	"testing"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/organization"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemberAuthorizerAllowsMembersToCreate(t *testing.T) {
	reader := &memberReaderStub{member: &organization.Member{
		OrganizationID: 42, UserID: 7, Role: organization.RoleMember,
	}}
	authorizer := NewMemberAuthorizer(reader)
	require.NoError(t, authorizer.AuthorizeCreate(
		context.Background(),
		orchestrationServiceScope(),
		orchestrationServiceTarget(),
	))
}

func TestMemberAuthorizerRequiresMembershipToList(t *testing.T) {
	reader := &memberReaderStub{member: &organization.Member{
		OrganizationID: 42, UserID: 7, Role: organization.RoleMember,
	}}
	authorizer := NewMemberAuthorizer(reader)
	require.NoError(t, authorizer.AuthorizeList(
		context.Background(),
		orchestrationServiceScope(),
	))

	reader.member = nil
	reader.err = organization.ErrMemberNotFound
	assert.ErrorIs(t, authorizer.AuthorizeList(
		context.Background(),
		orchestrationServiceScope(),
	), ErrForbidden)
}

func TestMemberAuthorizerAllowsCreatorOrAdminToUpdate(t *testing.T) {
	head := orchestrationServiceHead()
	reader := &memberReaderStub{member: &organization.Member{
		OrganizationID: 42, UserID: 7, Role: organization.RoleMember,
	}}
	authorizer := NewMemberAuthorizer(reader)
	require.NoError(t, authorizer.AuthorizeUpdate(
		context.Background(),
		orchestrationServiceScope(),
		head,
	))

	head.CreatedByID = 8
	assert.ErrorIs(t, authorizer.AuthorizeUpdate(
		context.Background(),
		orchestrationServiceScope(),
		head,
	), ErrForbidden)

	reader.member.Role = organization.RoleAdmin
	require.NoError(t, authorizer.AuthorizeUpdate(
		context.Background(),
		orchestrationServiceScope(),
		head,
	))
}

func TestMemberAuthorizerRejectsMissingMembership(t *testing.T) {
	authorizer := NewMemberAuthorizer(&memberReaderStub{
		err: organization.ErrMemberNotFound,
	})
	assert.ErrorIs(t, authorizer.AuthorizeCreate(
		context.Background(),
		orchestrationServiceScope(),
		orchestrationServiceTarget(),
	), ErrForbidden)
}

func TestMemberAuthorizerScopesReferencedResources(t *testing.T) {
	reader := &memberReaderStub{member: &organization.Member{
		OrganizationID: 42, UserID: 7, Role: organization.RoleMember,
	}}
	authorizer := NewMemberAuthorizer(reader)
	head := orchestrationServiceHead()
	require.NoError(t, authorizer.AuthorizeReference(
		context.Background(),
		orchestrationServiceScope(),
		head,
	))

	head.OrganizationID = 99
	assert.ErrorIs(t, authorizer.AuthorizeReference(
		context.Background(),
		orchestrationServiceScope(),
		head,
	), ErrForbidden)
}
