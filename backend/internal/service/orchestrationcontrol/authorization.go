package orchestrationcontrol

import (
	"context"
	"errors"

	control "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
	"github.com/anthropics/agentsmesh/backend/internal/domain/organization"
)

type MemberReader interface {
	GetMember(context.Context, int64, int64) (*organization.Member, error)
}

type MemberAuthorizer struct {
	members MemberReader
}

func NewMemberAuthorizer(members MemberReader) *MemberAuthorizer {
	return &MemberAuthorizer{members: members}
}

func (authorizer *MemberAuthorizer) AuthorizeList(
	ctx context.Context,
	scope control.Scope,
) error {
	_, err := authorizer.member(ctx, scope)
	return err
}

func (authorizer *MemberAuthorizer) AuthorizeCreate(
	ctx context.Context,
	scope control.Scope,
	_ control.ResourceTarget,
) error {
	_, err := authorizer.member(ctx, scope)
	return err
}

func (authorizer *MemberAuthorizer) AuthorizeUpdate(
	ctx context.Context,
	scope control.Scope,
	head control.ResourceHead,
) error {
	member, err := authorizer.member(ctx, scope)
	if err != nil {
		return err
	}
	if head.CreatedByID == scope.ActorID ||
		member.Role == organization.RoleOwner ||
		member.Role == organization.RoleAdmin {
		return nil
	}
	return ErrForbidden
}

func (authorizer *MemberAuthorizer) AuthorizeReference(
	ctx context.Context,
	scope control.Scope,
	head control.ResourceHead,
) error {
	if head.OrganizationID != scope.OrganizationID ||
		head.Identity.Namespace != scope.OrganizationSlug {
		return ErrForbidden
	}
	_, err := authorizer.member(ctx, scope)
	return err
}

func (authorizer *MemberAuthorizer) member(
	ctx context.Context,
	scope control.Scope,
) (*organization.Member, error) {
	if authorizer == nil || authorizer.members == nil {
		return nil, ErrUnavailable
	}
	if err := scope.Validate(); err != nil {
		return nil, err
	}
	member, err := authorizer.members.GetMember(
		ctx,
		scope.OrganizationID,
		scope.ActorID,
	)
	if errors.Is(err, organization.ErrMemberNotFound) ||
		(err == nil && member == nil) {
		return nil, ErrForbidden
	}
	if err != nil {
		return nil, err
	}
	if member.OrganizationID != scope.OrganizationID ||
		member.UserID != scope.ActorID {
		return nil, ErrForbidden
	}
	switch member.Role {
	case organization.RoleOwner, organization.RoleAdmin, organization.RoleMember:
		return member, nil
	default:
		return nil, ErrForbidden
	}
}
